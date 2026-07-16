CREATE TABLE IF NOT EXISTS oauth_states (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider TEXT NOT NULL,
    state_token TEXT NOT NULL UNIQUE,
    redirect_uri TEXT NOT NULL DEFAULT '',
    expires_at TEXT NOT NULL,
    created_at TEXT NOT NULL,
    consumed_at TEXT,
    CHECK (provider IN ('google'))
);

CREATE INDEX IF NOT EXISTS idx_oauth_states_provider_token ON oauth_states(provider, state_token);
CREATE INDEX IF NOT EXISTS idx_oauth_states_user_provider ON oauth_states(user_id, provider, created_at DESC);
