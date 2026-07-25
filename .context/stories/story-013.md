# Story 013: Add post-onboarding task management

**Status:** not-started  
**Type:** feature  
**Created:** 2026-07-25  
**Last accessed:** 2026-07-25  
**Completed:** —

---

## Goal
Let an authenticated tester maintain their routines after onboarding through a focused task-management page: view tasks, add new tasks, edit existing tasks and subtasks, pause or resume routines, and safely remove routines that are no longer relevant.

## Verification
A signed-in tester can open the task page, create and edit a routine using the existing guided editor, pause/resume it, and remove it without affecting another user or erasing completed history. The next schedule preview reflects active task changes while the normal day-to-day interaction remains in Telegram and the passive lookahead page.

## Scope — files this story may touch
- db/migrations/
- internal/tasks/
- internal/occurrences/
- internal/scheduler/
- cmd/server/
- web/src/lib/api/
- web/src/lib/components/TaskEditor.svelte
- web/src/lib/components/tasks/
- web/src/routes/tasks/
- web/src/routes/+page.svelte
- README.md

## Out of scope — do not touch
- A full interactive scheduling dashboard
- Direct editing of generated occurrences or completed history
- Drag-and-drop schedule placement
- Exposing required-chain versus soft-follow-up dependency semantics in the user-facing UI
- Household sharing, assignment, or roles
- Profile, Telegram, calendar, or email settings management

## Dependencies
- Story 003
- Story 006
- Story 009
- Story 012

---

## Checklist
- [ ] Add authenticated task APIs for list, create, update, pause/resume, and safe removal, deriving ownership only from the authenticated session
- [ ] Introduce archive/removal semantics that preserve completed occurrence and event history instead of cascading historical records away
- [ ] Ensure archived and paused tasks are handled correctly by due generation, previews, and future schedule runs
- [ ] Reuse and, where needed, extract the onboarding `TaskEditor` and task mapping logic rather than duplicating forms and validation
- [ ] Build a calm `/tasks` page with active, paused, and removed-state feedback plus clear add/edit actions
- [ ] Keep subtask dependency metadata internal while preserving it correctly through edits
- [ ] Add confirmation and explanatory copy for pause, resume, and remove actions
- [ ] Ensure task changes do not rewrite completed history and document how pending/future scheduling responds to edits
- [ ] Add backend tests for ownership isolation, validation, archive/history preservation, pause/resume, and scheduling effects
- [ ] Add frontend tests for listing, create/edit, pause/resume, remove confirmation, authentication failure, and main error states

---

## Issues

---

## Completion Summary
