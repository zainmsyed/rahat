# Story 008: Build the guided Google Calendar connection onboarding

**Status:** not-started  
**Type:** full-stack  
**Created:** 2026-07-21  
**Last accessed:** 2026-07-21  
**Completed:** —

---

## Goal
Let a beta tester connect Google Calendar during onboarding as an optional, clearly-explained step: one obvious connect action, a safe return into onboarding after the Google redirect, honest status when the server is not configured for Google OAuth, and clean disconnect/reconnect. The tester should always understand that calendar access is read-only, why it helps, and that skipping it is fine.

## Verification
On a server with Google OAuth configured, a tester can connect calendar in one action and land back in onboarding with a visible connected state. On a server without OAuth config, the screen states plainly that calendar is unavailable and optional, and the tester continues without confusion or dead ends.

## Scope — files this story may touch
- web/src/routes/onboarding/
- web/src/lib/components/
- web/src/lib/api/
- internal/calendar/
- cmd/server/

## Out of scope — do not touch
- Telegram connection flow (Story 007)
- Calendar sync and block classification logic (Story 005)
- Calendar write access of any kind
- Non-Google calendar providers

## Dependencies
- Story 005
- Story 006

---

## Checklist
- [ ] Present calendar connect as clearly optional and recommended, with a plain-language explanation of read-only access and why it helps
- [ ] Show calendar as clearly unavailable when server OAuth config is missing and let the tester continue without a dead end
- [ ] Provide one obvious connect action with a clear return path into onboarding after the Google redirect
- [ ] Support disconnecting and reconnecting cleanly, with state reflected accurately on-screen
- [ ] Reflect the final calendar connection state in the onboarding review step

---

## Issues

---

## Completion Summary
