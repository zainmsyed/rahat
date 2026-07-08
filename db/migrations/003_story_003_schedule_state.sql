CREATE TABLE IF NOT EXISTS schedule_checkpoints (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    schedule_date TEXT NOT NULL,
    next_checkpoint_at TEXT,
    scheduled_occurrence_count INTEGER NOT NULL DEFAULT 0,
    generated_at TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    UNIQUE(user_id, schedule_date)
);

CREATE INDEX IF NOT EXISTS idx_schedule_checkpoints_user_date ON schedule_checkpoints(user_id, schedule_date);
