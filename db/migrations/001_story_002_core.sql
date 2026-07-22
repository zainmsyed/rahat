CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    display_name TEXT NOT NULL,
    timezone TEXT NOT NULL,
    daily_time_budget_minutes INTEGER NOT NULL,
    telegram_chat_id TEXT,
    email TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS tasks (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    duration_minutes INTEGER NOT NULL,
    cadence_type TEXT NOT NULL,
    cadence_value INTEGER NOT NULL,
    priority TEXT NOT NULL,
    time_of_day_preference TEXT NOT NULL,
    is_multistep INTEGER NOT NULL DEFAULT 0,
    is_paused INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    CHECK (cadence_type IN ('interval', 'count')),
    CHECK (priority IN ('high', 'medium', 'low')),
    CHECK (time_of_day_preference IN ('any', 'morning', 'afternoon', 'evening'))
);

CREATE INDEX IF NOT EXISTS idx_tasks_user_id ON tasks(user_id);
CREATE INDEX IF NOT EXISTS idx_tasks_user_priority ON tasks(user_id, priority);

CREATE TABLE IF NOT EXISTS subtasks (
    id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    step_order INTEGER NOT NULL,
    name TEXT NOT NULL,
    duration_minutes INTEGER NOT NULL,
    time_of_day_preference TEXT NOT NULL,
    min_gap_after_previous_minutes INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    CHECK (time_of_day_preference IN ('any', 'morning', 'afternoon', 'evening')),
    UNIQUE(task_id, step_order)
);

CREATE INDEX IF NOT EXISTS idx_subtasks_task_id ON subtasks(task_id, step_order);

CREATE TABLE IF NOT EXISTS occurrences (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    subtask_id TEXT REFERENCES subtasks(id) ON DELETE CASCADE,
    status TEXT NOT NULL,
    scheduled_for_date TEXT NOT NULL,
    original_scheduled_for_date TEXT NOT NULL,
    scheduled_time_of_day TEXT NOT NULL,
    rollover_count INTEGER NOT NULL DEFAULT 0,
    consecutive_no_count INTEGER NOT NULL DEFAULT 0,
    snoozed_until_at TEXT,
    completed_at TEXT,
    skipped_at TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    CHECK (status IN ('pending', 'scheduled', 'completed', 'skipped')),
    CHECK (scheduled_time_of_day IN ('any', 'morning', 'afternoon', 'evening'))
);

CREATE INDEX IF NOT EXISTS idx_occurrences_task_id ON occurrences(task_id, scheduled_for_date);
CREATE INDEX IF NOT EXISTS idx_occurrences_user_status ON occurrences(user_id, status, scheduled_for_date);

CREATE TABLE IF NOT EXISTS channel_preferences (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    channel TEXT NOT NULL,
    enabled INTEGER NOT NULL DEFAULT 1,
    is_primary INTEGER NOT NULL DEFAULT 0,
    supports_interactive INTEGER NOT NULL DEFAULT 0,
    recap_enabled INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    CHECK (channel IN ('telegram', 'email')),
    UNIQUE(user_id, channel)
);

CREATE TABLE IF NOT EXISTS pauses (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    task_id TEXT REFERENCES tasks(id) ON DELETE CASCADE,
    scope TEXT NOT NULL,
    reason TEXT NOT NULL DEFAULT '',
    starts_at TEXT NOT NULL,
    ends_at TEXT NOT NULL,
    created_at TEXT NOT NULL,
    CHECK (scope IN ('global', 'task'))
);

CREATE INDEX IF NOT EXISTS idx_pauses_user_scope ON pauses(user_id, scope, starts_at, ends_at);
CREATE INDEX IF NOT EXISTS idx_pauses_task_id ON pauses(task_id, starts_at, ends_at);

CREATE TABLE IF NOT EXISTS event_logs (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    occurrence_id TEXT REFERENCES occurrences(id) ON DELETE SET NULL,
    channel TEXT NOT NULL,
    event_type TEXT NOT NULL,
    message_type TEXT NOT NULL,
    payload_json TEXT NOT NULL DEFAULT '{}',
    occurred_at TEXT NOT NULL,
    CHECK (channel IN ('telegram', 'email', 'system'))
);

CREATE INDEX IF NOT EXISTS idx_event_logs_user_time ON event_logs(user_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_event_logs_occurrence_id ON event_logs(occurrence_id, occurred_at DESC);
