## Architecture

The change enriches the `list` command's data pipeline end-to-end:

```
ListChanges() ──→ []ChangeInfo{Name, CompletedTasks, TotalTasks, LastModified}
    │                  │
    ├── TaskCompletion() (existing) reads tasks.md
    ├── GetLastModified() (new) walks dir for mtime
    │
    ▼
cmdList() ──→ sort by --sort flag ──→ format columns / JSON

ListSpecs() ──→ []SpecInfo{Name, RequirementCount}
    │
    ├── ParseMainSpec() (existing) reads spec.md
    │
    ▼
cmdList() ──→ sort alphabetical ──→ format columns / JSON
```

No new files. Everything lives in the existing `internal/change.go` (data enrichment) and `cmd/litespec/main.go` (display). JSON types go in `internal/json.go`.

## Decisions

**Return enriched structs from ListChanges/ListSpecs** — Changing the return type from `[]string` to `[]ChangeInfo`/`[]SpecInfo` is a breaking internal API change. The only callers are `cmdList()` and tests. This is cleaner than adding parallel enrichment functions. All callers get the data they need from one call.

**Recursive mtime walk over Created timestamp** — The `created` field in `.litespec.yaml` never changes. The mtime walk gives a "last touched" signal, which is more useful for "what's active" listing. Falls back to directory mtime if no files exist (empty change dir).

**Relative time formatting as a pure function** — `FormatRelativeTime(t time.Time) string` in `internal/change.go`. No special time package needed. Matches OpenSpec's thresholds exactly.

**Column padding with `padEnd`** — Compute `maxNameWidth` across all entries, then `fmt.Printf("%-*s", maxNameWidth, name)`. Same pattern as OpenSpec's `padEnd(nameWidth)`.

**Status derivation in JSON only** — The text display uses different labels (`✓ Complete` vs `No tasks` vs `3/5 tasks`). JSON collapses to three statuses: `complete`, `no-tasks`, `in-progress`. The derivation logic lives in the JSON builder, not in the type.

**--sort flag parsed alongside existing flags** — The CLI uses a hand-rolled arg parser. `--sort` takes the next arg as its value. Same pattern as existing `--type` for validate.

## File Changes

### `internal/change.go` — Enrich list functions and add helpers

- Add `ChangeInfo` struct: `Name string`, `CompletedTasks int`, `TotalTasks int`, `LastModified time.Time`
- Add `SpecInfo` struct: `Name string`, `RequirementCount int`
- Change `ListChanges(root string) ([]string, error)` to `ListChanges(root string) ([]ChangeInfo, error)` — for each change dir, read `tasks.md` via `TaskCompletion()`, call `GetLastModified()` on the change dir
- Change `ListSpecs(root string) ([]string, error)` to `ListSpecs(root string) ([]SpecInfo, error)` — for each spec dir, parse `spec.md` via `ParseMainSpec()`, count requirements
- Add `GetLastModified(dir string) (time.Time, error)` — recursive filepath.Walk, return max mtime, fall back to dir mtime
- Add `FormatRelativeTime(t time.Time) string` — convert to relative time string

Satisfies: Enriched Internal Types, Enriched Change Listing, Enriched Spec Listing

### `internal/json.go` — Update JSON types

- Change `ChangeListItemJSON` from `{Name string}` to `{Name, CompletedTasks, TotalTasks, LastModified, Status string}`
- Add `SpecListItemJSON{Name string, RequirementCount int}`
- Update any code that constructs `ChangeListItemJSON`

Satisfies: Enriched JSON Output

### `cmd/litespec/main.go` — Update cmdList display logic

- Parse `--sort` flag with value (`recent`/`name`)
- Update `cmdList()` to use new enriched return types
- Add column-aligned text output for changes: `name     status     time`
- Add column-aligned text output for specs: `name     requirements N`
- Update JSON output construction with enriched fields
- Sort changes by `LastModified` desc (default) or `Name` asc

Satisfies: Enriched Change Listing, Enriched Spec Listing, Sort Flag, Enriched JSON Output

### Tests to update

- Any tests calling `ListChanges()` or `ListSpecs()` need updating for new return types
- New tests: `TestGetLastModified`, `TestFormatRelativeTime`, `TestListChangesEnriched`, `TestListSpecsEnriched`
