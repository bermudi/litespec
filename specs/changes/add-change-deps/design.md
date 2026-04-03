## Architecture

This change adds a dependency graph layer on top of the existing change system. The dependency data lives in `.litespec.yaml`, resolution logic lives in a new `internal/deps.go`, and the existing `validate`, `list`, and `archive` commands integrate with it.

```
.litespec.yaml (ChangeMeta.DependsOn)
         │
         ▼
  internal/deps.go
  ├── ResolveDeps()      → map each dependsOn name to active/archived change
  ├── DetectCycles()     → DFS cycle detection across active changes
  ├── TopologicalSort()  → deterministic ordering for list --sort deps
  └── DetectOverlaps()   → scan delta targets across active changes
         │
         ▼
  cmd/litespec/main.go
  ├── validate  → calls ResolveDeps + DetectCycles + DetectOverlaps
  ├── list      → calls TopologicalSort when --sort deps
  ├── archive   → checks for active dependents before archiving
  └── view      → new command, renders dashboard + dependency graph
```

No new external dependencies. Graph algorithms are standard Go (DFS + topological sort via Kahn's algorithm or post-order DFS).

## Decisions

### `dependsOn` in `.litespec.yaml` (not a separate file)

**Chosen:** Add `dependsOn` to the existing metadata file.
**Why:** One file per change is the current pattern. A separate `deps.yaml` would add file-count noise for a single optional field.
**Constraint:** Backward compatible — `dependsOn` is optional, absent = no dependencies.

### Active-first resolution

**Chosen:** When both active and archived changes share a base name, resolve to active.
**Why:** The active change is the one being worked on. Archived is history. The user almost certainly means the live work.
**Constraint:** Archived change names include a date prefix (`2026-04-01-name`), so name extraction strips the date prefix for matching.

### Soft block on archive with `--allow-incomplete`

**Chosen:** Reuse the existing `--allow-incomplete` flag rather than adding `--force`.
**Why:** One escape hatch to remember. The flag already means "I know what I'm doing, proceed despite incompleteness." Active dependents are a form of incompleteness.
**Constraint:** Changes the semantics of `--allow-incomplete` slightly — it now covers both unchecked tasks and active dependents. Acceptable because both are "this isn't fully baked" signals.

### Overlap suppression when dep edge exists

**Chosen:** Don't warn about overlap between A and B if A dependsOn B or B dependsOn A.
**Why:** The dependency edge already acknowledges ordering awareness. Warning about overlap would be noise.
**Constraint:** Only direct edges suppress warnings. Transitive relationships (A→B→C) don't suppress A-C overlap warnings. Simple and predictable.

### `litespec view` as a new command

**Chosen:** Separate command, not a flag on `list`.
**Why:** `list` is tabular data. `view` is a dashboard with sections, progress bars, and a graph. Different output shape, different mental model.
**Constraint:** `view` mirrors OpenSpec's dashboard concept but is litespec-lean — no interactive mode, just a formatted terminal output.

### Topological sort with lexicographic tie-breaking

**Chosen:** When multiple changes have the same depth in the DAG, sort alphabetically.
**Why:** Deterministic output. Without it, the order would vary based on directory listing order or map iteration.
**Constraint:** Standard Kahn's algorithm with a sorted queue at each level.

## File Changes

### `internal/types.go`
- Add `DependsOn []string` field to `ChangeMeta` struct
- Affects: `ChangeMeta` YAML serialization/deserialization

### `internal/deps.go` (NEW)
- `ResolveDep(root, name string) (resolvedName string, isActive bool, found bool)` — resolve a single dependency name to active/archived
- `ResolveDeps(root string, deps []string) ([]ResolvedDep, error)` — resolve all deps for a change, error on missing
- `DetectCycles(root string) ([][]string, error)` — find all cycles in active change graph
- `TopologicalSort(changes []ChangeInfo, depMap map[string][]string) []ChangeInfo` — sort changes by dependency order
- `DetectOverlaps(root string, changes []ChangeInfo, depMap map[string][]string) []ValidationIssue` — find overlapping delta targets

### `internal/change.go`
- `ListArchivedChanges(root string) ([]string, error)` — list base names from archive directory (strip date prefix)
- `GetDependents(root, name string) ([]string, error)` — find active changes that depend on the given change

### `internal/validate.go`
- `ValidateChange` — add dependency resolution check after existing validation (resolve each `dependsOn`, error on missing)
- `ValidateAll` — add cycle detection + overlap detection after existing validation

### `internal/json.go`
- Add `DependsOn []string` to `ChangeListItemJSON` for JSON output

### `cmd/litespec/main.go`
- `cmdList` — add `deps` to `--sort` flag options, call `TopologicalSort`
- `cmdArchive` — add dependency check: find active dependents, block unless `--allow-incomplete`
- `cmdView` (NEW) — render dashboard with summary, changes, specs, dependency graph
- Add `view` to command dispatch switch

### `internal/paths.go`
- Add `ArchivedNameRe` regex for extracting base name from archived directory names (`YYYY-MM-DD-<name>`)
- Add `ParseArchivedName(name string) string` helper

### `DESIGN.md`
- Add `## Change Dependencies` section documenting `dependsOn` semantics
- Add `view` to CLI commands table
- Update `## Change Metadata` section to show new field

### `AGENTS.md`
- Add `view` to workflow awareness
- Mention `dependsOn` in core concepts
