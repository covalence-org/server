-- schema.sql

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE request_logs (
    request_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    api_key_id UUID,
    model TEXT NOT NULL,
    target_url TEXT NOT NULL,
    inputs JSONB[] NOT NULL,
    parameters JSONB,
    received_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    client_ip INET,
    archived BOOLEAN DEFAULT FALSE
);

CREATE TABLE response_logs (
    response_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID REFERENCES request_logs(request_id) ON DELETE CASCADE,
    response JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    latency_ms INTEGER
);

CREATE TABLE firewall_events (
    firewall_event_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID REFERENCES request_logs(request_id) ON DELETE CASCADE,
    firewall_id TEXT NOT NULL,
    firewall_type TEXT NOT NULL,
    blocked BOOLEAN DEFAULT FALSE,
    blocked_reason TEXT,
    risk_score NUMERIC(3, 2),
    evaluated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE audit_archives (
    archive_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID REFERENCES request_logs(request_id) ON DELETE CASCADE,
    s3_path TEXT NOT NULL,
    archived_at TIMESTAMPTZ DEFAULT now(),
    archive_hash TEXT
);

-- Indexes
CREATE INDEX idx_request_user ON request_logs(user_id);
CREATE INDEX idx_request_time ON request_logs(received_at);
CREATE INDEX idx_firewall_request ON firewall_events(request_id);
CREATE INDEX idx_response_request ON response_logs(request_id);
