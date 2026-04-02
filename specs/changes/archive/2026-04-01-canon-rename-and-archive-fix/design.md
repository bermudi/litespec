## Architecture

This is a two-part refactor touching the path layer and the archive flow. No new packages or commands are introduced. The change is mechanical but wide-reaching — the `SpecsDirName`/`SpecsPath` constant/function pair is referenced across the entire codebase.

```
paths.go (CanonDirName, CanonPath) ← all callers reference this
    ├── change.go   (InitProject, ListSpecs, PrepareArchiveWrites, ArchiveChange)
    ├── validate.go (ValidateSpec, ValidateChange, dangling delta checks)
    ├── archive_test.go
    ├── validate_test.go
    ├── main.go     (cmdArchive, cmdValidate, cmdInit)
    └── skill/artifact.go (string reference in template)
```

## Decisions

**Two constants: `CanonDirName` for the canonical directory, `ChangeSpecsDirName` for the change-internal delta staging area.** The current `SpecsDirName = "specs"` serves double duty — it names both `specs/specs/` (canon) and `specs/changes/<name>/specs/` (delta staging). These are distinct concepts and need distinct constants: `CanonDirName = "canon"` and `ChangeSpecsDirName = "specs"`. The `ChangeSpecsPath()` function switches to `ChangeSpecsDirName`.

**`os.RemoveAll` before `os.Rename` in `ArchiveChange`.** The simplest approach: delete the `specs/` subtree from the change directory, then move what remains. No need for selective copying or temporary directories. If the remove fails, the rename still proceeds — the worst case is the old behavior (stale specs in archive), which is acceptable.

**Physical directory rename at apply time.** The `specs/specs/` directory on disk becomes `specs/canon/` when we apply. No migration tooling — manual rename is fine at this stage.

## File Changes

### `internal/paths.go`
- Rename `SpecsDirName = "specs"` → `CanonDirName = "canon"` (line 9)
- Add new constant `ChangeSpecsDirName = "specs"` for the change-internal delta directory
- Rename `func SpecsPath(root)` → `func CanonPath(root)` (line 86-88)
- Update the function body to use `CanonDirName`
- Update `ChangeSpecsPath` to use `ChangeSpecsDirName` instead of the renamed constant

### `internal/change.go`
- Update `InitProject`: `SpecsPath(root)` → `CanonPath(root)` (line 21)
- Update `ListSpecs`: `SpecsPath(root)` → `CanonPath(root)` (line 79)
- Update `PrepareArchiveWrites`: `SpecsPath(root)` → `CanonPath(root)` (line 141)
- Update `ArchiveChange`: add `os.RemoveAll(ChangeSpecsPath(root, name))` before `os.Rename` (before line 196)

### `internal/validate.go`
- Update all `SpecsPath(root)` references → `CanonPath(root)` (lines 160, 264, 309)

### `internal/skill/artifact.go`
- Update string literal `specs/specs/` → `specs/canon/` (line 63)

### `internal/archive_test.go`
- Update all `SpecsPath(root)` references → `CanonPath(root)` (throughout)
- Add assertion: archived directory MUST NOT contain `specs/` subtree

### `internal/validate_test.go`
- Update `SpecsPath(root)` references → `CanonPath(root)` (lines 13, 49)

### `cmd/litespec/main.go`
- No changes needed — `main.go` has no direct references to `SpecsPath`; it calls internal functions that use the path layer

### `specs/specs/` → `specs/canon/` (physical rename)
- Rename the directory on disk at apply time

### `cmd/litespec/main_test.go`
- Update hardcoded `filepath.Join(root, "specs", "specs")` → `filepath.Join(root, "specs", "canon")` (line 26)
- Update hardcoded `filepath.Join(root, "specs", "specs", name)` → `filepath.Join(root, "specs", "canon", name)` (line 87)

### `AGENTS.md`
- Update line 26: `specs/specs/` → `specs/canon/`

### `BUGS.md`
- Update or remove entry 1 (this change fixes the described bug)
