package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"ant_colony_golang_backend/models"
)

func InitDB(dbPath string, sqlPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// Check if tables exist
	var count int
	err = db.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='wind_farm_map_points'").Scan(&count)
	if err != nil || count == 0 {
		if sqlPath != "" {
			sqlContent, err := os.ReadFile(sqlPath)
			if err == nil {
				_, err = db.Exec(string(sqlContent))
				if err != nil {
					return nil, fmt.Errorf("failed to execute database.sql: %w", err)
				}
			}
		}
	}

	return db, nil
}

func GetTurbinesMap(db *sql.DB) ([]models.AntColonyMap, error) {
	query := `SELECT 
		"TURBINE_ID" as turbine_id,
		"TURBINE_NAME" as turbine_name,
		"LATITUDE" as latitude,
		"LONGITUDE" as longitude
	FROM wind_farm_map_points`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.AntColonyMap
	for rows.Next() {
		var item models.AntColonyMap
		if err := rows.Scan(&item.TurbineID, &item.TurbineName, &item.Latitude, &item.Longitude); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, nil
}

func GetSubsystems(db *sql.DB) ([]models.Subsystems, error) {
	query := `SELECT 
		"SUBSYSTEM_ID" as subsystem_id,
		"SUBSYSTEM_NAME" as subsystem_name
	FROM subsystem`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.Subsystems
	for rows.Next() {
		var item models.Subsystems
		if err := rows.Scan(&item.SubsystemID, &item.SubsystemName); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, nil
}

func GetDowntimes(db *sql.DB) ([]models.Downtime, error) {
	query := `SELECT 
		sub."SUBSYSTEM_ID" as subsystem_id,
		sub."SUBSYSTEM_NAME" as subsystem_name,
		dwt."FAULT_TYPE" as fault_type,
		dwt."ANUAL_FAILURE_RATE" as anual_failure_rate,
		dwt."FAULT_DOWNTIME_DAYS" as fault_downtime_days
	FROM subsystem sub
	LEFT JOIN downtime as dwt on sub."SUBSYSTEM_ID" = dwt."SUBSYSTEM_ID"`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.Downtime
	for rows.Next() {
		var item models.Downtime
		var faultType sql.NullString
		var annualRate sql.NullFloat64
		var downtimeDays sql.NullFloat64

		if err := rows.Scan(&item.SubsystemID, &item.SubsystemName, &faultType, &annualRate, &downtimeDays); err != nil {
			return nil, err
		}
		if faultType.Valid {
			item.FaultType = faultType.String
		}
		if annualRate.Valid {
			item.AnualFailureRate = annualRate.Float64
		}
		if downtimeDays.Valid {
			item.FaultDowntimeDays = downtimeDays.Float64
		}
		result = append(result, item)
	}
	return result, nil
}
