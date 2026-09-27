# Automated API Load Testing Platform

An enterprise-ready, developer-centric platform to benchmark, stress test, and monitor API performance with real-time telemetry.

---

## Phase 1, 2 & 3: Load Engine, WebSocket Pipeline & Prometheus Exporter

### Architecture Highlights
1. **Core Load Generation Engine**: High-concurrency worker pool using bounded Go goroutines with graceful context cancellation and ramp-up scheduling.
2. **Real-Time WebSocket Pipeline**: Gorilla WebSocket Hub streaming live 500ms metric ticks to connected frontend and testing clients (`/ws/tests/:testId` or `/ws/metrics`).
3. **Prometheus Telemetry**: Native Prometheus `/metrics` exporter tracking requests, success/failures, latency histograms, live throughput gauges, active virtual users, and test statuses.
4. **Thread-Safe Metrics Collector**: Lock-minimized accumulator calculating Total Requests, Successful vs Failed requests, Error Rate %, Requests Per Second (RPS), Min/Max/Avg Latency, and exact percentiles (P50, P90, P95, P99).
5. **Safety Guardrails & Validator**: Limits concurrency (max 500 VUs), duration (max 10 mins), total requests (max 1M), and rejects non-HTTP schemes.
6. **SQLite Persistence**: Automatically stores completed test run summaries and percentile statistics in SQLite with WAL mode.
7. **REST API**: Clean JSON endpoints for creating, starting, monitoring, stopping, and inspecting load tests.
8. **Mock Companion API (`demo-api`)**: Target service on `:8081` with `/health`, `/api/users`, `/api/products`, and `/api/orders` supporting artificial latency, random jitter, and simulated error rates.

---

## Prerequisites

- **Go**: 1.22+ (Verified with `go1.27.0 windows/amd64`)
- **Node.js & npm**: Node 20+ (Verified with `v24.12.0`)

---

## Running the Services Locally

### 1. Start the Mock Demo API (Target)

Open a terminal and run:
```powershell
cd demo-api
go run main.go
```
The Mock API starts on `http://localhost:8081`.

#### Available Demo Endpoints:
| Method | Endpoint | Default Delay | Description |
| :--- | :--- | :--- | :--- |
| `GET` | `/health` | Instant | Demo API health status |
| `GET` | `/api/users` | ~15ms | Fast mock user list |
| `GET` | `/api/products` | ~60ms (+ jitter) | Catalog items with simulated query latency |
| `POST` | `/api/orders` | ~80ms (+ jitter) | Write operation simulating payment and database locks |

#### Configurable Latency & Failure Testing:
You can append query parameters to test how load testing handles various API behaviors:
- `?delay_ms=250`: Injects an exact artificial delay in milliseconds.
- `?error_rate=0.25`: Injects a 25% probability of HTTP 500 errors.

---

### 2. Start the Backend Server

Open a second terminal and run:
```powershell
cd backend
go run ./cmd/server/main.go
```
The Backend server starts on `http://localhost:8080`.

---

## API & Telemetry Reference

### REST Endpoints
| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `GET` | `/health` / `/api/v1/health` | Service and database health check |
| `GET` | `/metrics` / `/api/v1/metrics`| Prometheus metrics scrape endpoint |
| `POST` | `/api/v1/tests` | Start a new load test |
| `GET` | `/api/v1/tests/status` | Get current engine status and active configuration |
| `GET` | `/api/v1/tests/metrics` | Get current live or final metrics snapshot |
| `POST` | `/api/v1/tests/stop` | Stop the currently running load test |
| `GET` | `/api/v1/tests/:id` | Get persistent test run summary from SQLite |

### Real-Time WebSocket Streaming Endpoints
| Protocol | Endpoint | Description |
| :--- | :--- | :--- |
| `WS` | `/ws/tests/:testId` | Stream live metric ticks for a specific load test run |
| `WS` | `/ws/metrics` or `/ws/tests` | Stream live metric ticks across all active test runs |

#### Live WebSocket Metric Frame (JSON):
```json
{
  "test_id": "c0dc33b0-55c7-447d-a65c-1deae6e63079",
  "status": "running",
  "timestamp": "2026-09-27T07:40:24Z",
  "total_requests": 280,
  "successful_requests": 280,
  "failed_requests": 0,
  "error_rate": 0.0,
  "requests_per_second": 165.8,
  "average_latency_ms": 10.5,
  "p50_latency_ms": 10.5,
  "p95_latency_ms": 11.1,
  "p99_latency_ms": 11.4,
  "active_users": 5
}
```

---

## Automated Test Suites

### Run Backend Tests (Unit + Concurrency + WebSocket + Prometheus):
```powershell
cd backend
go test -count=1 -v ./...
```

### Run Mock API Tests:
```powershell
cd demo-api
go test -count=1 -v ./...
```

### Run Phase 3 End-to-End WebSocket & Prometheus Verification:
```powershell
cd backend
go run ./cmd/verify/main.go
```
