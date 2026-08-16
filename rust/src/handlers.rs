use std::collections::HashMap;
use std::fs::File;
use std::path::{Path, PathBuf};
use std::sync::{Arc, Mutex};
use std::time::Instant;

use axum::extract::{Query, State};
use axum::http::StatusCode;
use axum::response::Json;
use serde::Deserialize;

use crate::algorithms::{
    format_turbine_order_to_show, AntColony, GeneticAlgorithm,
};
use crate::database;
use crate::models::{
    AntColonyMap, AntColonyPath, Subsystems, TurbineFaultPoint, TurbineFaults,
};

pub struct AppState {
    pub db: Mutex<rusqlite::Connection>,
    pub base_dir: PathBuf,
}

#[derive(Debug, Deserialize)]
pub struct OptimizerQuery {
    pub algorithm: Option<String>,
}

pub async fn get_turbines_map(
    State(state): State<Arc<AppState>>,
) -> Result<Json<Vec<AntColonyMap>>, (StatusCode, String)> {
    let conn = state.db.lock().map_err(|e| {
        (
            StatusCode::INTERNAL_SERVER_ERROR,
            format!("Database lock error: {}", e),
        )
    })?;

    database::get_turbines_map(&conn)
        .map(Json)
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))
}

pub async fn get_subsystems(
    State(state): State<Arc<AppState>>,
) -> Result<Json<Vec<Subsystems>>, (StatusCode, String)> {
    let conn = state.db.lock().map_err(|e| {
        (
            StatusCode::INTERNAL_SERVER_ERROR,
            format!("Database lock error: {}", e),
        )
    })?;

    database::get_subsystems(&conn)
        .map(Json)
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))
}

pub async fn run_route_optimizer(
    State(state): State<Arc<AppState>>,
    Query(query): Query<OptimizerQuery>,
    payload: Option<Json<Vec<TurbineFaults>>>,
) -> Result<Json<AntColonyPath>, (StatusCode, String)> {
    let algo = query
        .algorithm
        .unwrap_or_else(|| "Ant Colony".to_string());

    let faults = payload.map(|Json(f)| f).unwrap_or_default();

    let mut points = if !faults.is_empty() {
        let conn = state.db.lock().map_err(|e| {
            (
                StatusCode::INTERNAL_SERVER_ERROR,
                format!("Database lock error: {}", e),
            )
        })?;
        build_points_from_faults(&conn, &faults)
            .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?
    } else {
        let csv_path = state
            .base_dir
            .join("tests")
            .join("inputs")
            .join("problem_5_turbines.csv");
        load_points_from_csv(&csv_path)
            .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?
    };

    normalize_coordinates(&mut points);

    let n_turbines = if !points.is_empty() { points.len() - 1 } else { 0 };
    let start_time = Instant::now();

    let mut rng = rand::thread_rng();

    let (turbine_order, best_path, best_path_len, best_downtime, best_path_len_dt) = match algo.as_str() {
        "Ant Colony" => {
            let (n_ants, alpha, beta, rho) = if n_turbines <= 10 {
                (3, 5.0, 1.5, 0.5)
            } else {
                (8, 5.0, 2.0, 0.5)
            };

            let mut aco = AntColony::new(points, n_ants, 200, alpha, beta, rho, 100.0);
            aco.optimize(&mut rng);

            (
                aco.turbine_order,
                aco.best_path,
                aco.best_path_length,
                aco.best_downtime_days,
                aco.best_path_len_downtime,
            )
        }
        "Genetic" | "Memetic" => {
            let local_search = algo == "Memetic";
            let (mutation_rate, population_size, n_generations) = if algo == "Genetic" {
                if n_turbines >= 15 {
                    (0.1, 100, 50)
                } else {
                    (0.2, 50, 50)
                }
            } else {
                // Memetic
                if n_turbines >= 40 {
                    (0.1, 150, 50)
                } else {
                    (0.2, 50, 10)
                }
            };

            let mut ga = GeneticAlgorithm::new(
                points,
                population_size,
                n_generations,
                mutation_rate,
                local_search,
                &mut rng,
            );
            ga.evolve(&mut rng);

            (
                ga.turbine_order,
                ga.best_path,
                ga.best_path_length,
                ga.best_downtime_days,
                ga.best_path_len_downtime,
            )
        }
        _ => {
            return Err((StatusCode::BAD_REQUEST, "Invalid algorithm".to_string()));
        }
    };

    let duration_sec = start_time.elapsed().as_secs_f64();
    let turbine_order_to_show = format_turbine_order_to_show(&turbine_order);

    let resp = AntColonyPath {
        turbine_order,
        turbine_order_to_show,
        best_path,
        best_path_length: best_path_len,
        best_downtime_days: best_downtime,
        best_path_len_downtime: Some(best_path_len_dt),
        time_to_run_sec: Some(duration_sec),
    };

    Ok(Json(resp))
}

