# LoadPulse: Automated API Load Testing Platform

An enterprise-ready, high-performance, developer-centric platform to benchmark, stress test, and monitor API performance with real-time WebSocket telemetry, Prometheus observability, SQLite persistence, and advanced analytics.

---

## 1. System Architecture

```
                                  [ Web Browser / User ]
                                            │
                                            ▼
                     ┌──────────────────────────────────────────────┐
                     │   LoadPulse Frontend (React + Vite + Nginx)  │
                     │          Port 3000 (Docker) / 5173 (Dev)     │
                     └──────────────────────┬───────────────────────┘
                                            │
                     ┌──────────────────────┴───────────────────────┐
                     │ REST API (/api/v1)   │ WebSocket (/ws/)      │
                     ▼                      ▼                       │
    ┌───────────────────────────────────────────────────────────┐   │
    │                    Go Load Engine Backend                 │   │
    │                        Port 8080                          │   │
    │  ┌──────────────────┐  ┌────────────────┐  ┌───────────┐  │   │
    │  │ Worker Pool      │  │ Real-Time Hub  │  │ Prometheus│  │   │
    │  │ (Goroutines + VU)│  │ (WebSocket)   │  │ Exporter  │──┼───┘ Scrapes :8080/metrics
    │  └────────┬─────────┘  └────────────────┘  └───────────┘  │
    └───────────┼─────────────────────┬─────────────────────────┘
                │ HTTP Requests       │ SQLite Read/Write
                ▼                     ▼
    ┌───────────────────────┐   ┌───────────────────────────────┐
    │     Demo Mock API     │   │   SQLite Database (Persistent)│
    │       Port 8081       │   │      /data/loadpulse.db       │
    └───────────────────────┘   └───────────────────────────────┘
```

---

## 2. Tech Stack

| Domain | Technologies & Libraries |
| :--- | :--- |
| **Backend Engine** | Go 1.24+, Gin Web Framework, Gorilla WebSocket, Prometheus Client Golang, ModernC SQLite |
| **Frontend UI** | React 18, Vite, Recharts, Lucide Icons, Vanilla Glassmorphism CSS |
| **Storage** | SQLite with WAL (Write-Ahead Logging) mode and automatic migrations |
| **Containerization** | Docker, Multi-Stage Builds, Docker Compose, Nginx Alpine Reverse Proxy |
| **Observability** | Native Prometheus `/metrics` exporter, real-time 500ms WebSocket metric ticks |

---

## 3. Project Structure

```
automated-api-load-testing/
├── docker-compose.yml          # Multi-container orchestration (Frontend, Backend, Demo API, Volumes)
├── .env.example                # Configuration template
├── .dockerignore               # Global docker ignore rules
├── .gitignore                  # Git ignore rules
│
├── backend/                    # Go Load Engine & REST/WS Server
│   ├── cmd/
│   │   └── server/main.go      # Backend entrypoint (port 8080)
│   ├── internal/
│   │   ├── config/             # Environment variable loading
│   │   ├── database/           # SQLite migrations, schema, and queries
│   │   ├── engine/             # Goroutine worker pool, metrics collector, and percentiles
│   │   ├── handlers/           # Gin REST, WebSocket, history, export, and compare handlers
│   │   ├── metrics/            # Prometheus collectors and live gauges
│   │   └── websocket/          # WebSocket hub and client connections
│   ├── Dockerfile              # Multi-stage Go production container
│   ├── .dockerignore           # Backend docker ignore rules
│   └── go.mod / go.sum
│
├── frontend/                   # React + Vite Production Dashboard
│   ├── src/
│   │   ├── components/         # Metric cards, config panel, progress bar, error boundary, modal
│   │   ├── pages/              # Dashboard, HistoryView, AnalyticsView, ComparisonView, ReportsView
│   │   ├── charts/             # Latency, throughput, error rate, and VU charts (Recharts)
│   │   ├── hooks/              # useWebSocket live telemetry hook
│   │   ├── services/           # REST client with export and compare helpers
│   │   └── styles/             # Modular CSS design system with glassmorphism
│   ├── nginx.conf              # Production Nginx reverse proxy with WS upgrade support
│   ├── Dockerfile              # Multi-stage Node builder + Nginx runtime container
│   ├── .dockerignore           # Frontend docker ignore rules
│   └── package.json
│
└── demo-api/                   # Target Mock API for Safe Benchmarking
    ├── main.go                 # Mock endpoints with configurable delay and error rate (port 8081)
    ├── main_test.go            # Demo API unit tests
    ├── Dockerfile              # Multi-stage container for Demo API
    ├── .dockerignore
    └── go.mod
```

