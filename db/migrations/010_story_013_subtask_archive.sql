ALTER TABLE subtasks ADD COLUMN archived_at TEXT;
CREATE INDEX IF NOT EXISTS idx_subtasks_task_archive ON subtasks(task_id, archived_at, step_order);
