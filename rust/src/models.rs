use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct AntColonyMap {
    pub turbine_id: i32,
    pub turbine_name: String,
    pub latitude: f64,
    pub longitude: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct Subsystems {
    pub subsystem_id: i32,
    pub subsystem_name: String,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Default)]
pub struct Downtime {
    pub subsystem_id: i32,
    pub subsystem_name: String,
    #[serde(default)]
    pub fault_type: String,
    #[serde(default)]
    pub anual_failure_rate: f64,
    #[serde(default)]
    pub fault_downtime_days: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct TurbineFaults {
    pub turbine_id: i32,
    #[serde(default)]
    pub turbine_name: Option<String>,
    pub subsystem_name: String,
    pub fault_type: String,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct AntColonyPath {
    pub turbine_order: Vec<String>,
    pub turbine_order_to_show: Vec<String>,
    pub best_path: Vec<usize>,
    pub best_path_length: f64,
    pub best_downtime_days: f64,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub best_path_len_downtime: Option<f64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub time_to_run_sec: Option<f64>,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Default)]
pub struct TurbineFaultPoint {
    pub turbine_id: i32,
    pub turbine_name: String,
    pub latitude: f64,
    pub longitude: f64,
    pub subsystem_name: String,
    pub fault_type: String,
    pub anual_failure_rate: f64,
    pub fault_downtime_days: f64,
    pub fault_downtime_days_norm: f64,
    pub latitude_norm: f64,
    pub longitude_norm: f64,
}
