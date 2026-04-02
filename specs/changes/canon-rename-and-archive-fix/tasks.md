## Phase 1: Rename path constants, function, and all callers

- [x] In `internal/paths.go`: rename `SpecsDirName` → `CanonDirName` with value `"canon"`, add `ChangeSpecsDirName = "specs"`, rename `SpecsPath()` → `CanonPath()`, update `ChangeSpecsPath` to use `ChangeSpecsDirName`
- [x] Update all callers in `internal/change.go`: `SpecsPath` → `CanonPath`
- [x] Update all callers in `internal/validate.go`: `SpecsPath` → `CanonPath`
- [x] Update `internal/skill/artifact.go`: string `specs/specs/` → `specs/canon/`
- [x] Update `internal/archive_test.go`: all `SpecsPath` → `CanonPath`
- [x] Update `internal/validate_test.go`: all `SpecsPath` → `CanonPath`
- [x] Update `cmd/litespec/main_test.go`: hardcoded `"specs", "specs"` → `"specs", "canon"` (lines 26, 87)
- [x] Run `go build ./...`, `go vet ./...`, `go test ./...` — must all pass

## Phase 2: Fix archive to strip specs subtree

- [ ] In `internal/change.go`: add `os.RemoveAll(ChangeSpecsPath(root, name))` in `ArchiveChange` before `os.Rename`
- [ ] In `internal/archive_test.go`: add test asserting archived directory does NOT contain `specs/` subtree
- [ ] Update `AGENTS.md` line 26: `specs/specs/` → `specs/canon/`
- [ ] Update or remove `BUGS.md` entry 1 (this change fixes the described bug)
- [ ] Run `go test ./...` — all tests pass

## Phase 3: Physical directory rename

- [ ] Rename physical directory: `mv specs/specs specs/canon`
- [ ] Run full `go test ./...` and `go vet ./...`
