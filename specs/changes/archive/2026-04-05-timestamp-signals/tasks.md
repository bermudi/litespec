## Phase 1: Data Layer
- [x] Add `Created time.Time` field to `ChangeInfo` struct in `internal/change.go` (Enriched Internal Types)
- [x] Update `ListChanges()` to call `ReadChangeMeta` for each change and populate `Created` on `ChangeInfo`
- [x] Add `Born string` field to `ChangeListItemJSON` in `internal/json.go` (Enriched JSON Output)
- [x] Add tests verifying `ListChanges` populates `Created` from `.litespec.yaml`
- [x] Run `go build`, `go test`, `go vet`

## Phase 2: List Output
- [x] Update text output in `cmd/litespec/list.go` to show four columns: name, status, born (YYYY-MM-DD), last-touched (Enriched Change Listing)
- [x] Update JSON output in `cmd/litespec/list.go` to populate `Born` field from `c.Created` (Enriched JSON Output)
- [x] Add CLI tests for `list` text output verifying born column appears
- [x] Add CLI tests for `list --json` verifying `born` field present
- [x] Run `go build`, `go test`, `go vet`

## Phase 3: View Dashboard
- [x] Update active changes section to show `(born YYYY-MM-DD, touched Xm ago)` after progress bar (Dashboard Display)
- [x] Update draft changes section to show `(born YYYY-MM-DD, touched Xm ago)` after name
- [x] Update completed changes section to show `(born YYYY-MM-DD, touched Xm ago)` after name
- [x] Update dependency graph nodes to show `(born YYYY-MM-DD, touched Xm ago)` after name (Dependency Graph Section)
- [x] Add CLI tests for `view` output verifying timestamps appear in each section
- [x] Run `go build`, `go test`, `go vet`
