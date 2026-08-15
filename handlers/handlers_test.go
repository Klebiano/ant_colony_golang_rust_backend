package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"

	"ant_colony_golang_backend/database"
	"ant_colony_golang_backend/models"
)

func setupTestDB(t *testing.T) (*AntColonyHandler, string) {
	_, filename, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(filepath.Dir(filename))

	sqlPath := filepath.Join(baseDir, "database", "database.sql")
	db, err := database.InitDB(":memory:", sqlPath)
	if err != nil {
		t.Fatalf("Failed to init in-memory db: %v", err)
	}

	handler := NewAntColonyHandler(db, baseDir)
	return handler, baseDir
}

func TestGetTurbinesMap(t *testing.T) {
	handler, _ := setupTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/ant-colony/get-turbines-map", nil)
	rr := httptest.NewRecorder()

	handler.GetTurbinesMap(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", rr.Code)
	}

	var data []models.AntColonyMap
	if err := json.Unmarshal(rr.Body.Bytes(), &data); err != nil {
		t.Fatalf("Failed to decode JSON: %v", err)
	}

	if len(data) == 0 {
		t.Fatalf("Expected non-empty turbines map list")
	}

	if data[0].TurbineID == 0 || data[0].TurbineName == "" {
		t.Fatalf("Invalid turbine map entry: %+v", data[0])
	}
}

func TestGetSubsystems(t *testing.T) {
	handler, _ := setupTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/ant-colony/get-subsystems", nil)
	rr := httptest.NewRecorder()

	handler.GetSubsystems(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", rr.Code)
	}

	var data []models.Subsystems
	if err := json.Unmarshal(rr.Body.Bytes(), &data); err != nil {
		t.Fatalf("Failed to decode JSON: %v", err)
	}

	if len(data) == 0 {
		t.Fatalf("Expected non-empty subsystems list")
	}

	if data[0].SubsystemID == 0 || data[0].SubsystemName == "" {
		t.Fatalf("Invalid subsystem entry: %+v", data[0])
	}
}

func TestRunRouteOptimizerAntColony(t *testing.T) {
	handler, _ := setupTestDB(t)

	payload := []models.TurbineFaults{
		{TurbineID: 2, SubsystemName: "Electrical System", FaultType: "Minor"},
		{TurbineID: 3, SubsystemName: "Rotor Hub", FaultType: "Major"},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/ant-colony/run-route-optimizer?algorithm=Ant%20Colony", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	handler.RunRouteOptimizer(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	var res models.AntColonyPath
	if err := json.Unmarshal(rr.Body.Bytes(), &res); err != nil {
		t.Fatalf("Failed to decode response JSON: %v", err)
	}

	if len(res.TurbineOrder) == 0 || len(res.TurbineOrderToShow) == 0 {
		t.Fatalf("Expected non-empty turbine order")
	}
	if res.BestPathLength <= 0 {
		t.Fatalf("Expected positive best path length")
	}
}

func TestRunRouteOptimizerGenetic(t *testing.T) {
	handler, _ := setupTestDB(t)

	payload := []models.TurbineFaults{
		{TurbineID: 2, SubsystemName: "Electrical System", FaultType: "Minor"},
		{TurbineID: 3, SubsystemName: "Rotor Hub", FaultType: "Major"},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/ant-colony/run-route-optimizer?algorithm=Genetic", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	handler.RunRouteOptimizer(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", rr.Code)
	}

	var res models.AntColonyPath
	if err := json.Unmarshal(rr.Body.Bytes(), &res); err != nil {
		t.Fatalf("Failed to decode response JSON: %v", err)
	}

	if len(res.TurbineOrder) == 0 {
		t.Fatalf("Expected non-empty turbine order")
	}
}

func TestRunRouteOptimizerMemetic(t *testing.T) {
	handler, _ := setupTestDB(t)

	payload := []models.TurbineFaults{
		{TurbineID: 2, SubsystemName: "Electrical System", FaultType: "Minor"},
		{TurbineID: 3, SubsystemName: "Rotor Hub", FaultType: "Major"},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/ant-colony/run-route-optimizer?algorithm=Memetic", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	handler.RunRouteOptimizer(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", rr.Code)
	}

	var res models.AntColonyPath
	if err := json.Unmarshal(rr.Body.Bytes(), &res); err != nil {
		t.Fatalf("Failed to decode response JSON: %v", err)
	}

	if len(res.TurbineOrder) == 0 {
		t.Fatalf("Expected non-empty turbine order")
	}
	if len(res.TurbineOrderToShow) == 0 || res.TurbineOrderToShow[0] != "Doca" || res.TurbineOrderToShow[len(res.TurbineOrderToShow)-1] != "Doca" {
		t.Fatalf("Expected TurbineOrderToShow to start and end with Doca: %v", res.TurbineOrderToShow)
	}
}

func TestRunRouteOptimizerSingleTurbine(t *testing.T) {
	handler, _ := setupTestDB(t)

	payload := []models.TurbineFaults{
		{TurbineID: 2, SubsystemName: "Electrical System", FaultType: "Minor"},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/ant-colony/run-route-optimizer?algorithm=Ant%20Colony", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	handler.RunRouteOptimizer(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", rr.Code)
	}

	var res models.AntColonyPath
	if err := json.Unmarshal(rr.Body.Bytes(), &res); err != nil {
		t.Fatalf("Failed to decode response JSON: %v", err)
	}

	if len(res.TurbineOrderToShow) != 3 {
		t.Fatalf("Expected 3 items in TurbineOrderToShow (Doca -> Turbine -> Doca), got %d: %v", len(res.TurbineOrderToShow), res.TurbineOrderToShow)
	}
	if res.TurbineOrderToShow[0] != "Doca" || res.TurbineOrderToShow[2] != "Doca" {
		t.Fatalf("Expected start and end at Doca: %v", res.TurbineOrderToShow)
	}
}

func TestRunRouteOptimizerEmptyPayload(t *testing.T) {
	handler, _ := setupTestDB(t)

	body, _ := json.Marshal([]models.TurbineFaults{})
	req := httptest.NewRequest(http.MethodPost, "/ant-colony/run-route-optimizer?algorithm=Genetic", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	handler.RunRouteOptimizer(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", rr.Code)
	}

	var res models.AntColonyPath
	if err := json.Unmarshal(rr.Body.Bytes(), &res); err != nil {
		t.Fatalf("Failed to decode response JSON: %v", err)
	}

	if len(res.TurbineOrderToShow) == 0 {
		t.Fatalf("Expected non-empty turbine order to show")
	}
	if res.TurbineOrderToShow[0] != "Doca" || res.TurbineOrderToShow[len(res.TurbineOrderToShow)-1] != "Doca" {
		t.Fatalf("Expected start and end with Doca: %v", res.TurbineOrderToShow)
	}
}
