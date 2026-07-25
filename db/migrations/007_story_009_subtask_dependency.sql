ALTER TABLE subtasks ADD COLUMN dependency_type TEXT NOT NULL DEFAULT 'required_same_day' CHECK (dependency_type IN ('required_same_day', 'soft_followup'));

ALTER TABLE starter_subtask_templates ADD COLUMN dependency_type TEXT NOT NULL DEFAULT 'required_same_day' CHECK (dependency_type IN ('required_same_day', 'soft_followup'));

UPDATE starter_subtask_templates
SET dependency_type = 'soft_followup'
WHERE id = 'starter-laundry-fold';
