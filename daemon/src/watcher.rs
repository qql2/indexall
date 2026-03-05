use crate::client::ApiClient;
use crate::config::WatchDir;
use anyhow::Result;
use notify::RecursiveMode;
use notify_debouncer_mini::new_debouncer;
use std::collections::HashMap;
use std::path::{Path, PathBuf};
use std::sync::{Arc, Mutex};
use std::time::{Duration, Instant, SystemTime};
use tracing::{error, info, warn};

const PENDING_MOVE_TIMEOUT: Duration = Duration::from_secs(2);

#[derive(Clone, PartialEq, Eq, Hash)]
struct FileMeta {
    size: u64,
    mtime: SystemTime,
}

pub struct FSWatcher {
    client: Arc<ApiClient>,
    watch_dir: WatchDir,
    machine_id: String,
    // path → (size, mtime): populated on startup scan + every create/modify event
    file_cache: Mutex<HashMap<PathBuf, FileMeta>>,
    // (size, mtime) → (old_external_id, removal_time): waits for matching Create
    pending_moves: Mutex<HashMap<FileMeta, (String, Instant)>>,
}

impl FSWatcher {
    pub fn new(client: Arc<ApiClient>, watch_dir: WatchDir, machine_id: String) -> Self {
        FSWatcher {
            client,
            watch_dir,
            machine_id,
            file_cache: Mutex::new(HashMap::new()),
            pending_moves: Mutex::new(HashMap::new()),
        }
    }

    pub async fn watch(self) -> Result<()> {
        // Populate cache before watching so Remove events have metadata to work with
        self.scan_directory(Path::new(&self.watch_dir.path), self.watch_dir.recursive);

        let (tx, mut rx) = tokio::sync::mpsc::channel(100);
        let mut debouncer = new_debouncer(Duration::from_millis(500), move |event| {
            let _ = tx.blocking_send(event);
        })?;

        let mode = if self.watch_dir.recursive {
            RecursiveMode::Recursive
        } else {
            RecursiveMode::NonRecursive
        };
        debouncer.watcher().watch(Path::new(&self.watch_dir.path), mode)?;

        info!(
            "Watching directory: {} (recursive: {})",
            self.watch_dir.path, self.watch_dir.recursive
        );

        while let Some(result) = rx.recv().await {
            if let Err(e) = self.handle_event(result).await {
                warn!("Error handling event: {}", e);
            }
        }
        Ok(())
    }

    fn scan_directory(&self, dir: &Path, recursive: bool) {
        let Ok(entries) = std::fs::read_dir(dir) else { return };
        for entry in entries.flatten() {
            let path = entry.path();
            if self.should_ignore(&path) {
                continue;
            }
            if path.is_dir() {
                if recursive {
                    self.scan_directory(&path, recursive);
                }
            } else if path.is_file() {
                self.cache_file_meta(&path);
            }
        }
    }

    fn cache_file_meta(&self, path: &Path) {
        if let Ok(meta) = std::fs::metadata(path) {
            if let Ok(mtime) = meta.modified() {
                self.file_cache
                    .lock()
                    .unwrap()
                    .insert(path.to_path_buf(), FileMeta { size: meta.len(), mtime });
            }
        }
    }

    async fn handle_event(&self, result: notify_debouncer_mini::DebounceEventResult) -> Result<()> {
        // Expire pending_moves older than timeout → treat as real deletions
        self.flush_expired_pending_moves().await;

        let Ok(events) = result else {
            warn!("Debouncer error");
            return Ok(());
        };

        for event in events {
            let path = &event.path;
            if self.should_ignore(path) {
                continue;
            }
            let abs_path = path.canonicalize().unwrap_or_else(|_| path.clone());
            let external_id = format!("{}:{}", self.machine_id, abs_path.display());
            let kind_str = format!("{:?}", event.kind);

            if kind_str.contains("Remove") || kind_str.contains("Name(From)") {
                self.handle_remove(path, &external_id).await?;
            } else if kind_str.contains("Create") || kind_str.contains("Name(To)") {
                self.handle_create(&abs_path, &external_id).await?;
            } else if kind_str.contains("Name(Any)") {
                // Platform-ambiguous rename: check existence to decide direction
                if abs_path.is_file() {
                    self.handle_create(&abs_path, &external_id).await?;
                } else {
                    self.handle_remove(path, &external_id).await?;
                }
            } else if kind_str.contains("Modify") {
                self.handle_modify(&abs_path, &external_id).await?;
            }
        }
        Ok(())
    }

