# System Rules

## Rules
- Follow existing project conventions.
- Write directly to real project files.
- Ask before changing ambiguous areas.
- Commit `.context` changes whenever they are part of the work, unless the user explicitly says not to commit them.
- Treat .git/, .jj/, .fslckout, .fossil-settings/ as protected VCS metadata targets.
- Never delete, reset, clean, reinitialize, or overwrite VCS metadata without explicit user approval for that exact action.
- If Vazir blocks a destructive VCS action, wait for the user to send the exact `VCS_APPROVE <token>` phrase before retrying that same action.

## Learned Rules
### From failures
- When a SQLite pragma is connection-scoped, do not set it only once on a pooled *sql.DB; apply it for every connection or deliberately constrain the pool. <!-- source: story-001 --> <!-- confidence: low — no signal in last 5 stories -->
- Build commands documented for fresh-clone use must be self-sufficient: they should generate required derived config/state and run on a clean install without compatibility or missing-file warnings. <!-- source: story-001 --> <!-- confidence: low — no signal in last 5 stories -->
- When a story adds migrations, wire them into the real app/bootstrap path in the same story; test-only migration execution is not sufficient. <!-- source: story-002 --> <!-- confidence: low — no signal in last 5 stories -->
- Multi-row repository operations that represent one logical write must be transactional so partial state is not persisted on mid-operation failure. <!-- source: story-002 --> <!-- confidence: low — no signal in last 5 stories -->
- When cadence is defined at the parent task level, scheduler due-logic must count task-level completion units rather than raw subtask occurrences. <!-- source: story-003 --> <!-- confidence: low — no signal in last 5 stories -->
- When subtask spacing rules exist, scheduler placement must enforce the actual gap value, not just broad window order. <!-- source: story-003 --> <!-- confidence: low — no signal in last 5 stories -->
- When an integration supports both webhooks and polling, default to the transport that works in local/dev without public infrastructure and enable webhook mode only when deployment settings clearly support it. <!-- source: story-004 --> <!-- confidence: low — no signal in last 5 stories -->
- When runtime config accepts an external webhook URL, validate that the application actually serves the same webhook path or derive the served route from that URL before enabling webhook mode. <!-- source: story-004 --> <!-- confidence: low — no signal in last 5 stories -->
- OAuth account-link flows must verify the returned state against stored initiation context before exchanging the code or binding tokens to a local user. <!-- source: story-005 --> <!-- confidence: high -->
- When a scheduling rule is intended for whole-day constraints, do not trigger it from a single-window event unless the implementation explicitly proves full-day coverage. <!-- source: story-005 --> <!-- confidence: high -->
- Production-bound HTTP handlers must not use wildcard CORS; allowed origins must be configurable and default to the local dev origin rather than '*'. <!-- source: story-006 --> <!-- confidence: high -->
- New backend handlers and service methods must ship with unit or integration tests covering the happy path and the main error paths. <!-- source: story-006 --> <!-- confidence: high -->
- Generated dependency/build/runtime directories and local env files must be covered by committed VCS ignore rules before broad staging or closeout flows run, so package installs and local runs do not flood commits with machine-generated files. <!-- source: story-006 --> <!-- confidence: high -->
- Avoid third-party services for rendering sensitive onboarding credentials; generate QR codes and similar one-user secrets locally. <!-- source: story-007 --> <!-- confidence: high -->
- Every webhook handler must have automated tests for each update type it routes. <!-- source: story-007 --> <!-- confidence: high -->
- Keep the API base URL and similar endpoint configuration in a single exported frontend constant; do not duplicate it across route modules. <!-- source: story-008 --> <!-- confidence: high -->
- OAuth account-link flows should reuse a valid, unconsumed initiation state for the same user/session instead of generating a new state row on every page load. <!-- source: story-008 --> <!-- confidence: high -->
- Multi-day preview endpoints must simulate state transitions between days instead of previewing each day independently from the same persisted base state. <!-- source: story-009 --> <!-- confidence: high -->
- Dev-only token or admin helper endpoints should default off outside explicit local development and require an opt-in flag when enabled. <!-- source: story-009 --> <!-- confidence: high -->
