use std::path::PathBuf;
use std::sync::{Arc, Mutex};
use axum::body::Body;
use axum::http::{Request, StatusCode};
use tower::ServiceExt;

use ant_colony_rust_backend::models::{
    AntColonyMap, AntColonyPath, Subsystems, TurbineFaults,
};
use ant_colony_rust_backend::{create_app, database, AppState};

fn setup_test_app() -> axum::Router {
    let manifest_dir = PathBuf::from(env!("CARGO_MANIFEST_DIR"));
    let base_dir = manifest_dir.parent().unwrap().to_path_buf();
    let sql_path = base_dir.join("database").join("database.sql");

    let conn = database::init_db(":memory:", sql_path.to_str().unwrap())
        .expect("Failed to init in-memory DB");

    let state = Arc::new(AppState {
        db: Mutex::new(conn),
        base_dir,
    });

    create_app(state)
}

#[tokio::test]
async fn test_get_turbines_map() {
    let app = setup_test_app();

    let req = Request::builder()
        .uri("/ant-colony/get-turbines-map")
        .body(Body::empty())
        .unwrap();

    let resp = app.oneshot(req).await.unwrap();
    assert_eq!(resp.status(), StatusCode::OK);

    let body = axum::body::to_bytes(resp.into_body(), usize::MAX).await.unwrap();
    let data: Vec<AntColonyMap> = serde_json::from_slice(&body).unwrap();

    assert!(!data.is_empty());
    assert_ne!(data[0].turbine_id, 0);
    assert!(!data[0].turbine_name.is_empty());
}

#[tokio::test]
async fn test_get_subsystems() {
    let app = setup_test_app();

    let req = Request::builder()
        .uri("/ant-colony/get-subsystems")
        .body(Body::empty())
        .unwrap();

    let resp = app.oneshot(req).await.unwrap();
    assert_eq!(resp.status(), StatusCode::OK);

    let body = axum::body::to_bytes(resp.into_body(), usize::MAX).await.unwrap();
    let data: Vec<Subsystems> = serde_json::from_slice(&body).unwrap();

    assert!(!data.is_empty());
    assert_ne!(data[0].subsystem_id, 0);
    assert!(!data[0].subsystem_name.is_empty());
}

#[tokio::test]
async fn test_run_route_optimizer_ant_colony() {
    let app = setup_test_app();

    let payload = vec![
        TurbineFaults {
            turbine_id: 2,
            turbine_name: None,
            subsystem_name: "Electrical System".to_string(),
            fault_type: "Minor".to_string(),
        },
        TurbineFaults {
            turbine_id: 3,
            turbine_name: None,
            subsystem_name: "Rotor Hub".to_string(),
            fault_type: "Major".to_string(),
        },
    ];

    let body_bytes = serde_json::to_vec(&payload).unwrap();
    let req = Request::builder()
        .method("POST")
        .uri("/ant-colony/run-route-optimizer?algorithm=Ant%20Colony")
        .header("content-type", "application/json")
        .body(Body::from(body_bytes))
        .unwrap();

    let resp = app.oneshot(req).await.unwrap();
    assert_eq!(resp.status(), StatusCode::OK);

    let body = axum::body::to_bytes(resp.into_body(), usize::MAX).await.unwrap();
    let res: AntColonyPath = serde_json::from_slice(&body).unwrap();

    assert!(!res.turbine_order.is_empty());
    assert!(!res.turbine_order_to_show.is_empty());
    assert!(res.best_path_length > 0.0);
}

