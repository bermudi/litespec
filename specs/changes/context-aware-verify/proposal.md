## Motivation

The verify skill currently assumes implementation has started — it reads code and compares against specs. But there's a natural checkpoint missing: reviewing the planning artifacts (proposal, specs, design, tasks) for quality and consistency *before* any code is written. Running verify on a change with zero tasks checked produces a useless report about missing implementation.

This happened in practice: running `verify spec-requirements-wrapper` on a fully-planned but unimplemented change produced 11 CRITICAL issues about missing code, when the user actually wanted a review of the artifacts themselves.

## Scope

- Make verify context-aware by detecting task completion state from `tasks.md`
- **All tasks unchecked** → artifact review mode: evaluate proposal/specs/design/tasks for quality, consistency, and readiness
- **Some tasks checked** → implementation review mode: current behavior (code vs specs)
- Update the verify skill template (`internal/skill/verify.go`) with branching instructions
- Update the verify skill description in `internal/paths.go` to reflect the dual mode
- Regenerate the skill file in `.agents/skills/litespec-verify/SKILL.md`

## Non-Goals

- Adding new CLI commands or flags
- Changing the workflow order or adding new skills
- Modifying how task checkbox state is parsed (use existing `TaskCompletion`)
- Changing any other skill's behavior
