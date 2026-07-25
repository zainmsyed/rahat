CREATE TABLE IF NOT EXISTS beta_access_grants (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    selector TEXT NOT NULL UNIQUE,
    token_hash TEXT NOT NULL,
    expires_at TEXT NOT NULL,
    used_at TEXT,
    created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_beta_access_grants_user_id ON beta_access_grants(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_beta_access_grants_expires_at ON beta_access_grants(expires_at);

CREATE TABLE IF NOT EXISTS web_sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    selector TEXT NOT NULL UNIQUE,
    token_hash TEXT NOT NULL,
    expires_at TEXT NOT NULL,
    revoked_at TEXT,
    created_at TEXT NOT NULL,
    last_seen_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_web_sessions_user_id ON web_sessions(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_web_sessions_expires_at ON web_sessions(expires_at);
