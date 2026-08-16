use rand::rngs::StdRng;
use rand::SeedableRng;
use ant_colony_rust_backend::algorithms::{
    format_turbine_order_to_show, AntColony, GeneticAlgorithm,
};
use ant_colony_rust_backend::models::TurbineFaultPoint;

fn sample_points() -> Vec<TurbineFaultPoint> {
    vec![
        TurbineFaultPoint {
            turbine_id: 1,
            turbine_name: "Doca".to_string(),
            latitude_norm: 0.0,
            longitude_norm: 0.0,
            fault_downtime_days_norm: 0.0,
            ..Default::default()
        },
        TurbineFaultPoint {
            turbine_id: 2,
            turbine_name: "T1".to_string(),
            latitude_norm: 0.2,
            longitude_norm: 0.5,
            fault_downtime_days_norm: 0.1,
            ..Default::default()
        },
        TurbineFaultPoint {
            turbine_id: 3,
            turbine_name: "T2".to_string(),
            latitude_norm: 0.8,
            longitude_norm: 0.3,
            fault_downtime_days_norm: 0.2,
            ..Default::default()
        },
        TurbineFaultPoint {
            turbine_id: 4,
            turbine_name: "T3".to_string(),
            latitude_norm: 0.5,
            longitude_norm: 0.9,
            fault_downtime_days_norm: 0.05,
            ..Default::default()
        },
        TurbineFaultPoint {
            turbine_id: 5,
            turbine_name: "T4".to_string(),
            latitude_norm: 0.9,
            longitude_norm: 0.1,
            fault_downtime_days_norm: 0.3,
            ..Default::default()
        },
    ]
}

#[test]
fn test_ant_colony() {
    let pts = sample_points();
    let mut rng = StdRng::seed_from_u64(123);

    let mut aco = AntColony::new(pts.clone(), 3, 50, 5.0, 1.5, 0.5, 100.0);
    aco.optimize(&mut rng);

    assert_eq!(aco.best_path.len(), pts.len() + 1);
    assert!(aco.best_path_length > 0.0);
    assert_eq!(aco.turbine_order.len(), pts.len() + 1);
}

#[test]
fn test_genetic_algorithm() {
    let pts = sample_points();
    let mut rng = StdRng::seed_from_u64(123);

    let mut ga = GeneticAlgorithm::new(pts.clone(), 20, 30, 0.2, false, &mut rng);
    ga.evolve(&mut rng);

    assert_eq!(ga.best_path.len(), pts.len() + 1);
    assert!(ga.best_path_length > 0.0);
}

#[test]
fn test_memetic_algorithm() {
    let pts = sample_points();
    let mut rng = StdRng::seed_from_u64(123);

    let mut ga = GeneticAlgorithm::new(pts.clone(), 20, 10, 0.2, true, &mut rng);
    ga.evolve(&mut rng);

    assert_eq!(ga.best_path.len(), pts.len() + 1);
    assert!(ga.best_path_length > 0.0);
}

#[test]
fn test_format_turbine_order_to_show() {
    // Case 1: Closed loop starting/ending at T1 with Doca in middle
    let order1 = vec![
        "T1".to_string(),
        "T2".to_string(),
        "Doca".to_string(),
        "T3".to_string(),
        "T1".to_string(),
    ];
    let res1 = format_turbine_order_to_show(&order1);
    let expected1 = vec!["Doca", "T3", "T1", "T2", "Doca"];
    assert_eq!(res1, expected1);

    // Case 2: Closed loop already starting/ending at Doca
    let order2 = vec![
        "Doca".to_string(),
        "T1".to_string(),
        "T2".to_string(),
        "Doca".to_string(),
    ];
    let res2 = format_turbine_order_to_show(&order2);
    let expected2 = vec!["Doca", "T1", "T2", "Doca"];
    assert_eq!(res2, expected2);

    // Case 3: Single turbine with Doca
    let order3 = vec!["T1".to_string(), "Doca".to_string(), "T1".to_string()];
    let res3 = format_turbine_order_to_show(&order3);
    let expected3 = vec!["Doca", "T1", "Doca"];
    assert_eq!(res3, expected3);

    // Case 4: Empty order
    let res4 = format_turbine_order_to_show(&[]);
    let expected4 = vec!["Doca", "Doca"];
    assert_eq!(res4, expected4);
}

#[test]
fn test_aco_unit_square_optimal_tour() {
    let square_points = vec![
        TurbineFaultPoint {
            turbine_id: 0,
            turbine_name: "T0".to_string(),
            latitude_norm: 0.0,
            longitude_norm: 0.0,
            fault_downtime_days_norm: 0.0,
            ..Default::default()
        },
        TurbineFaultPoint {
            turbine_id: 1,
            turbine_name: "T1".to_string(),
            latitude_norm: 0.0,
            longitude_norm: 1.0,
            fault_downtime_days_norm: 0.0,
            ..Default::default()
        },
        TurbineFaultPoint {
            turbine_id: 2,
            turbine_name: "T2".to_string(),
            latitude_norm: 1.0,
            longitude_norm: 1.0,
            fault_downtime_days_norm: 0.0,
            ..Default::default()
        },
        TurbineFaultPoint {
            turbine_id: 3,
            turbine_name: "T3".to_string(),
            latitude_norm: 1.0,
            longitude_norm: 0.0,
            fault_downtime_days_norm: 0.0,
            ..Default::default()
        },
    ];

    let mut rng = StdRng::seed_from_u64(42);
    let mut aco = AntColony::new(square_points, 10, 30, 2.0, 3.0, 0.5, 100.0);
    aco.optimize(&mut rng);

    assert_eq!(aco.best_path.len(), 5);
    assert_eq!(aco.best_path.first(), aco.best_path.last());
    assert!((aco.best_path_length - 4.0).abs() < 0.01);
}

#[test]
fn test_algorithms_single_point() {
    let single_pt = vec![TurbineFaultPoint {
        turbine_id: 1,
        turbine_name: "Doca".to_string(),
        latitude_norm: 0.0,
        longitude_norm: 0.0,
        fault_downtime_days_norm: 0.0,
        ..Default::default()
    }];

    let mut rng = StdRng::seed_from_u64(42);
    let mut aco = AntColony::new(single_pt.clone(), 3, 10, 1.0, 2.0, 0.5, 100.0);
    aco.optimize(&mut rng);
    assert_eq!(aco.best_path.len(), 2);
    assert_eq!(aco.best_path_length, 0.0);

    let mut ga = GeneticAlgorithm::new(single_pt, 10, 5, 0.1, false, &mut rng);
    ga.evolve(&mut rng);
    assert_eq!(ga.best_path.len(), 2);
    assert_eq!(ga.best_path_length, 0.0);
}
