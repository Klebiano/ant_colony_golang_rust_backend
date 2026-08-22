import os
import sys
import time
import subprocess
import requests
import json
import pandas as pd
import numpy as np
import pulp
from pathlib import Path

# Roots
WORKSPACE_DIR = Path(__file__).resolve().parent.parent
GO_DIR = WORKSPACE_DIR / "go"
RUST_DIR = WORKSPACE_DIR / "rust"
PY_DIR = Path("/home/klebs/Documentos/ant_colony_fast_api_backend")
sys.path.insert(0, str(PY_DIR))
sys.path.insert(0, str(PY_DIR / "scripts"))

from scripts.ant_colony import AntColony
from scripts.genetic_algorithm import GeneticAlgorithm

TIMEOUT_PER_PROBLEM_SEC = 300  # 5 minutes limit per algorithm per problem

PROBLEMS = [
    "problem_5_turbines.csv",
    "problem_10_turbines.csv",
    "problem_15_turbines.csv",
    "problem_20_turbines.csv",
    "problem_40_turbines.csv",
    "problem_100_turbines.csv",
    "problem_200_turbines.csv",
]


def min_max_scale(series: pd.Series) -> pd.Series:
    s_min = series.min()
    s_max = series.max()
    denom = s_max - s_min
    if denom == 0 or pd.isna(denom):
        return pd.Series(0.0, index=series.index)
    return (series - s_min) / denom


def load_problem_data(file_path: Path):
    df = pd.read_csv(file_path, index_col=0).reset_index(drop=True)
    df["latitude_norm"] = min_max_scale(df["latitude"])
    df["longitude_norm"] = min_max_scale(df["longitude"])
    if "fault_downtime_days_norm" not in df.columns:
        df["fault_downtime_days_norm"] = min_max_scale(df.get("fault_downtime_days", pd.Series(0.0, index=df.index)))

    pts = df[["latitude_norm", "longitude_norm"]].values
    dts = df["fault_downtime_days_norm"].values
    names = df["turbine_name"].tolist() if "turbine_name" in df.columns else [f"Point_{i}" for i in range(len(df))]
    data_dict = df.to_dict("index")
    return df, pts, dts, names, data_dict


def solve_cbc(points: np.ndarray, dts: np.ndarray, names: list, max_seconds: int = TIMEOUT_PER_PROBLEM_SEC):
    """
    Solves the exact TSP with Downtime Objective using Coin-OR CBC via PuLP
    with iterative subtour elimination constraints (Dantzig-Fulkerson-Johnson / DFJ).
    """
    n = len(points)
    dist_matrix = np.zeros((n, n), dtype=float)
    for i in range(n):
        for j in range(n):
            if i != j:
                dist = float(np.linalg.norm(points[i] - points[j]))
                dt = float(dts[i] + dts[j])
                dist_matrix[i, j] = dist + dt

    prob = pulp.LpProblem("TSP_Downtime_CBC", pulp.LpMinimize)
    x = pulp.LpVariable.dicts("x", ((i, j) for i in range(n) for j in range(n) if i != j), cat=pulp.LpBinary)

    # Objective: Minimize total edge cost (scaled distance + downtime cost)
    prob += pulp.lpSum(dist_matrix[i, j] * x[i, j] for i in range(n) for j in range(n) if i != j)

    # In-degree and Out-degree constraints (each node visited exactly once)
    for i in range(n):
        prob += pulp.lpSum(x[i, j] for j in range(n) if j != i) == 1
    for j in range(n):
        prob += pulp.lpSum(x[i, j] for i in range(n) if i != j) == 1

    solver = pulp.PULP_CBC_CMD(msg=False, timeLimit=max_seconds, threads=4)

    start_time = time.time()
    iteration = 0
    while True:
        iteration += 1
        elapsed = time.time() - start_time
        remaining_time = max_seconds - elapsed
        if remaining_time <= 0:
            break

        solver.timeLimit = max(1, int(remaining_time))
        prob.solve(solver)

        # Build successor map
        succ = {}
        for i in range(n):
            for j in range(n):
                if i != j and pulp.value(x[i, j]) is not None and pulp.value(x[i, j]) > 0.5:
                    succ[i] = j

        # Identify all disjoint cycles
        visited = [False] * n
        cycles = []
        for i in range(n):
            if not visited[i]:
                cycle = []
                curr = i
                while not visited[curr]:
                    visited[curr] = True
                    cycle.append(curr)
                    curr = succ.get(curr, curr)
                if cycle and succ.get(cycle[-1]) == cycle[0]:
                    cycles.append(cycle)

        # If exactly one cycle of length N is found, it is the globally optimal Hamiltonian tour!
        if len(cycles) == 1 and len(cycles[0]) == n:
            tour = cycles[0] + [cycles[0][0]]
            solve_time = time.time() - start_time
            obj_val = pulp.value(prob.objective)
            total_dist = sum(np.linalg.norm(points[tour[k]] - points[tour[k + 1]]) for k in range(len(tour) - 1))
            total_dt = sum(dts[tour[k]] + dts[tour[k + 1]] for k in range(len(tour) - 1))

            # Rotate tour so Doca (node 0) is at start and end
            doca_idx = tour.index(0) if 0 in tour else 0
            rotated_tour = tour[doca_idx:-1] + tour[:doca_idx] + [0]
            turbine_order = [names[idx] for idx in rotated_tour]

            return {
                "status": "Optimal",
                "cost": float(obj_val),
                "distance": float(total_dist),
                "downtime": float(total_dt),
                "time_sec": float(solve_time),
                "time_ms": float(solve_time * 1000.0),
                "iterations": iteration,
                "tour": rotated_tour,
                "turbine_order": turbine_order,
            }

        # Add subtour elimination cuts
        for cycle in cycles:
            if len(cycle) < n:
                prob += pulp.lpSum(x[i, j] for i in cycle for j in cycle if i != j) <= len(cycle) - 1

    # Fallback if timeout reached
    solve_time = time.time() - start_time
    return {
        "status": "TimeLimit",
        "cost": float(pulp.value(prob.objective) or np.nan),
        "distance": np.nan,
        "downtime": np.nan,
        "time_sec": float(solve_time),
        "time_ms": float(solve_time * 1000.0),
        "iterations": iteration,
        "tour": [],
        "turbine_order": [],
    }


