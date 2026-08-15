package algorithms

import (
	"math"
	"math/rand/v2"
	"testing"
	"ant_colony_golang_backend/models"
)

func samplePoints() []models.TurbineFaultPoint {
	return []models.TurbineFaultPoint{
		{TurbineID: 1, TurbineName: "Doca", LatitudeNorm: 0.0, LongitudeNorm: 0.0, FaultDowntimeDaysNorm: 0.0},
		{TurbineID: 2, TurbineName: "T1", LatitudeNorm: 0.2, LongitudeNorm: 0.5, FaultDowntimeDaysNorm: 0.1},
		{TurbineID: 3, TurbineName: "T2", LatitudeNorm: 0.8, LongitudeNorm: 0.3, FaultDowntimeDaysNorm: 0.2},
		{TurbineID: 4, TurbineName: "T3", LatitudeNorm: 0.5, LongitudeNorm: 0.9, FaultDowntimeDaysNorm: 0.05},
		{TurbineID: 5, TurbineName: "T4", LatitudeNorm: 0.9, LongitudeNorm: 0.1, FaultDowntimeDaysNorm: 0.3},
	}
}

func TestAntColony(t *testing.T) {
	pts := samplePoints()
	rng := rand.New(rand.NewPCG(123, 0))

	aco := NewAntColony(pts, 3, 50, 5.0, 1.5, 0.5, 100.0, rng)
	aco.Optimize()

	if len(aco.BestPath) != len(pts)+1 {
		t.Fatalf("Expected best path of length %d, got %d", len(pts)+1, len(aco.BestPath))
	}
	if aco.BestPathLength <= 0 {
		t.Fatalf("Expected positive best path length, got %f", aco.BestPathLength)
	}
	if len(aco.TurbineOrder) != len(pts)+1 {
		t.Fatalf("Expected turbine order of length %d, got %d", len(pts)+1, len(aco.TurbineOrder))
	}
}

func TestGeneticAlgorithm(t *testing.T) {
	pts := samplePoints()
	rng := rand.New(rand.NewPCG(123, 0))

	ga := NewGeneticAlgorithm(pts, 20, 30, 0.2, false, rng)
	ga.Evolve()

	if len(ga.BestPath) != len(pts)+1 {
		t.Fatalf("Expected best path of length %d, got %d", len(pts)+1, len(ga.BestPath))
	}
	if ga.BestPathLength <= 0 {
		t.Fatalf("Expected positive best path length, got %f", ga.BestPathLength)
	}
}

func TestMemeticAlgorithm(t *testing.T) {
	pts := samplePoints()
	rng := rand.New(rand.NewPCG(123, 0))

	ga := NewGeneticAlgorithm(pts, 20, 10, 0.2, true, rng)
	ga.Evolve()

	if len(ga.BestPath) != len(pts)+1 {
		t.Fatalf("Expected best path of length %d, got %d", len(pts)+1, len(ga.BestPath))
	}
	if ga.BestPathLength <= 0 {
		t.Fatalf("Expected positive best path length, got %f", ga.BestPathLength)
	}
}

