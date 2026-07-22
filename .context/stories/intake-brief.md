# Intake Brief

**Last updated:** 2026-07-21

## Planning brief
I would like to update the onboarding flow story and mayb ebreak it up into 2 one for telegram and one for gmail. The issue is for beta testers they are moms new moms that are not tech savvy and we need to really hand hold them through this process. right now it is not doing that

## Distilled answers (replan 2026-07-21)
- Split onboarding into three stories, not two: **core onboarding** (profile, timezone, budget, optional email, tasks, finish), **Telegram onboarding** (guided connection), and **Google Calendar onboarding** (guided connection). "gmail" in the original brief meant the Google Calendar step, not the email channel.
- Story 006 is retired as superseded: its implementation proved the plumbing but failed the UX bar for the target audience, and the code was rolled back to the pre-Story-006 checkout before this replan.
- Target tester is a non-technical new mother with near-zero time: onboarding must hand-hold step by step, state required vs optional explicitly, and never expose internal IDs, infra state, or operator setup (e.g. BotFather is operator docs, not tester UX).
- Telegram must be connectable by sending a short code to the bot via a one-tap deep link, with automatic on-screen confirmation — no raw chat ID entry. Email remains a simple optional field inside core onboarding.
- Google Calendar stays optional and read-only, must fail gracefully when unconfigured, and needs a clean connect → redirect → return loop.
- Telegram model for v1: one shared bot operated by the project owner (no BotFather in tester UX); every tester gets a private 1:1 chat with the shared bot and a per-tester link code; testers cannot see each other's messages, which is acceptable for the trusted beta group.
- Final numbering after renumber: 006 (guided core onboarding), 007 (guided Telegram onboarding), 008 (guided Google Calendar onboarding), with lookahead/email/ops shifted to 009/010/011. The retired original Story 006 is preserved at .context/stories/archive/story-006.md.

## Source files
- .context/intake/prd/rahat-prd.md (14465 bytes)

## Distilled notes
### .context/intake/prd/rahat-prd.md
Large file (14465 bytes). Read enough of it to extract evidence for every planning field before asking questions.

## Planning rules
- Treat listed source files as user-authored planning inputs unless they are explicitly marked as generated artifacts.
- Vazir-generated files in .context/stories/ are replan context, not primary intake.
- Read all text-based planning sources before asking questions.
- Ask only implementation-blocking delta questions after reviewing this brief and any raw files you actually need.
- State safe default assumptions briefly so the user can correct them.
- Surface contradictions instead of resolving them silently.