def benchmark_python_direct(problems):
    results = []
    print("\nRunning Python Algorithm Benchmarks...")

    for p in problems:
        csv_path = WORKSPACE_DIR / "tests" / "inputs" / p
        df, pts, dts, names, points = load_problem_data(csv_path)
        n_turbines = len(points.keys()) - 1

        # ACO
        if n_turbines <= 10:
            n_ants, alpha, beta, n_iter = 3, 5.0, 1.5, 200
        elif n_turbines <= 40:
            n_ants, alpha, beta, n_iter = 8, 5.0, 2.0, 200
        else:
            n_ants, alpha, beta, n_iter = 10, 5.0, 2.0, 100

        t0 = time.time()
        for _ in range(3):
            aco = AntColony(points, n_ants=n_ants, n_iterations=n_iter, alpha=alpha, beta=beta, evaporation_rate=0.5, Q=100)
            aco.ant_colony_optimization()
        aco_ms = ((time.time() - t0) / 3.0) * 1000.0
        aco_cost = aco.best_path_len_downtime

        # Genetic
        if n_turbines >= 15:
            mut_rate, pop_size, n_gen = 0.1, 100, 50
        else:
            mut_rate, pop_size, n_gen = 0.2, 50, 50

        t1 = time.time()
        for _ in range(3):
            ga = GeneticAlgorithm(points, population_size=pop_size, n_generations=n_gen, mutation_rate=mut_rate, implement_local_search=False)
            ga.evolve()
        ga_ms = ((time.time() - t1) / 3.0) * 1000.0
        ga_cost = ga.best_path_len_downtime

        # Memetic
        if n_turbines >= 100:
            mut_rate, pop_size, n_gen = 0.1, 60, 20
        elif n_turbines >= 40:
            mut_rate, pop_size, n_gen = 0.1, 80, 30
        else:
            mut_rate, pop_size, n_gen = 0.2, 50, 10

        t2 = time.time()
        for _ in range(3):
            mem = GeneticAlgorithm(points, population_size=pop_size, n_generations=n_gen, mutation_rate=mut_rate, implement_local_search=True)
            mem.evolve()
        mem_ms = ((time.time() - t2) / 3.0) * 1000.0
        mem_cost = mem.best_path_len_downtime

        results.append({
            "problem": p,
            "py_aco_ms": round(aco_ms, 2),
            "py_ga_ms": round(ga_ms, 2),
            "py_mem_ms": round(mem_ms, 2),
            "py_aco_cost": aco_cost,
            "py_ga_cost": ga_cost,
            "py_mem_cost": mem_cost,
        })
        print(f"  -> Python {p} done: ACO={aco_ms:.2f}ms, GA={ga_ms:.2f}ms, Mem={mem_ms:.2f}ms")

    return pd.DataFrame(results)


