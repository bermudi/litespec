## Motivation

Two related problems:

1. **`specs/specs/` is confusing.** The double-nested `specs/specs/` directory makes paths hard to read and harder to explain. Every conversation about where specs live requires clarifying "the inner specs, not the outer specs."

2. **Archive retains stale delta specs.** `litespec archive` moves the entire change directory — including its `specs/` subtree — into `specs/changes/archive/`. After merging, those delta specs are obsolete. Their presence in the archive creates a misleading parallel spec tree and contradicts the principle that `canon/` is the single source of truth.

## Scope

- Rename `specs/specs/` to `specs/canon/` throughout the codebase (constant, function, tests, skill templates)
- Update `ArchiveChange()` to strip the change's `specs/` subtree before moving to archive
- Rename the `SpecsDirName` constant to `CanonDirName` with value `"canon"`
- Rename the `SpecsPath()` function to `CanonPath()`
- Update all callers and tests to use the new names
- Rename the physical directory `specs/specs/` → `specs/canon/` on disk
- Update the archive skill template (`internal/skill/artifact.go`) which references `specs/specs/`

## Non-Goals

- No changes to the delta spec format or merge logic
- No changes to the `specs/changes/<name>/specs/` structure (that stays as-is — it's the delta staging area)
- No changes to CLI flags or command names
- No migration tooling for existing projects — this is early enough that a manual rename suffices
