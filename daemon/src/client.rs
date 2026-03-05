use anyhow::Result;
use reqwest::Client;
use serde_json::json;

fn path_to_file_url(path: &str) -> String {
    if cfg!(windows) {
        format!("file:///{}", path.replace('\\', "/"))
    } else {
        path_to_file_url(path)
    }
}

pub struct ApiClient {
    base_url: String,
    api_key: Option<String>,
    client: Client,
}

impl ApiClient {
    pub fn new(base_url: String, api_key: Option<String>) -> Self {
        ApiClient {
            base_url,
            api_key,
            client: Client::new(),
        }
    }

    fn request(&self, method: reqwest::Method, url: &str) -> reqwest::RequestBuilder {
        let builder = self.client.request(method, url);
        if let Some(key) = &self.api_key {
            builder.header("Authorization", format!("Bearer {}", key))
        } else {
            builder
        }
    }

    pub async fn get_by_external_id(&self, external_id: &str) -> Result<Option<String>> {
        let url = format!(
            "{}/v1/resources/by-external-id?source=filesystem&external_id={}",
            self.base_url,
            urlencoding::encode(external_id)
        );
        let response = self.request(reqwest::Method::GET, &url).send().await?;
        if response.status().is_success() {
            let body: serde_json::Value = response.json().await?;
            Ok(body.get("id").and_then(|v| v.as_str()).map(|s| s.to_string()))
        } else {
            Ok(None)
        }
    }

    pub async fn create_resource(
        &self,
        path: &str,
        title: &str,
        external_id: &str,
    ) -> Result<String> {
        let url = format!("{}/v1/resources", self.base_url);
        let payload = json!({
            "title": title,
            "url": path_to_file_url(path),
            "source": "filesystem",
            "external_id": external_id,
        });

        let response = self.request(reqwest::Method::POST, &url).json(&payload).send().await?;
        let body: serde_json::Value = response.json().await?;
        let id = body
            .get("id")
            .and_then(|v| v.as_str())
            .ok_or_else(|| anyhow::anyhow!("No id in response"))?;
        Ok(id.to_string())
    }

    pub async fn update_resource(
        &self,
        id: &str,
        new_path: Option<&str>,
        new_external_id: Option<&str>,
        new_title: Option<&str>,
    ) -> Result<()> {
        let url = format!("{}/v1/resources/{}", self.base_url, id);
        let mut payload = serde_json::Map::new();

        if let Some(path) = new_path {
            payload.insert("url".to_string(), json!(path_to_file_url(path)));
        }
        if let Some(ext_id) = new_external_id {
            payload.insert("external_id".to_string(), json!(ext_id));
        }
        if let Some(title) = new_title {
            payload.insert("title".to_string(), json!(title));
        }

        self.request(reqwest::Method::PATCH, &url)
            .json(&serde_json::Value::Object(payload))
            .send()
            .await?;

        Ok(())
    }

    pub async fn delete_resource(&self, id: &str) -> Result<()> {
        let url = format!("{}/v1/resources/{}", self.base_url, id);
        self.request(reqwest::Method::DELETE, &url).send().await?;
        Ok(())
    }
}
