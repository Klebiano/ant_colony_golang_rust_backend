use std::fs;
use std::path::Path;
use rusqlite::{Connection, Result};
use crate::models::{AntColonyMap, Downtime, Subsystems};

pub fn init_db(db_path: &str, sql_path: &str) -> Result<Connection> {
    let conn = if db_path == ":memory:" {
        Connection::open_in_memory()?
    } else {
        Connection::open(db_path)?
    };

    conn.execute_batch("PRAGMA foreign_keys = OFF;")?;

    let count: i64 = {
        let mut stmt = conn.prepare("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='wind_farm_map_points'")?;
        stmt.query_row([], |row| row.get(0)).unwrap_or(0)
    };

    if count == 0 && !sql_path.is_empty() && Path::new(sql_path).exists() {
        if let Ok(sql_content) = fs::read_to_string(sql_path) {
            conn.execute_batch(&sql_content)?;
        }
    }

    Ok(conn)
}

pub fn get_turbines_map(conn: &Connection) -> Result<Vec<AntColonyMap>> {
    let query = r#"SELECT 
        "TURBINE_ID" as turbine_id,
        "TURBINE_NAME" as turbine_name,
        "LATITUDE" as latitude,
        "LONGITUDE" as longitude
    FROM wind_farm_map_points"#;

    let mut stmt = conn.prepare(query)?;
    let rows = stmt.query_map([], |row| {
        Ok(AntColonyMap {
            turbine_id: row.get(0)?,
            turbine_name: row.get(1)?,
            latitude: row.get(2)?,
            longitude: row.get(3)?,
        })
    })?;

    let mut result = Vec::new();
    for row in rows {
        result.push(row?);
    }
    Ok(result)
}

pub fn get_subsystems(conn: &Connection) -> Result<Vec<Subsystems>> {
    let query = r#"SELECT 
        "SUBSYSTEM_ID" as subsystem_id,
        "SUBSYSTEM_NAME" as subsystem_name
    FROM subsystem"#;

    let mut stmt = conn.prepare(query)?;
    let rows = stmt.query_map([], |row| {
        Ok(Subsystems {
            subsystem_id: row.get(0)?,
            subsystem_name: row.get(1)?,
        })
    })?;

    let mut result = Vec::new();
    for row in rows {
        result.push(row?);
    }
    Ok(result)
}

pub fn get_downtimes(conn: &Connection) -> Result<Vec<Downtime>> {
    let query = r#"SELECT 
        sub."SUBSYSTEM_ID" as subsystem_id,
        sub."SUBSYSTEM_NAME" as subsystem_name,
        dwt."FAULT_TYPE" as fault_type,
        dwt."ANUAL_FAILURE_RATE" as anual_failure_rate,
        dwt."FAULT_DOWNTIME_DAYS" as fault_downtime_days
    FROM subsystem sub
    LEFT JOIN downtime as dwt on sub."SUBSYSTEM_ID" = dwt."SUBSYSTEM_ID""#;

    let mut stmt = conn.prepare(query)?;
    let rows = stmt.query_map([], |row| {
        let fault_type: Option<String> = row.get(2)?;
        let annual_rate: Option<f64> = row.get(3)?;
        let dt_days: Option<f64> = row.get(4)?;

        Ok(Downtime {
            subsystem_id: row.get(0)?,
            subsystem_name: row.get(1)?,
            fault_type: fault_type.unwrap_or_default(),
            anual_failure_rate: annual_rate.unwrap_or(0.0),
            fault_downtime_days: dt_days.unwrap_or(0.0),
        })
    })?;

    let mut result = Vec::new();
    for row in rows {
        result.push(row?);
    }
    Ok(result)
}
