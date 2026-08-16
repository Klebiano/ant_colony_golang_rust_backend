use rand::Rng;
use crate::models::TurbineFaultPoint;
use super::common::{distance, downtime_cost, objective_function};

#[derive(Debug, Clone)]
pub struct AntColony {
    pub points: Vec<TurbineFaultPoint>,
    pub n_ants: usize,
    pub n_iterations: usize,
    pub n_points: usize,
    pub pheromone: Vec<Vec<f64>>,
    pub alpha: f64,
    pub beta: f64,
    pub evaporation_rate: f64,
    pub q: f64,
    pub best_path: Vec<usize>,
    pub turbine_order: Vec<String>,
    pub best_path_length: f64,
    pub best_downtime_days: f64,
    pub best_path_len_downtime: f64,
}

impl AntColony {
    pub fn new(
        points: Vec<TurbineFaultPoint>,
        n_ants: usize,
        n_iterations: usize,
        alpha: f64,
        beta: f64,
        evaporation_rate: f64,
        q: f64,
    ) -> Self {
        let n_points = points.len();
        let pheromone = vec![vec![1.0; n_points]; n_points];

        Self {
            points,
            n_ants,
            n_iterations,
            n_points,
            pheromone,
            alpha,
            beta,
            evaporation_rate,
            q,
            best_path: Vec::new(),
            turbine_order: Vec::new(),
            best_path_length: f64::INFINITY,
            best_downtime_days: f64::INFINITY,
            best_path_len_downtime: f64::INFINITY,
        }
    }

    pub fn update_pheromone(&mut self, paths: &[Vec<usize>], objectives: &[(f64, f64)]) {
        // Evaporation on all edges
        for i in 0..self.n_points {
            for j in 0..self.n_points {
                self.pheromone[i][j] *= 1.0 - self.evaporation_rate;
            }
        }

        // Deposit pheromone strictly on edges traversed by each ant
        for (k, path) in paths.iter().enumerate() {
            let total_cost = objectives[k].0 + objectives[k].1;
            let delta = if total_cost > 1e-9 {
                self.q / total_cost
            } else {
                self.q
            };

            for t in 0..path.len() - 1 {
                let u = path[t];
                let v = path[t + 1];
                self.pheromone[u][v] += delta;
                self.pheromone[v][u] += delta;
            }
        }

        // Avoid pheromone vanishing to exact 0
        for i in 0..self.n_points {
            for j in 0..self.n_points {
                if self.pheromone[i][j] < 1e-6 {
                    self.pheromone[i][j] = 1e-6;
                }
            }
        }
    }

    pub fn optimize<R: Rng>(&mut self, rng: &mut R) {
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

        let mut unvisited_buf = Vec::with_capacity(self.n_points);
        let mut probs_buf = Vec::with_capacity(self.n_points);

        for _iter in 0..self.n_iterations {
            let mut paths = Vec::with_capacity(self.n_ants);
            let mut objectives = Vec::with_capacity(self.n_ants);

            for _ant in 0..self.n_ants {
                let mut visited = vec![false; self.n_points];
                let current_point = rng.gen_range(0..self.n_points);
                visited[current_point] = true;
                let mut path = Vec::with_capacity(self.n_points + 1);
                path.push(current_point);

                let mut curr_node = current_point;
                let mut visited_count = 1;

                while visited_count < self.n_points {
                    unvisited_buf.clear();
                    for i in 0..self.n_points {
                        if !visited[i] {
                            unvisited_buf.push(i);
                        }
                    }

                    probs_buf.clear();
                    let mut sum_prob = 0.0;

                    let curr_lat = self.points[curr_node].latitude_norm;
                    let curr_lon = self.points[curr_node].longitude_norm;
                    let curr_dt = self.points[curr_node].fault_downtime_days_norm;

                    for &unvis in &unvisited_buf {
                        let p_alpha = self.pheromone[curr_node][unvis].powf(self.alpha);

                        let u_lat = self.points[unvis].latitude_norm;
                        let u_lon = self.points[unvis].longitude_norm;
                        let u_dt = self.points[unvis].fault_downtime_days_norm;

                        let dist = distance([curr_lat, curr_lon], [u_lat, u_lon]);
                        let dt_cost = downtime_cost(curr_dt, u_dt);
                        let combined_cost = dist + dt_cost;

                        let heuristic = (1.0 / (combined_cost + 1e-10)).powf(self.beta);
                        let prob = p_alpha * heuristic;
                        probs_buf.push(prob);
                        sum_prob += prob;
                    }

                    let next_point = if sum_prob > 0.0 && !sum_prob.is_nan() {
                        let r = rng.gen::<f64>() * sum_prob;
                        let mut cum_sum = 0.0;
                        let mut selected = unvisited_buf[unvisited_buf.len() - 1];
                        for (i, &prob) in probs_buf.iter().enumerate() {
                            cum_sum += prob;
                            if r <= cum_sum {
                                selected = unvisited_buf[i];
                                break;
                            }
                        }
                        selected
                    } else {
                        let idx = rng.gen_range(0..unvisited_buf.len());
                        unvisited_buf[idx]
                    };

                    path.push(next_point);
                    visited[next_point] = true;
                    curr_node = next_point;
                    visited_count += 1;
                }

                // Return to start
                path.push(path[0]);

                let (dist, dt_cost) = objective_function(&path, &self.points);
                let total_cost = dist + dt_cost;
                paths.push(path.clone());
                objectives.push((dist, dt_cost));

                if total_cost < self.best_path_len_downtime {
                    self.best_path = path.clone();
                    self.best_path_length = dist;
                    self.best_downtime_days = dt_cost;
                    self.best_path_len_downtime = total_cost;
                    self.turbine_order = path
                        .iter()
                        .map(|&idx| self.points[idx].turbine_name.clone())
                        .collect();
                }
            }

            self.update_pheromone(&paths, &objectives);
        }
    }
}
