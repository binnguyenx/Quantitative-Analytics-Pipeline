# Benchmark Report — Kafka Ingestion Throughput/Latency

## Environment
- Date:
- OS:
- CPU / RAM:
- Go version:
- Docker version:
- Kafka / PostgreSQL / Redis versions:

## Test Configuration
- API endpoint: `PUT /api/v1/users/:id/profile`
- k6 script: `benchmarks/k6/ingestion.js`
- Warm-up:
- Steady-state:
- Ramp-down:
- `KAFKA_CONSUMER_BATCH_SIZE`:
- `KAFKA_CONSUMER_BATCH_WAIT_MS`:
- `KAFKA_CONSUMER_MAX_RETRIES`:

## API Load Results (k6)
- RPS:
- Error rate:
- p50 latency:
- p95 latency:
- p99 latency:

## Ingestion Pipeline Results
- Consumer throughput (events/sec):
- `ingest_events_total{status="success"}` rate:
- `ingest_events_total{status="failed"}` rate:
- `ingest_events_total{status="dlq"}` rate:
- `ingest_processing_latency_ms` p95:
- `ingest_processing_latency_ms` p99:
- Consumer lag max/avg:

## End-to-End Latency
- Method: compare event timestamp vs `processed_at` in `ingested_events`.
- SQL sample:
  - `SELECT percentile_cont(0.95) WITHIN GROUP (ORDER BY EXTRACT(EPOCH FROM (processed_at - event_timestamp))*1000) AS p95_ms FROM ingested_events;`
- p50:
- p95:
- p99:

## Bottlenecks Observed
- 

## Tuning Actions Taken
- 

## Recommendations
- 

## TODO
- Add distributed tracing for producer->consumer path.
- Add separate benchmark profile for peak bursts (short high rate).
- Add autoscaling tests with multiple consumer replicas.
