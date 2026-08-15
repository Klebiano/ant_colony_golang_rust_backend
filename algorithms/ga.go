package algorithms

import (
	"math"
	"math/rand/v2"
	"sort"
	"ant_colony_golang_backend/models"
)

type GeneticAlgorithm struct {
	Points               []models.TurbineFaultPoint
	PopulationSize       int
	NGenerations         int
	NPoints              int
	MutationRate         float64
	ImplementLocalSearch bool
	Population           [][]int
	TurbineOrder         []string
	BestPath             []int
	BestPathLength       float64
	BestDowntimeDays     float64
	BestPathLenDowntime  float64
	Rnd                  *rand.Rand
}

func NewGeneticAlgorithm(points []models.TurbineFaultPoint, populationSize, nGenerations int, mutationRate float64, implementLocalSearch bool, rng *rand.Rand) *GeneticAlgorithm {
	if rng == nil {
		rng = rand.New(rand.NewPCG(42, 0))
	}
	if populationSize < 4 {
		populationSize = 4
	}
	ga := &GeneticAlgorithm{
		Points:               points,
		PopulationSize:       populationSize,
		NGenerations:         nGenerations,
		NPoints:              len(points),
		MutationRate:         mutationRate,
		ImplementLocalSearch: implementLocalSearch,
		BestPathLength:       math.Inf(1),
		BestDowntimeDays:     math.Inf(1),
		BestPathLenDowntime:  math.Inf(1),
		Rnd:                  rng,
	}
	ga.Population = ga.InitializePopulation()
	return ga
}

func (ga *GeneticAlgorithm) InitializePopulation() [][]int {
	if ga.NPoints == 0 {
		return [][]int{}
	}
	if ga.NPoints == 1 {
		pop := make([][]int, ga.PopulationSize)
		for i := 0; i < ga.PopulationSize; i++ {
			pop[i] = []int{0}
		}
		return pop
	}

	pop := make([][]int, ga.PopulationSize)
	for i := 0; i < ga.PopulationSize; i++ {
		ind := make([]int, ga.NPoints)
		for j := 0; j < ga.NPoints; j++ {
			ind[j] = j
		}
		ga.Rnd.Shuffle(ga.NPoints, func(a, b int) {
			ind[a], ind[b] = ind[b], ind[a]
		})
		pop[i] = ind
	}
	return pop
}

func (ga *GeneticAlgorithm) RouteCost(route []int) (float64, float64) {
	if len(route) <= 1 {
		return 0.0, 0.0
	}
	var fullPath []int
	if len(route) == ga.NPoints && ga.NPoints > 1 && route[0] != route[len(route)-1] {
		fullPath = make([]int, len(route)+1)
		copy(fullPath, route)
		fullPath[len(route)] = route[0]
	} else {
		fullPath = route
	}
	return ObjectiveFunction(fullPath, ga.Points)
}

func (ga *GeneticAlgorithm) CalculateTotalDistance(route []int) float64 {
	dist, _ := ga.RouteCost(route)
	return dist
}

type indEval struct {
	ind  []int
	cost float64
}

func (ga *GeneticAlgorithm) Selection() [][]int {
	evaluated := make([]indEval, len(ga.Population))
	for i, ind := range ga.Population {
		d, c := ga.RouteCost(ind)
		evaluated[i] = indEval{ind: ind, cost: d + c}
	}

	sort.Slice(evaluated, func(i, j int) bool {
		return evaluated[i].cost < evaluated[j].cost
	})

	numSelected := max(2, ga.PopulationSize/2)
	if numSelected > len(evaluated) {
		numSelected = len(evaluated)
	}
	selected := make([][]int, numSelected)
	for i := 0; i < numSelected; i++ {
		selected[i] = evaluated[i].ind
	}
	return selected
}

func (ga *GeneticAlgorithm) Crossover(parent1, parent2 []int) []int {
	if ga.NPoints <= 2 {
		res := make([]int, len(parent1))
		copy(res, parent1)
		return res
	}

	r1 := int(ga.Rnd.UintN(uint(ga.NPoints)))
	r2 := int(ga.Rnd.UintN(uint(ga.NPoints)))
	for r1 == r2 {
		r2 = int(ga.Rnd.UintN(uint(ga.NPoints)))
	}
	start, end := r1, r2
	if start > end {
		start, end = end, start
	}

	child := make([]int, ga.NPoints)
	inChild := make(map[int]bool)
	for i := range child {
		child[i] = -1
	}

	for i := start; i < end; i++ {
		child[i] = parent1[i]
		inChild[parent1[i]] = true
	}

	pointer := end % ga.NPoints
	for i := 0; i < ga.NPoints; i++ {
		gene := parent2[(end+i)%ga.NPoints]
		if !inChild[gene] {
			child[pointer] = gene
			inChild[gene] = true
			pointer = (pointer + 1) % ga.NPoints
		}
	}

	return child
}

