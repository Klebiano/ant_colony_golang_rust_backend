# Ant Colony & Metaheuristics Golang & Rust Backend

High-performance Go and Rust microservices for optimizing offshore wind farm turbine maintenance routing. Replicated from the original Python FastAPI backend, these services compute optimal navigation routes using bio-inspired metaheuristic algorithms—balancing travel distance against turbine fault downtime costs with up to **854x faster algorithm execution speeds** and **48x reduced HTTP API latency**.

---

## ✨ Features

- ⚡ **Ultra-High Performance & Low Latency**: Compiled Go and Rust implementations delivering sub-millisecond to low-millisecond route optimization.
- 🛸 **Bio-Inspired Metaheuristics**: Implements **Ant Colony Optimization (ACO)**, **Genetic Algorithm (GA)**, and **Memetic Algorithm (GA + 2-opt local search)** across Python, Go, and Rust with full algorithmic parity.
- 🗺️ **Geographical Route Normalization**: Computes optimal navigation paths starting and ending at the dock (`Doca`), applying Min-Max scaling to turbine coordinates and downtime costs.
- 🗄️ **Shared SQLite Storage & Datasets**: Common SQLite schema and seed datasets located in root `database/` (`database.sql`, `wind_farm_map_points.csv`, `subsystem.csv`, `downtime.csv`).
- 🌐 **REST API & CORS Enabled**: Standardized HTTP endpoints with cross-origin resource sharing middleware for easy integration with web dashboards (React, Vite, Next.js).
- 🔬 **Comprehensive 3-Way Benchmarks**: End-to-end benchmarking suite comparing Python FastAPI, Go, and Rust backends.

---

## 🛠️ Tech Stack

