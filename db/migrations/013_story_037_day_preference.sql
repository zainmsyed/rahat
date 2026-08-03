ALTER TABLE tasks ADD COLUMN day_preference TEXT NOT NULL DEFAULT 'any';
ALTER TABLE starter_task_templates ADD COLUMN day_preference TEXT NOT NULL DEFAULT 'any';

CREATE INDEX IF NOT EXISTS idx_tasks_user_day_preference ON tasks(user_id, day_preference);
