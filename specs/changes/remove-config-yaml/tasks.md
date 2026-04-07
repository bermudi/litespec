## Phase 1: Core removal and detection logic

- [x] Add `DetectActiveAdapters(root string) []string` to `internal/adapter.go` — scans each registered adapter's skill directory for symlinks whose target is inside `.agents/skills/`
- [x] Add tests for `DetectActiveAdapters` in `internal/adapter_test.go` — cover: active adapter detected, no adapters, adapter dir exists but empty, adapter dir does not exist
- [x] Remove `ProjectConfig` struct from `internal/types.go`
- [x] Remove `ConfigFileName` constant from `internal/paths.go`
- [x] Delete `internal/config.go` and `internal/config_test.go`

## Phase 2: Wire auto-detection into init and update

- [x] Update `cmd/litespec/init.go` — replace `ReadProjectConfig` fallback with `DetectActiveAdapters`, remove `saveToolIDs` call (refs: Auto-Detection of Active Tool Adapters, No Config File for Tool Persistence)
- [x] Update `cmd/litespec/update.go` — same changes as init.go
- [x] Remove `saveToolIDs` from `cmd/litespec/helpers.go`
- [x] Update tests in `cmd/litespec/main_test.go` — remove assertions that check for `specs/config.yaml` creation, replace `ReadProjectConfig` calls with `DetectActiveAdapters` calls, add test cases for auto-detection scenarios (adapter detected, no adapter, adapter dir exists but empty)
- [x] Delete `specs/config.yaml` from the repo

## Phase 3: Docs and cleanup

- [x] Update `docs/cli-reference.md` — remove config persistence references, explain auto-detection behavior for `init` and `update`
- [x] Update `docs/getting-started.md` — update `--tools` section to explain that `update` auto-detects active adapters
- [x] Update `docs/project-structure.md` — remove `specs/config.yaml` from project structure
- [x] Verify `AGENTS.md` and `DESIGN.md` are consistent with config removal — remove or update any references to config-based tool persistence
- [x] Run `go build`, `go test ./...`, `go vet` to verify everything passes