func (ga *GeneticAlgorithm) Mutate(individual []int) []int {
	if ga.NPoints >= 2 && ga.Rnd.Float64() < ga.MutationRate {
		i := int(ga.Rnd.UintN(uint(ga.NPoints)))
		j := int(ga.Rnd.UintN(uint(ga.NPoints)))
		for i == j {
			j = int(ga.Rnd.UintN(uint(ga.NPoints)))
		}
		individual[i], individual[j] = individual[j], individual[i]
	}
	return individual
}

func (ga *GeneticAlgorithm) TwoOpt(route []int, maxIterations int) []int {
	if ga.NPoints < 4 {
		return route
	}

	best := make([]int, len(route))
	copy(best, route)
	d, c := ga.RouteCost(best)
	bestCost := d + c
	count := 0

	for count < maxIterations {
		improved := false
		for i := 0; i < ga.NPoints-1; i++ {
			for j := i + 1; j < ga.NPoints; j++ {
				if i == 0 && j == ga.NPoints-1 {
					continue
				}
				newRoute := make([]int, len(route))
				copy(newRoute[:i], best[:i])
				for k := 0; k <= j-i; k++ {
					newRoute[i+k] = best[j-k]
				}
				copy(newRoute[j+1:], best[j+1:])

				nd, nc := ga.RouteCost(newRoute)
				newCost := nd + nc
				if newCost < bestCost-1e-9 {
					best = newRoute
					bestCost = newCost
					improved = true
					break
				}
			}
			if improved {
				break
			}
		}
		if !improved {
			break
		}
		count++
	}
	return best
}

func (ga *GeneticAlgorithm) Evolve() {
	if ga.NPoints == 0 {
		ga.BestPath = []int{}
		ga.TurbineOrder = []string{}
		ga.BestPathLength = 0.0
		ga.BestDowntimeDays = 0.0
		ga.BestPathLenDowntime = 0.0
		return
	}

	if ga.NPoints == 1 {
		ga.BestPath = []int{0, 0}
		name := ga.Points[0].TurbineName
		if name == "" {
			name = "Point_0"
		}
		ga.TurbineOrder = []string{name, name}
		dist, dt := ObjectiveFunction(ga.BestPath, ga.Points)
		ga.BestPathLength = dist
		ga.BestDowntimeDays = dt
		ga.BestPathLenDowntime = dist + dt
		return
	}

	// Initial evaluation
	for _, individual := range ga.Population {
		closedInd := make([]int, len(individual)+1)
		copy(closedInd, individual)
		closedInd[len(individual)] = individual[0]

		dist, dtCost := ObjectiveFunction(closedInd, ga.Points)
		totalCost := dist + dtCost

		if totalCost < ga.BestPathLenDowntime {
			ga.BestPath = closedInd
			ga.BestPathLength = dist
			ga.BestDowntimeDays = dtCost
			ga.BestPathLenDowntime = totalCost

			order := make([]string, len(closedInd))
			for i, idx := range closedInd {
				order[i] = ga.Points[idx].TurbineName
			}
			ga.TurbineOrder = order
		}
	}

	for gen := 0; gen < ga.NGenerations; gen++ {
		selected := ga.Selection()
		newPop := make([][]int, 0, ga.PopulationSize)

		// Elitism: preserve top 2 individuals
		for i := 0; i < 2 && i < len(selected); i++ {
			preserved := make([]int, len(selected[i]))
			copy(preserved, selected[i])
			newPop = append(newPop, preserved)
		}

		for len(newPop) < ga.PopulationSize {
			p1Idx := int(ga.Rnd.UintN(uint(len(selected))))
			p2Idx := int(ga.Rnd.UintN(uint(len(selected))))
			for len(selected) > 1 && p1Idx == p2Idx {
				p2Idx = int(ga.Rnd.UintN(uint(len(selected))))
			}

			child := ga.Crossover(selected[p1Idx], selected[p2Idx])
			child = ga.Mutate(child)

			if ga.ImplementLocalSearch && ga.Rnd.Float64() < 0.2 {
				child = ga.TwoOpt(child, 10)
			}
			newPop = append(newPop, child)
		}
		ga.Population = newPop

		for _, individual := range ga.Population {
			closedInd := make([]int, len(individual)+1)
			copy(closedInd, individual)
			closedInd[len(individual)] = individual[0]

			dist, dtCost := ObjectiveFunction(closedInd, ga.Points)
			totalCost := dist + dtCost

			if totalCost < ga.BestPathLenDowntime {
				ga.BestPath = closedInd
				ga.BestPathLength = dist
				ga.BestDowntimeDays = dtCost
				ga.BestPathLenDowntime = totalCost

				order := make([]string, len(closedInd))
				for i, idx := range closedInd {
					order[i] = ga.Points[idx].TurbineName
				}
				ga.TurbineOrder = order
			}
		}
	}
}