- **Languages & Frameworks**: 
  - **Go (`/go`)**: Standard Library `net/http`, custom CORS middleware, `database/sql`, [`github.com/mattn/go-sqlite3`](https://github.com/mattn/go-sqlite3), `math/rand/v2`
  - **Rust (`/rust`)**: `axum 0.8`, `tokio 1.x`, `tower-http 0.6` (CORS), `rusqlite 0.33`, `serde`, `rand 0.8`
  - **Python 3.12+**: FastAPI, Uvicorn, NumPy, Pandas, SQLAlchemy (reference backend)
- **Database & Storage**: SQLite (`sql_app.db`), seeded via `database/database.sql`
- **Data Serialization**: JSON (`encoding/json`, `serde_json`), CSV (`encoding/csv`, `csv`)
- **Testing & Benchmarking**: Go `testing` package, Rust `cargo test`, Python comparative runner (`tests/compare_backends.py`)

---

## 📁 Repository Structure

```
ant_colony_golang_rust_backend/
├── database/                   # Shared SQLite datasets & SQL seeding script
│   ├── database.sql            # Seed ANSI SQL DDL & DML script
│   ├── downtime.csv            # Subsystem fault downtime dataset
│   ├── subsystem.csv           # Subsystem definitions dataset
│   └── wind_farm_map_points.csv# Geographical turbine map dataset
├── go/                         # Golang microservice backend
│   ├── algorithms/             # Core metaheuristic solvers (ACO, GA, Memetic)
│   ├── database/               # SQLite database queries & connection logic
│   ├── handlers/               # HTTP request handlers & CORS middleware
│   ├── models/                 # Struct definitions & JSON DTOs
│   ├── tests/                  # Go benchmark suites (testing.B)
│   ├── go.mod                  # Go module definition
│   ├── go.sum                  # Dependency checksums
│   └── main.go                 # Go server entrypoint (port 8080)
├── rust/                       # Rust microservice backend
│   ├── src/
│   │   ├── algorithms/         # Rust metaheuristic solvers (ACO, GA, Memetic)
│   │   ├── database.rs         # SQLite queries (rusqlite)
│   │   ├── handlers.rs         # Axum HTTP request handlers
│   │   ├── models.rs           # Rust data models & DTOs
│   │   ├── lib.rs              # Router & AppState definitions
│   │   └── main.rs             # Axum HTTP server entrypoint (port 8082 / 8080)
│   ├── tests/                  # Unit, handler integration & benchmark tests
│   └── Cargo.toml              # Rust crate manifest & dependencies
├── tests/                      # Shared problem datasets & comparison suite
│   ├── inputs/                 # Benchmark CSV datasets (5 to 100 turbines)
│   └── compare_backends.py     # 3-way comparative benchmark runner
├── sql_app.db                  # Shared SQLite database file
├── README.md                   # Project documentation & benchmark overview
└── CONTEXT.md                  # Comprehensive technical reference
```

---

## 🚀 Quick Start

### 1. Prerequisites

Ensure Go 1.24+ and Rust (Cargo) 1.80+ are installed:

```bash
go version
cargo --version
```

### 2. Running the Go Backend

```bash
# From workspace root
PORT=8080 go run ./go/main.go

# Or inside the go directory
cd go
PORT=8080 go run main.go
```

### 3. Running the Rust Backend

```bash
# From workspace root
PORT=8082 cargo run --release --manifest-path rust/Cargo.toml

# Or inside the rust directory
cd rust
PORT=8082 cargo run --release
```

---

## 🧪 Running Tests & Benchmarks

### Go Test Suite & Benchmarks
```bash
# Run Go unit & handler tests
cd go
go test -v ./...

# Run Go algorithm benchmark suite
go test -v -run TestPrintSpeedBenchmarkResults ./tests
```

### Rust Test Suite & Benchmarks
```bash
# Run Rust unit & handler integration tests
cargo test --manifest-path rust/Cargo.toml

# Run Rust algorithm benchmark suite (release mode)
cargo test --release --manifest-path rust/Cargo.toml --test benchmarks_test -- --nocapture
```

### 3-Way Comparative Benchmark (Python vs Go vs Rust)
```bash
# Run full direct algorithm and HTTP API benchmarks across Python, Go, and Rust
/home/klebs/Documentos/ant_colony_fast_api_backend/.venv/bin/python tests/compare_backends.py
```

---

## 📡 API Reference & Endpoints

Both Go (port `8080`) and Rust (port `8082` / `8080`) expose identical REST endpoints:

### 1. `GET /ant-colony/get-turbines-map`
Fetches all wind farm turbines and geographical coordinates from the SQLite database.

**Response Example (JSON):**
```json
[
  {
    "turbine_id": 1,
    "turbine_name": "Doca",
    "latitude": -4.9525,
    "longitude": -36.8828
  },
  {
    "turbine_id": 2,
    "turbine_name": "BETA-01",
    "latitude": -4.9262,
    "longitude": -36.7855
  }
]
```

### 2. `GET /ant-colony/get-subsystems`
Retrieves all wind turbine subsystem categories from the database.

**Response Example (JSON):**
```json
[
  {
    "subsystem_id": 1,
    "subsystem_name": "Electrical System"
  },
  {
    "subsystem_id": 2,
    "subsystem_name": "Electronic Control"
  }
]
```

### 3. `POST /ant-colony/run-route-optimizer`
Optimizes the maintenance traversal path for faulted turbines.

* **Query Parameters**:
  * `algorithm` (*optional*): `Ant Colony` (default), `Genetic`, or `Memetic`.
* **Request Body** (*optional* JSON array; falls back to `tests/inputs/problem_5_turbines.csv` when omitted/empty):

```json
[
  {
    "turbine_id": 2,
    "subsystem_name": "Electrical System",
    "fault_type": "Minor"
  },
  {
    "turbine_id": 3,
    "subsystem_name": "Rotor Hub",
    "fault_type": "Major"
  }
]
```

* **Response Example (JSON)**:

```json
{
  "turbine_order": ["Doca", "BETA-01", "BETA-02", "Doca"],
  "turbine_order_to_show": ["Doca", "BETA-01", "BETA-02", "Doca"],
  "best_path": [0, 1, 2, 0],
  "best_path_length": 1.4285,
  "best_downtime_days": 0.8542,
  "best_path_len_downtime": 2.2827,
  "time_to_run_sec": 0.00044
}
```

---

## 🧬 Algorithm Heuristics & Adaptive Tuning

The engines dynamically adjust hyperparameters based on the problem scale ($N$ faulted turbines):

| Algorithm | Scale Condition | Population / Ants | Iterations / Generations | Exploration & Mutation | Local Search |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Ant Colony (ACO)** | $N \le 10$ turbines | $m = 3$ ants | 200 iterations | $\alpha = 5.0, \beta = 1.5, \rho = 0.5, Q = 100$ | — |
| **Ant Colony (ACO)** | $N > 10$ turbines | $m = 8$ ants | 200 iterations | $\alpha = 5.0, \beta = 2.0, \rho = 0.5, Q = 100$ | — |
| **Genetic (GA)** | $N < 15$ turbines | 50 individuals | 50 generations | Mutation rate: 0.2 | — |
| **Genetic (GA)** | $N \ge 15$ turbines | 100 individuals | 50 generations | Mutation rate: 0.1 | — |
| **Memetic (GA + 2-opt)** | $N < 40$ turbines | 50 individuals | 10 generations | Mutation rate: 0.2 | 2-opt (20% chance, 10 iter) |
| **Memetic (GA + 2-opt)** | $N \ge 40$ turbines | 150 individuals | 50 generations | Mutation rate: 0.1 | 2-opt (20% chance, 10 iter) |

---

## 📁 Benchmark Datasets

Located in `tests/inputs/`:
* `problem_5_turbines.csv` (5 turbines + Doca)
* `problem_10_turbines.csv` (10 turbines + Doca)
* `problem_15_turbines.csv` (15 turbines + Doca)
* `problem_20_turbines.csv` (20 turbines + Doca)
* `problem_40_turbines.csv` (40 turbines + Doca)
* `problem_100_turbines.csv` (100 turbines + Doca)

---

## 📊 Performance Comparison (Python FastAPI vs Go vs Rust)

### 1. Direct Algorithm Execution Times

| Problem Dataset | Python ACO | Go ACO | Rust ACO | Python GA | Go GA | Rust GA | Python Memetic | Go Memetic | Rust Memetic | Rust vs Py Speedup |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **5 Turbines** | `115.15 ms` | `1.00 ms` | **`0.44 ms`** | `115.87 ms` | `1.40 ms` | **`0.48 ms`** | `70.11 ms` | `0.60 ms` | **`0.25 ms`** | **261x – 280x** |
| **10 Turbines** | `271.13 ms` | `3.40 ms` | **`1.25 ms`** | `198.74 ms` | `3.00 ms` | **`0.60 ms`** | `467.17 ms` | `3.80 ms` | **`1.10 ms`** | **217x – 425x** |
| **15 Turbines** | `1314.80 ms` | `12.20 ms` | **`7.56 ms`** | `571.59 ms` | `10.00 ms` | **`1.57 ms`** | `1617.55 ms` | `12.00 ms` | **`2.66 ms`** | **174x – 608x** |
| **20 Turbines** | `2108.59 ms` | `20.80 ms` | **`12.61 ms`** | `851.95 ms` | `10.80 ms` | **`2.64 ms`** | `2237.56 ms` | `16.60 ms` | **`4.47 ms`** | **167x – 500x** |
| **40 Turbines** | `6372.23 ms` | `77.20 ms` | **`45.06 ms`** | `1479.42 ms` | `22.40 ms` | **`2.82 ms`** | `142633.81 ms` | `954.60 ms` | **`166.94 ms`** | **141x – 854x** |

> **Key Takeaways**: 
> * **Rust vs Python**: Rust provides up to **854x speedup** over Python FastAPI (Memetic on 40 turbines executes in 166.9 ms in Rust vs 142.6 seconds in Python).
> * **Rust vs Go**: Rust delivers an additional **1.6x to 7.9x speedup over Go** across all problem scales due to zero-cost abstractions, direct cache locality, and aggressive compiler vectorization.

### 2. HTTP API Response Latency

| Endpoint | FastAPI (Python) | Golang Backend | Rust Backend | Best Latency Reduction |
| :--- | :--- | :--- | :--- | :--- |
| **`GET /ant-colony/get-turbines-map`** | `2.53 ms` | `1.26 ms` | **`1.00 ms`** | **2.5x faster** |
| **`GET /ant-colony/get-subsystems`** | `1.47 ms` | `0.86 ms` | **`0.86 ms`** | **1.7x faster** |
| **`POST /run-route-optimizer` (ACO)** | `58.82 ms` | `1.56 ms` | **`1.23 ms`** | **47.8x faster** |
| **`POST /run-route-optimizer` (Genetic)** | `79.77 ms` | `2.24 ms` | **`1.68 ms`** | **47.5x faster** |
| **`POST /run-route-optimizer` (Memetic)** | `24.74 ms` | `1.47 ms` | **`1.16 ms`** | **21.3x faster** |

---

## 📄 License

Distributed under the MIT License.
