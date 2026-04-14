# FinBud Backend — Makefile
# Common development commands.

.PHONY: help run consumer telemetry telemetry-publisher dashboard dashboard-build dashboard-test test-integration test-ci build test lint infra-up infra-down metrics-up metrics-down bench tidy

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

tidy: ## Download and tidy Go modules
	go mod tidy

build: ## Build the binary
	go build -o bin/finbud-api ./cmd/api

run: ## Run the API server locally
	go run ./cmd/api

consumer: ## Run the Kafka ingestion consumer
	go run ./cmd/consumer

telemetry: ## Run telemetry streaming service
	go run ./cmd/telemetry

telemetry-publisher: ## Publish sample telemetry metrics to Redis pub/sub
	go run ./cmd/telemetry-publisher

dashboard: ## Run Vue telemetry dashboard
	cd telemetry-dashboard && npm install && npm run dev

dashboard-build: ## Build Vue telemetry dashboard
	cd telemetry-dashboard && npm install && npm run build

dashboard-test: ## Run Vue telemetry dashboard tests
	cd telemetry-dashboard && npm install && npm run test

test-integration: ## Run backend integration tests (telemetry endpoints)
	go test -count=1 -run 'TestSnapshotEndpoint|TestHealthAndReadinessEndpoints|TestStreamEndpointSendsSnapshotEvent' ./internal/telemetry

test-ci: ## Run CI-equivalent local checks
	go test -race -count=1 ./...
	cd telemetry-dashboard && npm install && npm run test && npm run build

test: ## Run all tests
	go test -race -count=1 ./...

lint: ## Run golangci-lint
	golangci-lint run ./...

infra-up: ## Start PostgreSQL, Redis and Kafka via Docker
	docker-compose up -d

infra-down: ## Stop infrastructure containers
	docker-compose down -v

metrics-up: ## Start Prometheus and Grafana
	docker-compose up -d prometheus grafana

metrics-down: ## Stop Prometheus and Grafana
	docker-compose stop prometheus grafana

bench: ## Run k6 ingestion benchmark
	k6 run ./benchmarks/k6/ingestion.js

