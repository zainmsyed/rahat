ALTER TABLE tasks ADD COLUMN archived_at TEXT;
CREATE INDEX IF NOT EXISTS idx_tasks_user_archive ON tasks(user_id, archived_at, is_paused);