pub fn build_points_from_faults(
    conn: &rusqlite::Connection,
    faults: &[TurbineFaults],
) -> rusqlite::Result<Vec<TurbineFaultPoint>> {
    let turbines_map = database::get_turbines_map(conn)?;
    let downtimes = database::get_downtimes(conn)?;

    let mut min_downtime = f64::INFINITY;
    let mut max_downtime = f64::NEG_INFINITY;
    for d in &downtimes {
        if d.fault_downtime_days < min_downtime {
            min_downtime = d.fault_downtime_days;
        }
        if d.fault_downtime_days > max_downtime {
            max_downtime = d.fault_downtime_days;
        }
    }

    struct DtInfo {
        annual_rate: f64,
        downtime_days: f64,
        downtime_days_norm: f64,
    }

    let mut downtimes_norm_map: HashMap<String, HashMap<String, DtInfo>> = HashMap::new();
    for d in &downtimes {
        let norm = if max_downtime > min_downtime {
            (d.fault_downtime_days - min_downtime) / (max_downtime - min_downtime)
        } else {
            0.0
        };

        downtimes_norm_map
            .entry(d.subsystem_name.clone())
            .or_default()
            .insert(
                d.fault_type.clone(),
                DtInfo {
                    annual_rate: d.anual_failure_rate,
                    downtime_days: d.fault_downtime_days,
                    downtime_days_norm: norm,
                },
            );
    }

    let mut turbines_map_by_id: HashMap<i32, AntColonyMap> = HashMap::new();
    let mut doca_turbine: Option<AntColonyMap> = None;
    for t in turbines_map {
        if t.turbine_name == "Doca" {
            doca_turbine = Some(t.clone());
        }
        turbines_map_by_id.insert(t.turbine_id, t);
    }

    let mut points = Vec::new();
    for f in faults {
        if let Some(t_info) = turbines_map_by_id.get(&f.turbine_id) {
            let mut pt = TurbineFaultPoint {
                turbine_id: f.turbine_id,
                turbine_name: t_info.turbine_name.clone(),
                latitude: t_info.latitude,
                longitude: t_info.longitude,
                subsystem_name: f.subsystem_name.clone(),
                fault_type: f.fault_type.clone(),
                anual_failure_rate: 0.0,
                fault_downtime_days: 0.0,
                fault_downtime_days_norm: 0.0,
                latitude_norm: 0.0,
                longitude_norm: 0.0,
            };

            if let Some(sub_map) = downtimes_norm_map.get(&f.subsystem_name) {
                if let Some(dt_info) = sub_map.get(&f.fault_type) {
                    pt.anual_failure_rate = dt_info.annual_rate;
                    pt.fault_downtime_days = dt_info.downtime_days;
                    pt.fault_downtime_days_norm = dt_info.downtime_days_norm;
                }
            }
            points.push(pt);
        }
    }

    // Append Doca
    if let Some(doca) = doca_turbine {
        points.push(TurbineFaultPoint {
            turbine_id: doca.turbine_id,
            turbine_name: doca.turbine_name,
            latitude: doca.latitude,
            longitude: doca.longitude,
            subsystem_name: "0".to_string(),
            fault_type: "0".to_string(),
            anual_failure_rate: 0.0,
            fault_downtime_days: 0.0,
            fault_downtime_days_norm: 0.0,
            latitude_norm: 0.0,
            longitude_norm: 0.0,
        });
    }

    // Sort by TurbineID
    points.sort_by_key(|p| p.turbine_id);

    Ok(points)
}

