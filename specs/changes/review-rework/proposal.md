## Motivation

The `verify` skill's own vocabulary already says "review" — it describes itself as a "QA reviewer" performing "context-aware review" in three modes: artifact review, implementation review, and pre-archive review. The name `verify` was a poor fit from the start, sounding more like a CI check than a judgment-based review.

Additionally, the workflow is missing a natural gate: reviewing planning artifacts before implementation begins. Currently the workflow jumps straight from `propose` to `apply` with no checkpoint. Users should review their plan before committing to code.

Finally, the `continue` skill exists to create one artifact at a time, but `propose` already handles this via its resume behavior (`propose.go:20`: "pick up where it left off"). `continue` is a subset of `propose` that adds a skill registration, a directory, tests, and a template — all to save one re-invocation.

## Scope

- Rename `verify` skill to `review` across the entire codebase: skill registration, template, generated skill directory, canon spec directory, and all documentation
- Reorder the canonical workflow to `explore → grill → propose → review → apply → review → archive`
- Remove the `continue` skill entirely — `propose` absorbs incremental artifact creation
- Update `propose` skill template to reference `review` instead of `verify` in suggested next steps
- Update all docs (AGENTS.md, DESIGN.md, README.md, docs/*) to reflect the new workflow and naming
- Regenerate `.agents/skills/litespec-review/SKILL.md` from the renamed template (covered by existing skill-generation spec's template registration requirement)
- Update `internal/skill/skill_test.go` expected skill list
- Rename `specs/canon/verify/` to `specs/canon/review/`

## Non-Goals

- Changing the three-mode review detection logic (artifact/impl/pre-archive remains identical)
- Updating archived change directories under `specs/changes/archive/` — those are historical records
- Adding enforcement that blocks `apply` without a prior `review` — it remains a hard suggestion, not a hard gate
- Changing any CLI commands — the CLI is a read-only context provider and has no `verify` or `continue` subcommands
