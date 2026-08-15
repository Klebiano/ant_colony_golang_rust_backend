package models

type AntColonyMap struct {
	TurbineID   int     `json:"turbine_id"`
	TurbineName string  `json:"turbine_name"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
}

type Subsystems struct {
	SubsystemID   int    `json:"subsystem_id"`
	SubsystemName string `json:"subsystem_name"`
}

type Downtime struct {
	SubsystemID       int     `json:"subsystem_id"`
	SubsystemName     string  `json:"subsystem_name"`
	FaultType         string  `json:"fault_type"`
	AnualFailureRate  float64 `json:"anual_failure_rate"`
	FaultDowntimeDays float64 `json:"fault_downtime_days"`
}

type TurbineFaults struct {
	TurbineID     int     `json:"turbine_id"`
	TurbineName   *string `json:"turbine_name"`
	SubsystemName string  `json:"subsystem_name"`
	FaultType     string  `json:"fault_type"`
}

type AntColonyPath struct {
	TurbineOrder       []string `json:"turbine_order"`
	TurbineOrderToShow []string `json:"turbine_order_to_show"`
	BestPath           []int    `json:"best_path"`
	BestPathLength     float64  `json:"best_path_length"`
	BestDowntimeDays   float64  `json:"best_downtime_days"`
	BestPathLenDowntime *float64 `json:"best_path_len_downtime,omitempty"`
	TimeToRunSec       *float64 `json:"time_to_run_sec,omitempty"`
}

type TurbineFaultPoint struct {
	TurbineID             int     `json:"turbine_id"`
	TurbineName           string  `json:"turbine_name"`
	Latitude              float64 `json:"latitude"`
	Longitude             float64 `json:"longitude"`
	SubsystemName         string  `json:"subsystem_name"`
	FaultType             string  `json:"fault_type"`
	AnualFailureRate      float64 `json:"anual_failure_rate"`
	FaultDowntimeDays     float64 `json:"fault_downtime_days"`
	FaultDowntimeDaysNorm float64 `json:"fault_downtime_days_norm"`
	LatitudeNorm          float64 `json:"latitude_norm"`
	LongitudeNorm         float64 `json:"longitude_norm"`
}
