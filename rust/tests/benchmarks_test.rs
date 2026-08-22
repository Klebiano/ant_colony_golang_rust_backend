use std::path::PathBuf;
use std::time::Instant;
use rand::rngs::StdRng;
use rand::SeedableRng;

use ant_colony_rust_backend::algorithms::{AntColony, GeneticAlgorithm};
use ant_colony_rust_backend::handlers::{load_points_from_csv, normalize_coordinates};

#[test]
fn test_print_speed_benchmark_results() {
    let manifest_dir = PathBuf::from(env!("CARGO_MANIFEST_DIR"));
    let base_dir = manifest_dir.parent().unwrap();

    let problems = [
        "problem_5_turbines.csv",
        "problem_10_turbines.csv",
        "problem_15_turbines.csv",
        "problem_20_turbines.csv",
        "problem_40_turbines.csv",
        "problem_100_turbines.csv",
        "problem_200_turbines.csv",
    ];

    println!("\n=================================================================");
    println!("RUST DIRECT ALGORITHM BENCHMARK EXECUTION TIMES");
    println!("=================================================================");
    println!("{:<25} | {:<12} | {:<12} | {:<12}", "Problem Dataset", "ACO (ms)", "Genetic (ms)", "Memetic (ms)");
    println!("-----------------------------------------------------------------");

    for p in problems {
        let csv_path = base_dir.join("tests").join("inputs").join(p);
        let mut pts = load_points_from_csv(&csv_path).expect("Failed to load CSV");
        normalize_coordinates(&mut pts);

        let n_turbines = pts.len() - 1;

        // ACO
        let (n_ants, alpha, beta, rho, n_iter) = if n_turbines <= 10 {
            (3, 5.0, 1.5, 0.5, 200)
        } else if n_turbines <= 40 {
            (8, 5.0, 2.0, 0.5, 200)
        } else {
            (10, 5.0, 2.0, 0.5, 100)
        };

        let t0 = Instant::now();
        for i in 0..5 {
            let mut rng = StdRng::seed_from_u64(i as u64);
            let mut aco = AntColony::new(pts.clone(), n_ants, n_iter, alpha, beta, rho, 100.0);
            aco.optimize(&mut rng);
        }
        let aco_ms = (t0.elapsed().as_secs_f64() * 1000.0) / 5.0;

        // Genetic
        let (mut_rate, pop_size, n_gen) = if n_turbines >= 15 {
            (0.1, 100, 50)
        } else {
            (0.2, 50, 50)
        };

        let t1 = Instant::now();
        for i in 0..5 {
            let mut rng = StdRng::seed_from_u64(i as u64);
            let mut ga = GeneticAlgorithm::new(pts.clone(), pop_size, n_gen, mut_rate, false, &mut rng);
            ga.evolve(&mut rng);
        }
        let ga_ms = (t1.elapsed().as_secs_f64() * 1000.0) / 5.0;

        // Memetic
        let (mut_rate_m, pop_size_m, n_gen_m) = if n_turbines >= 100 {
            (0.1, 60, 20)
        } else if n_turbines >= 40 {
            (0.1, 80, 30)
        } else {
            (0.2, 50, 10)
        };

        let t2 = Instant::now();
        for i in 0..5 {
            let mut rng = StdRng::seed_from_u64(i as u64);
            let mut mem = GeneticAlgorithm::new(pts.clone(), pop_size_m, n_gen_m, mut_rate_m, true, &mut rng);
            mem.evolve(&mut rng);
        }
        let mem_ms = (t2.elapsed().as_secs_f64() * 1000.0) / 5.0;

        println!("{:<25} | {:<12.2} | {:<12.2} | {:<12.2}", p, aco_ms, ga_ms, mem_ms);
    }
    println!("=================================================================\n");
}
