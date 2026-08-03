# Story 034: Restore clickable Telegram `/edit` management links

**Status:** in-progress  
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
- [x] Reproduce the current `/edit` response with a linked test chat and identify why Telegram falls back from the inline button.
- [x] Ensure the generated management URL uses an explicit configured reachable IP address and port for local/LAN deployments, or the configured public host in production, and is valid for Telegram's URL button requirements.
- [x] Keep the inline **Manage my routines** button as the primary response and retain a safe plain-text fallback only when button delivery fails.
- [x] Add or update tests for configured host handling, clickable markup delivery, fallback behavior, and single-use access grants.
- [x] Update deployment examples/runbook with the local/LAN IP-address-and-port or public-host requirement for clickable `/edit` links.
- [x] Manually verify `/edit` from a linked Telegram chat using a reachable host.

---

## Issues

- **Initial no-response report was caused by the service not running.** There was no `rahat` container, so Telegram long polling was inactive. The current container is now running with the real bot token, `TELEGRAM_LINK_HOST=192.168.86.232:8080`, and the existing local database; logs confirm `/edit` updates were received and inline link responses were issued.
- **Login initially failed because the server origin did not match the LAN link host.** The link used `192.168.86.232:8080`, but `WEB_ORIGIN` was `http://localhost:8080`, causing the phone's LAN `Origin` to receive HTTP 403. The container was restarted with both `WEB_ORIGIN=http://192.168.86.232:8080` and `TELEGRAM_LINK_HOST=192.168.86.232:8080`; an access-link exchange then returned HTTP 200 for the existing user/database.
- **Live Telegram verification completed.** The user confirmed that the link is correct and successfully opened the Manage routines screen from Telegram using the reachable LAN host. No Story 034 blockers remain.

---

## Completion Summary

Implemented explicit reachable-host handling for Telegram `/edit` links. Local/loopback `WEB_ORIGIN` values now require `TELEGRAM_LINK_HOST` with a non-loopback IP address and port, such as `192.168.1.20:8080`; automatic interface selection and `localhost` links are no longer used. Public HTTP(S) origins continue to work without the override. The inline **Manage my routines** button remains the primary response, with the existing plain-text fallback retained only when Telegram rejects button delivery. Added tests covering LAN IP URLs, public origins, missing/invalid local hosts, inline markup, fallback delivery, access-grant exchange, and single-use behavior. Updated deployment documentation and `.env.example` with the phone-reachable IP/port requirement.

The Telegram package tests and full Go test suite pass in the Go 1.25 Docker toolchain. A live linked Telegram chat was also verified: the `/edit` link used the configured reachable LAN host and opened the Manage routines screen successfully. The story is ready for Vazir closeout; it remains in-progress until that workflow runs.
