## Phase 1: Backlog Parsing and View Integration

- [ ] Add `BacklogFileName` constant to `internal/paths.go`
- [ ] Create `internal/backlog.go` with `BacklogPath`, `BacklogSummary` struct, and `ParseBacklog` function
- [ ] Create `internal/backlog_test.go` with tests for all categories, unknown sections as other, nested bullet exclusion, missing file, and empty file
- [ ] Add backlog summary line to `cmd/litespec/view.go` in the summary section after decisions
- [ ] Verify with `go build && go test ./... && go vet ./...`

## Phase 2: Skill Template Updates

- [ ] Add backlog awareness directive to `internal/skill/explore.go` template
- [ ] Add backlog graduation directive to `internal/skill/propose.go` template
- [ ] Add backlog deferral directive to `internal/skill/review.go` template
- [ ] Add backlog scope challenge directive to `internal/skill/grill.go` template
- [ ] Run `litespec update` to regenerate skills
- [ ] Verify with `go build && go test ./... && go vet ./...`
