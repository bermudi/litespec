## Phase 1: Backlog Parsing and View Integration

- [x] Add `BacklogFileName` constant to `internal/paths.go`
- [x] Create `internal/backlog.go` with `BacklogPath`, `BacklogSummary` struct, and `ParseBacklog` function
- [x] Create `internal/backlog_test.go` with tests for all categories, unknown sections as other, nested bullet exclusion, missing file, and empty file
- [x] Add backlog summary line to `cmd/litespec/view.go` in the summary section after decisions
- [x] Verify with `go build && go test ./... && go vet ./...`

## Phase 2: Skill Template Updates

- [x] Add backlog awareness directive to `internal/skill/explore.go` template
- [x] Add backlog graduation directive to `internal/skill/propose.go` template
- [x] Add backlog deferral directive to `internal/skill/review.go` template
- [x] Add backlog scope challenge directive to `internal/skill/grill.go` template
- [x] Run `litespec update` to regenerate skills
- [x] Verify with `go build && go test ./... && go vet ./...`
