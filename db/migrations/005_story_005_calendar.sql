CREATE TABLE IF NOT EXISTS calendar_connections (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider TEXT NOT NULL,
    calendar_id TEXT NOT NULL DEFAULT 'primary',
    access_token TEXT NOT NULL,
    refresh_token TEXT,
    token_type TEXT NOT NULL DEFAULT 'Bearer',
    expiry TEXT,
    scope TEXT NOT NULL DEFAULT '',
    timezone TEXT NOT NULL DEFAULT 'UTC',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    CHECK (provider IN ('google')),
    UNIQUE(user_id, provider)
);

CREATE INDEX IF NOT EXISTS idx_calendar_connections_user_provider ON calendar_connections(user_id, provider);

CREATE TABLE IF NOT EXISTS calendar_blocks (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider TEXT NOT NULL,
    external_event_id TEXT NOT NULL,
    local_date TEXT NOT NULL,
    timezone TEXT NOT NULL,
    title TEXT NOT NULL DEFAULT '',
    detail TEXT NOT NULL DEFAULT '',
    start_at TEXT,
    end_at TEXT,
    is_all_day INTEGER NOT NULL DEFAULT 0,
    classification TEXT NOT NULL,
    window TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    CHECK (provider IN ('google')),
    CHECK (classification IN ('small', 'medium', 'large')),
    CHECK (window IN ('none', 'morning', 'afternoon', 'evening', 'all-day')),
    UNIQUE(user_id, provider, external_event_id, local_date)
);

CREATE INDEX IF NOT EXISTS idx_calendar_blocks_user_date ON calendar_blocks(user_id, local_date);
