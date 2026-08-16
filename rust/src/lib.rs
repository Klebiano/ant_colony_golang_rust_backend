pub mod models;
pub mod algorithms;
pub mod database;
pub mod handlers;

use std::sync::Arc;
use axum::{
    routing::{get, post},
    Router,
};
use tower_http::cors::{Any, CorsLayer};

pub use handlers::AppState;

pub fn create_app(state: Arc<AppState>) -> Router {
    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods(Any)
        .allow_headers(Any);

    Router::new()
        .route("/ant-colony/get-turbines-map", get(handlers::get_turbines_map))
        .route("/ant-colony/get-subsystems", get(handlers::get_subsystems))
        .route("/ant-colony/run-route-optimizer", post(handlers::run_route_optimizer))
        .route("/ant-colony/run-route-optimizer/", post(handlers::run_route_optimizer))
        .layer(cors)
        .with_state(state)
}
