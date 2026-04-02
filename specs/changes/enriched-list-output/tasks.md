## Phase 1: Internal enrichment types and helpers

- [x] Add `ChangeInfo` struct to `internal/change.go` with `Name`, `CompletedTasks`, `TotalTasks`, `LastModified` fields
- [x] Add `SpecInfo` struct to `internal/change.go` with `Name`, `RequirementCount` fields
- [x] Implement `GetLastModified(dir string) (time.Time, error)` in `internal/change.go` — recursive walk, max mtime, directory mtime fallback
- [x] Implement `FormatRelativeTime(t time.Time) string` in `internal/change.go` — just now / Xm ago / Xh ago / Xd ago / date
- [x] Change `ListChanges()` return type to `[]ChangeInfo` — read tasks.md via `TaskCompletion()`, call `GetLastModified()` for each change
- [x] Change `ListSpecs()` return type to `[]SpecInfo` — parse spec.md via `ParseMainSpec()`, count requirements, 0 on parse failure

## Phase 2: JSON output types

- [x] Update `ChangeListItemJSON` in `internal/json.go` to include `CompletedTasks`, `TotalTasks`, `LastModified`, `Status` fields
- [x] Add `SpecListItemJSON` struct with `Name` and `RequirementCount` fields
- [x] Add `ChangeListStatus(completed, total int) string` helper that returns `no-tasks` / `complete` / `in-progress`

## Phase 3: CLI display and sort

- [x] Add `--sort` flag parsing to `cmdList()` accepting `recent` (default) or `name` with next-arg value extraction
- [x] Update changes text output to column-aligned format: padded name + status + relative time
- [x] Update specs text output to column-aligned format: padded name + `requirements N`
- [x] Update JSON output construction using enriched types
- [x] Add sort logic: `--sort recent` → by `LastModified` desc, `--sort name` → by `Name` asc, specs always alphabetical

## Phase 4: Tests

- [x] Write `TestGetLastModified` — nested files, empty directory, single file
- [x] Write `TestFormatRelativeTime` — just now, minutes, hours, days, >30 days
- [x] Write `TestListChangesEnriched` — verify task counts and lastModified populated
- [x] Write `TestListSpecsEnriched` — verify requirement counts, parse failure returns 0
- [x] Update any existing tests broken by the `ListChanges`/`ListSpecs` signature change
- [x] Run `go build`, `go test`, `go vet`
