import os
import sys
import time
import subprocess
import requests
import pandas as pd
from pathlib import Path

# Roots
GO_DIR = Path(__file__).resolve().parent.parent
PY_DIR = Path("/home/klebs/Documentos/ant_colony_fast_api_backend")
sys.path.insert(0, str(PY_DIR))

from scripts.ant_colony import AntColony
from scripts.genetic_algorithm import GeneticAlgorithm

def load_py_problem(csv_name):
    csv_path = GO_DIR / "tests" / "inputs" / csv_name
    df = pd.read_csv(csv_path, index_col=0)
    df[['latitude_norm', 'longitude_norm']] = df[['latitude', 'longitude']].apply(lambda x: (x - x.min()) / (x.max() - x.min()))
    return df.reset_index(drop=True).to_dict('index')

def benchmark_python_direct():
    problems = [
        "problem_5_turbines.csv",
        "problem_10_turbines.csv",
        "problem_15_turbines.csv",
        "problem_20_turbines.csv",
        "problem_40_turbines.csv",
    ]
    
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

def start_servers():
    print("Starting FastAPI (port 8000) and Go Backend (port 8080)...")
    py_env = os.environ.copy()
    py_env["PYTHONPATH"] = str(PY_DIR)
    py_proc = subprocess.Popen(
        [str(PY_DIR / ".venv" / "bin" / "python"), "-m", "uvicorn", "main:app", "--port", "8000"],
        cwd=str(PY_DIR),
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
        env=py_env
    )
    
    go_proc = subprocess.Popen(
        ["go", "run", "main.go"],
        cwd=str(GO_DIR),
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
        env=dict(os.environ, PORT="8080")
    )
    
    time.sleep(3) # Wait for startup
    return py_proc, go_proc

func_payload = [
    {"turbine_id": 2, "subsystem_name": "Electrical System", "fault_type": "Minor"},
    {"turbine_id": 3, "subsystem_name": "Rotor Hub", "fault_type": "Major"}
]

def benchmark_http():
    urls = {
        "Python FastAPI (8000)": "http://localhost:8000",
        "Golang Backend (8080)": "http://localhost:8080",
    }
    
    results = []
    
    endpoints = [
        ("GET Turbines Map", "GET", "/ant-colony/get-turbines-map", None),
        ("GET Subsystems", "GET", "/ant-colony/get-subsystems", None),
        ("POST Optimizer (ACO)", "POST", "/ant-colony/run-route-optimizer?algorithm=Ant%20Colony", func_payload),
        ("POST Optimizer (Genetic)", "POST", "/ant-colony/run-route-optimizer?algorithm=Genetic", func_payload),
        ("POST Optimizer (Memetic)", "POST", "/ant-colony/run-route-optimizer?algorithm=Memetic", func_payload),
    ]
    
    for ep_name, method, path, body in endpoints:
        row = {"endpoint": ep_name}
        for name, base in urls.items():
            url = base + path
            durations = []
            for _ in range(5):
                t0 = time.time()
                if method == "GET":
                    resp = requests.get(url)
                else:
                    resp = requests.post(url, json=body)
                durations.append((time.time() - t0) * 1000.0)
            avg_ms = sum(durations) / len(durations)
            row[name] = round(avg_ms, 2)
        results.append(row)
        
    return pd.DataFrame(results)

if __name__ == "__main__":
    py_df = benchmark_python_direct()
    print("\n--- Python Direct Execution Results ---")
    print(py_df.to_string(index=False))
    
    py_proc, go_proc = start_servers()
    try:
        http_df = benchmark_http()
        print("\n--- HTTP API Latency Benchmark (FastAPI vs Go) ---")
        print(http_df.to_string(index=False))
    finally:
        py_proc.terminate()
        go_proc.terminate()
