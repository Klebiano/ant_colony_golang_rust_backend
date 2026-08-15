package algorithms

import (
	"math"
	"math/rand/v2"
	"ant_colony_golang_backend/models"
)

type AntColony struct {
	Points              []models.TurbineFaultPoint
	NAnts               int
	NIterations         int
	NPoints             int
	Pheromone           [][]float64
	Alpha               float64
	Beta                float64
	EvaporationRate     float64
	Q                   float64
	BestPath            []int
	TurbineOrder        []string
	BestPathLength      float64
	BestDowntimeDays    float64
	BestPathLenDowntime float64
	Rnd                 *rand.Rand
}

func NewAntColony(points []models.TurbineFaultPoint, nAnts int, nIterations int, alpha, beta, evaporationRate, q float64, rng *rand.Rand) *AntColony {
	nPoints := len(points)
	pheromone := make([][]float64, nPoints)
	for i := 0; i < nPoints; i++ {
		pheromone[i] = make([]float64, nPoints)
		for j := 0; j < nPoints; j++ {
			pheromone[i][j] = 1.0
		}
	}
	if rng == nil {
		rng = rand.New(rand.NewPCG(42, 0))
	}
	return &AntColony{
		Points:              points,
		NAnts:               nAnts,
		NIterations:         nIterations,
		NPoints:             nPoints,
		Pheromone:           pheromone,
		Alpha:               alpha,
		Beta:                beta,
		EvaporationRate:     evaporationRate,
		Q:                   q,
		BestPathLength:      math.Inf(1),
		BestDowntimeDays:    math.Inf(1),
		BestPathLenDowntime: math.Inf(1),
		Rnd:                 rng,
	}
}

func (ac *AntColony) UpdatePheromone(paths [][]int, objectives [][2]float64) {
	// Evaporation on all edges
	for i := 0; i < ac.NPoints; i++ {
		for j := 0; j < ac.NPoints; j++ {
			ac.Pheromone[i][j] = (1.0 - ac.EvaporationRate) * ac.Pheromone[i][j]
		}
	}

	// Deposit pheromone strictly on edges traversed by each ant
	for k := 0; k < len(paths); k++ {
		totalCost := objectives[k][0] + objectives[k][1]
		delta := ac.Q
		if totalCost > 1e-9 {
			delta = ac.Q / totalCost
		}

		path := paths[k]
		for t := 0; t < len(path)-1; t++ {
			u := path[t]
			v := path[t+1]
			ac.Pheromone[u][v] += delta
			ac.Pheromone[v][u] += delta
		}
	}

	// Avoid pheromone vanishing to exact 0
	for i := 0; i < ac.NPoints; i++ {
		for j := 0; j < ac.NPoints; j++ {
			if ac.Pheromone[i][j] < 1e-6 {
				ac.Pheromone[i][j] = 1e-6
			}
		}
	}
}

func (ac *AntColony) Optimize() {
	if ac.NPoints == 0 {
		ac.BestPath = []int{}
		ac.TurbineOrder = []string{}
		ac.BestPathLength = 0.0
		ac.BestDowntimeDays = 0.0
		ac.BestPathLenDowntime = 0.0
		return
	}

	if ac.NPoints == 1 {
		ac.BestPath = []int{0, 0}
		name := ac.Points[0].TurbineName
		if name == "" {
			name = "Point_0"
		}
		ac.TurbineOrder = []string{name, name}
		dist, dt := ObjectiveFunction(ac.BestPath, ac.Points)
		ac.BestPathLength = dist
		ac.BestDowntimeDays = dt
		ac.BestPathLenDowntime = dist + dt
		return
	}

	for iter := 0; iter < ac.NIterations; iter++ {
		paths := make([][]int, 0, ac.NAnts)
		objectives := make([][2]float64, 0, ac.NAnts)

		for ant := 0; ant < ac.NAnts; ant++ {
			visited := make([]bool, ac.NPoints)
			currentPoint := int(ac.Rnd.UintN(uint(ac.NPoints)))
			visited[currentPoint] = true
			path := []int{currentPoint}

			visitedCount := 1
			for visitedCount < ac.NPoints {
				unvisited := make([]int, 0, ac.NPoints-visitedCount)
				for i := 0; i < ac.NPoints; i++ {
					if !visited[i] {
						unvisited = append(unvisited, i)
					}
				}

				probs := make([]float64, len(unvisited))
				var sumProb float64

				currLatNorm := ac.Points[currentPoint].LatitudeNorm
				currLonNorm := ac.Points[currentPoint].LongitudeNorm
				currDtNorm := ac.Points[currentPoint].FaultDowntimeDaysNorm

				for i, unvis := range unvisited {
					pAlpha := math.Pow(ac.Pheromone[currentPoint][unvis], ac.Alpha)

					uLatNorm := ac.Points[unvis].LatitudeNorm
					uLonNorm := ac.Points[unvis].LongitudeNorm
					uDtNorm := ac.Points[unvis].FaultDowntimeDaysNorm

					dist := Distance([2]float64{currLatNorm, currLonNorm}, [2]float64{uLatNorm, uLonNorm})
					dtCost := DowntimeCost(currDtNorm, uDtNorm)
					combinedCost := dist + dtCost

					heuristic := math.Pow(1.0/(combinedCost+1e-10), ac.Beta)
					prob := pAlpha * heuristic
					probs[i] = prob
					sumProb += prob
				}

				var nextPoint int
				if sumProb > 0 && !math.IsNaN(sumProb) {
					r := ac.Rnd.Float64() * sumProb
					var cumSum float64
					nextPoint = unvisited[len(unvisited)-1]
					for i, prob := range probs {
						cumSum += prob
						if r <= cumSum {
							nextPoint = unvisited[i]
							break
						}
					}
				} else {
					nextPoint = unvisited[int(ac.Rnd.UintN(uint(len(unvisited))))]
				}

				path = append(path, nextPoint)
				visited[nextPoint] = true
				currentPoint = nextPoint
				visitedCount++
			}

			// Return to start
			path = append(path, path[0])

			dist, dtCost := ObjectiveFunction(path, ac.Points)
			paths = append(paths, path)
			objectives = append(objectives, [2]float64{dist, dtCost})

			totalCost := dist + dtCost
			if totalCost < ac.BestPathLenDowntime {
				ac.BestPath = path
				ac.BestPathLength = dist
				ac.BestDowntimeDays = dtCost
				ac.BestPathLenDowntime = totalCost

				order := make([]string, len(path))
				for i, idx := range path {
					order[i] = ac.Points[idx].TurbineName
				}
				ac.TurbineOrder = order
			}
		}

		ac.UpdatePheromone(paths, objectives)
	}
}
