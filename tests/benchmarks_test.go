package tests

import (
	"encoding/csv"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"ant_colony_golang_backend/algorithms"
	"ant_colony_golang_backend/models"
)

func loadCSVPoints(problemFile string) ([]models.TurbineFaultPoint, error) {
	_, filename, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(filepath.Dir(filename))
	path := filepath.Join(baseDir, "tests", "inputs", problemFile)

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
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

	// Normalize
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
		}
		if maxLon > minLon {
			points[i].LongitudeNorm = (points[i].Longitude - minLon) / (maxLon - minLon)
		}
	}

	return points, nil
}

func BenchmarkAntColony_5Turbines(b *testing.B) {
	pts, err := loadCSVPoints("problem_5_turbines.csv")
	if err != nil {
		b.Fatalf("Failed to load points: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rng := rand.New(rand.NewPCG(uint64(i), 0))
		aco := algorithms.NewAntColony(pts, 3, 200, 5.0, 1.5, 0.5, 100.0, rng)
		aco.Optimize()
	}
}

func BenchmarkAntColony_20Turbines(b *testing.B) {
	pts, err := loadCSVPoints("problem_20_turbines.csv")
	if err != nil {
		b.Fatalf("Failed to load points: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rng := rand.New(rand.NewPCG(uint64(i), 0))
		aco := algorithms.NewAntColony(pts, 8, 200, 5.0, 2.0, 0.5, 100.0, rng)
		aco.Optimize()
	}
}

func BenchmarkGenetic_5Turbines(b *testing.B) {
	pts, err := loadCSVPoints("problem_5_turbines.csv")
	if err != nil {
		b.Fatalf("Failed to load points: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rng := rand.New(rand.NewPCG(uint64(i), 0))
		ga := algorithms.NewGeneticAlgorithm(pts, 50, 50, 0.2, false, rng)
		ga.Evolve()
	}
}

func BenchmarkGenetic_20Turbines(b *testing.B) {
	pts, err := loadCSVPoints("problem_20_turbines.csv")
	if err != nil {
		b.Fatalf("Failed to load points: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rng := rand.New(rand.NewPCG(uint64(i), 0))
		ga := algorithms.NewGeneticAlgorithm(pts, 100, 50, 0.1, false, rng)
		ga.Evolve()
	}
}

func BenchmarkMemetic_5Turbines(b *testing.B) {
	pts, err := loadCSVPoints("problem_5_turbines.csv")
	if err != nil {
		b.Fatalf("Failed to load points: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rng := rand.New(rand.NewPCG(uint64(i), 0))
		ga := algorithms.NewGeneticAlgorithm(pts, 50, 10, 0.2, true, rng)
		ga.Evolve()
	}
}

func BenchmarkMemetic_20Turbines(b *testing.B) {
	pts, err := loadCSVPoints("problem_20_turbines.csv")
	if err != nil {
		b.Fatalf("Failed to load points: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rng := rand.New(rand.NewPCG(uint64(i), 0))
		ga := algorithms.NewGeneticAlgorithm(pts, 50, 10, 0.2, true, rng)
		ga.Evolve()
	}
}

func TestPrintSpeedBenchmarkResults(t *testing.T) {
	problems := []string{
		"problem_5_turbines.csv",
		"problem_10_turbines.csv",
		"problem_15_turbines.csv",
		"problem_20_turbines.csv",
		"problem_40_turbines.csv",
	}

	fmt.Println("=================================================================")
	fmt.Println("GOLANG DIRECT ALGORITHM BENCHMARK EXECUTION TIMES")
	fmt.Println("=================================================================")
	fmt.Printf("%-25s | %-12s | %-12s | %-12s\n", "Problem Dataset", "ACO (ms)", "Genetic (ms)", "Memetic (ms)")
	fmt.Println("-----------------------------------------------------------------")

	for _, p := range problems {
		pts, err := loadCSVPoints(p)
		if err != nil {
			t.Fatalf("Error loading %s: %v", p, err)
		}
		nTurbines := len(pts) - 1

		// ACO
		var nAnts int
		var alpha, beta float64
		if nTurbines <= 10 {
			nAnts = 3
			alpha = 5.0
			beta = 1.5
		} else {
			nAnts = 8
			alpha = 5.0
			beta = 2.0
		}
		t0 := time.Now()
		for i := 0; i < 5; i++ {
			rng := rand.New(rand.NewPCG(uint64(i), 0))
			aco := algorithms.NewAntColony(pts, nAnts, 200, alpha, beta, 0.5, 100.0, rng)
			aco.Optimize()
		}
		acoMs := float64(time.Since(t0).Milliseconds()) / 5.0

		// Genetic
		var mutRate float64
		var popSize, nGen int
		if nTurbines >= 15 {
			mutRate = 0.1
			popSize = 100
			nGen = 50
		} else {
			mutRate = 0.2
			popSize = 50
			nGen = 50
		}
		t1 := time.Now()
		for i := 0; i < 5; i++ {
			rng := rand.New(rand.NewPCG(uint64(i), 0))
			ga := algorithms.NewGeneticAlgorithm(pts, popSize, nGen, mutRate, false, rng)
			ga.Evolve()
		}
		gaMs := float64(time.Since(t1).Milliseconds()) / 5.0

		// Memetic
		if nTurbines >= 40 {
			mutRate = 0.1
			popSize = 150
			nGen = 50
		} else {
			mutRate = 0.2
			popSize = 50
			nGen = 10
		}
		t2 := time.Now()
		for i := 0; i < 5; i++ {
			rng := rand.New(rand.NewPCG(uint64(i), 0))
			ga := algorithms.NewGeneticAlgorithm(pts, popSize, nGen, mutRate, true, rng)
			ga.Evolve()
		}
		memMs := float64(time.Since(t2).Milliseconds()) / 5.0

		fmt.Printf("%-25s | %-12.2f | %-12.2f | %-12.2f\n", p, acoMs, gaMs, memMs)
	}
	fmt.Println("=================================================================")
}