pub fn load_points_from_csv<P: AsRef<Path>>(csv_path: P) -> Result<Vec<TurbineFaultPoint>, Box<dyn std::error::Error + Send + Sync>> {
    let file = File::open(csv_path)?;
    let mut rdr = csv::Reader::from_reader(file);

    let headers = rdr.headers()?.clone();
    let mut col_idx: HashMap<String, usize> = HashMap::new();
    for (i, h) in headers.iter().enumerate() {
        col_idx.insert(h.trim().to_string(), i);
    }

    let mut points = Vec::new();
    for result in rdr.records() {
        let record = result?;
        if record.is_empty() {
            continue;
        }

        let t_id = record
            .get(*col_idx.get("turbine_id").unwrap_or(&0))
            .unwrap_or("0")
            .parse::<i32>()
            .unwrap_or(0);
        let t_name = record
            .get(*col_idx.get("turbine_name").unwrap_or(&0))
            .unwrap_or("")
            .to_string();
        let sub_name = record
            .get(*col_idx.get("subsystem_name").unwrap_or(&0))
            .unwrap_or("")
            .to_string();
        let fault_type = record
            .get(*col_idx.get("fault_type").unwrap_or(&0))
            .unwrap_or("")
            .to_string();
        let lat = record
            .get(*col_idx.get("latitude").unwrap_or(&0))
            .unwrap_or("0")
            .parse::<f64>()
            .unwrap_or(0.0);
        let lon = record
            .get(*col_idx.get("longitude").unwrap_or(&0))
            .unwrap_or("0")
            .parse::<f64>()
            .unwrap_or(0.0);
        let annual_rate = record
            .get(*col_idx.get("anual_failure_rate").unwrap_or(&0))
            .unwrap_or("0")
            .parse::<f64>()
            .unwrap_or(0.0);
        let dt_days = record
            .get(*col_idx.get("fault_downtime_days").unwrap_or(&0))
            .unwrap_or("0")
            .parse::<f64>()
            .unwrap_or(0.0);
        let dt_norm = record
            .get(*col_idx.get("fault_downtime_days_norm").unwrap_or(&0))
            .unwrap_or("0")
            .parse::<f64>()
            .unwrap_or(0.0);

        points.push(TurbineFaultPoint {
            turbine_id: t_id,
            turbine_name: t_name,
            latitude: lat,
            longitude: lon,
            subsystem_name: sub_name,
            fault_type,
            anual_failure_rate: annual_rate,
            fault_downtime_days: dt_days,
            fault_downtime_days_norm: dt_norm,
            latitude_norm: 0.0,
            longitude_norm: 0.0,
        });
    }

    Ok(points)
}

pub fn normalize_coordinates(points: &mut [TurbineFaultPoint]) {
    if points.is_empty() {
        return;
    }

    let mut min_lat = points[0].latitude;
    let mut max_lat = points[0].latitude;
    let mut min_lon = points[0].longitude;
    let mut max_lon = points[0].longitude;

    for p in points.iter() {
        if p.latitude < min_lat {
            min_lat = p.latitude;
        }
        if p.latitude > max_lat {
            max_lat = p.latitude;
        }
        if p.longitude < min_lon {
            min_lon = p.longitude;
        }
        if p.longitude > max_lon {
            max_lon = p.longitude;
        }
    }

    for p in points.iter_mut() {
        p.latitude_norm = if max_lat > min_lat {
            (p.latitude - min_lat) / (max_lat - min_lat)
        } else {
            0.0
        };

        p.longitude_norm = if max_lon > min_lon {
            (p.longitude - min_lon) / (max_lon - min_lon)
        } else {
            0.0
        };
    }
}
