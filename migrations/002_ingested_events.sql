-- FinBud — Kafka ingested events table

CREATE TABLE IF NOT EXISTS ingested_events (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id         VARCHAR(128) NOT NULL UNIQUE,
    event_type       VARCHAR(64)  NOT NULL,
    user_id          UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    correlation_id   VARCHAR(128) NOT NULL,
    event_timestamp  TIMESTAMPTZ  NOT NULL,
    kafka_key        VARCHAR(128),
    kafka_topic      VARCHAR(128),
    kafka_partition  INTEGER      NOT NULL,
    kafka_offset     BIGINT       NOT NULL,
    payload_json     JSONB        NOT NULL,
    processed_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ingested_events_user_id ON ingested_events (user_id);
CREATE INDEX IF NOT EXISTS idx_ingested_events_event_type ON ingested_events (event_type);
CREATE INDEX IF NOT EXISTS idx_ingested_events_correlation_id ON ingested_events (correlation_id);
CREATE INDEX IF NOT EXISTS idx_ingested_events_kafka_topic ON ingested_events (kafka_topic);
