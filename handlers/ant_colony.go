package handlers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"math/rand/v2"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"database/sql"
	"ant_colony_golang_backend/algorithms"
	"ant_colony_golang_backend/database"
	"ant_colony_golang_backend/models"
)

type AntColonyHandler struct {
	DB      *sql.DB
	BaseDir string
}

func NewAntColonyHandler(db *sql.DB, baseDir string) *AntColonyHandler {
	return &AntColonyHandler{
		DB:      db,
		BaseDir: baseDir,
	}
}

func (h *AntColonyHandler) GetTurbinesMap(w http.ResponseWriter, r *http.Request) {
	turbines, err := database.GetTurbinesMap(h.DB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(turbines)
}

func (h *AntColonyHandler) GetSubsystems(w http.ResponseWriter, r *http.Request) {
	subsystems, err := database.GetSubsystems(h.DB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subsystems)
}

func (h *AntColonyHandler) RunRouteOptimizer(w http.ResponseWriter, r *http.Request) {
	// Query params: algorithm
	algoQuery := r.URL.Query()["algorithm"]
	algo := "Ant Colony"
	if len(algoQuery) > 0 {
		algo = algoQuery[0]
	}

	var faults []models.TurbineFaults
	if r.Body != nil {
		json.NewDecoder(r.Body).Decode(&faults)
	}

	var points []models.TurbineFaultPoint
	var err error

	if len(faults) > 0 {
		points, err = h.buildPointsFromFaults(faults)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error building points: %v", err), http.StatusInternalServerError)
			return
		}
	} else {
		// Fallback to problem_5_turbines.csv
		csvPath := filepath.Join(h.BaseDir, "tests", "inputs", "problem_5_turbines.csv")
		points, err = h.loadPointsFromCSV(csvPath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error loading CSV: %v", err), http.StatusInternalServerError)
			return
		}
	}

	normalizeCoordinates(points)

	nTurbines := len(points) - 1
	startTime := time.Now()

	var turbineOrder []string
	var turbineOrderToShow []string
	var bestPath []int
	var bestPathLen, bestDowntime, bestPathLenDt float64

	rng := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 1))

	switch algo {
	case "Ant Colony":
		var nAnts int
		var alpha, beta, rho float64
		if nTurbines <= 10 {
			nAnts = 3
			alpha = 5.0
			beta = 1.5
			rho = 0.5
		} else {
			nAnts = 8
			alpha = 5.0
			beta = 2.0
			rho = 0.5
		}
		aco := algorithms.NewAntColony(points, nAnts, 200, alpha, beta, rho, 100.0, rng)
		aco.Optimize()

		turbineOrder = aco.TurbineOrder
		bestPath = aco.BestPath
		bestPathLen = aco.BestPathLength
		bestDowntime = aco.BestDowntimeDays
		bestPathLenDt = aco.BestPathLenDowntime

	case "Genetic", "Memetic":
		var mutationRate float64
		var populationSize, nGenerations int
		localSearch := (algo == "Memetic")

		if algo == "Genetic" {
			if nTurbines >= 15 {
				mutationRate = 0.1
				populationSize = 100
				nGenerations = 50
			} else {
				mutationRate = 0.2
				populationSize = 50
				nGenerations = 50
			}
		} else { // Memetic
			if nTurbines >= 40 {
				mutationRate = 0.1
				populationSize = 150
				nGenerations = 50
			} else {
				mutationRate = 0.2
				populationSize = 50
				nGenerations = 10
			}
		}

		ga := algorithms.NewGeneticAlgorithm(points, populationSize, nGenerations, mutationRate, localSearch, rng)
		ga.Evolve()

		turbineOrder = ga.TurbineOrder
		bestPath = ga.BestPath
		bestPathLen = ga.BestPathLength
		bestDowntime = ga.BestDowntimeDays
		bestPathLenDt = ga.BestPathLenDowntime

	default:
		http.Error(w, "Invalid algorithm", http.StatusBadRequest)
		return
	}

	durationSec := time.Since(startTime).Seconds()
	turbineOrderToShow = algorithms.FormatTurbineOrderToShow(turbineOrder)

	resp := models.AntColonyPath{
		TurbineOrder:       turbineOrder,
		TurbineOrderToShow: turbineOrderToShow,
		BestPath:           bestPath,
		BestPathLength:     bestPathLen,
		BestDowntimeDays:   bestDowntime,
		BestPathLenDowntime: &bestPathLenDt,
		TimeToRunSec:       &durationSec,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *AntColonyHandler) buildPointsFromFaults(faults []models.TurbineFaults) ([]models.TurbineFaultPoint, error) {
	turbinesMap, err := database.GetTurbinesMap(h.DB)
	if err != nil {
		return nil, err
	}
	downtimes, err := database.GetDowntimes(h.DB)
	if err != nil {
		return nil, err
	}

	// Calculate normalized downtime days
	minDowntime := math.Inf(1)
	maxDowntime := math.Inf(-1)
	for _, d := range downtimes {
		if d.FaultDowntimeDays < minDowntime {
			minDowntime = d.FaultDowntimeDays
		}
		if d.FaultDowntimeDays > maxDowntime {
			maxDowntime = d.FaultDowntimeDays
		}
	}

	downtimesNormMap := make(map[string]map[string]struct {
		AnnualRate       float64
		DowntimeDays     float64
		DowntimeDaysNorm float64
	})

	for _, d := range downtimes {
		if _, ok := downtimesNormMap[d.SubsystemName]; !ok {
			downtimesNormMap[d.SubsystemName] = make(map[string]struct {
				AnnualRate       float64
				DowntimeDays     float64
				DowntimeDaysNorm float64
			})
		}
		norm := 0.0
		if maxDowntime > minDowntime {
			norm = (d.FaultDowntimeDays - minDowntime) / (maxDowntime - minDowntime)
		}
		downtimesNormMap[d.SubsystemName][d.FaultType] = struct {
			AnnualRate       float64
			DowntimeDays     float64
			DowntimeDaysNorm float64
		}{
			AnnualRate:       d.AnualFailureRate,
			DowntimeDays:     d.FaultDowntimeDays,
			DowntimeDaysNorm: norm,
		}
	}

	turbinesMapByID := make(map[int]models.AntColonyMap)
	var docaTurbine models.AntColonyMap
	for _, t := range turbinesMap {
		turbinesMapByID[t.TurbineID] = t
		if t.TurbineName == "Doca" {
			docaTurbine = t
		}
	}

	var points []models.TurbineFaultPoint
	for _, f := range faults {
		tInfo, ok := turbinesMapByID[f.TurbineID]
		if !ok {
			continue
		}
		pt := models.TurbineFaultPoint{
			TurbineID:     f.TurbineID,
			TurbineName:   tInfo.TurbineName,
			Latitude:      tInfo.Latitude,
			Longitude:     tInfo.Longitude,
			SubsystemName: f.SubsystemName,
			FaultType:     f.FaultType,
		}
		if subMap, ok := downtimesNormMap[f.SubsystemName]; ok {
			if dtInfo, ok := subMap[f.FaultType]; ok {
				pt.AnualFailureRate = dtInfo.AnnualRate
				pt.FaultDowntimeDays = dtInfo.DowntimeDays
				pt.FaultDowntimeDaysNorm = dtInfo.DowntimeDaysNorm
			}
		}
		points = append(points, pt)
	}

	// Append Doca
	docaPt := models.TurbineFaultPoint{
		TurbineID:             docaTurbine.TurbineID,
		TurbineName:           docaTurbine.TurbineName,
		Latitude:              docaTurbine.Latitude,
		Longitude:             docaTurbine.Longitude,
		SubsystemName:         "0",
		FaultType:             "0",
		AnualFailureRate:      0,
		FaultDowntimeDays:     0,
		FaultDowntimeDaysNorm: 0,
	}
	points = append(points, docaPt)

	// Sort by TurbineID
	sort.Slice(points, func(i, j int) bool {
		return points[i].TurbineID < points[j].TurbineID
	})

	return points, nil
}

