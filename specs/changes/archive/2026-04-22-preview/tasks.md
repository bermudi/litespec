# Tasks — Delta Preview Command

## Phase 1: Internal Package

- [x] Create `internal/preview.go` with `PreviewOperation`, `PreviewCapability`, `PreviewTotals`, `PreviewResult` types
- [x] Implement `ComputePreviewResult(writes []PendingWrite, root string) (*PreviewResult, error)` — diff current canon against merged content
- [x] Implement text formatter `FormatPreviewText(result *PreviewResult) string`
- [x] Implement JSON formatter `FormatPreviewJSON(result *PreviewResult) ([]byte, error)`
- [x] Unit tests: new capability, modified capability, mixed operations, empty change, no net change (MODIFIED with identical content omitted)
- [x] Unit tests: JSON shape matches spec

## Phase 2: CLI Command

- [x] Create `cmd/litespec/preview.go` with `cmdPreview(args []string) error`
- [x] Parse positional change name and `--json` flag
- [x] Reject empty name, missing change, archived change with clear errors
- [x] Call `PrepareArchiveWrites` and `ComputePreviewResult`
- [x] Print text or JSON output to stdout
- [x] Register `preview` in `cmd/litespec/main.go` dispatcher
- [x] Unit tests: happy path, empty change, non-existent change, archived change, JSON flag
- [x] Integration test: preview output matches archive merge result for sample changes

## Phase 3: Completions and Documentation

- [x] Add `preview` to `cmd/litespec/completion.go`
- [x] Add dynamic change-name completion for `preview <name>` in `internal/completion.go`
- [x] Update `DESIGN.md` CLI commands table to include `preview`
- [x] Update `AGENTS.md` if preview workflow guidance is needed
- [x] Regenerate skills with `litespec update` if any skill templates reference the new command

## Phase 4: Verification

- [x] Run `go build ./...`, `go vet ./...`, `go test ./...`
- [x] Manual E2E: `litespec preview <active-change>` on a real change
- [x] Manual E2E: `litespec preview <active-change> --json` and verify JSON structure
- [x] Manual E2E: attempt `litespec preview <archived-change>` and verify error
- [x] Manual E2E: attempt `litespec preview` with no name and verify error
