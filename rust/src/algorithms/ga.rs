use rand::seq::SliceRandom;
use rand::Rng;
use crate::models::TurbineFaultPoint;
use super::common::objective_function;

#[derive(Debug, Clone)]
pub struct GeneticAlgorithm {
    pub points: Vec<TurbineFaultPoint>,
    pub population_size: usize,
    pub n_generations: usize,
    pub n_points: usize,
    pub mutation_rate: f64,
    pub implement_local_search: bool,
    pub population: Vec<Vec<usize>>,
    pub turbine_order: Vec<String>,
    pub best_path: Vec<usize>,
    pub best_path_length: f64,
    pub best_downtime_days: f64,
    pub best_path_len_downtime: f64,
}

impl GeneticAlgorithm {
    pub fn new<R: Rng>(
        points: Vec<TurbineFaultPoint>,
        population_size: usize,
        n_generations: usize,
        mutation_rate: f64,
        implement_local_search: bool,
        rng: &mut R,
    ) -> Self {
        let pop_size = population_size.max(4);
        let n_points = points.len();

        let mut ga = Self {
            points,
            population_size: pop_size,
            n_generations,
            n_points,
            mutation_rate,
            implement_local_search,
            population: Vec::new(),
            turbine_order: Vec::new(),
            best_path: Vec::new(),
            best_path_length: f64::INFINITY,
            best_downtime_days: f64::INFINITY,
            best_path_len_downtime: f64::INFINITY,
        };

        ga.population = ga.initialize_population(rng);
        ga
    }

    pub fn initialize_population<R: Rng>(&self, rng: &mut R) -> Vec<Vec<usize>> {
        if self.n_points == 0 {
            return Vec::new();
        }
        if self.n_points == 1 {
            return vec![vec![0]; self.population_size];
        }

        let mut pop = Vec::with_capacity(self.population_size);
        for _ in 0..self.population_size {
            let mut ind: Vec<usize> = (0..self.n_points).collect();
            ind.shuffle(rng);
            pop.push(ind);
        }
        pop
    }

    pub fn route_cost(&self, route: &[usize]) -> (f64, f64) {
        if route.len() <= 1 {
            return (0.0, 0.0);
        }
        if route.len() == self.n_points && self.n_points > 1 && route[0] != route[route.len() - 1] {
            let mut full_path = Vec::with_capacity(route.len() + 1);
            full_path.extend_from_slice(route);
            full_path.push(route[0]);
            objective_function(&full_path, &self.points)
        } else {
            objective_function(route, &self.points)
        }
    }

    pub fn calculate_total_distance(&self, route: &[usize]) -> f64 {
        let (dist, _) = self.route_cost(route);
        dist
    }

    pub fn selection(&self) -> Vec<Vec<usize>> {
        let mut evaluated: Vec<(usize, f64)> = self
            .population
            .iter()
            .enumerate()
            .map(|(idx, ind)| {
                let (d, c) = self.route_cost(ind);
                (idx, d + c)
            })
            .collect();

        evaluated.sort_by(|a, b| a.1.partial_cmp(&b.1).unwrap_or(std::cmp::Ordering::Equal));

        let num_selected = (self.population_size / 2).max(2).min(evaluated.len());
        let mut selected = Vec::with_capacity(num_selected);
        for &(idx, _) in &evaluated[..num_selected] {
            selected.push(self.population[idx].clone());
        }
        selected
    }

    pub fn crossover<R: Rng>(&self, parent1: &[usize], parent2: &[usize], rng: &mut R) -> Vec<usize> {
        if self.n_points <= 2 {
            return parent1.to_vec();
        }

        let r1 = rng.gen_range(0..self.n_points);
        let mut r2 = rng.gen_range(0..self.n_points);
        while r1 == r2 {
            r2 = rng.gen_range(0..self.n_points);
        }

        let (start, end) = if r1 < r2 { (r1, r2) } else { (r2, r1) };

        let mut child = vec![usize::MAX; self.n_points];
        let mut in_child = vec![false; self.n_points];

        for i in start..end {
            let gene = parent1[i];
            child[i] = gene;
            in_child[gene] = true;
        }

        let mut pointer = end % self.n_points;
        for i in 0..self.n_points {
            let gene = parent2[(end + i) % self.n_points];
            if !in_child[gene] {
                child[pointer] = gene;
                in_child[gene] = true;
                pointer = (pointer + 1) % self.n_points;
            }
        }

        child
    }

