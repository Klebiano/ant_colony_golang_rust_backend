# Project Context & Architecture

This document serves as the primary technical reference and architectural blueprint for the **Ant Colony & Metaheuristics Golang & Rust Backend** workspace.

---

## 1. Project Overview

* **Purpose**: Ultra-high-performance Go and Rust microservices designed for optimizing offshore wind farm turbine maintenance routing. Replicated from the original Python FastAPI backend, they compute optimal navigation routes using bio-inspired metaheuristic algorithms (Ant Colony Optimization, Genetic Algorithms, and Memetic Algorithms), balancing travel distance against turbine fault downtime costs.
* **Target Audience**: Operational planners, wind farm maintenance engineers, and researchers needing sub-millisecond route optimization and scheduling.
* **Context**: Developed as part of an offshore wind farm maintenance research project (TCC) to evaluate language execution speeds (Python vs Go vs Rust) and provide a production-ready API for web dashboards (e.g., React / Vite / Next.js).

---

## 2. Tech Stack

### Languages & Frameworks
* **Go (`/go`)**:
  * Web Server & Routing: Standard Library `net/http` with custom CORS middleware
  * Optimization: Pure Go vector math (`math`, `math/rand/v2`, `sort`)
  * Database: SQLite (`database/sql`, [`github.com/mattn/go-sqlite3`](https://github.com/mattn/go-sqlite3))
* **Rust (`/rust`)**:
  * Web Server & Routing: `axum 0.8`, `tokio 1.x`, `tower-http 0.6` (CORS middleware)
  * Optimization: Pure Rust vector calculations (`rand 0.8`, `std::cmp`)
  * Database: SQLite (`rusqlite 0.33` with bundled sqlite3)
  * Serialization: `serde`, `serde_json`, `csv`
* **Python 3.12+**: FastAPI, Uvicorn, NumPy, Pandas, SQLAlchemy (reference implementation)

---

## 3. Architecture & Directory Structure

```
ant_colony_golang_rust_backend/
├── database/                   # Shared SQLite datasets & SQL seeding script
│   ├── database.sql            # Seed ANSI SQL DDL & DML script
│   ├── downtime.csv            # Subsystem fault downtime dataset
│   ├── subsystem.csv           # Subsystem definitions dataset
│   └── wind_farm_map_points.csv# Geographical turbine map dataset
├── go/                         # Golang microservice backend
│   ├── algorithms/             # Go metaheuristic solvers (ACO, GA, Memetic)
│   │   ├── aco.go              # Go Ant Colony Optimization (ACO)
│   │   ├── ga.go               # Go Genetic & Memetic (GA + 2-opt)
│   │   ├── common.go           # Go distance metrics, cost functions, rotation
│   │   └── algorithms_test.go  # Go unit tests for solvers
│   ├── database/               # Go SQLite connection & query functions
│   │   └── database.go         # Go database queries (wind_farm_map_points, etc.)
│   ├── handlers/               # Go HTTP Request Handlers package
│   │   ├── ant_colony.go       # Go route handlers (/ant-colony/*) & CSV parser
│   │   └── handlers_test.go    # Go HTTP integration tests (httptest)
│   ├── models/                 # Go Data structures & JSON DTO definitions
│   │   └── models.go           # Struct definitions (AntColonyMap, TurbineFaults, etc.)
│   ├── tests/                  # Go benchmark suites
│   │   └── benchmarks_test.go  # Go benchmark test suite (testing.B)
│   ├── go.mod                  # Go module definition
│   ├── go.sum                  # Go dependency checksums
│   └── main.go                 # Go server entrypoint (port 8080)
├── rust/                       # Rust microservice backend
│   ├── Cargo.toml              # Rust crate manifest & dependencies
│   ├── src/
│   │   ├── lib.rs              # App router & exports
│   │   ├── main.rs             # Axum HTTP server entrypoint
│   │   ├── models.rs           # Rust data models & DTOs
│   │   ├── database.rs         # Rust SQLite initialization & queries
│   │   ├── handlers.rs         # Rust Axum handlers & coordinate normalization
│   │   └── algorithms/
│   │       ├── mod.rs          # Algorithms module root
│   │       ├── common.rs       # Shared Euclidean distance & objective evaluation
│   │       ├── aco.rs          # Rust Ant Colony Optimization
│   │       └── ga.rs           # Rust Genetic & Memetic (GA + 2-opt)
│   └── tests/
│       ├── algorithms_test.rs  # Rust unit tests
│       ├── handlers_test.rs    # Rust Axum HTTP endpoint integration tests
│       └── benchmarks_test.rs  # Rust direct algorithm benchmark suite
├── tests/                      # Problem datasets & benchmark suites
│   ├── inputs/                 # Problem datasets (5 to 100 turbines CSVs)
│   │   ├── problem_5_turbines.csv
│   │   ├── problem_10_turbines.csv
│   │   ├── problem_15_turbines.csv
│   │   ├── problem_20_turbines.csv
│   │   ├── problem_40_turbines.csv
│   │   └── problem_100_turbines.csv
│   └── compare_backends.py     # 3-way comparative benchmark runner (Python vs Go vs Rust)
├── sql_app.db                  # Shared SQLite database file
├── README.md                   # Documentation & benchmark summary
└── CONTEXT.md                  # Comprehensive technical reference (this file)
```

---

## 4. Component & Data Models Breakdown

### Data Models (`go/models/models.go` and `rust/src/models.rs`)
* **`AntColonyMap`**: Struct representing wind farm turbine geographical entries (`TurbineID`, `TurbineName`, `Latitude`, `Longitude`).
* **`Subsystems`**: Struct defining wind turbine subsystem metadata (`SubsystemID`, `SubsystemName`).
* **`Downtime`**: Struct capturing failure downtime profiles (`SubsystemID`, `SubsystemName`, `FaultType`, `AnualFailureRate`, `FaultDowntimeDays`).
* **`TurbineFaults`**: Incoming API request payload DTO representing reported faults on turbines (`TurbineID`, `TurbineName`, `SubsystemName`, `FaultType`).
* **`AntColonyPath`**: API response DTO containing optimization results (`TurbineOrder`, `TurbineOrderToShow`, `BestPath`, `BestPathLength`, `BestDowntimeDays`, `BestPathLenDowntime`, `TimeToRunSec`).
* **`TurbineFaultPoint`**: Internal normalized operational entity carrying both raw and Min-Max normalized coordinates (`LatitudeNorm`, `LongitudeNorm`, `FaultDowntimeDaysNorm`).

### Core Algorithms (`go/algorithms/` and `rust/src/algorithms/`)
* **ACO (`aco.go` / `aco.rs`)**: Multi-colony Ant Colony Optimization with probabilistic roulette-wheel selection and pheromone matrix evaporation/deposition.
* **GA & Memetic (`ga.go` / `ga.rs`)**: Genetic Algorithm with Order Crossover (OX), swap mutation, elitist selection, and optional 2-Opt local search.
* **Common (`common.go` / `common.rs`)**: Shared Euclidean distance, downtime cost calculation, objective evaluation, and `FormatTurbineOrderToShow` cycle rotation logic.

---

## 5. Mathematical Formulations & Heuristics

### Min-Max Feature Scaling
$$x_{\text{norm}} = \frac{x - x_{\text{min}}}{x_{\text{max}} - x_{\text{min}}}$$

### Objective Cost Function
$$f(P) = \sum_{i=0}^{k-1} d(p_i, p_{i+1}) + \sum_{i=0}^{k-1} c(downtime_i, downtime_{i+1})$$

Where Euclidean distance $d(p_1, p_2)$ is:
$$d(p_1, p_2) = \sqrt{(lat_1 - lat_2)^2 + (lon_1 - lon_2)^2}$$

And downtime cost $c(d_1, d_2)$ is:
$$c(d_1, d_2) = d_1 + d_2$$

### Ant Colony Transition Probability
$$p_{ij}^k = \frac{[\tau_{ij}]^\alpha \cdot [\eta_{ij}]^\beta}{\sum_{l \in \text{unvisited}} [\tau_{il}]^\alpha \cdot [\eta_{il}]^\beta}$$
$$\eta_{ij} = \frac{1}{d(p_i, p_j) + c(d_i, d_j) + \epsilon}$$

### Pheromone Evaporation & Deposition
$$\tau_{ij} \leftarrow (1 - \rho)\tau_{ij} + \sum_{k: (i,j) \in P_k} \frac{Q}{f(P_k)}$$
$$\tau_{ij} \leftarrow \max(\tau_{ij}, 10^{-6})$$

### Adaptive Hyperparameter Rules

| Algorithm | Problem Scale ($N$) | Population / Ants | Iterations / Generations | Mutation / Exploration | Local Search (2-Opt) |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Ant Colony** | $N \le 10$ | $m = 3$ | 200 | $\alpha = 5.0, \beta = 1.5, \rho = 0.5, Q = 100$ | — |
| **Ant Colony** | $N > 10$ | $m = 8$ | 200 | $\alpha = 5.0, \beta = 2.0, \rho = 0.5, Q = 100$ | — |
| **Genetic** | $N < 15$ | 50 | 50 | Mutation rate: 0.2 | — |
| **Genetic** | $N \ge 15$ | 100 | 50 | Mutation rate: 0.1 | — |
| **Memetic** | $N < 40$ | 50 | 10 | Mutation rate: 0.2 | 2-opt ($p=0.2$, 10 iter) |
| **Memetic** | $N \ge 40$ | 150 | 50 | Mutation rate: 0.1 | 2-opt ($p=0.2$, 10 iter) |

---

## 6. Development & Operations Guide

### Building and Running Locally
```bash
# Go server (port 8080)
PORT=8080 go run ./go/main.go

# Rust server (port 8082 or custom)
PORT=8082 cargo run --release --manifest-path rust/Cargo.toml
```

### Running Test Suites
```bash
# Go unit & handler tests
cd go && go test -v ./...

# Rust unit & handler tests
cargo test --manifest-path rust/Cargo.toml

# Comparative benchmark suite across Python, Go, and Rust
/home/klebs/Documentos/ant_colony_fast_api_backend/.venv/bin/python tests/compare_backends.py
```

---

## 7. Performance Benchmarks Summary

### Direct Algorithm Execution Times (Python vs Go vs Rust)

| Problem Dataset | Python ACO | Go ACO | Rust ACO | Python GA | Go GA | Rust GA | Python Memetic | Go Memetic | Rust Memetic | Rust vs Py Speedup |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **5 Turbines** | `115.15 ms` | `1.00 ms` | **`0.44 ms`** | `115.87 ms` | `1.40 ms` | **`0.48 ms`** | `70.11 ms` | `0.60 ms` | **`0.25 ms`** | **261x – 280x** |
| **10 Turbines** | `271.13 ms` | `3.40 ms` | **`1.25 ms`** | `198.74 ms` | `3.00 ms` | **`0.60 ms`** | `467.17 ms` | `3.80 ms` | **`1.10 ms`** | **217x – 425x** |
| **15 Turbines** | `1314.80 ms` | `12.20 ms` | **`7.56 ms`** | `571.59 ms` | `10.00 ms` | **`1.57 ms`** | `1617.55 ms` | `12.00 ms` | **`2.66 ms`** | **174x – 608x** |
| **20 Turbines** | `2108.59 ms` | `20.80 ms` | **`12.61 ms`** | `851.95 ms` | `10.80 ms` | **`2.64 ms`** | `2237.56 ms` | `16.60 ms` | **`4.47 ms`** | **167x – 500x** |
| **40 Turbines** | `6372.23 ms` | `77.20 ms` | **`45.06 ms`** | `1479.42 ms` | `22.40 ms` | **`2.82 ms`** | `142633.81 ms` | `954.60 ms` | **`166.94 ms`** | **141x – 854x** |

### HTTP API Latency Comparison

| Endpoint | FastAPI (Python) | Golang Backend | Rust Backend | Best Latency Reduction |
| :--- | :--- | :--- | :--- | :--- |
| **`GET /ant-colony/get-turbines-map`** | `2.53 ms` | `1.26 ms` | **`1.00 ms`** | **2.5x faster** |
| **`GET /ant-colony/get-subsystems`** | `1.47 ms` | `0.86 ms` | **`0.86 ms`** | **1.7x faster** |
| **`POST /run-route-optimizer` (ACO)** | `58.82 ms` | `1.56 ms` | **`1.23 ms`** | **47.8x faster** |
| **`POST /run-route-optimizer` (Genetic)** | `79.77 ms` | `2.24 ms` | **`1.68 ms`** | **47.5x faster** |
| **`POST /run-route-optimizer` (Memetic)** | `24.74 ms` | `1.47 ms` | **`1.16 ms`** | **21.3x faster** |
