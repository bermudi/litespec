## Architecture

This is a removal change — no new components. The archive *CLI command* (`cmd/litespec/archive.go`, `internal/change.go`) remains untouched. We are only removing the archive *skill* (the AI-facing instructions).

The skill system (`internal/skill/`) uses a registry pattern: each skill registers a template via `init()`. Removing a skill means deleting its registration file and removing it from the `Skills` slice in `internal/paths.go`.

## Decisions

**Remove, not deprecate.** No feature flags, no transition period. The archive skill adds no value, so we cut it cleanly. Anyone who wants to run archive just runs the CLI command directly — which is what they should have been doing all along.

**Keep the CLI command.** `litespec archive` does real work (validate, merge deltas, move files). It stays. The skill was just a wrapper around two CLI commands with no AI judgment layer.

**Pre-archive review lives in review skill.** The review skill's Section C (Pre-Archive Review Mode) already covers AI review before archiving. Nothing needs to replace the archive skill's review-like behavior because it already has a proper home.

## File Changes

- `internal/paths.go` — remove the `archive` entry from the `Skills` slice (lines 50-54)
- `internal/skill/archive.go` — delete the entire file (registers the archive template)
- `.agents/skills/litespec-archive/SKILL.md` — delete the generated skill file
- `internal/skill/skill_test.go` — remove `"archive"` from `knownIDs` in `TestGet_ReturnsNonEmptyForKnownIDs`
- `DESIGN.md` — remove `archive` from the skills table, remove `litespec-archive/` from directory listing
- `AGENTS.md` — remove archive skill references from the workflow description and skills listing
