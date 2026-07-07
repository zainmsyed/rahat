# Intake Brief

**Last updated:** 2026-07-07

## Planning brief
Rahat is a greenfield single-user task-scheduling and follow-through assistant for overwhelmed parents, starting with new mothers and mothers with multiple young children. v1 should prove the scheduling engine and adaptive check-in loop using Telegram for real-time reminders and check-ins, email for overview/recap delivery only, read-only Google Calendar integration, a lightweight SvelteKit onboarding flow, and a passive today/tomorrow web lookahead page.

## Source files
- .context/intake/prd/rahat-prd.md (user-authored PRD)
- Direct user clarifications from this planning conversation

## Distilled notes
### Users
- Primary users: new mothers and mothers with multiple young children
- v1 is single-user only
- Initial testing group is small and trusted: wife, friends, family

### Scope decisions confirmed
- This repo is brand new / greenfield
- Telegram is the interactive v1 channel for batched daily lists, reminders, Done / Not Yet responses, snoozes, reschedules, and pause actions
- Email stays in v1 only as an overview/recap channel, not an interactive task-handling surface
- Keep the lightweight SvelteKit onboarding flow from the PRD
- Keep the read-only today + tomorrow web page from the PRD
- Keep Google Calendar read-only integration
- SMS was considered, then removed from v1 scope

### Explicit v1 exclusions
- SMS / Twilio delivery
- Multi-user household assignment
- Calendar write access
- Push notifications
- Yearly cadence tasks
- Full dashboard-style day-to-day task management
- Machine-learned personalization beyond the PRD rules

### Resolved ambiguity
- An earlier preference for "let the user choose the channel behavior" was superseded by the final scope decision: Telegram handles the real-time loop, while email is recap-only

## Planning rules
- Treat listed source files as user-authored planning inputs unless they are explicitly marked as generated artifacts.
- Vazir-generated files in .context/stories/ are replan context, not primary intake.
- Read all text-based planning sources before asking questions.
- Ask only implementation-blocking delta questions after reviewing this brief and any raw files you actually need.
- State safe default assumptions briefly so the user can correct them.
- Surface contradictions instead of resolving them silently.
