mod client;
mod config;
mod server;
mod watcher;

use anyhow::Result;
use config::Config;
use std::sync::Arc;
use tracing::info;
use watcher::FSWatcher;

#[tokio::main]
async fn main() -> Result<()> {
    tracing_subscriber::fmt()
        .with_max_level(tracing::Level::INFO)
        .init();

    let config = Config::load()?;

    if config.watch_dirs.is_empty() {
        info!("No watch directories configured. Please edit config.toml");
        return Ok(());
    }

    info!("Starting IndexAll Daemon");
    info!("API URL: {}", config.api_url);
    info!("Machine ID: {}", config.machine_id);
    info!("Watching {} directories", config.watch_dirs.len());

    for dir in &config.watch_dirs {
        info!("  - {} (recursive: {})", dir.path, dir.recursive);
    }

    let client = Arc::new(client::ApiClient::new(config.api_url, config.api_key));

    let mut handles = vec![];

    let server_state = server::AppState {
        client: Arc::clone(&client),
        machine_id: config.machine_id.clone(),
    };
    let http_port = config.http_port;
    handles.push(tokio::spawn(async move {
        if let Err(e) = server::start(http_port, server_state).await {
            tracing::error!("HTTP server error: {}", e);
        }
    }));

    for watch_dir in config.watch_dirs {
        let client = Arc::clone(&client);
        let machine_id = config.machine_id.clone();
        let handle = tokio::spawn(async move {
            let watcher = FSWatcher::new(client, watch_dir, machine_id);
            if let Err(e) = watcher.watch().await {
                tracing::error!("Watcher error: {}", e);
            }
        });
        handles.push(handle);
    }

    tokio::signal::ctrl_c().await?;
    info!("Shutting down...");

    Ok(())
}
