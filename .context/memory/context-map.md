# Context Map

- Project: Pi workspace that customizes Vazir for this Fossil checkout (`rahat.fossil`), with runtime context in `.context/` and agent code in `.pi/`.
- Stack: TypeScript Pi extensions, Pi TUI helpers, and Fossil/Git/JJ CLI integration.
- Key dirs: `.pi/extensions`, `.pi/lib`, `.context/{memory,design,reviews,stories,settings}`.
- Fragile: active VCS mode is Fossil; tracker VCS logic is duplicated across `.pi/lib/vazir-vcs-helpers.ts` and `.pi/extensions/vazir-tracker/vcs.ts`; Fossil theme output is generated from `.context/design/*.md`; live reload watches `.pi/extensions`.
