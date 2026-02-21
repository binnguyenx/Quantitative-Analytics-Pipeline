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
| Architecture | Clean Architecture / Hexagonal |

## Project Structure

```
FINBUD/
├── cmd/api/main.go                  # Entry point — DI wiring & graceful shutdown
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
├── migrations/001_init_schema.sql   # Reference DDL (also auto-migrated by GORM)
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

### 5. Verify

```bash
curl http://localhost:8080/health
# → {"service":"finbud-backend","status":"ok"}
```

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
make test        # Run all tests with race detector
make lint        # Run golangci-lint
make infra-up    # Start Docker infrastructure
make infra-down  # Stop and remove Docker infrastructure
```

## License

MIT

