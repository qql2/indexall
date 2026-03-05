use anyhow::{anyhow, Result};
use serde::{Deserialize, Serialize};
use std::fs;
use std::path::PathBuf;
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Config {
    pub api_url: String,
    pub machine_id: String,
    pub watch_dirs: Vec<WatchDir>,
    #[serde(default = "default_http_port")]
    pub http_port: u16,
}

fn default_http_port() -> u16 {
    47832
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct WatchDir {
    pub path: String,
    pub recursive: bool,
    pub auto_index_new: bool,
    pub ignore_patterns: Vec<String>,
}

impl Config {
    pub fn load() -> Result<Self> {
        let config_path = Self::config_path()?;

        if config_path.exists() {
            let content = fs::read_to_string(&config_path)?;
            let mut config: Config = toml::from_str(&content)?;
            config.machine_id = config.machine_id.trim().to_string();
            Ok(config)
        } else {
            let config = Self::default_config()?;
            config.save(&config_path)?;
            println!(
                "Created default config at: {}",
                config_path.display()
            );
            println!("Please edit the config to add watch directories.");
            Ok(config)
        }
    }

    fn default_config() -> Result<Self> {
        Ok(Config {
            api_url: "http://localhost:8080".to_string(),
            machine_id: Uuid::new_v4().to_string(),
            watch_dirs: vec![],
            http_port: default_http_port(),
        })
    }

    fn config_path() -> Result<PathBuf> {
        let config_dir =
            dirs::config_dir().ok_or_else(|| anyhow!("Cannot find config directory"))?;
        let daemon_dir = config_dir.join("indexall-daemon");
        fs::create_dir_all(&daemon_dir)?;
        Ok(daemon_dir.join("config.toml"))
    }

    fn save(&self, path: &PathBuf) -> Result<()> {
        let content = toml::to_string_pretty(self)?;
        fs::write(path, content)?;
        Ok(())
    }
}
