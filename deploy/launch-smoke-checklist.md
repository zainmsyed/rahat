# Launch smoke checklist

Run these checks before letting testers rely on the environment. These commands assume the container is named `rahat`; use the actual Coolify service/container name where necessary.

## Container and web UI
- Confirm the service is built from the repo-root `Dockerfile` and exposes port `8080`.
- Confirm the persistent volume is mounted at `/data` and the container is not relying on ephemeral SQLite storage.
- Confirm `curl -fsS https://<your-domain>/healthz` returns JSON with `"status":"ok"`.
- Confirm `curl -fsS -H 'Accept: text/html' https://<your-domain>/` returns the landing-page HTML.
- Open `https://<your-domain>/` in a browser and confirm the landing page renders.
- Confirm a static asset referenced by the landing page under `/_app/immutable/` returns HTTP 200.
- Confirm the container health status is `healthy` and startup logs show the expected production address.

## Onboarding and sessions
- Open `/onboarding?invite=<configured invite>` and confirm it advances to the profile step.
- Complete a test profile and confirm the browser receives a `rahat_session` cookie after onboarding finishes.
- Confirm the authenticated `/tasks` page loads after onboarding.

## Telegram long polling
- Set `TELEGRAM_BOT_TOKEN` and leave `TELEGRAM_WEBHOOK_SECRET` and `TELEGRAM_WEBHOOK_URL` empty.
- Confirm startup logs report `"transport":"long_polling"`.
- Link a test Telegram chat during onboarding and confirm the bot sends the welcome message.
- Confirm the bot can send the daily list with:
  `docker exec rahat rahat-api ops:run-job telegram-daily`
- Confirm a window reminder can send with:
  `docker exec rahat rahat-api ops:run-job telegram-window`
- Send `/edit` from a linked chat and confirm the one-time management link uses the configured public/local host.
- Click a callback button and verify a `user_response` event is logged.
- Confirm no Telegram webhook configuration is required for this mode.

## Scheduling
- Run `docker exec rahat rahat-api ops:run-job schedule-daily` and verify new schedule checkpoints appear for today in the database/logs.

## Email recap
- Run `docker exec rahat rahat-api ops:run-job email-recap` and confirm recap files are written under `EMAIL_RECAP_OUTBOX_DIR` when configured.
- Verify a matching `channel=email`, `event_type=message_sent`, `message_type=daily_recap` event is present.

## Calendar sync
- Confirm at least one tester can connect Google Calendar when the optional Google variables are configured.
- Run `docker exec rahat rahat-api ops:run-job calendar-sync` and verify calendar blocks exist for today.

## Read-only view
- Generate a lookahead token in a controlled non-production environment when the issuer helper is enabled.
- Open `/lookahead?token=...` and confirm today/tomorrow, blocked-window explanations, and omitted-item reasons render.
- In a controlled non-production environment, request `/lookahead/plan?token=...&days=7` and confirm seven days are returned, recurring tasks stay within their cadence/day preferences, and repeated requests do not change occurrence or checkpoint counts.

## Backups
- Run `docker exec rahat rahat-api ops:run-job backup-daily` when `BACKUP_TARGET_URI` is configured.
- Confirm a WAL-safe SQLite backup appears at the configured target or S3 path and can be restored.

## Reporting
- Run `docker exec rahat rahat-api ops:report-events`.
- Verify the summary includes `message_sent`, `user_response`, and relevant `message_type` counts.