def benchmark_cbc_direct(problems, max_seconds=TIMEOUT_PER_PROBLEM_SEC):
    results = {}
    print(f"\nRunning Coin-OR CBC Exact Solver Benchmarks (Timeout = {max_seconds}s / 5 min)...")

    for p in problems:
        csv_path = WORKSPACE_DIR / "tests" / "inputs" / p
        df, pts, dts, names, points = load_problem_data(csv_path)
        print(f"  -> Solving {p} with CBC (MILP + DFJ cuts)...")
        cbc_res = solve_cbc(pts, dts, names, max_seconds=max_seconds)
        cost_str = f"{cbc_res['cost']:.4f}" if not np.isnan(cbc_res['cost']) else "N/A"
        print(f"     Status: {cbc_res['status']} | Cost: {cost_str} | Time: {cbc_res['time_ms']:.2f} ms ({cbc_res['time_sec']:.4f}s)")
        results[p] = cbc_res
    return results


def benchmark_go_direct():
    print("\nRunning Golang Algorithm Benchmarks...")
    cmd = ["go", "test", "-v", "-run", "TestPrintSpeedBenchmarkResults", "./tests"]
    proc = subprocess.run(cmd, cwd=str(GO_DIR), capture_output=True, text=True)

    results = {}
    for line in proc.stdout.splitlines():
        if "problem_" in line and "|" in line:
            parts = [p.strip() for p in line.split("|")]
            if len(parts) == 4:
                prob = parts[0]
                aco = float(parts[1])
                ga = float(parts[2])
                mem = float(parts[3])
                results[prob] = {
                    "go_aco_ms": aco,
                    "go_ga_ms": ga,
                    "go_mem_ms": mem,
                }
                print(f"  -> Golang {prob}: ACO={aco:.2f}ms, GA={ga:.2f}ms, Mem={mem:.2f}ms")
    return results


def benchmark_rust_direct():
    print("\nRunning Rust Algorithm Benchmarks...")
    cmd = ["cargo", "test", "--release", "--manifest-path", "rust/Cargo.toml", "--test", "benchmarks_test", "--", "--nocapture"]
    proc = subprocess.run(cmd, cwd=str(WORKSPACE_DIR), capture_output=True, text=True)

    results = {}
    for line in proc.stdout.splitlines():
        if "problem_" in line and "|" in line:
            parts = [p.strip() for p in line.split("|")]
            if len(parts) == 4:
                prob = parts[0]
                aco = float(parts[1])
                ga = float(parts[2])
                mem = float(parts[3])
                results[prob] = {
                    "rust_aco_ms": aco,
                    "rust_ga_ms": ga,
                    "rust_mem_ms": mem,
                }
                print(f"  -> Rust {prob}: ACO={aco:.2f}ms, GA={ga:.2f}ms, Mem={mem:.2f}ms")
    return results


def start_servers():
    print("\nStarting FastAPI (port 8000), Go Backend (port 8080), and Rust Backend (port 8082)...")

    # 1. Python FastAPI
    py_env = os.environ.copy()
    py_env["PYTHONPATH"] = str(PY_DIR)
    py_proc = subprocess.Popen(
        [str(PY_DIR / ".venv" / "bin" / "python"), "-m", "uvicorn", "main:app", "--port", "8000"],
        cwd=str(PY_DIR),
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
        env=py_env,
    )

    # 2. Go Backend
    go_proc = subprocess.Popen(
        ["go", "run", "main.go"],
        cwd=str(GO_DIR),
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
        env=dict(os.environ, PORT="8080"),
    )

    # 3. Rust Backend
    rust_bin = RUST_DIR / "target" / "release" / "ant_colony_rust_backend"
    if not rust_bin.exists():
        subprocess.run(["cargo", "build", "--release", "--manifest-path", "rust/Cargo.toml"], cwd=str(WORKSPACE_DIR))

    rust_proc = subprocess.Popen(
        [str(rust_bin)],
        cwd=str(WORKSPACE_DIR),
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
        env=dict(os.environ, PORT="8082"),
    )

    # Wait for all servers to become responsive
    servers = [
        ("FastAPI", "http://localhost:8000/ant-colony/get-subsystems"),
        ("Go", "http://localhost:8080/ant-colony/get-subsystems"),
        ("Rust", "http://localhost:8082/ant-colony/get-subsystems"),
    ]

    for name, url in servers:
        ready = False
        for _ in range(30):
            try:
                r = requests.get(url, timeout=1.0)
                if r.status_code == 200:
                    ready = True
                    break
            except Exception:
                pass
            time.sleep(0.3)
        if not ready:
            print(f"Warning: {name} server did not respond on {url}")

    return py_proc, go_proc, rust_proc


