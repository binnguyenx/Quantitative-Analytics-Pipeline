-- FinBud — Initial Database Schema
-- This file is executed automatically when the PostgreSQL container
-- starts for the first time (mounted into /docker-entrypoint-initdb.d).
--
-- GORM's AutoMigrate also creates these tables at boot, so this SQL
-- serves as a reference / documentation of the canonical schema.

-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ─────────────────────────────────────────────────────────────────
-- USERS
-- ─────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS users (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    email       VARCHAR     NOT NULL UNIQUE,
    username    VARCHAR(50) NOT NULL UNIQUE,
    full_name   VARCHAR(120) NOT NULL,
    age_group   VARCHAR(10),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);

-- ─────────────────────────────────────────────────────────────────
-- FINANCIAL PROFILES  (one-to-one with users)
-- ─────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS financial_profiles (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID        NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    monthly_income   NUMERIC     NOT NULL DEFAULT 0,
    monthly_expenses NUMERIC     NOT NULL DEFAULT 0,
    total_debt       NUMERIC     NOT NULL DEFAULT 0,
    debt_apr         NUMERIC     NOT NULL DEFAULT 0,
    savings_balance  NUMERIC     NOT NULL DEFAULT 0,
    credit_score     INTEGER     DEFAULT 0,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_financial_profiles_user_id ON financial_profiles (user_id);

-- ─────────────────────────────────────────────────────────────────
-- DECISION SCENARIOS  (simulation history)
-- ─────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS decision_scenarios (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type        VARCHAR(30) NOT NULL,
    title       VARCHAR(200) NOT NULL,
    input_json  JSONB       NOT NULL,
    output_json JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_decision_scenarios_user_id ON decision_scenarios (user_id);
CREATE INDEX IF NOT EXISTS idx_decision_scenarios_type    ON decision_scenarios (type);

