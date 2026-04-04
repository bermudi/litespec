## Phase 1: Core import logic and spec transformations

- [ ] 1.1 Create `internal/importer/transforms.go` with H1 normalization, H2 stripping, and FROM/TO rename conversion functions (Delta Spec Rename Format Conversion requirement)
- [ ] 1.2 Create `internal/importer/transforms_test.go` with tests for all transformation functions
- [ ] 1.3 Create `internal/importer/openspec.go` with OpenSpec project detection, canon spec migration, change migration, and metadata conversion (OpenSpec Project Detection, Canon Spec Migration, Change Migration, Metadata Format Conversion requirements)
- [ ] 1.4 Create `internal/importer/openspec_test.go` with tests for detection and migration logic

## Phase 2: CLI command and integration

- [ ] 2.1 Create `cmd/litespec/import.go` with flag parsing (`--dry-run`, `--source`, `--force`), conflict detection, streaming progress output, and orchestration (Import Dry Run Mode, Import Target Directory, Conflict Detection, Config and Context File Warnings, Post-Import Suggestion requirements)
- [ ] 2.2 Add `import` case to `cmd/litespec/main.go` command switch and update `printUsage()` (also fix `--sort` flag alignment)
- [ ] 2.3 Create `cmd/litespec/import_test.go` with CLI-level tests including dry-run, source flag, force flag, and conflict scenarios
- [ ] 2.4 Run `go build`, `go test`, `go vet` and fix any failures