func (h *AntColonyHandler) loadPointsFromCSV(csvPath string) ([]models.TurbineFaultPoint, error) {
	file, err := os.Open(csvPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("empty CSV")
	}

	header := records[0]
	colIdx := make(map[string]int)
	for i, hName := range header {
		colIdx[strings.TrimSpace(hName)] = i
	}

	var points []models.TurbineFaultPoint
	for _, row := range records[1:] {
		if len(row) == 0 {
			continue
		}
		tID, _ := strconv.Atoi(row[colIdx["turbine_id"]])
		lat, _ := strconv.ParseFloat(row[colIdx["latitude"]], 64)
		lon, _ := strconv.ParseFloat(row[colIdx["longitude"]], 64)
		annualRate, _ := strconv.ParseFloat(row[colIdx["anual_failure_rate"]], 64)
		dtDays, _ := strconv.ParseFloat(row[colIdx["fault_downtime_days"]], 64)
		dtNorm, _ := strconv.ParseFloat(row[colIdx["fault_downtime_days_norm"]], 64)

		pt := models.TurbineFaultPoint{
			TurbineID:             tID,
			TurbineName:           row[colIdx["turbine_name"]],
			SubsystemName:         row[colIdx["subsystem_name"]],
			FaultType:             row[colIdx["fault_type"]],
			Latitude:              lat,
			Longitude:             lon,
			AnualFailureRate:      annualRate,
			FaultDowntimeDays:     dtDays,
			FaultDowntimeDaysNorm: dtNorm,
		}
		points = append(points, pt)
	}

	return points, nil
}

func normalizeCoordinates(points []models.TurbineFaultPoint) {
	if len(points) == 0 {
		return
	}
	minLat, maxLat := points[0].Latitude, points[0].Latitude
	minLon, maxLon := points[0].Longitude, points[0].Longitude

	for _, p := range points {
		if p.Latitude < minLat {
			minLat = p.Latitude
		}
		if p.Latitude > maxLat {
			maxLat = p.Latitude
		}
		if p.Longitude < minLon {
			minLon = p.Longitude
		}
		if p.Longitude > maxLon {
			maxLon = p.Longitude
		}
	}

	for i := range points {
		if maxLat > minLat {
			points[i].LatitudeNorm = (points[i].Latitude - minLat) / (maxLat - minLat)
		} else {
			points[i].LatitudeNorm = 0
		}

		if maxLon > minLon {
			points[i].LongitudeNorm = (points[i].Longitude - minLon) / (maxLon - minLon)
		} else {
			points[i].LongitudeNorm = 0
		}
	}
}
