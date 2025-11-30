-- 002_api_keys.sql
-- API key management and rate limiting tables

-- API keys table for authentication and rate limiting
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_hash VARCHAR(64) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    rate_limit INT NOT NULL DEFAULT 100,
    rate_window_seconds INT NOT NULL DEFAULT 60,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_used_at TIMESTAMP WITH TIME ZONE
);

-- Index on key_hash for fast lookups during authentication
CREATE INDEX idx_api_keys_hash ON api_keys(key_hash);

-- Rate limit tracking table using fixed time windows
CREATE TABLE rate_limit_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    api_key_id UUID NOT NULL REFERENCES api_keys(id) ON DELETE CASCADE,
    window_start TIMESTAMP WITH TIME ZONE NOT NULL,
    request_count INT NOT NULL DEFAULT 1,
    UNIQUE(api_key_id, window_start)
);

-- Index for efficient rate limit queries
CREATE INDEX idx_rate_limit_window ON rate_limit_records(api_key_id, window_start);