---

## 4. Key Features

1. **High-Performance Load Engine**
   - Controlled concurrent HTTP load generation using lightweight Go goroutines.
   - Configurable Virtual Users (VUs), duration, total requests, request timeouts, and ramp-up schedules.
   - Thread-safe metrics aggregator tracking RPS, total, successful, and failed requests.
   - Precise latency percentiles: **P50 (Median)**, **P90**, **P95**, and **P99 (Tail Outlier)**.

2. **Real-Time Telemetry & WebSocket Pipeline**
   - Live 500ms metric updates pushed over WebSockets to `/ws/metrics` or `/ws/tests/:testId`.
   - Real-time Recharts visualization of latency, throughput, error rate, and active virtual users.

3. **Prometheus Observability**
   - Native Prometheus `/metrics` endpoint.
   - Tracks counters (`loadtest_requests_total`), gauges (`loadtest_active_users`, `loadtest_current_rps`), and histograms (`loadtest_request_duration_seconds`).

4. **Persistent Test History & Filtering**
   - Every completed test is automatically persisted in SQLite.
   - Comprehensive test history table with search, status filtering, HTTP method filtering (`GET`, `POST`, `PUT`, `DELETE`), and multi-column sorting.

5. **Advanced Analytics**
   - Peak throughput and virtual user KPI cards.
   - Latency percentiles benchmarked against SLA thresholds.
   - Min / Avg / Max latency spread breakdown.
   - Historical multi-run trend lines for RPS and response times.

6. **Benchmark Comparison**
   - Select 2 or more tests to view a side-by-side benchmark matrix.
   - Differential KPI cards with colored improvement/degradation badges (`RPS Delta`, `Avg Latency Delta`, `P95 Delta`, `Error Rate Delta`).

7. **Multi-Format Report Downloads**
   - Export test results directly from the UI or REST API in **JSON**, **Markdown (GFM)**, or **CSV** formats.
   - Built-in print stylesheet for printable audit reports.

---

## 5. Docker Setup & Deployment

### Quick Start with Docker Compose
Run the entire production stack with a single command:

```bash
docker compose up --build -d
```

### Services & Port Mappings
| Service | Container Name | Port Mapping | Healthcheck Endpoint |
| :--- | :--- | :--- | :--- |
| **Frontend** | `loadpulse-frontend` | `http://localhost:3000` | `http://localhost:80/` |
| **Go Backend** | `loadpulse-backend` | `http://localhost:8080` | `http://localhost:8080/health` |
| **Demo API** | `loadpulse-demo-api` | `http://localhost:8081` | `http://localhost:8081/health` |

### Verifying Persistent Storage
The SQLite database is stored in a named Docker volume (`loadpulse-data` mounted at `/data/loadpulse.db`).

To verify persistence across restarts:
```bash
# 1. Run a test from the UI or API
# 2. Stop the stack
docker compose down

# 3. Start the stack again
docker compose up -d

# 4. Check test history: all previous test runs remain preserved!
```

---

## 6. Local Development (Without Docker)

### Prerequisites
- **Go**: 1.22+
- **Node.js**: 20+

### Step 1: Start the Demo Mock API
```bash
cd demo-api
go run main.go
# Running on http://localhost:8081
```

### Step 2: Start the Go Backend
```bash
cd backend
go run ./cmd/server/main.go
# Running on http://localhost:8080
```

### Step 3: Start the React Frontend
```bash
cd frontend
npm install
npm run dev
# Running on http://localhost:5173
```

Open `http://localhost:5173` in your browser.

---

## 7. Environment Variables Reference

