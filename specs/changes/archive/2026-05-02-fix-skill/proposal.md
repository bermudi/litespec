# Proposal: Fix Skill

## Motivation

The review skill produces structured findings (CRITICAL, WARNING, SUGGESTION) but litespec has no structured skill to address them. The review skill currently says "use apply" — but apply is designed for phased task implementation from `tasks.md`, not for ingesting and resolving review findings. Each phase is one agent session, so the original phase agent that wrote the code is gone.

This creates a dead end after review:

```
apply (phase N) → review → [findings] → ???
```

The user must feed findings back to an agent ad hoc, with no litespec structure around verification, prioritization, or closure. This invites compounding booboos: findings get partially addressed, fixes introduce regressions, and the human has no confidence that "all findings are resolved" vs. "the agent stopped mentioning them."

The wiki concepts of backpressure, verification loops, and agent quality loops all point to the same conclusion: review findings are a feedback circuit that needs a structured response path, not an ad hoc handoff.

## Scope

- A new `fix` skill (`litespec-fix`) that ingests review findings, addresses them systematically, verifies each fix, and commits
- The fix skill operates on the same change — no new proposal, no new delta. It closes the review loop within the existing change
- Review skill updated to reference the fix skill instead of telling the user to "use apply"
- `SkillInfo` entry added to `internal/paths.go` Skills list
- Template registered in `internal/skill/fix.go`
- Workflow updated in DESIGN.md to show `fix` between `review` and `archive`

## Non-Goals

- Fix skill does NOT replace the review skill — review remains pure review, fix is pure correction
- Fix skill does NOT create new changes or modify specs — it only addresses implementation-level findings within an existing change
- Fix skill does NOT handle build failures from CI — that's a separate concern
- Fix skill does NOT modify `tasks.md` or create new phases — it operates on review output, not the task plan
- No CLI changes — this is purely a skill template addition, like all other skills
