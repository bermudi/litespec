## Phase 1: Remove archive skill from codebase

- [x] Delete `internal/skill/archive.go` (removes the archive template registration)
- [x] Remove the `archive` entry from the `Skills` slice in `internal/paths.go`
- [x] Remove `"archive"` from `knownIDs` in `internal/skill/skill_test.go`
- [x] Delete `.agents/skills/litespec-archive/SKILL.md`
- [x] Run `go build ./...` and `go test ./...` to verify no breakage

## Phase 2: Update documentation

- [x] Update `DESIGN.md` — remove `archive` row from skills table, remove `litespec-archive/` from directory tree listing
- [x] Update `AGENTS.md` — remove archive skill references from workflow description and skills listing
- [x] Update `docs/project-structure.md` — remove `litespec-archive/` from skills directory tree
- [x] Update `docs/cli-reference.md` — remove `litespec-archive/` from skills directory tree
- [x] Update `docs/workflow.md` — reframe "archive: Finalization Mode" to describe CLI usage without a dedicated skill
- [x] Run `go build ./...` and `go test ./...` to confirm everything still passes
