use std::env;
use std::net::SocketAddr;
use std::sync::{Arc, Mutex};
use ant_colony_rust_backend::{create_app, database, AppState};

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let mut base_dir = env::current_dir()?;
    if !base_dir.join("database").join("database.sql").exists() {
        if let Some(parent) = base_dir.parent() {
            if parent.join("database").join("database.sql").exists() {
                base_dir = parent.to_path_buf();
            }
        }
    }

    let db_path = base_dir.join("sql_app.db");
    let sql_path = base_dir.join("database").join("database.sql");

    println!("Initializing SQLite database at {:?} ...", db_path);
    let conn = database::init_db(
        db_path.to_str().unwrap_or("sql_app.db"),
        sql_path.to_str().unwrap_or("database/database.sql"),
    )?;

    let state = Arc::new(AppState {
        db: Mutex::new(conn),
        base_dir: base_dir.clone(),
    });

    let app = create_app(state);

    let port: u16 = env::var("PORT")
        .unwrap_or_else(|_| "8080".to_string())
        .parse()
        .unwrap_or(8080);

    let addr = SocketAddr::from(([0, 0, 0, 0], port));
    println!("Starting Rust Ant Colony Backend on http://{} ...", addr);

    let listener = tokio::net::TcpListener::bind(addr).await?;
    axum::serve(listener, app).await?;

    Ok(())
}
