## Phase 1: Refactor Exit Pattern

- [ ] Change all command function signatures from `func cmdXxx(args []string)` to `func cmdXxx(args []string) error`
- [ ] Replace all `os.Exit(1)` calls inside command functions with `return fmt.Errorf(...)`
- [ ] Create `run() error` function and update `main()` to call it and handle exit codes
- [ ] Verify `go build ./...` and `go test ./...` still pass

## Phase 2: Split Command Files

- [ ] Create `cmd/litespec/init.go` and move `cmdInit` implementation
- [ ] Create `cmd/litespec/new.go` and move `cmdNew` implementation
- [ ] Create `cmd/litespec/status.go` and move `cmdStatus` implementation
- [ ] Create `cmd/litespec/validate.go` and move `cmdValidate` implementation
- [ ] Create `cmd/litespec/list.go` and move `cmdList` implementation
- [ ] Create `cmd/litespec/instructions.go` and move `cmdInstructions` implementation
- [ ] Create `cmd/litespec/archive.go` and move `cmdArchive` implementation
- [ ] Create `cmd/litespec/view.go` and move `cmdView` implementation
- [ ] Create `cmd/litespec/update.go` and move `cmdUpdate` implementation
- [ ] Create `cmd/litespec/completion.go` and move completion commands
- [ ] Create `cmd/litespec/helpers.go` for shared flag helpers and usage printers
- [ ] Strip `main.go` to dispatch, `run()`, and `main()` only
- [ ] Verify `go build ./...` and `go test ./...` still pass

## Phase 3: Add Missing Tests

- [ ] Create `internal/adapter_test.go` with tests for `GetAdapter` and unknown tool ID error
- [ ] Add CLI happy path tests for core commands (new, status, validate, archive, list)
- [ ] Add CLI error path tests (missing name, nonexistent change, invalid flags)
- [ ] Run `go test ./...` and verify all tests pass

## Phase 4: Verify

- [ ] Run `go vet ./...`
- [ ] Run `go test -cover ./cmd/litespec/` and verify coverage improvement
- [ ] Run `go test -cover ./internal/` and verify no regression
- [ ] Run `go build ./...` and verify binary builds correctly