func TestFormatTurbineOrderToShow(t *testing.T) {
	// Case 1: Closed loop starting/ending at T1 with Doca in middle
	order1 := []string{"T1", "T2", "Doca", "T3", "T1"}
	res1 := FormatTurbineOrderToShow(order1)
	expected1 := []string{"Doca", "T3", "T1", "T2", "Doca"}
	if len(res1) != len(expected1) {
		t.Fatalf("Expected len %d, got %d for case 1", len(expected1), len(res1))
	}
	for i := range res1 {
		if res1[i] != expected1[i] {
			t.Fatalf("Case 1 index %d: expected %s, got %s", i, expected1[i], res1[i])
		}
	}

	// Case 2: Closed loop already starting/ending at Doca
	order2 := []string{"Doca", "T1", "T2", "Doca"}
	res2 := FormatTurbineOrderToShow(order2)
	expected2 := []string{"Doca", "T1", "T2", "Doca"}
	if len(res2) != len(expected2) {
		t.Fatalf("Expected len %d, got %d for case 2", len(expected2), len(res2))
	}
	for i := range res2 {
		if res2[i] != expected2[i] {
			t.Fatalf("Case 2 index %d: expected %s, got %s", i, expected2[i], res2[i])
		}
	}

	// Case 3: Single turbine with Doca
	order3 := []string{"T1", "Doca", "T1"}
	res3 := FormatTurbineOrderToShow(order3)
	expected3 := []string{"Doca", "T1", "Doca"}
	if len(res3) != len(expected3) {
		t.Fatalf("Expected len %d, got %d for case 3", len(expected3), len(res3))
	}
	for i := range res3 {
		if res3[i] != expected3[i] {
			t.Fatalf("Case 3 index %d: expected %s, got %s", i, expected3[i], res3[i])
		}
	}

	// Case 4: Empty order
	res4 := FormatTurbineOrderToShow([]string{})
	expected4 := []string{"Doca", "Doca"}
	if len(res4) != len(expected4) || res4[0] != expected4[0] || res4[1] != expected4[1] {
		t.Fatalf("Expected %v, got %v for empty case", expected4, res4)
	}
}

func TestACOUnitSquareOptimalTour(t *testing.T) {
	// 4 turbines on unit square corners: (0,0), (0,1), (1,1), (1,0)
	squarePoints := []models.TurbineFaultPoint{
		{TurbineID: 0, TurbineName: "T0", LatitudeNorm: 0.0, LongitudeNorm: 0.0, FaultDowntimeDaysNorm: 0.0},
		{TurbineID: 1, TurbineName: "T1", LatitudeNorm: 0.0, LongitudeNorm: 1.0, FaultDowntimeDaysNorm: 0.0},
		{TurbineID: 2, TurbineName: "T2", LatitudeNorm: 1.0, LongitudeNorm: 1.0, FaultDowntimeDaysNorm: 0.0},
		{TurbineID: 3, TurbineName: "T3", LatitudeNorm: 1.0, LongitudeNorm: 0.0, FaultDowntimeDaysNorm: 0.0},
	}
	rng := rand.New(rand.NewPCG(42, 0))
	aco := NewAntColony(squarePoints, 10, 30, 2.0, 3.0, 0.5, 100.0, rng)
	aco.Optimize()

	if len(aco.BestPath) != 5 {
		t.Fatalf("Expected best path of length 5, got %d", len(aco.BestPath))
	}
	if aco.BestPath[0] != aco.BestPath[len(aco.BestPath)-1] {
		t.Fatalf("Expected closed loop where start == end, got %v", aco.BestPath)
	}
	// Optimal square perimeter tour length is 4.0
	if math.Abs(aco.BestPathLength-4.0) > 0.01 {
		t.Fatalf("Expected optimal square route length ~4.0, got %f", aco.BestPathLength)
	}
}

func TestAlgorithmsSinglePoint(t *testing.T) {
	singlePt := []models.TurbineFaultPoint{
		{TurbineID: 1, TurbineName: "Doca", LatitudeNorm: 0.0, LongitudeNorm: 0.0, FaultDowntimeDaysNorm: 0.0},
	}
	rng := rand.New(rand.NewPCG(42, 0))

	aco := NewAntColony(singlePt, 3, 10, 1.0, 2.0, 0.5, 100.0, rng)
	aco.Optimize()
	if len(aco.BestPath) != 2 || aco.BestPathLength != 0.0 {
		t.Fatalf("Unexpected single point ACO result: %+v", aco)
	}

	ga := NewGeneticAlgorithm(singlePt, 10, 5, 0.1, false, rng)
	ga.Evolve()
	if len(ga.BestPath) != 2 || ga.BestPathLength != 0.0 {
		t.Fatalf("Unexpected single point GA result: %+v", ga)
	}
}