    pub fn mutate<R: Rng>(&self, individual: &mut [usize], rng: &mut R) {
        if self.n_points >= 2 && rng.gen::<f64>() < self.mutation_rate {
            let i = rng.gen_range(0..self.n_points);
            let mut j = rng.gen_range(0..self.n_points);
            while i == j {
                j = rng.gen_range(0..self.n_points);
            }
            individual.swap(i, j);
        }
    }

    pub fn two_opt(&self, route: &[usize], max_iterations: usize) -> Vec<usize> {
        if self.n_points < 4 {
            return route.to_vec();
        }

        let mut best = route.to_vec();
        let (d, c) = self.route_cost(&best);
        let mut best_cost = d + c;
        let mut count = 0;

        while count < max_iterations {
            let mut improved = false;
            for i in 0..self.n_points - 1 {
                for j in (i + 1)..self.n_points {
                    if i == 0 && j == self.n_points - 1 {
                        continue;
                    }

                    let mut new_route = Vec::with_capacity(self.n_points);
                    new_route.extend_from_slice(&best[..i]);
                    let mut reversed_segment = best[i..=j].to_vec();
                    reversed_segment.reverse();
                    new_route.extend(reversed_segment);
                    new_route.extend_from_slice(&best[j + 1..]);

                    let (nd, nc) = self.route_cost(&new_route);
                    let new_cost = nd + nc;
                    if new_cost < best_cost - 1e-9 {
                        best = new_route;
                        best_cost = new_cost;
                        improved = true;
                        break;
                    }
                }
                if improved {
                    break;
                }
            }
            if !improved {
                break;
            }
            count += 1;
        }

        best
    }

    pub fn evolve<R: Rng>(&mut self, rng: &mut R) {
        if self.n_points == 0 {
            self.best_path = Vec::new();
            self.turbine_order = Vec::new();
            self.best_path_length = 0.0;
            self.best_downtime_days = 0.0;
            self.best_path_len_downtime = 0.0;
            return;
        }

        if self.n_points == 1 {
            self.best_path = vec![0, 0];
            let name = if self.points[0].turbine_name.is_empty() {
                "Point_0".to_string()
            } else {
                self.points[0].turbine_name.clone()
            };
            self.turbine_order = vec![name.clone(), name];
            let (dist, dt) = objective_function(&self.best_path, &self.points);
            self.best_path_length = dist;
            self.best_downtime_days = dt;
            self.best_path_len_downtime = dist + dt;
            return;
        }

        // Initial evaluation
        for individual in &self.population {
            let mut closed_ind = Vec::with_capacity(individual.len() + 1);
            closed_ind.extend_from_slice(individual);
            closed_ind.push(individual[0]);

            let (dist, dt_cost) = objective_function(&closed_ind, &self.points);
            let total_cost = dist + dt_cost;

            if total_cost < self.best_path_len_downtime {
                self.best_path = closed_ind.clone();
                self.best_path_length = dist;
                self.best_downtime_days = dt_cost;
                self.best_path_len_downtime = total_cost;
                self.turbine_order = closed_ind
                    .iter()
                    .map(|&idx| self.points[idx].turbine_name.clone())
                    .collect();
            }
        }

        for _gen in 0..self.n_generations {
            let selected = self.selection();
            let mut new_pop = Vec::with_capacity(self.population_size);

            // Elitism: preserve top 2 individuals
            for item in selected.iter().take(2) {
                new_pop.push(item.clone());
            }

            while new_pop.len() < self.population_size {
                let p1_idx = rng.gen_range(0..selected.len());
                let mut p2_idx = rng.gen_range(0..selected.len());
                while selected.len() > 1 && p1_idx == p2_idx {
                    p2_idx = rng.gen_range(0..selected.len());
                }

                let mut child = self.crossover(&selected[p1_idx], &selected[p2_idx], rng);
                self.mutate(&mut child, rng);

                if self.implement_local_search && rng.gen::<f64>() < 0.2 {
                    child = self.two_opt(&child, 10);
                }
                new_pop.push(child);
            }

            self.population = new_pop;

            for individual in &self.population {
                let mut closed_ind = Vec::with_capacity(individual.len() + 1);
                closed_ind.extend_from_slice(individual);
                closed_ind.push(individual[0]);

                let (dist, dt_cost) = objective_function(&closed_ind, &self.points);
                let total_cost = dist + dt_cost;

                if total_cost < self.best_path_len_downtime {
                    self.best_path = closed_ind.clone();
                    self.best_path_length = dist;
                    self.best_downtime_days = dt_cost;
                    self.best_path_len_downtime = total_cost;
                    self.turbine_order = closed_ind
                        .iter()
                        .map(|&idx| self.points[idx].turbine_name.clone())
                        .collect();
                }
            }
        }
    }
}
