# Tasks — Delta Preview Command

## Phase 1: Internal Package

- [x] Create `internal/preview.go` with `PreviewOperation`, `PreviewCapability`, `PreviewTotals`, `PreviewResult` types
- [x] Implement `ComputePreviewResult(writes []PendingWrite, root string) (*PreviewResult, error)` — diff current canon against merged content
- [x] Implement text formatter `FormatPreviewText(result *PreviewResult) string`
- [x] Implement JSON formatter `FormatPreviewJSON(result *PreviewResult) ([]byte, error)`
- [x] Unit tests: new capability, modified capability, mixed operations, empty change, no net change (MODIFIED with identical content omitted)
- [x] Unit tests: JSON shape matches spec

## Phase 2: CLI Command

- [ ] Create `cmd/litespec/preview.go` with `cmdPreview(args []string) error`
- [ ] Parse positional change name and `--json` flag
- [ ] Reject empty name, missing change, archived change with clear errors
- [ ] Call `PrepareArchiveWrites` and `ComputePreviewResult`
- [ ] Print text or JSON output to stdout
- [ ] Register `preview` in `cmd/litespec/main.go` dispatcher
- [ ] Unit tests: happy path, empty change, non-existent change, archived change, JSON flag
- [ ] Integration test: preview output matches archive merge result for sample changes

## Phase 3: Completions and Documentation

- [ ] Add `preview` to `cmd/litespec/completion.go`
- [ ] Add dynamic change-name completion for `preview <name>` in `internal/completion.go`
- [ ] Update `DESIGN.md` CLI commands table to include `preview`
- [ ] Update `AGENTS.md` if preview workflow guidance is needed
- [ ] Regenerate skills with `litespec update` if any skill templates reference the new command

## Phase 4: Verification

- [ ] Run `go build ./...`, `go vet ./...`, `go test ./...`
- [ ] Manual E2E: `litespec preview <active-change>` on a real change
- [ ] Manual E2E: `litespec preview <active-change> --json` and verify JSON structure
- [ ] Manual E2E: attempt `litespec preview <archived-change>` and verify error
- [ ] Manual E2E: attempt `litespec preview` with no name and verify error
