use crate::models::TurbineFaultPoint;

#[inline]
pub fn distance(p1: [f64; 2], p2: [f64; 2]) -> f64 {
    let dx = p1[0] - p2[0];
    let dy = p1[1] - p2[1];
    (dx * dx + dy * dy).sqrt()
}

#[inline]
pub fn downtime_cost(d1: f64, d2: f64) -> f64 {
    d1 + d2
}

pub fn objective_function(path: &[usize], points: &[TurbineFaultPoint]) -> (f64, f64) {
    if path.len() <= 1 {
        return (0.0, 0.0);
    }

    let mut total_distance = 0.0;
    let mut total_downtime_cost = 0.0;

    for i in 0..path.len() - 1 {
        let idx1 = path[i];
        let idx2 = path[i + 1];

        let p1 = [points[idx1].latitude_norm, points[idx1].longitude_norm];
        let p2 = [points[idx2].latitude_norm, points[idx2].longitude_norm];
        total_distance += distance(p1, p2);

        let d1 = points[idx1].fault_downtime_days_norm;
        let d2 = points[idx2].fault_downtime_days_norm;
        total_downtime_cost += downtime_cost(d1, d2);
    }

    (total_distance, total_downtime_cost)
}

pub fn format_turbine_order_to_show(turbine_order: &[String]) -> Vec<String> {
    if turbine_order.is_empty() {
        return vec!["Doca".to_string(), "Doca".to_string()];
    }

    let unique_nodes: Vec<String> = if turbine_order.len() > 1 && turbine_order.first() == turbine_order.last() {
        turbine_order[..turbine_order.len() - 1].to_vec()
    } else {
        turbine_order.to_vec()
    };

    let doca_idx = unique_nodes.iter().position(|name| name == "Doca");

    if let Some(idx) = doca_idx {
        let mut ordered = Vec::with_capacity(unique_nodes.len() + 1);
        ordered.extend_from_slice(&unique_nodes[idx..]);
        ordered.extend_from_slice(&unique_nodes[..idx]);
        ordered.push("Doca".to_string());
        return ordered;
    }

    if !unique_nodes.is_empty() {
        let mut res = Vec::with_capacity(unique_nodes.len() + 2);
        res.push("Doca".to_string());
        res.extend(unique_nodes);
        res.push("Doca".to_string());
        return res;
    }

    vec!["Doca".to_string(), "Doca".to_string()]
}
