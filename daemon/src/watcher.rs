use crate::client::ApiClient;
use crate::config::WatchDir;
use anyhow::Result;
use notify_debouncer_mini::new_debouncer;
use notify::RecursiveMode;
use std::sync::Arc;
use std::time::Duration;
use tracing::{info, warn, error};

pub struct FSWatcher {
    client: Arc<ApiClient>,
    watch_dir: WatchDir,
    machine_id: String,
}

impl FSWatcher {
    pub fn new(client: Arc<ApiClient>, watch_dir: WatchDir, machine_id: String) -> Self {
        FSWatcher {
            client,
            watch_dir,
            machine_id,
        }
    }

    pub async fn watch(self) -> Result<()> {
        let (tx, mut rx) = tokio::sync::mpsc::channel(100);

        let mut debouncer = new_debouncer(Duration::from_millis(500), move |event| {
            let _ = tx.blocking_send(event);
        })?;

        let mode = if self.watch_dir.recursive {
            RecursiveMode::Recursive
        } else {
            RecursiveMode::NonRecursive
        };

        debouncer
            .watcher()
            .watch(std::path::Path::new(&self.watch_dir.path), mode)?;

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

    async fn handle_event(
        &self,
        result: notify_debouncer_mini::DebounceEventResult,
    ) -> Result<()> {
        match result {
            Ok(events) => {
                for event in events {
                    let path = &event.path;
                    if self.should_ignore(path) {
                        continue;
                    }

                    let abs_path = path.canonicalize().unwrap_or_else(|_| path.clone());
                    let external_id = format!("{}:{}", self.machine_id, abs_path.display());

                    let kind_str = format!("{:?}", event.kind);
                    if kind_str.contains("Create") {
                        self.handle_create(&abs_path, &external_id).await?;
                    } else if kind_str.contains("Modify") {
                        self.handle_modify(&abs_path, &external_id).await?;
                    } else if kind_str.contains("Remove") {
                        self.handle_remove(&external_id).await?;
                    }
                }
            }
            Err(e) => warn!("Debouncer error: {}", e),
        }
        Ok(())
    }

    async fn handle_create(&self, path: &std::path::Path, external_id: &str) -> Result<()> {
        if !path.is_file() {
            return Ok(());
        }

        match self.client.get_by_external_id(external_id).await? {
            Some(_) => info!("Resource already indexed: {}", external_id),
            None => {
                let title = path
                    .file_name()
                    .and_then(|n| n.to_str())
                    .unwrap_or("unknown");
                match self.client.create_resource(path.to_str().unwrap_or(""), title, external_id).await {
                    Ok(id) => info!("Created resource: {} -> {}", external_id, id),
                    Err(e) => error!("Failed to create resource: {}", e),
                }
            }
        }
        Ok(())
    }

    async fn handle_modify(&self, path: &std::path::Path, external_id: &str) -> Result<()> {
        if !path.is_file() {
            return Ok(());
        }

        let title = path
            .file_name()
            .and_then(|n| n.to_str())
            .unwrap_or("unknown");

        match self.client.get_by_external_id(external_id).await? {
            Some(id) => {
                match self.client.update_resource(
                    &id,
                    Some(path.to_str().unwrap_or("")),
                    Some(external_id),
                    Some(title),
                ).await {
                    Ok(_) => info!("Updated resource: {}", external_id),
                    Err(e) => error!("Failed to update resource: {}", e),
                }
            }
            None => {
                info!("Resource not found for update, creating: {}", external_id);
                match self.client.create_resource(path.to_str().unwrap_or(""), title, external_id).await {
                    Ok(id) => info!("Created resource: {} -> {}", external_id, id),
                    Err(e) => error!("Failed to create resource: {}", e),
                }
            }
        }
        Ok(())
    }

    async fn handle_remove(&self, external_id: &str) -> Result<()> {
        match self.client.get_by_external_id(external_id).await? {
            Some(id) => {
                match self.client.delete_resource(&id).await {
                    Ok(_) => info!("Deleted resource: {}", external_id),
                    Err(e) => error!("Failed to delete resource: {}", e),
                }
            }
            None => info!("Resource not found for deletion: {}", external_id),
        }
        Ok(())
    }

    fn should_ignore(&self, path: &std::path::Path) -> bool {
        if let Some(name) = path.file_name().and_then(|n| n.to_str()) {
            for pattern in &self.watch_dir.ignore_patterns {
                if name.contains(pattern) {
                    return true;
                }
            }
        }
        false
    }
}
