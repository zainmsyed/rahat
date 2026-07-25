-- Story 014: enforce one Telegram chat per user and support self-service /edit links.

-- Preserve any existing duplicate Telegram chat IDs before adding the unique index.
CREATE TABLE IF NOT EXISTS telegram_identity_conflicts (
    id TEXT PRIMARY KEY,
    chat_id TEXT NOT NULL,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    resolved_user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
    detected_at TEXT NOT NULL,
    resolved_at TEXT
);

CREATE INDEX IF NOT EXISTS idx_telegram_identity_conflicts_chat_id ON telegram_identity_conflicts(chat_id);

INSERT INTO telegram_identity_conflicts (id, chat_id, user_id, detected_at)
SELECT lower(hex(randomblob(16))), u1.telegram_chat_id, u1.id, datetime('now')
FROM users u1
WHERE u1.telegram_chat_id IS NOT NULL AND u1.telegram_chat_id != ''
  AND EXISTS (
      SELECT 1 FROM users u2
      WHERE u2.telegram_chat_id = u1.telegram_chat_id
        AND u2.id != u1.id
        AND (
            u2.created_at < u1.created_at
            OR (u2.created_at = u1.created_at AND u2.id < u1.id)
        )
  );

-- Clear the duplicate chat IDs, keeping the earliest-created user linked.
UPDATE users
SET telegram_chat_id = NULL
WHERE id IN (
    SELECT u1.id FROM users u1
    WHERE u1.telegram_chat_id IS NOT NULL AND u1.telegram_chat_id != ''
      AND EXISTS (
          SELECT 1 FROM users u2
          WHERE u2.telegram_chat_id = u1.telegram_chat_id
            AND u2.id != u1.id
            AND (
                u2.created_at < u1.created_at
                OR (u2.created_at = u1.created_at AND u2.id < u1.id)
            )
      )
);

-- Enforce one non-empty Telegram chat ID per user at the database level.
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_telegram_chat_id
ON users(telegram_chat_id)
WHERE telegram_chat_id IS NOT NULL AND telegram_chat_id != '';
