## Phase 1: Input Validation

- [x] Add `validateChangeName(name string) error` in `cmd/litespec/main.go` that rejects empty, path-containing, and whitespace-padded names
- [x] Call `validateChangeName` at top of `cmdNew`
- [x] Add `ValidToolIDs() []string` to `internal/paths.go` deriving from `Adapters` slice
- [x] Validate `--tools` flag values in `cmdInit` and `cmdUpdate` against `ValidToolIDs()`
- [x] Update error message in `internal/adapter.go:15` to use `ValidToolIDs()`

## Phase 2: Error Propagation

- [x] Replace all `data, _ := internal.MarshalJSON(out)` with proper error handling across all commands
- [x] Add error collection in `cmdStatus` and `cmdList` iteration loops — collect warnings instead of silent `continue`
- [x] Surface collected warnings in both text and JSON output modes

## Phase 3: Testing

- [x] Add tests for `validateChangeName`: empty, path separator, traversal, whitespace, valid names
- [x] Add tests for tools flag validation: unknown tool, known tool, multiple tools
- [x] Add test verifying JSON marshal error propagation
