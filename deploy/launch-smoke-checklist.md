# Launch smoke checklist

Run these checks before letting testers rely on the environment.

## Scheduling
- Confirm `bash scripts/run-daily-schedule.sh` completes without errors.
- Verify new schedule checkpoints appear for today in the database/logs.

## Telegram
- Confirm the bot can send the daily list via `bash scripts/run-telegram-daily.sh`.
- Confirm a window reminder can send via `bash scripts/run-telegram-window.sh`.
- Click a callback button and verify a `user_response` event is logged.

## Email recap
- Confirm `bash scripts/run-email-recap.sh` writes recap files into `EMAIL_RECAP_OUTBOX_DIR`.
- Verify a matching `channel=email`, `event_type=message_sent`, `message_type=daily_recap` event is present.

## Calendar sync
- Confirm at least one tester can connect Google Calendar.
- Run `bash scripts/run-calendar-sync.sh` and verify calendar blocks exist for today.

## Read-only view
- Generate a lookahead token in a controlled non-production environment when the issuer helper is enabled.
- Open `/lookahead?token=...` and confirm today/tomorrow, blocked-window explanations, and omitted-item reasons render.

## Backups
- Run `bash scripts/run-backup.sh`.
- Confirm a gzipped SQLite backup appears at the configured target or S3 path.

## Reporting
- Run `bash scripts/report-events.sh`.
- Verify the summary includes `message_sent`, `user_response`, and relevant `message_type` counts.
