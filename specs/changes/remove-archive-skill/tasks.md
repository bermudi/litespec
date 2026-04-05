## Phase 1: Remove archive skill from codebase

- [ ] Delete `internal/skill/archive.go` (removes the archive template registration)
- [ ] Remove the `archive` entry from the `Skills` slice in `internal/paths.go`
- [ ] Remove `"archive"` from `knownIDs` in `internal/skill/skill_test.go`
- [ ] Delete `.agents/skills/litespec-archive/SKILL.md`
- [ ] Run `go build ./...` and `go test ./...` to verify no breakage

## Phase 2: Update documentation

- [ ] Update `DESIGN.md` — remove `archive` row from skills table, remove `litespec-archive/` from directory tree listing
- [ ] Update `AGENTS.md` — remove archive skill references from workflow description and skills listing
- [ ] Run `go build ./...` and `go test ./...` to confirm everything still passes
