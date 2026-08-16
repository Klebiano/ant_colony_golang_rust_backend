import os
import sys
import time
import subprocess
import requests
import re
import pandas as pd
from pathlib import Path

# Roots
WORKSPACE_DIR = Path(__file__).resolve().parent.parent
GO_DIR = WORKSPACE_DIR / "go"
RUST_DIR = WORKSPACE_DIR / "rust"
PY_DIR = Path("/home/klebs/Documentos/ant_colony_fast_api_backend")
sys.path.insert(0, str(PY_DIR))

from scripts.ant_colony import AntColony
from scripts.genetic_algorithm import GeneticAlgorithm

def load_py_problem(csv_name):
    csv_path = WORKSPACE_DIR / "tests" / "inputs" / csv_name
    df = pd.read_csv(csv_path, index_col=0)
    df[['latitude_norm', 'longitude_norm']] = df[['latitude', 'longitude']].apply(lambda x: (x - x.min()) / (x.max() - x.min()))
    return df.reset_index(drop=True).to_dict('index')

def benchmark_python_direct(problems):
    results = []
    print("Running Python Algorithm Benchmarks...")
    
    for p in problems:
        points = load_py_problem(p)
        n_turbines = len(points.keys()) - 1
        
        # ACO
        if n_turbines <= 10:
            n_ants, alpha, beta = 3, 5.0, 1.5
        else:
            n_ants, alpha, beta = 8, 5.0, 2.0
            
        t0 = time.time()
        for _ in range(3):
            aco = AntColony(points, n_ants=n_ants, n_iterations=200, alpha=alpha, beta=beta, evaporation_rate=0.5, Q=100)
            aco.ant_colony_optimization()
        aco_ms = ((time.time() - t0) / 3.0) * 1000.0

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

        # Memetic
        if n_turbines >= 40:
            mut_rate, pop_size, n_gen = 0.1, 150, 50
        else:
            mut_rate, pop_size, n_gen = 0.2, 50, 10
            
        t2 = time.time()
        for _ in range(3):
            mem = GeneticAlgorithm(points, population_size=pop_size, n_generations=n_gen, mutation_rate=mut_rate, implement_local_search=True)
            mem.evolve()
        mem_ms = ((time.time() - t2) / 3.0) * 1000.0

        results.append({
            "problem": p,
            "py_aco_ms": round(aco_ms, 2),
            "py_ga_ms": round(ga_ms, 2),
            "py_mem_ms": round(mem_ms, 2)
        })
        
    return pd.DataFrame(results)

def benchmark_go_direct():
    print("Running Golang Algorithm Benchmarks...")
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
                    "go_mem_ms": mem
                }
    return results

def benchmark_rust_direct():
    print("Running Rust Algorithm Benchmarks...")
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
                    "rust_mem_ms": mem
                }
    return results

def start_servers():
    print("Starting FastAPI (port 8000), Go Backend (port 8080), and Rust Backend (port 8082)...")
    
    # 1. Python FastAPI
    py_env = os.environ.copy()
    py_env["PYTHONPATH"] = str(PY_DIR)
    py_proc = subprocess.Popen(
        [str(PY_DIR / ".venv" / "bin" / "python"), "-m", "uvicorn", "main:app", "--port", "8000"],
        cwd=str(PY_DIR),
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
        env=py_env
    )
    
    # 2. Go Backend
    go_proc = subprocess.Popen(
        ["go", "run", "main.go"],
        cwd=str(GO_DIR),
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
        env=dict(os.environ, PORT="8080")
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
        env=dict(os.environ, PORT="8082")
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
    {"turbine_id": 3, "subsystem_name": "Rotor Hub", "fault_type": "Major"}
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
            durations = []
            # Warm up
            try:
                if method == "GET":
                    requests.get(url, timeout=5)
                else:
                    requests.post(url, json=body, timeout=5)
            except Exception as e:
                print(f"Warmup error for {url}: {e}")
                
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
    problems = [
        "problem_5_turbines.csv",
        "problem_10_turbines.csv",
        "problem_15_turbines.csv",
        "problem_20_turbines.csv",
        "problem_40_turbines.csv",
    ]
    
    py_df = benchmark_python_direct(problems)
    go_res = benchmark_go_direct()
    rust_res = benchmark_rust_direct()
    
    combined = []
    for _, row in py_df.iterrows():
        p = row["problem"]
        p_name = p.replace("problem_", "").replace(".csv", "").replace("_", " ").title()
        go_data = go_res.get(p, {"go_aco_ms": 0.0, "go_ga_ms": 0.0, "go_mem_ms": 0.0})
        rust_data = rust_res.get(p, {"rust_aco_ms": 0.0, "rust_ga_ms": 0.0, "rust_mem_ms": 0.0})
        
        combined.append({
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
        })
        
    comb_df = pd.DataFrame(combined)
    print("\n=========================================================================================================")
    print("DIRECT ALGORITHM EXECUTION TIMES (PYTHON vs GOLANG vs RUST)")
    print("=========================================================================================================")
    print(comb_df.to_string(index=False))
    
    py_proc, go_proc, rust_proc = start_servers()
    try:
        http_df = benchmark_http()
        print("\n=========================================================================================================")
        print("HTTP API LATENCY BENCHMARK (FASTAPI vs GOLANG vs RUST)")
        print("=========================================================================================================")
        print(http_df.to_string(index=False))
    finally:
        py_proc.terminate()
        go_proc.terminate()
        rust_proc.terminate()
