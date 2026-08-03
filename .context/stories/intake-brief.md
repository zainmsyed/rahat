# Intake Brief

**Last updated:** 2026-08-03

## Planning brief
Add a day-of-week preference to tasks so users can say a task belongs on weekdays, weekends, or any day, both during onboarding and in post-onboarding task management. The scheduler must honor the preference as a hard day filter, computed in the user's timezone, with honest overflow when allowed days cannot fit the task.

## Source files
- Interactive planning conversation (2026-08-03); no new intake documents.
- Prior replan context: .context/stories/plan.md and existing story files.

## Distilled notes
### Day-preference decisions (clarified 2026-08-03)
- **Three options, user-facing wording:** "Any day is fine" (default), "Weekdays only", "Weekends only".
- **New field** `day_preference` on tasks (`any` | `weekday` | `weekend`), mirroring the existing `TimeOfDayPreference` enum pattern; existing rows default to `any` with unchanged behavior.
- **Hard scheduler constraint, not soft:** weekday tasks plan Mon–Fri only, weekend tasks Sat–Sun only, in the user's timezone; overflow is reported honestly rather than spilling onto forbidden days.
- **Weekend cadence rule:** weekend tasks use weekly-count cadence capped at 2 (one per weekend day). Picking weekends in the editor auto-switches cadence to "a few times each week" with an explanatory note; editing an existing interval-cadence task to weekend auto-switches rather than failing validation. Interval cadence stays allowed for weekday tasks (the scheduler skips Sat/Sun).
- **UI variants:** onboarding tasks page uses radio cards with plain-language hints; the task-management page uses a compact segmented control. Both come from one new `DayPreferencePicker` design primitive with a `variant` prop, surfaced through the shared TaskEditor.
- **Weekend is fixed as Saturday/Sunday for v1**; per-user weekend definitions are out of scope.
- Day-preference tags on task-list tiles are deferred as a possible follow-up story.

## Planning rules
- Treat listed source files as user-authored planning inputs unless they are explicitly marked as generated artifacts.
- Vazir-generated files in .context/stories/ are replan context, not primary intake.
- Read all text-based planning sources before asking questions.
- Ask only implementation-blocking delta questions after reviewing this brief and any raw files you actually need.
- State safe default assumptions briefly so the user can correct them.
- Surface contradictions instead of resolving them silently.