func_payload = [
    {"turbine_id": 2, "subsystem_name": "Electrical System", "fault_type": "Minor"},
    {"turbine_id": 3, "subsystem_name": "Rotor Hub", "fault_type": "Major"},
]


def benchmark_http():
    urls = {
        "FastAPI (Python)": "http://localhost:8000",
        "Golang Backend": "http://localhost:8080",
        "Rust Backend": "http://localhost:8082",
    }

    results = []

    endpoints = [
        ("GET /ant-colony/get-turbines-map", "GET", "/ant-colony/get-turbines-map", None),
        ("GET /ant-colony/get-subsystems", "GET", "/ant-colony/get-subsystems", None),
        ("POST /run-route-optimizer (ACO)", "POST", "/ant-colony/run-route-optimizer?algorithm=Ant%20Colony", func_payload),
        ("POST /run-route-optimizer (Genetic)", "POST", "/ant-colony/run-route-optimizer?algorithm=Genetic", func_payload),
        ("POST /run-route-optimizer (Memetic)", "POST", "/ant-colony/run-route-optimizer?algorithm=Memetic", func_payload),
    ]

    for ep_name, method, path, body in endpoints:
        row = {"Endpoint": ep_name}
        for name, base in urls.items():
            url = base + path
            # Warm up
            try:
                if method == "GET":
                    requests.get(url, timeout=5)
                else:
                    requests.post(url, json=body, timeout=5)
            except Exception as e:
                print(f"Warmup error for {url}: {e}")

            durations = []
            for _ in range(10):
                t0 = time.perf_counter()
                if method == "GET":
                    resp = requests.get(url, timeout=10)
                else:
                    resp = requests.post(url, json=body, timeout=10)
                durations.append((time.perf_counter() - t0) * 1000.0)
            avg_ms = sum(durations) / len(durations)
            row[name] = round(avg_ms, 2)
        results.append(row)

    return pd.DataFrame(results)


