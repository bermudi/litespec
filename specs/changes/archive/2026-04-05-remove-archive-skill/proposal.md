## Motivation

The `litespec-archive` skill exists to tell the agent "run `litespec validate`, then `litespec archive`." That is two sequential CLI commands with zero AI judgment involved. Every other skill in the system exists because the agent needs to exercise judgment — asking the right questions (explore, grill), creating structured artifacts from understanding (propose, adopt), evaluating quality (review), or translating specs into code (apply). Archive is purely mechanical.

Worse, the skill's "fix errors before proceeding" step is a dead end: it tells the agent to fix issues but provides no structure or loop for doing so, and fixes belong back in `apply` anyway. The pre-archive AI review is already fully covered by the `review` skill's Section C (Pre-Archive Review Mode).

A skill that wraps two CLI commands adds overhead with no value. It should be removed. Archive remains a CLI command — it just doesn't need a skill wrapping it.

## Scope

- Delete `.agents/skills/litespec-archive/SKILL.md`
- Update `DESIGN.md` — remove `archive` from the skills table and directory listing
- Update `AGENTS.md` — remove archive from workflow description and skills listing
- Remove archive skill references from `docs/project-structure.md`, `docs/cli-reference.md`, and `docs/workflow.md`
- Ensure `litespec init` and `litespec update` no longer generate the archive skill (removing from `internal.Skills` handles both)


## Non-Goals

- Changing the `litespec archive` CLI command behavior — it stays exactly as-is
- Changing the review skill's pre-archive review mode — it already covers this

