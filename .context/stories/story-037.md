# Story 037: Add weekday/weekend day preference for tasks

**Status:** in-progress  
**Type:** feature  
**Created:** 2026-08-03  
**Last accessed:** 2026-08-03  
**Completed:** —

---

## Goal
Let users say which days a task belongs on — **Any day is fine** (default), **Weekdays only**, or **Weekends only** — both during onboarding and in post-onboarding task management. The preference is stored as a new `day_preference` field on tasks (mirroring the existing `TimeOfDayPreference` pattern), and the scheduler treats it as a hard day-candidate filter: weekday tasks only plan Monday–Friday, weekend tasks only Saturday–Sunday, computed in the user's timezone. Weekend tasks are planned per week with a count cadence capped at 2 (one per weekend day); when a user picks weekends in the editor, the cadence switches to weekly-count automatically with an explanatory note. Tasks that cannot fit their allowed days overflow honestly rather than spilling onto forbidden days.

## Verification
A user creating or editing a task during onboarding sees a radio-card day picker with plain-language hints; the same field appears on the task-management page as a compact segmented control. Picking "Weekends only" switches cadence to "a few times each week" capped at 2 and shows the explanatory note. Generated schedules place weekday tasks only on Mon–Fri and weekend tasks only on Sat–Sun in the user's timezone, with honest overflow when the allowed days' budget is full. Existing tasks default to "any" and behave exactly as before. Automated tests cover service validation, scheduler filtering and overflow, API payloads, the new picker primitive, TaskEditor behavior, and both pages; the user manually verifies both surfaces at desktop and narrow widths.

## Scope — files this story may touch
- db/migrations/013_story_037_day_preference.sql (new)
- internal/tasks/types.go
- internal/tasks/repository.go
- internal/tasks/service.go
- internal/tasks/service_test.go
- internal/scheduler/service.go
- internal/scheduler/service_test.go
- cmd/server/onboarding.go
- cmd/server/onboarding_test.go
- cmd/server/tasks.go
- cmd/server/tasks_test.go
- web/src/lib/api/onboarding.ts
- web/src/lib/api/tasks.ts
- web/src/lib/api/tasks.test.ts
- web/src/lib/components/design/DayPreferencePicker.svelte (new)
- web/src/lib/components/design/DayPreferencePicker.test.ts (new)
- web/src/lib/components/TaskEditor.svelte
- web/src/lib/components/TaskEditor.test.ts
- web/src/routes/onboarding/tasks/+page.svelte
- web/src/routes/onboarding/tasks/page.test.ts
- web/src/routes/tasks/+page.svelte
- web/src/routes/tasks/page.test.ts
- .context/design/components.md (component registry entry)

## Out of scope — do not touch
- Day-preference tags or badges on task-list tiles (possible follow-up story)
- Per-user weekend definitions; weekend is fixed as Saturday/Sunday for v1
- Telegram messages, bot commands, or notification content
- Scheduler priority, cadence-generation, or budget logic beyond the day-candidate filter
- Starter template library content beyond carrying the new field's default
- Global stylesheet or unrelated pages

## Dependencies
- Story 003
- Story 013
- Story 024
- Story 028

---

## Checklist
- [x] Add the `day_preference` migration (default `any`) and thread the field through task and starter-template types, repositories, and services, with server-side validation that weekend tasks use weekly-count cadence of at most 2.
- [x] Filter scheduler day candidates by day preference (weekday = Mon–Fri, weekend = Sat–Sun) anchored in the user's timezone, keep overflow honest, and cover filtering and overflow with scheduler tests.
- [x] Expose `day_preference` in onboarding and task-management API create/update/read payloads with backend handler tests.
- [x] Create the `DayPreferencePicker` design primitive with `cards` and `segmented` variants, render/state tests, and a component registry entry.
- [x] Integrate the picker into TaskEditor with the weekend cadence auto-switch, the cap-at-2 rule, and the explanatory note; update TaskEditor tests.
- [x] Wire the onboarding tasks page (cards variant) and task-management page (segmented variant), mapping the field in drafts and updating both page-level tests.
- [ ] Manually verify both pages at desktop and narrow widths and confirm weekday/weekend scheduling and overflow behavior end to end.

---

## Issues

- The Go toolchain (`go`/`gofmt`) is not installed or available in this environment, so backend formatting and Go test execution could not be performed. Backend changes were reviewed statically; Go verification remains required before closeout.
- Browser verification at desktop and narrow widths remains pending; a live preview is available at `http://localhost:5200/day-preference-preview` and `http://192.168.86.232:5200/day-preference-preview` for user confirmation. The preview uses the real TaskEditor and DayPreferencePicker components and is a temporary, explicitly approved preview route outside the original story scope.
- The only remaining checklist blocker is verification, not implementation: the Go toolchain is unavailable for backend test execution, and the user must confirm the live UI and end-to-end weekday/weekend scheduling behavior.

---

## Completion Summary

Implemented the weekday/weekend task preference across persistence, task types, starter-template transport, onboarding and task-management APIs, scheduler day filtering, and overflow rollover. Added the reusable `DayPreferencePicker` with guided radio-card and compact segmented variants, integrated it into the shared TaskEditor, and made weekend selection switch to weekly count cadence capped at two with an explanatory note. Added frontend component, editor, API, page, backend validation, and scheduler scenario coverage. `npm run check` passes with zero diagnostics and the web suite passes 24 files / 87 tests. The implementation is otherwise ready; the final checklist item remains unchecked pending user confirmation of the live preview at desktop and narrow widths, plus backend Go test execution when the toolchain is available.
