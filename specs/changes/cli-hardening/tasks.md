## Phase 1: Input Validation

- [ ] Add `validateChangeName(name string) error` in `cmd/litespec/main.go` that rejects empty, path-containing, and whitespace-padded names
- [ ] Call `validateChangeName` at top of `cmdNew`
- [ ] Add `ValidToolIDs() []string` to `internal/paths.go` deriving from `Adapters` slice
- [ ] Validate `--tools` flag values in `cmdInit` and `cmdUpdate` against `ValidToolIDs()`
- [ ] Update error message in `internal/adapter.go:15` to use `ValidToolIDs()`

## Phase 2: Error Propagation

- [ ] Replace all `data, _ := internal.MarshalJSON(out)` with proper error handling across all commands
- [ ] Add error collection in `cmdStatus` and `cmdList` iteration loops — collect warnings instead of silent `continue`
- [ ] Surface collected warnings in both text and JSON output modes

## Phase 3: Testing

- [ ] Add tests for `validateChangeName`: empty, path separator, traversal, whitespace, valid names
- [ ] Add tests for tools flag validation: unknown tool, known tool, multiple tools
- [ ] Add test verifying JSON marshal error propagation
