# Story 034: Restore clickable Telegram `/edit` management links

**Status:** not-started  
**Type:** bug  
**Created:** 2026-08-03  
**Last accessed:** 2026-08-03  
**Completed:** —

---

## Goal
Restore the Telegram `/edit` response as a reliable clickable management action. A linked user should receive an inline button with a valid, reachable login URL. For local/LAN deployments, the URL must contain the explicitly configured reachable IP address and port (for example `192.168.1.20:8080`), not `localhost`, so opening it from a phone reaches the development machine.

The implementation must preserve the single-use, short-lived access-grant security behavior and keep a safe fallback when Telegram rejects an inline keyboard message.

## Verification
A linked Telegram user sends `/edit` and receives a clickable **Manage my routines** button whose URL contains the configured public hostname or local/LAN IP address and port. In local/LAN mode, the link must not use `localhost` or an automatically selected unreachable interface. The button opens the login flow and exchanges only once. Automated tests cover successful inline-button delivery, configured IP/host handling, Telegram send failure fallback, and single-use token behavior.

## Scope — files this story may touch
- internal/notifications/telegram/edit_command.go
- internal/notifications/telegram/edit_command_test.go
- internal/notifications/telegram/client.go (only if required for message markup handling)
- deploy/README.md
- .env.example

## Out of scope — do not touch
- Telegram onboarding linking flow
- Authentication grant storage or expiration policy
- Frontend page design
- Dockerfile or deployment image structure

## Dependencies
- Story 014
- Story 033

---

## Checklist
- [ ] Reproduce the current `/edit` response with a linked test chat and identify why Telegram falls back from the inline button.
- [ ] Ensure the generated management URL uses an explicit configured reachable IP address and port for local/LAN deployments, or the configured public host in production, and is valid for Telegram's URL button requirements.
- [ ] Keep the inline **Manage my routines** button as the primary response and retain a safe plain-text fallback only when button delivery fails.
- [ ] Add or update tests for configured host handling, clickable markup delivery, fallback behavior, and single-use access grants.
- [ ] Update deployment examples/runbook with the local/LAN IP-address-and-port or public-host requirement for clickable `/edit` links.
- [ ] Manually verify `/edit` from a linked Telegram chat using a reachable host.

---

## Issues

---

## Completion Summary