#[tokio::test]
async fn test_run_route_optimizer_genetic() {
    let app = setup_test_app();

    let payload = vec![
        TurbineFaults {
            turbine_id: 2,
            turbine_name: None,
            subsystem_name: "Electrical System".to_string(),
            fault_type: "Minor".to_string(),
        },
        TurbineFaults {
            turbine_id: 3,
            turbine_name: None,
            subsystem_name: "Rotor Hub".to_string(),
            fault_type: "Major".to_string(),
        },
    ];

    let body_bytes = serde_json::to_vec(&payload).unwrap();
    let req = Request::builder()
        .method("POST")
        .uri("/ant-colony/run-route-optimizer?algorithm=Genetic")
        .header("content-type", "application/json")
        .body(Body::from(body_bytes))
        .unwrap();

    let resp = app.oneshot(req).await.unwrap();
    assert_eq!(resp.status(), StatusCode::OK);

    let body = axum::body::to_bytes(resp.into_body(), usize::MAX).await.unwrap();
    let res: AntColonyPath = serde_json::from_slice(&body).unwrap();

    assert!(!res.turbine_order.is_empty());
}

#[tokio::test]
async fn test_run_route_optimizer_memetic() {
    let app = setup_test_app();

    let payload = vec![
        TurbineFaults {
            turbine_id: 2,
            turbine_name: None,
            subsystem_name: "Electrical System".to_string(),
            fault_type: "Minor".to_string(),
        },
        TurbineFaults {
            turbine_id: 3,
            turbine_name: None,
            subsystem_name: "Rotor Hub".to_string(),
            fault_type: "Major".to_string(),
        },
    ];

    let body_bytes = serde_json::to_vec(&payload).unwrap();
    let req = Request::builder()
        .method("POST")
        .uri("/ant-colony/run-route-optimizer?algorithm=Memetic")
        .header("content-type", "application/json")
        .body(Body::from(body_bytes))
        .unwrap();

    let resp = app.oneshot(req).await.unwrap();
    assert_eq!(resp.status(), StatusCode::OK);

    let body = axum::body::to_bytes(resp.into_body(), usize::MAX).await.unwrap();
    let res: AntColonyPath = serde_json::from_slice(&body).unwrap();

    assert!(!res.turbine_order.is_empty());
    assert!(!res.turbine_order_to_show.is_empty());
    assert_eq!(res.turbine_order_to_show.first().unwrap(), "Doca");
    assert_eq!(res.turbine_order_to_show.last().unwrap(), "Doca");
}

#[tokio::test]
async fn test_run_route_optimizer_single_turbine() {
    let app = setup_test_app();

    let payload = vec![TurbineFaults {
        turbine_id: 2,
        turbine_name: None,
        subsystem_name: "Electrical System".to_string(),
        fault_type: "Minor".to_string(),
    }];

    let body_bytes = serde_json::to_vec(&payload).unwrap();
    let req = Request::builder()
        .method("POST")
        .uri("/ant-colony/run-route-optimizer?algorithm=Ant%20Colony")
        .header("content-type", "application/json")
        .body(Body::from(body_bytes))
        .unwrap();

    let resp = app.oneshot(req).await.unwrap();
    assert_eq!(resp.status(), StatusCode::OK);

    let body = axum::body::to_bytes(resp.into_body(), usize::MAX).await.unwrap();
    let res: AntColonyPath = serde_json::from_slice(&body).unwrap();

    assert_eq!(res.turbine_order_to_show.len(), 3);
    assert_eq!(res.turbine_order_to_show[0], "Doca");
    assert_eq!(res.turbine_order_to_show[2], "Doca");
}

#[tokio::test]
async fn test_run_route_optimizer_empty_payload() {
    let app = setup_test_app();

    let payload: Vec<TurbineFaults> = vec![];
    let body_bytes = serde_json::to_vec(&payload).unwrap();

    let req = Request::builder()
        .method("POST")
        .uri("/ant-colony/run-route-optimizer?algorithm=Genetic")
        .header("content-type", "application/json")
        .body(Body::from(body_bytes))
        .unwrap();

    let resp = app.oneshot(req).await.unwrap();
    assert_eq!(resp.status(), StatusCode::OK);

    let body = axum::body::to_bytes(resp.into_body(), usize::MAX).await.unwrap();
    let res: AntColonyPath = serde_json::from_slice(&body).unwrap();

    assert!(!res.turbine_order_to_show.is_empty());
    assert_eq!(res.turbine_order_to_show.first().unwrap(), "Doca");
    assert_eq!(res.turbine_order_to_show.last().unwrap(), "Doca");
}
