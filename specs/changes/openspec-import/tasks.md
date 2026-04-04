## Phase 1: Core import logic

- [x] 1.1 Create `internal/importer/openspec.go` with OpenSpec project detection, canon spec copy with H1 normalization, change migration with loose file filtering, archive migration with spec stripping and metadata synthesis, metadata conversion (including quoted dates), and task phase label normalization (OpenSpec Project Detection, Canon Spec Migration, Change Migration, Archive Migration with Metadata Synthesis, Metadata Format Conversion, Task Phase Label Normalization requirements)
- [x] 1.2 Create `internal/importer/openspec_test.go` with tests for detection, H1 normalization, loose file filtering, archive stripping, metadata synthesis, metadata conversion, and task phase label normalization logic

## Phase 2: CLI command and integration

- [x] 2.1 Create `cmd/litespec/import.go` with flag parsing (`--dry-run`, `--source`, `--force`), conflict detection, streaming progress output, and orchestration (Import Dry Run Mode, Import Target Directory, Conflict Detection, Skipped Directory and File Warnings, Post-Import Suggestion requirements)
- [x] 2.2 Add `import` case to `cmd/litespec/main.go` command switch and update `printUsage()`
- [x] 2.3 Create `cmd/litespec/import_test.go` with CLI-level tests including dry-run, source flag, force flag, and conflict scenarios
- [x] 2.4 Run `go build`, `go test`, `go vet` and fix any failures