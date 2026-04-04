## Architecture

The `import` command follows the same pattern as other litespec CLI commands: a single function in `cmd/litespec/import.go` that orchestrates file operations, with conversion logic in `internal/` packages. The command is read-write (unlike most CLI commands which are read-only) because it creates the litespec project structure from an external source.

Data flow:
1. Detect OpenSpec project structure at source path
2. Enumerate canon specs and changes
3. For each spec: read, transform (H1 normalization, H2 stripping), write to `specs/canon/`
4. For each change: copy directory tree, transform metadata (`.openspec.yaml` → `.litespec.yaml`), transform delta specs (FROM/TO → arrow)
5. Report warnings for skipped items (archive, config.yaml, conflicts)
6. Suggest running `litespec update` to generate skills

## Decisions

### File-based migration, not in-place

The import copies from source to target rather than moving in-place. This preserves the original OpenSpec project as a safety net. The source directory is never modified.

### Streaming output over summary-only

The command prints progress for each spec and change as it processes them, giving the user visibility into what's happening. A final summary counts successes, warnings, and skips.

### Spec transformation via regex, not full parser

H1 normalization and H2 stripping use regex-based text transformation rather than parsing and re-serializing the full spec. This preserves formatting, comments, and whitespace that the litespec parser doesn't care about. The FROM/TO rename conversion also uses pattern matching since the structure is well-defined.

### Conflict detection before any writes

The command scans for conflicts (existing files at target paths) before writing anything. If conflicts are found and `--force` is not set, it aborts without modifying the project.

## File Changes

### Created

- `cmd/litespec/import.go` — CLI entry point for `litespec import`, flag parsing (`--dry-run`, `--source`, `--force`), orchestration (OpenSpec Detection requirement, Import Target Directory requirement, Import Dry Run Mode requirement)
- `internal/importer/openspec.go` — Core import logic: project detection, canon spec migration, change migration, metadata conversion (Canon Spec Migration requirement, Change Migration requirement, Metadata Format Conversion requirement)
- `internal/importer/transforms.go` — Spec file transformations: H1 normalization, H2 stripping, FROM/TO rename conversion (Delta Spec Rename Format Conversion requirement)
- `cmd/litespec/import_test.go` — Tests for the import command
- `internal/importer/openspec_test.go` — Tests for import logic
- `internal/importer/transforms_test.go` — Tests for spec transformations

 ### Modified
 
- `cmd/litespec/main.go` — Add `import` case to command switch, add entry to `printUsage()` (sort flag alignment fix included)