if __name__ == "__main__":
    print("=" * 105)
    print(f"COMPREHENSIVE BACKEND & ALGORITHM BENCHMARK (PYTHON vs GO vs RUST vs COIN-OR CBC)")
    print(f"CBC Timeout: {TIMEOUT_PER_PROBLEM_SEC}s (5 minutes)")
    print("=" * 105)

    # 1. Coin-OR CBC Exact Solver
    cbc_res = benchmark_cbc_direct(PROBLEMS, max_seconds=TIMEOUT_PER_PROBLEM_SEC)

    # 2. Python Direct Benchmarks
    py_df = benchmark_python_direct(PROBLEMS)

    # 3. Golang Direct Benchmarks
    go_res = benchmark_go_direct()

    # 4. Rust Direct Benchmarks
    rust_res = benchmark_rust_direct()

    # Combine Execution Times Table
    combined_time = []
    combined_optimality = []
    full_records = []

    for _, row in py_df.iterrows():
        p = row["problem"]
        p_name = p.replace("problem_", "").replace(".csv", "").replace("_", " ").title()
        go_data = go_res.get(p, {"go_aco_ms": 0.0, "go_ga_ms": 0.0, "go_mem_ms": 0.0})
        rust_data = rust_res.get(p, {"rust_aco_ms": 0.0, "rust_ga_ms": 0.0, "rust_mem_ms": 0.0})
        cbc_data = cbc_res.get(p, {"status": "N/A", "cost": np.nan, "time_ms": np.nan, "time_sec": np.nan, "distance": np.nan, "downtime": np.nan})

        cbc_cost = cbc_data["cost"]
        aco_cost = row["py_aco_cost"]
        ga_cost = row["py_ga_cost"]
        mem_cost = row["py_mem_cost"]

        aco_gap = ((aco_cost - cbc_cost) / cbc_cost * 100.0) if not np.isnan(cbc_cost) and cbc_cost > 0 else 0.0
        ga_gap = ((ga_cost - cbc_cost) / cbc_cost * 100.0) if not np.isnan(cbc_cost) and cbc_cost > 0 else 0.0
        mem_gap = ((mem_cost - cbc_cost) / cbc_cost * 100.0) if not np.isnan(cbc_cost) and cbc_cost > 0 else 0.0

        combined_time.append({
            "Dataset": p_name,
            "Py ACO (ms)": row["py_aco_ms"],
            "Go ACO (ms)": go_data["go_aco_ms"],
            "Rust ACO (ms)": rust_data["rust_aco_ms"],
            "Py GA (ms)": row["py_ga_ms"],
            "Go GA (ms)": go_data["go_ga_ms"],
            "Rust GA (ms)": rust_data["rust_ga_ms"],
            "Py Mem (ms)": row["py_mem_ms"],
            "Go Mem (ms)": go_data["go_mem_ms"],
            "Rust Mem (ms)": rust_data["rust_mem_ms"],
            "CBC Time (ms)": round(cbc_data["time_ms"], 2) if not np.isnan(cbc_data["time_ms"]) else "Timeout",
        })

        combined_optimality.append({
            "Dataset": p_name,
            "CBC Opt Cost": round(cbc_cost, 4) if not np.isnan(cbc_cost) else "TimeLimit",
            "CBC Status": cbc_data["status"],
            "ACO Best Cost": round(aco_cost, 4),
            "ACO Gap (%)": f"{aco_gap:+.2f}%",
            "GA Best Cost": round(ga_cost, 4),
            "GA Gap (%)": f"{ga_gap:+.2f}%",
            "Memetic Best Cost": round(mem_cost, 4),
            "Memetic Gap (%)": f"{mem_gap:+.2f}%",
        })

        full_records.append({
            "problem": p,
            "dataset": p_name,
            "cbc": cbc_data,
            "python": {
                "aco_ms": row["py_aco_ms"],
                "ga_ms": row["py_ga_ms"],
                "mem_ms": row["py_mem_ms"],
                "aco_cost": aco_cost,
                "ga_cost": ga_cost,
                "mem_cost": mem_cost,
                "aco_gap_pct": round(aco_gap, 2),
                "ga_gap_pct": round(ga_gap, 2),
                "mem_gap_pct": round(mem_gap, 2),
            },
            "golang": go_data,
            "rust": rust_data,
        })

    comb_time_df = pd.DataFrame(combined_time)
    comb_opt_df = pd.DataFrame(combined_optimality)

    print("\n" + "=" * 105)
    print("1. DIRECT ALGORITHM EXECUTION TIMES (PYTHON vs GOLANG vs RUST vs CBC)")
    print("=" * 105)
    print(comb_time_df.to_string(index=False))

    print("\n" + "=" * 105)
    print("2. SOLUTION QUALITY & OPTIMALITY GAP vs COIN-OR CBC EXACT SOLVER")
    print("=" * 105)
    print(comb_opt_df.to_string(index=False))

    # Save output artifacts
    output_dir = WORKSPACE_DIR / "tests" / "output"
    output_dir.mkdir(parents=True, exist_ok=True)

    csv_path = output_dir / "comparison_results.csv"
    comb_time_df.to_csv(csv_path, index=False)
    print(f"\n[+] Saved execution times CSV to {csv_path}")

    opt_csv_path = output_dir / "cbc_comparison_results.csv"
    comb_opt_df.to_csv(opt_csv_path, index=False)
    print(f"[+] Saved optimality comparison CSV to {opt_csv_path}")

    json_path = output_dir / "comparison_summary.json"
    with open(json_path, "w") as f:
        json.dump(full_records, f, indent=2)
    print(f"[+] Saved summary JSON to {json_path}")

    # 5. HTTP Latency Benchmark
    py_proc, go_proc, rust_proc = start_servers()
    try:
        http_df = benchmark_http()
        print("\n" + "=" * 105)
        print("3. HTTP API LATENCY BENCHMARK (FASTAPI vs GOLANG vs RUST)")
        print("=" * 105)
        print(http_df.to_string(index=False))

        http_csv_path = output_dir / "http_latency_results.csv"
        http_df.to_csv(http_csv_path, index=False)
        print(f"\n[+] Saved HTTP latency CSV to {http_csv_path}")
    finally:
        py_proc.terminate()
        go_proc.terminate()
        rust_proc.terminate()
