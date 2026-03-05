use axum::{
    extract::State,
    http::StatusCode,
    routing::post,
    Json, Router,
};
use serde::{Deserialize, Serialize};
use std::sync::Arc;

use crate::client::ApiClient;

#[derive(Clone)]
pub struct AppState {
    pub client: Arc<ApiClient>,
    pub machine_id: String,
}

#[derive(Deserialize)]
pub struct IndexRequest {
    pub path: String,
}

#[derive(Serialize)]
pub struct IndexResponse {
    pub id: String,
    pub created: bool,
}

async fn index_handler(
    State(state): State<AppState>,
    Json(req): Json<IndexRequest>,
) -> Result<Json<IndexResponse>, (StatusCode, String)> {
    let raw = std::path::PathBuf::from(&req.path);
    if !raw.is_absolute() {
        return Err((StatusCode::BAD_REQUEST, "Path must be absolute".to_string()));
    }
    let path = dunce::canonicalize(&raw).unwrap_or(raw);

    let external_id = format!("{}:{}", state.machine_id, path.display());
    let title = path
        .file_name()
        .and_then(|n| n.to_str())
        .unwrap_or("Unknown");

    match state.client.get_by_external_id(&external_id).await {
        Ok(Some(id)) => Ok(Json(IndexResponse { id, created: false })),
        Ok(None) => {
            let path_str = path.to_str().unwrap_or("");
            match state.client.create_resource(path_str, title, &external_id).await {
                Ok(id) => Ok(Json(IndexResponse { id, created: true })),
                Err(e) => Err((StatusCode::INTERNAL_SERVER_ERROR, e.to_string())),
            }
        }
        Err(e) => Err((StatusCode::INTERNAL_SERVER_ERROR, e.to_string())),
    }
}

pub async fn start(port: u16, state: AppState) -> anyhow::Result<()> {
    let app = Router::new()
        .route("/index", post(index_handler))
        .with_state(state);

    let addr = format!("127.0.0.1:{}", port);
    tracing::info!("Connector HTTP server listening on {}", addr);
    let listener = tokio::net::TcpListener::bind(&addr).await?;
    axum::serve(listener, app).await?;
    Ok(())
}
