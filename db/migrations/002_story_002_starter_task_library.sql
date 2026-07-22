CREATE TABLE IF NOT EXISTS starter_task_templates (
    id TEXT PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    duration_minutes INTEGER NOT NULL,
    cadence_type TEXT NOT NULL,
    cadence_value INTEGER NOT NULL,
    priority TEXT NOT NULL,
    time_of_day_preference TEXT NOT NULL,
    is_multistep INTEGER NOT NULL DEFAULT 0,
    sort_order INTEGER NOT NULL,
    created_at TEXT NOT NULL,
    CHECK (cadence_type IN ('interval', 'count')),
    CHECK (priority IN ('high', 'medium', 'low')),
    CHECK (time_of_day_preference IN ('any', 'morning', 'afternoon', 'evening'))
);

CREATE TABLE IF NOT EXISTS starter_subtask_templates (
    id TEXT PRIMARY KEY,
    starter_task_template_id TEXT NOT NULL REFERENCES starter_task_templates(id) ON DELETE CASCADE,
    step_order INTEGER NOT NULL,
    name TEXT NOT NULL,
    duration_minutes INTEGER NOT NULL,
    time_of_day_preference TEXT NOT NULL,
    min_gap_after_previous_minutes INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    CHECK (time_of_day_preference IN ('any', 'morning', 'afternoon', 'evening')),
    UNIQUE(starter_task_template_id, step_order)
);

INSERT OR IGNORE INTO starter_task_templates (id, slug, name, description, duration_minutes, cadence_type, cadence_value, priority, time_of_day_preference, is_multistep, sort_order, created_at) VALUES
('starter-laundry', 'laundry', 'Laundry', 'New-parent default laundry routine split across the day.', 25, 'count', 2, 'medium', 'any', 1, 1, '2026-07-07T00:00:00Z'),
('starter-grocery-run', 'grocery-run', 'Grocery run', 'Weekly grocery pickup or store run.', 60, 'interval', 7, 'medium', 'afternoon', 0, 2, '2026-07-07T00:00:00Z'),
('starter-meal-prep', 'meal-prep', 'Meal prep', 'Prepare a few meals or components ahead of time.', 45, 'count', 2, 'medium', 'afternoon', 0, 3, '2026-07-07T00:00:00Z'),
('starter-clean-kitchen', 'clean-kitchen', 'Clean kitchen', 'Reset counters, dishes, and surfaces.', 20, 'interval', 1, 'medium', 'evening', 0, 4, '2026-07-07T00:00:00Z'),
('starter-pediatrician-follow-ups', 'pediatrician-follow-ups', 'Pediatrician follow-ups', 'Calls, forms, and check-ins related to pediatric care.', 15, 'interval', 7, 'high', 'morning', 0, 5, '2026-07-07T00:00:00Z'),
('starter-thank-you-notes', 'thank-you-notes', 'Thank-you notes', 'Short gratitude note batch for helpers and gifts.', 20, 'interval', 14, 'low', 'evening', 0, 6, '2026-07-07T00:00:00Z'),
('starter-bottle-pump-cleaning', 'bottle-pump-cleaning', 'Bottle/pump cleaning', 'Reset bottles and pump parts.', 15, 'interval', 1, 'high', 'evening', 0, 7, '2026-07-07T00:00:00Z');

INSERT OR IGNORE INTO starter_subtask_templates (id, starter_task_template_id, step_order, name, duration_minutes, time_of_day_preference, min_gap_after_previous_minutes, created_at) VALUES
('starter-laundry-wash', 'starter-laundry', 1, 'Wash', 5, 'morning', 0, '2026-07-07T00:00:00Z'),
('starter-laundry-dry', 'starter-laundry', 2, 'Move to dryer', 5, 'afternoon', 45, '2026-07-07T00:00:00Z'),
('starter-laundry-fold', 'starter-laundry', 3, 'Fold', 15, 'evening', 45, '2026-07-07T00:00:00Z');