    async fn flush_expired_pending_moves(&self) {
        let expired: Vec<(FileMeta, String)> = {
            let pending = self.pending_moves.lock().unwrap();
            pending
                .iter()
                .filter(|(_, (_, t))| t.elapsed() > PENDING_MOVE_TIMEOUT)
                .map(|(meta, (ext_id, _))| (meta.clone(), ext_id.clone()))
                .collect()
        };

        for (meta, external_id) in expired {
            self.pending_moves.lock().unwrap().remove(&meta);
            if let Ok(Some(id)) = self.client.get_by_external_id(&external_id).await {
                match self.client.delete_resource(&id).await {
                    Ok(_) => info!("Deleted resource (confirmed not a move): {}", external_id),
                    Err(e) => error!("Failed to delete resource: {}", e),
                }
            }
        }
    }

    async fn handle_create(&self, path: &Path, external_id: &str) -> Result<()> {
        if !path.is_file() {
            return Ok(());
        }
        self.cache_file_meta(path);

        // Try to match a pending Remove by size+mtime → file was moved
        if let Ok(meta) = std::fs::metadata(path) {
            if let Ok(mtime) = meta.modified() {
                let file_meta = FileMeta { size: meta.len(), mtime };
                let pending = self.pending_moves.lock().unwrap().remove(&file_meta);

                if let Some((old_external_id, _)) = pending {
                    match self.client.get_by_external_id(&old_external_id).await? {
                        Some(id) => {
                            let title = path
                                .file_name()
                                .and_then(|n| n.to_str())
                                .unwrap_or("unknown");
                            match self
                                .client
                                .update_resource(&id, path.to_str(), Some(external_id), Some(title))
                                .await
                            {
                                Ok(_) => info!(
                                    "Moved resource: {} → {}",
                                    old_external_id, external_id
                                ),
                                Err(e) => error!("Failed to update moved resource: {}", e),
                            }
                        }
                        None => {
                            // Old resource not in index, just create normally
                            self.create_if_auto_index(path, external_id).await;
                        }
                    }
                    return Ok(());
                }
            }
        }

        self.create_if_auto_index(path, external_id).await;
        Ok(())
    }

    async fn handle_modify(&self, path: &Path, external_id: &str) -> Result<()> {
        if !path.is_file() {
            return Ok(());
        }
        self.cache_file_meta(path);

        if let Ok(Some(id)) = self.client.get_by_external_id(external_id).await {
            let title = path.file_name().and_then(|n| n.to_str()).unwrap_or("unknown");
            match self.client.update_resource(&id, path.to_str(), None, Some(title)).await {
                Ok(_) => info!("Updated resource: {}", external_id),
                Err(e) => error!("Failed to update resource: {}", e),
            }
        }
        Ok(())
    }

    async fn handle_remove(&self, path: &Path, external_id: &str) -> Result<()> {
        let cached_meta = self.file_cache.lock().unwrap().remove(path);

        if let Some(meta) = cached_meta {
            // Park in pending_moves; will be resolved when matching Create arrives
            // or deleted after PENDING_MOVE_TIMEOUT
            self.pending_moves
                .lock()
                .unwrap()
                .insert(meta, (external_id.to_string(), Instant::now()));
            info!("Pending move detection for: {}", external_id);
        } else {
            // No cache entry (file not seen before this event) → immediate delete
            if let Ok(Some(id)) = self.client.get_by_external_id(external_id).await {
                match self.client.delete_resource(&id).await {
                    Ok(_) => info!("Deleted resource: {}", external_id),
                    Err(e) => error!("Failed to delete resource: {}", e),
                }
            }
        }
        Ok(())
    }

    async fn create_if_auto_index(&self, path: &Path, external_id: &str) {
        if !self.watch_dir.auto_index_new {
            return;
        }
        let title = path.file_name().and_then(|n| n.to_str()).unwrap_or("unknown");
        match self.client.get_by_external_id(external_id).await {
            Ok(Some(_)) => {}
            Ok(None) => {
                match self
                    .client
                    .create_resource(path.to_str().unwrap_or(""), title, external_id)
                    .await
                {
                    Ok(id) => info!("Created resource: {} → {}", external_id, id),
                    Err(e) => error!("Failed to create resource: {}", e),
                }
            }
            Err(e) => error!("Failed to check resource: {}", e),
        }
    }

    fn should_ignore(&self, path: &Path) -> bool {
        const BUILTIN_IGNORES: &[&str] = &[
            "node_modules", ".git", "target", "build", "dist",
            "__pycache__", ".cache", ".tox", "vendor", ".gradle",
            "coverage", ".next", ".nuxt", "venv", ".venv",
        ];

        for component in path.components() {
            if let Some(name) = component.as_os_str().to_str() {
                if BUILTIN_IGNORES.contains(&name) {
                    return true;
                }
                for pattern in &self.watch_dir.ignore_patterns {
                    if name.contains(pattern.as_str()) {
                        return true;
                    }
                }
            }
        }
        false
    }
}
