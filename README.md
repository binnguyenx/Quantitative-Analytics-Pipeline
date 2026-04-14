# 💰 FinBud — Financial Decision-Making Backend for Gen Z

A high-performance Go backend that helps Gen-Z users make smarter financial decisions through simulations, profiling, and event-driven insights.

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.22+ |
| Framework | [Gin](https://github.com/gin-gonic/gin) |
| Database | PostgreSQL 16 (via [GORM](https://gorm.io)) |
| Cache | Redis 7 |
| Messaging | Apache Kafka 3.7 (KRaft — no Zookeeper) |
| Observability | Prometheus + Grafana |
| Load Testing | k6 |
| Architecture | Clean Architecture / Hexagonal |

## Project Structure

```
FINBUD/
├── cmd/api/main.go                  # API entry point
├── cmd/consumer/main.go             # Kafka ingestion consumer entry point
├── api/router.go                    # Gin route definitions (driving adapter)
├── internal/
│   ├── config/config.go             # Env-based configuration loader
│   ├── domain/models.go             # Core entities & DTOs (framework-agnostic)
│   ├── handler/                     # HTTP handlers (driving adapters)
│   │   ├── health_handler.go
│   │   ├── user_handler.go
│   │   ├── profile_handler.go
│   │   └── simulator_handler.go
│   ├── middleware/middleware.go      # Request logger & CORS
│   ├── repository/                  # Data access layer (driven adapters)
│   │   ├── interfaces.go            # Port definitions (contracts)
│   │   ├── user_repo.go
│   │   ├── profile_repo.go
│   │   └── scenario_repo.go
│   └── service/                     # Business logic (use-case layer)
│       ├── user_service.go
│       ├── profile_service.go       # Profile CRUD + cache + Kafka events
│       └── simulator_service.go     # Decision Engine (compound interest, debt payoff)
├── pkg/                             # Reusable infrastructure clients
│   ├── database/postgres.go         # GORM connection + pool tuning
│   ├── redis/redis.go               # Redis client factory
│   └── kafka/producer.go            # Kafka writer wrapper
├── migrations/001_init_schema.sql   # Base schema
├── migrations/002_ingested_events.sql
├── benchmarks/k6/ingestion.js       # Load profile updates to stress ingestion
├── monitoring/                      # Prometheus + Grafana provisioning
├── docker-compose.yml               # Local dev infrastructure
├── Makefile                         # Dev commands
├── env.example                      # Sample environment variables
└── go.mod
```

## Getting Started

### Prerequisites

- [Go 1.22+](https://go.dev/dl/)
- [Docker & Docker Compose](https://docs.docker.com/get-docker/)

### 1. Clone & configure

```bash
git clone https://github.com/finbud/finbud-backend.git
cd finbud-backend
cp env.example .env   # edit values if needed
```

### 2. Start infrastructure

```bash
make infra-up
# or: docker-compose up -d
```

This spins up **PostgreSQL**, **Redis**, and **Kafka** (KRaft mode).

### 3. Install dependencies

```bash
make tidy
# or: go mod tidy
```

### 4. Run the server

```bash
make run
# or: go run ./cmd/api
```

The API starts on `http://localhost:8080`.

### 5. Run the Kafka ingestion consumer

```bash
make consumer
# or: go run ./cmd/consumer
```

Consumer metrics are exposed at `http://localhost:2113/metrics` by default.

### 6. Verify

```bash
curl http://localhost:8080/health
# → {"service":"finbud-backend","status":"ok"}
```

## Ingestion & Benchmark

### Ingestion pipeline

- Consumer group reads from `EVENT_PROFILE_UPDATED`
- Event schema validation + idempotent `event_id` handling
- Batch persist to `ingested_events` (unique `event_id`)
- Retry with exponential backoff
- Invalid events routed to DLQ topic `EVENT_PROFILE_UPDATED_DLQ`

### Metrics

- API metrics: `http://localhost:8080/metrics`
- Consumer metrics: `http://localhost:2113/metrics`
- Start dashboards:

```bash
make metrics-up
```

- Grafana: `http://localhost:3000` (admin/admin)
- Prometheus: `http://localhost:9090`

### Run benchmark

```bash
make bench
```

You can tune k6 stages with env vars:

```bash
BASE_URL=http://localhost:8080 STEADY_VUS=50 STEADY_DURATION=3m k6 run ./benchmarks/k6/ingestion.js
```

Use `BENCHMARK_REPORT.md` as the report template.

## API Endpoints

### Health

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Liveness probe |

### Users

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/v1/users` | Create a new user |
| `GET` | `/api/v1/users/:id` | Get user by ID (includes profile) |
| `DELETE` | `/api/v1/users/:id` | Delete user (cascades to profile & scenarios) |

### Financial Profile

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/v1/users/:id/profile` | Get financial profile (Redis-cached) |
| `PUT` | `/api/v1/users/:id/profile` | Create or update profile → emits Kafka event |

### Decision Engine — Simulations

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/v1/users/:id/simulate/compound-interest` | Compound interest projection |
| `POST` | `/api/v1/users/:id/simulate/debt-payoff` | Debt payoff timeline |
| `GET` | `/api/v1/users/:id/scenarios?limit=20&offset=0` | Paginated simulation history |

## API Examples

### Create a user

```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alex@example.com",
    "username": "alex_z",
    "full_name": "Alex Zhang",
    "age_group": "18-24"
  }'
```

### Update financial profile

```bash
curl -X PUT http://localhost:8080/api/v1/users/{USER_ID}/profile \
  -H "Content-Type: application/json" \
  -d '{
    "monthly_income": 3500,
    "monthly_expenses": 2200,
    "total_debt": 8000,
    "debt_apr": 22.0,
    "savings_balance": 1500,
    "credit_score": 680
  }'
```

### Simulate compound interest

```bash
curl -X POST http://localhost:8080/api/v1/users/{USER_ID}/simulate/compound-interest \
  -H "Content-Type: application/json" \
  -d '{
    "principal": 1000,
    "monthly_contribution": 200,
    "annual_rate_percent": 7,
    "years": 10
  }'
# → {"future_value":36610.28,"total_contributed":25000,"total_interest":11610.28}
```

### Simulate debt payoff

```bash
curl -X POST http://localhost:8080/api/v1/users/{USER_ID}/simulate/debt-payoff \
  -H "Content-Type: application/json" \
  -d '{
    "total_debt": 8000,
    "annual_apr": 22,
    "monthly_payment": 300
  }'
# → {"months_to_payoff":34,"total_paid":10103.45,"total_interest":2103.45,"is_feasible":true}
```

## Decision Engine

The simulator service (`internal/service/simulator_service.go`) is the core of FinBud. It implements two models:

### Compound Interest

Uses the future-value formula with monthly compounding:

```
FV = P × (1 + r)^n + PMT × [((1 + r)^n − 1) / r]
```

Where `P` = principal, `PMT` = monthly contribution, `r` = monthly rate, `n` = total months.

### Debt Payoff

Runs a month-by-month amortisation loop:

```
Each month:
  interest   = balance × (APR / 12 / 100)
  principal  = payment − interest
  balance   -= principal
```

If the monthly payment doesn't cover the first month's interest, the debt is flagged as **infeasible**.

Both simulations are **cached in Redis** (10 min TTL) and **persisted** as `DecisionScenario` records for the user's history.

## Event-Driven Architecture

When a user updates their profile (`PUT /api/v1/users/:id/profile`), three things happen:

1. **PostgreSQL** — profile is upserted via `ON CONFLICT`
2. **Redis** — cached profile for that user is invalidated
3. **Kafka** — an `EVENT_PROFILE_UPDATED` message is published (async, non-blocking)

Downstream consumers (analytics, notifications, ML pipelines) can subscribe to this topic independently.

## Configuration

All configuration is driven by environment variables. See `env.example` for the full list:

| Variable | Default | Description |
|---|---|---|
| `SERVER_PORT` | `8080` | HTTP listen port |
| `GIN_MODE` | `debug` | Gin mode (`debug` / `release`) |
| `POSTGRES_HOST` | `localhost` | PostgreSQL host |
| `POSTGRES_PORT` | `5432` | PostgreSQL port |
| `POSTGRES_USER` | `finbud` | Database user |
| `POSTGRES_PASSWORD` | `finbud_secret` | Database password |
| `POSTGRES_DB` | `finbud_db` | Database name |
| `REDIS_ADDR` | `localhost:6379` | Redis address |
| `REDIS_DB` | `0` | Redis database number |
| `KAFKA_BROKERS` | `localhost:9092` | Kafka broker addresses |
| `KAFKA_TOPIC_PROFILE_UPDATED` | `EVENT_PROFILE_UPDATED` | Kafka topic name |

## Makefile Commands

```bash
make help        # Show all available commands
make tidy        # go mod tidy
make build       # Build binary to bin/finbud-api
make run         # Run the API server
make consumer    # Run Kafka ingestion consumer
make telemetry   # Run realtime telemetry service (SSE + snapshot API)
make telemetry-publisher # Publish sample telemetry metrics to Redis
make dashboard   # Run Vue telemetry dashboard on :5173
make test        # Run all tests with race detector
make lint        # Run golangci-lint
make infra-up    # Start Docker infrastructure
make infra-down  # Stop and remove Docker infrastructure
make metrics-up  # Start Prometheus and Grafana
make metrics-down # Stop Prometheus and Grafana
make bench       # Run k6 ingestion benchmark
```

## Telemetry Service + Vue Dashboard

The repository includes a realtime telemetry stack:

- Backend service: `cmd/telemetry` + `internal/telemetry`
- Frontend dashboard: `telemetry-dashboard` (Vue 3 + TypeScript + Pinia + Chart.js)
- Stream protocol: **SSE** (`/telemetry/stream`) for simple one-way realtime updates with native browser reconnect.

### Why SSE (instead of WebSocket)

- Native browser auto-reconnect behavior for server->client metrics stream.
- Lower complexity for read-only telemetry feeds (no bidirectional requirement).
- Easy to debug with plain HTTP and proxy support.

### Telemetry architecture

1. Ingestion:
   - Kafka consumer reads `KAFKA_TOPIC_TELEMETRY` (default `EVENT_PROFILE_UPDATED`).
   - Optional Redis Pub/Sub reads `TELEMETRY_REDIS_CHANNEL` (default `telemetry.metrics`).
2. Aggregation:
   - Normalizes payload into `timestamp/source/metric_name/value/tags`.
   - Computes rolling windows `1m/5m/15m`: throughput, error_rate, latency p50/p95/p99.
   - Stores latest snapshot into Redis key `telemetry:snapshot`.
3. Transport:
   - REST snapshot: `GET /telemetry/snapshot`
   - SSE stream: `GET /telemetry/stream`
   - Health/readiness: `GET /health`, `GET /ready`
   - Prometheus metrics: `:2114/metrics`

### Data contract (schema_version = v1)

Envelope:

```json
{
  "schema_version": "v1",
  "event_type": "snapshot | delta_update | heartbeat | error",
  "generated_at": "2026-04-13T10:00:00Z",
  "payload": {}
}
```

Snapshot / delta payload (shape):

```json
{
  "schema_version": "v1",
  "generated_at": "2026-04-13T10:00:00Z",
  "windows": {
    "1m": {
      "event_count": 43,
      "error_count": 3,
      "throughput": 0.7167,
      "error_rate": 6.9767,
      "latency_ms": { "p50": 90, "p95": 180, "p99": 240 },
      "window_label": "1m"
    }
  },
  "consumer_lag": 4,
  "recent_points": [
    {
      "timestamp": "2026-04-13T10:00:00Z",
      "throughput": 0.7167,
      "error_rate": 6.9767,
      "latency_p95": 180,
      "consumer_lag": 4
    }
  ]
}
```

### Run end-to-end locally

1) Start infra:

```bash
make infra-up
```

2) Run telemetry backend:

```bash
make telemetry
```

3) (Optional) Generate sample realtime metrics via Redis:

```bash
make telemetry-publisher
```

4) Run dashboard:

```bash
make dashboard
```

Then open [http://localhost:5173](http://localhost:5173).

### Expected dashboard output

- Connection badge switches among `connected/reconnecting/disconnected`.
- KPI cards show current:
  - throughput (1m)
  - error rate (1m)
  - latency p95 (1m)
  - consumer lag
- Charts update continuously for throughput, error rate, latency p95, and lag.
- If stream disconnects, dashboard falls back to polling snapshot every 5s.

### Tests

Backend:

```bash
go test ./internal/telemetry/...
```

Frontend:

```bash
make dashboard-test
```

Integration (telemetry endpoints):

```bash
make test-integration
```

### Observability stack

- Prometheus scrapes:
  - API: `:8080/metrics`
  - Consumer: `:2113/metrics`
  - Telemetry: `:2114/metrics`
- Grafana dashboards (provisioned from `monitoring/grafana/dashboards`):
  - `FinBud Ingestion Overview`
  - `FinBud Platform Observability`

Start metrics stack:

```bash
make metrics-up
```

### CI (GitHub Actions)

Workflow: `.github/workflows/ci.yml`

Jobs:

- `Go Lint and Test` (golangci-lint + `go test -race`)
- `Frontend Test and Build` (`npm ci`, test, build)
- `Telemetry Integration Tests` (health/ready/snapshot/stream checks)

Run equivalent checks locally:

```bash
make test-ci
```

### Debugging CI failures

- Go test/lint failure:
  - run `go test -race -count=1 ./...`
  - run `golangci-lint run ./...`
- Frontend failure:
  - `cd telemetry-dashboard && npm ci && npm run test && npm run build`
- Integration failure:
  - run `make test-integration`
  - inspect telemetry endpoint tests in `internal/telemetry/transport_test.go`

### Troubleshooting

- Empty dashboard:
  - check `GET http://localhost:8081/telemetry/snapshot`
  - run `make telemetry-publisher` to emit sample metrics
- Stream disconnect loops:
  - verify telemetry backend on `:8081`
  - check browser devtools EventSource errors
- No Kafka metrics:
  - confirm topic exists and telemetry service can reach `KAFKA_BROKERS`

## License

MIT

## Analytics ML Service (Python)

This repository now includes a standalone Python service at `analytics_ml_service/` for time-series forecasting and error-improvement tracking.

### Assumed Input Schema

CSV columns:

- `target` (required, numeric)
- `timestamp` (optional, parseable datetime)

If `timestamp` is missing, the service keeps row order and still computes lag/rolling features leakage-safely.

### What the service does

- Load data -> clean -> generate features (lags, rolling mean/std, calendar)
- Train `XGBoostRegressor`
- Run walk-forward backtesting across multiple time folds
- Compare and log MAPE before/after:
  - `baseline_mape` (naive: `y_t = y_(t-1)`)
  - `xgboost_mape`
  - `delta_mape_abs`
  - `delta_mape_pct`
- Save model artifact + metadata + logs (CSV/JSON)

### Run in one command

```bash
python -m analytics_ml_service.run_service --generate-sample
```

Optional custom file:

```bash
python -m analytics_ml_service.run_service --input path/to/your.csv
```

### Install Python dependencies

```bash
pip install -r analytics_ml_service/requirements.txt
```

### Unit tests

```bash
python -m unittest discover -s analytics_ml_service/tests
```

### Example console output

```text
[INFO] Walk-forward MAPE (before/after):
  baseline_mape  : 2.9831
  xgboost_mape   : 1.9447
  delta_mape_abs : 1.0384
  delta_mape_pct : 34.81%
```