Copy `.env.example` to `.env` to customize settings:

| Variable | Default Value | Description |
| :--- | :--- | :--- |
| `PORT` | `8080` | HTTP port for the Go backend server |
| `ENV` | `development` / `production` | Gin runtime mode (`debug` vs `release`) |
| `DB_PATH` | `loadtest.db` (local) / `/data/loadpulse.db` (docker) | File path to SQLite database file |
| `ALLOWED_ORIGINS` | `http://localhost:3000,http://localhost:5173` | Allowed CORS origins (comma-separated) |
| `DEMO_API_HOST` | `demo-api:8081` | Container alias for automatic target translation in Docker |
| `DEMO_PORT` | `8081` | Port for the companion Demo Mock API |

---

## 8. API & Telemetry Reference

### REST Endpoints
| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `GET` | `/health` | System and SQLite connectivity health status |
| `GET` | `/metrics` | Prometheus metrics scrape endpoint |
| `POST` | `/api/v1/tests` | Start a new load test |
| `GET` | `/api/v1/tests` | List historical test runs with search, filtering, and pagination |
| `GET` | `/api/v1/tests/status` | Current load engine status (`idle`, `running`, `completed`, `stopped`) |
| `GET` | `/api/v1/tests/metrics` | Current live snapshot or final test metrics |
| `POST` | `/api/v1/tests/stop` | Gracefully terminate an ongoing test run |
| `GET` | `/api/v1/tests/summary` | Global aggregated benchmark statistics |
| `GET` | `/api/v1/tests/compare?ids=id1,id2` | Side-by-side benchmark comparison and delta calculations |
| `GET` | `/api/v1/tests/:id` | Detailed test record by UUID |
| `DELETE`| `/api/v1/tests/:id` | Delete a single test record |
| `DELETE`| `/api/v1/tests` | Clear entire test history |
| `GET` | `/api/v1/tests/:id/export?format=json` | Export test report as JSON |
| `GET` | `/api/v1/tests/:id/export?format=markdown` | Export test report as GitHub Flavored Markdown |
| `GET` | `/api/v1/tests/:id/export?format=csv` | Export test report as CSV |

### WebSocket Endpoints
| Protocol | Endpoint | Description |
| :--- | :--- | :--- |
| `WS` | `/ws/metrics` | Stream live metric updates across all active test runs |
| `WS` | `/ws/tests/:testId` | Stream live metric updates for a specific test run |

### Demo API Endpoints
| Method | Endpoint | Default Delay | Description |
| :--- | :--- | :--- | :--- |
| `GET` | `/health` | Instant | Demo API health check |
| `GET/POST/PUT/DELETE` | `/api/users` | ~15ms | Fast mock user directory |
| `GET/POST/PUT/DELETE` | `/api/products` | ~50ms (+ jitter) | Catalog items with simulated latency |
| `GET/POST` | `/api/orders` | ~70ms (+ jitter) | Order transactions with stock validation |

*Query Parameters*:
- `?delay_ms=<ms>`: Inject explicit artificial latency (e.g. `?delay_ms=200`).
- `?error_rate=<0.0-1.0>`: Inject simulated HTTP 500 error probability (e.g. `?error_rate=0.25`).

---

## 9. Testing & Quality Assurance

### Run Backend Unit & Integration Tests:
```bash
cd backend
go vet ./...
go test -v ./...
```

### Run Demo API Tests:
```bash
cd demo-api
go test -v ./...
```

### Run Frontend Unit Tests:
```bash
cd frontend
npm test
```

### Run Frontend Production Build:
```bash
cd frontend
npm run build
```

---

## 10. Security & Responsible Testing Safeguards

This platform is strictly designed for authorized, local, and staging environment load testing:
- **Concurrency Caps**: Virtual users bounded to 500 max to prevent unintended client-side exhaustion.
- **Duration Limits**: Test duration strictly capped at 600 seconds (10 minutes).
- **Scheme Validation**: Restricted to valid `http://` and `https://` endpoints.
- **Safety**: Contains no capabilities for authentication bypass, credential stuffing, rate-limit evasion, WAF bypass, or DDoS attacks.
