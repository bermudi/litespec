## Architecture

The `import` command follows the same pattern as other litespec CLI commands: a single function in `cmd/litespec/import.go` that orchestrates file operations, with migration logic in `internal/` packages. The command is read-write (unlike most CLI commands which are read-only) because it creates the litespec project structure from an external source.

Data flow:
1. Detect OpenSpec project structure at source path
2. Enumerate canon specs and changes (active and archived)
3. For each spec: copy to `specs/canon/`, strip " Specification" suffix from H1 if present
4. For each active change: copy directory tree (skip loose files), convert metadata (`.openspec.yaml` → `.litespec.yaml`), normalize task phase labels
5. For each archived change: copy directory tree, convert/synthesize metadata, normalize task phase labels, strip `specs/` subdirectories
6. Report warnings for skipped items (config.yaml, project.md, AGENTS.md, explorations/, loose files, conflicts)
7. Suggest running `litespec update` to generate skills

## Decisions

### File-based migration, not in-place

The import copies from source to target rather than moving in-place. This preserves the original OpenSpec project as a safety net. The source directory is never modified.

### Streaming output over summary-only

The command prints progress for each spec and change as it processes them, giving the user visibility into what's happening. A final summary counts successes, warnings, and skips.

### Minimal content transformation

Spec file content is preserved with two targeted normalizations:
1. **H1 " Specification" suffix** — stripped if present (e.g., `# cli-init Specification` → `# cli-init`). This is pure boilerplate in OpenSpec conventions.
2. **Task phase labels** — `## 1. Name` → `## Phase 1: Name`. litespec's task tracking depends on the `Phase N:` prefix to derive progress.

No other content transformation is performed. Spec files with non-standard H2 sections (other than `## Purpose` or `## Requirements`) will fail litespec validation after import and must be fixed manually. This edge case is documented in the proposal.

### Archive metadata synthesis

Archived changes in OpenSpec may lack `.openspec.yaml` (older archives). The importer synthesizes a minimal `.litespec.yaml` by parsing the date prefix from the directory name (e.g., `2026-04-01-my-change/` yields `created: 2026-04-01T00:00:00Z`, `schema: spec-driven`). Archives with existing metadata are converted normally.

### Archive spec stripping

Archived changes in OpenSpec retain their full `specs/` subtrees. litespec convention strips specs from archived changes. The importer strips `specs/` subdirectories during archive migration to match litespec convention.

### Conflict detection before any writes

The command scans for conflicts (existing files at target paths) before writing anything. If conflicts are found and `--force` is not set, it aborts without modifying the project.

## File Changes

### Created

- `cmd/litespec/import.go` — CLI entry point for `litespec import`, flag parsing (`--dry-run`, `--source`, `--force`), conflict detection, streaming output, orchestration
- `internal/importer/openspec.go` — Core import logic: project detection, canon spec copy with H1 normalization, change migration, archive migration with spec stripping and metadata synthesis, loose file filtering, task phase label normalization
- `cmd/litespec/import_test.go` — Tests for the import command
- `internal/importer/openspec_test.go` — Tests for import logic

### Modified

- `cmd/litespec/main.go` — Add `import` case to command switch and update `printUsage()`
