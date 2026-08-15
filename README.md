# Ant Colony & Metaheuristics Golang Backend

A high-performance Go microservice for optimizing offshore wind farm turbine maintenance routing. Replicated from the Python FastAPI backend, this service computes optimal navigation routes using bio-inspired metaheuristic algorithms—balancing travel distance against turbine fault downtime costs with up to **127x faster algorithm execution speeds** and **26x reduced HTTP latency**.

---

## ✨ Features

- ⚡ **High Performance & Low Latency**: Compiled Go implementation delivering sub-millisecond to low-millisecond route optimization.
- 🛸 **Bio-Inspired Metaheuristics**: Implements **Ant Colony Optimization (ACO)**, **Genetic Algorithm (GA)**, and **Memetic Algorithm (GA + 2-opt local search)**.
- 🗺️ **Geographical Route Normalization**: Computes optimal navigation paths starting and ending at the dock (`Doca`), applying Min-Max scaling to turbine coordinates and downtime costs.
- 🗄️ **Embedded SQLite Storage**: Built-in SQLite database support using `database/sql` and `github.com/mattn/go-sqlite3`, auto-seeding schema and data from ANSI SQL scripts (`database/database.sql`).
- 🌐 **REST API & CORS Enabled**: Standardized HTTP endpoints with cross-origin resource sharing middleware for easy integration with web dashboards (React, Vite, Next.js).

---

## 🛠️ Tech Stack

- **Core Language**: Go 1.24+
- **Web Framework**: Standard Library `net/http` with custom CORS middleware
- **Database & Storage**: SQLite, `database/sql`, [`github.com/mattn/go-sqlite3`](https://github.com/mattn/go-sqlite3)
- **Data Serialization**: Standard `encoding/json`, `encoding/csv`
- **Testing & Benchmarking**: Standard Go `testing` package and comparative Python benchmark runner (`compare_backends.py`)

---

## 🚀 Quick Start

### 1. Prerequisites

Ensure Go 1.24 or higher is installed:

```bash
go version
```

### 2. Download Dependencies

Clone the repository and fetch dependencies:

```bash
git clone https://github.com/Klebiano/ant_colony_golang_backend.git
cd ant_colony_golang_backend

go mod download
```

### 3. Running the Development Server

Start the Go backend server:

```bash
go run main.go
```

By default, the server runs on port `8080`. You can specify a custom port via environment variable:

```bash
PORT=8000 go run main.go
```

---

## 🧪 Running Tests & Benchmarks

Run unit and handler integration tests:

```bash
go test -v ./...
```

Run internal Go algorithm benchmark suites:

```bash
go test -v ./tests
```

Run end-to-end comparative benchmarks against the Python FastAPI backend:

```bash
/home/klebs/Documentos/ant_colony_fast_api_backend/.venv/bin/python tests/compare_backends.py
```

---

## 📡 API Reference & Endpoints

### 1. `GET /ant-colony/get-turbines-map`
Fetches all wind farm turbines and geographical coordinates from the SQLite database.

**Response Example (JSON):**
```json
[
  {
    "turbine_id": 1,
    "turbine_name": "Doca",
    "latitude": -23.0123,
    "longitude": -43.1234
  },
  {
    "turbine_id": 2,
    "turbine_name": "T01",
    "latitude": -23.0155,
    "longitude": -43.1288
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
    "subsystem_name": "Rotor Hub"
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
  "turbine_order": ["Doca", "T01", "T02", "Doca"],
  "turbine_order_to_show": ["Doca", "T01", "T02", "Doca"],
  "best_path": [0, 1, 2, 0],
  "best_path_length": 1.4285,
  "best_downtime_days": 0.8542,
  "best_path_len_downtime": 2.2827,
  "time_to_run_sec": 0.0024
}
```

---

## 🧬 Algorithm Heuristics & Adaptive Tuning

The engine dynamically adjusts hyperparameters based on the problem scale ($N$ faulted turbines):

| Algorithm | Scale Condition | Population / Ants | Iterations / Generations | Exploration & Mutation | Local Search |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Ant Colony (ACO)** | $N \le 10$ turbines | $m = 3$ ants | 200 iterations | $\alpha = 5.0, \beta = 1.5, \rho = 0.5$ | — |
| **Ant Colony (ACO)** | $N > 10$ turbines | $m = 8$ ants | 200 iterations | $\alpha = 5.0, \beta = 2.0, \rho = 0.5$ | — |
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

## 📊 Performance Comparison (Python FastAPI vs Go Backend)

### Direct Algorithm Execution Times

| Dataset | Python ACO (ms) | Go ACO (ms) | Speedup | Python GA (ms) | Go GA (ms) | Speedup | Python Memetic (ms) | Go Memetic (ms) | Speedup |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **5 Turbines** | `125.84 ms` | **`1.60 ms`** | **~78x** | `105.83 ms` | **`1.20 ms`** | **~88x** | `57.61 ms` | **`0.80 ms`** | **~72x** |
| **10 Turbines** | `315.86 ms` | **`5.40 ms`** | **~58x** | `188.36 ms` | **`2.60 ms`** | **~72x** | `710.61 ms` | **`5.40 ms`** | **~131x** |
| **15 Turbines** | `1450.00 ms` | **`16.00 ms`** | **~90x** | `548.31 ms` | **`9.60 ms`** | **~57x** | `2315.29 ms` | **`21.20 ms`** | **~109x** |
| **20 Turbines** | `2361.06 ms` | **`28.00 ms`** | **~84x** | `750.02 ms` | **`11.60 ms`** | **~64x** | `4063.17 ms` | **`29.40 ms`** | **~138x** |
| **40 Turbines** | `7602.03 ms` | **`100.00 ms`** | **~76x** | `1407.40 ms` | **`22.00 ms`** | **~64x** | `148236.24 ms` | **`1210.20 ms`** | **~122x** |

### HTTP API Response Latency

| Endpoint | FastAPI (Python) | Golang Backend | Latency Reduction |
| :--- | :--- | :--- | :--- |
| **`GET /ant-colony/get-turbines-map`** | `6.22 ms` | **`1.39 ms`** | **4.5x faster** |
| **`GET /ant-colony/get-subsystems`** | `2.08 ms` | **`1.08 ms`** | **1.9x faster** |
| **`POST /run-route-optimizer` (ACO)** | `60.29 ms` | **`2.61 ms`** | **23.1x faster** |
| **`POST /run-route-optimizer` (Genetic)** | `70.94 ms` | **`2.72 ms`** | **26.1x faster** |
| **`POST /run-route-optimizer` (Memetic)** | `24.57 ms` | **`2.09 ms`** | **11.7x faster** |

---

## 📄 License

Distributed under the MIT License.
