## Phase 1: Metadata and Resolution
- [x] Add `DependsOn []string` field to `ChangeMeta` in `internal/types.go`
- [x] Add `ParseArchivedName` and `ArchivedNameRe` to `internal/paths.go`
- [x] Create `internal/deps.go` with `ResolveDep` and `ResolveDeps` functions
- [x] Add `ListArchivedChanges` to `internal/change.go`
- [x] Update tests: verify `ChangeMeta` round-trips with and without `dependsOn`

## Phase 2: Validation Integration
- [x] Add dependency reference validation to `ValidateChange` in `internal/validate.go`
- [x] Implement `DetectCycles` in `internal/deps.go` (DFS across active change graph)
- [x] Add cycle detection to `ValidateAll`
- [x] Implement `DetectOverlaps` in `internal/deps.go` (scan delta targets, suppress when dep edge exists)
- [x] Add overlap detection to `ValidateAll` (only during bulk validation, not single-change `ValidateChange`)
- [x] Update tests: missing dep produces error, cycle produces error, overlap produces warning, overlap suppressed by dep edge

## Phase 3: Archive Guard and List Sorting
- [x] Implement `GetDependents` in `internal/change.go` (find active changes that declare `dependsOn: [name]`)
- [x] Add dependency check to `cmdArchive` in `cmd/litespec/main.go`: block if active dependents exist, bypass with `--allow-incomplete`
- [x] Implement `TopologicalSort` in `internal/deps.go` (Kahn's algorithm with lexicographic tie-breaking)
- [x] Add `deps` option to `--sort` flag in `cmdList`
- [x] Add `DependsOn` to `ChangeListItemJSON` and populate in list output
- [x] Update tests: archive blocks on active dependent, list --sort deps produces correct order

## Phase 4: View Command
- [x] Implement `cmdView` in `cmd/litespec/main.go`: render summary section, active changes section, specs section
- [x] Add dependency graph rendering to `cmdView` (tree-style with box-drawing characters, omitted when no deps exist)
- [x] Add `view` to command dispatch switch and usage text
- [x] Update tests: view renders dashboard, graph section appears only when deps exist

## Phase 5: Documentation
- [ ] Update `DESIGN.md`: add Change Dependencies section, add view to CLI table, update Change Metadata section
- [ ] Update `AGENTS.md`: add view command, mention dependsOn in core concepts
