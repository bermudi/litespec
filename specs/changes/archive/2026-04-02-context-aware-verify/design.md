## Architecture

This is a skill-template-only change. The verify skill template (`internal/skill/verify.go`) gains a branching section that detects task state and selects the appropriate review mode. The skill description in `internal/paths.go` is updated. The generated skill file at `.agents/skills/litespec-verify/SKILL.md` is regenerated.

No changes to parser, merge logic, CLI commands, or data structures.

```
verify skill template (after)
┌──────────────────────────────────────┐
│ Setup: read tasks.md                 │
│         ↓                            │
│ Count checked vs total tasks         │
│         ↓                            │
│ 0 checked → Artifact review mode     │
│ some     → Implementation review     │
│ all      → Pre-archive review        │
│            (both artifacts + code)    │
└──────────────────────────────────────┘
```

## Decisions

**Heuristic, not flag.** The mode is determined by task checkbox state, not by a CLI flag or user input. This keeps the workflow simple — the skill just does the right thing based on where the change is in its lifecycle. A `--mode` flag would add ceremony for no gain.

**Task state is the source of truth.** `TaskCompletion()` already parses checkbox state. The skill template instructs the AI to read `tasks.md` and count checked vs total. Three outcomes: zero checked (including zero total, e.g. empty or malformed tasks.md) → artifact review; some but not all → implementation review; all → pre-archive review. No new infrastructure needed.

**Same output format for all modes.** All modes use the existing CRITICAL/WARNING/SUGGESTION/Scorecard format. Only the dimensions and what gets evaluated change per mode.

**Artifact review is judgment-based, not structural.** `litespec validate` catches syntax and structural issues (missing scenarios, bad delta format). Artifact review catches judgment gaps: vague requirements, non-goal contradictions, untestable scenarios, design-scope mismatches. The two are complementary.

## File Changes

- `internal/skill/verify.go` — Add three-way branch to the verify template: artifact review (0 checked), implementation review (some checked), pre-archive review (all checked)
- `internal/paths.go` — Update verify skill description to mention all three review modes
- `.agents/skills/litespec-verify/SKILL.md` — Regenerate from updated template
- `DESIGN.md` — Update verify row in Skills table to describe context-aware three-mode behavior
- `AGENTS.md` — Update verify description in Workflow section to reflect artifact review, implementation review, and pre-archive review modes
