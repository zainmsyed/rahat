CREATE TABLE IF NOT EXISTS onboarding_telegram_confirmations (
    user_id TEXT PRIMARY KEY,
    delivered INTEGER NOT NULL,
    failed_reason TEXT,
    sent_at TEXT NOT NULL
);
