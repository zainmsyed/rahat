-- Keep one open occurrence for each logical task/subtask occurrence.
-- Terminal history remains append-only and is intentionally excluded.
CREATE TEMP TABLE occurrence_idempotency_duplicates AS
SELECT
    id AS duplicate_id,
    FIRST_VALUE(id) OVER (
        PARTITION BY user_id, task_id, COALESCE(subtask_id, ''), original_scheduled_for_date
        ORDER BY
            CASE WHEN status = 'scheduled' THEN 0 ELSE 1 END,
            updated_at DESC,
            id
    ) AS survivor_id
FROM occurrences
WHERE status IN ('pending', 'scheduled');

-- Keep event history attached to the surviving occurrence before removing duplicates.
UPDATE event_logs
SET occurrence_id = (
    SELECT survivor_id
    FROM occurrence_idempotency_duplicates
    WHERE duplicate_id = event_logs.occurrence_id
)
WHERE occurrence_id IN (
    SELECT duplicate_id
    FROM occurrence_idempotency_duplicates
    WHERE duplicate_id != survivor_id
);

DELETE FROM occurrences
WHERE id IN (
    SELECT duplicate_id
    FROM occurrence_idempotency_duplicates
    WHERE duplicate_id != survivor_id
);

DROP TABLE occurrence_idempotency_duplicates;

CREATE UNIQUE INDEX IF NOT EXISTS idx_occurrences_open_identity
ON occurrences (user_id, task_id, COALESCE(subtask_id, ''), original_scheduled_for_date)
WHERE status IN ('pending', 'scheduled');
