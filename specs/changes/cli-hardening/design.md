## Architecture

This change adds input validation at CLI command boundaries and error propagation in JSON output paths. No new packages are introduced.

## Decisions

- **Validate early, fail fast**: Input validation happens at the top of each command function, before any filesystem operations.
- **Derive, don't hardcode**: The adapter list comes from the existing `Adapters` slice — add a helper `ValidToolIDs() []string` to `paths.go`.
- **Collect, don't skip**: Iteration errors accumulate in a warnings slice and are reported alongside results, matching the existing `ValidationIssue` pattern.

## File Changes

- `cmd/litespec/main.go`: Add `validateChangeName()` helper, call at top of `cmdNew`. Replace all `data, _ := MarshalJSON(out)` with error-checking. Add `ValidToolIDs()` call in `cmdInit`/`cmdUpdate`. Collect iteration errors in status/list commands.
- `internal/paths.go`: Add `ValidToolIDs() []string` function that returns IDs from the `Adapters` slice.
- `internal/adapter.go`: Update error message in `GenerateAdapterCommands` to use `ValidToolIDs()` instead of hardcoded string.
- `cmd/litespec/main_test.go`: Add tests for change name validation, tools flag validation, and JSON error propagation.
