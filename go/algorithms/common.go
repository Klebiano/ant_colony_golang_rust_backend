package algorithms

import (
	"math"
	"ant_colony_golang_backend/models"
)

func Distance(p1, p2 [2]float64) float64 {
	dx := p1[0] - p2[0]
	dy := p1[1] - p2[1]
	return math.Sqrt(dx*dx + dy*dy)
}

func DowntimeCost(d1, d2 float64) float64 {
	return d1 + d2
}

func ObjectiveFunction(path []int, points []models.TurbineFaultPoint) (float64, float64) {
	var totalDistance float64
	var totalDowntimeCost float64

	for i := 0; i < len(path)-1; i++ {
		p1 := [2]float64{points[path[i]].LatitudeNorm, points[path[i]].LongitudeNorm}
		p2 := [2]float64{points[path[i+1]].LatitudeNorm, points[path[i+1]].LongitudeNorm}
		totalDistance += Distance(p1, p2)

		d1 := points[path[i]].FaultDowntimeDaysNorm
		d2 := points[path[i+1]].FaultDowntimeDaysNorm
		totalDowntimeCost += DowntimeCost(d1, d2)
	}

	return totalDistance, totalDowntimeCost
}

func FormatTurbineOrderToShow(turbineOrder []string) []string {
	if len(turbineOrder) == 0 {
		return []string{"Doca", "Doca"}
	}

	// Extract unique nodes if the order is already a closed loop (starts and ends with same element)
	var uniqueNodes []string
	if len(turbineOrder) > 1 && turbineOrder[0] == turbineOrder[len(turbineOrder)-1] {
		uniqueNodes = make([]string, len(turbineOrder)-1)
		copy(uniqueNodes, turbineOrder[:len(turbineOrder)-1])
	} else {
		uniqueNodes = make([]string, len(turbineOrder))
		copy(uniqueNodes, turbineOrder)
	}

	docaIdx := -1
	for i, name := range uniqueNodes {
		if name == "Doca" {
			docaIdx = i
			break
		}
	}

	if docaIdx != -1 {
		ordered := append([]string{}, uniqueNodes[docaIdx:]...)
		ordered = append(ordered, uniqueNodes[:docaIdx]...)
		return append(ordered, "Doca")
	}

	if len(uniqueNodes) > 0 {
		res := make([]string, 0, len(uniqueNodes)+2)
		res = append(res, "Doca")
		res = append(res, uniqueNodes...)
		res = append(res, "Doca")
		return res
	}

	return []string{"Doca", "Doca"}
}
