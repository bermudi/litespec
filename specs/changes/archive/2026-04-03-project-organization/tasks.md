## Phase 1: Refactor Exit Pattern

- [x] Change all command function signatures from `func cmdXxx(args []string)` to `func cmdXxx(args []string) error`
- [x] Replace all `os.Exit(1)` calls inside command functions with `return fmt.Errorf(...)`
- [x] Create `run() error` function and update `main()` to call it and handle exit codes
- [x] Verify `go build ./...` and `go test ./...` still pass

## Phase 2: Split Command Files

- [x] Create `cmd/litespec/init.go` and move `cmdInit` implementation
- [x] Create `cmd/litespec/new.go` and move `cmdNew` implementation
- [x] Create `cmd/litespec/status.go` and move `cmdStatus` implementation
- [x] Create `cmd/litespec/validate.go` and move `cmdValidate` implementation
- [x] Create `cmd/litespec/list.go` and move `cmdList` implementation
- [x] Create `cmd/litespec/instructions.go` and move `cmdInstructions` implementation
- [x] Create `cmd/litespec/archive.go` and move `cmdArchive` implementation
- [x] Create `cmd/litespec/view.go` and move `cmdView` implementation
- [x] Create `cmd/litespec/update.go` and move `cmdUpdate` implementation
- [x] Create `cmd/litespec/completion.go` and move completion commands
- [x] Create `cmd/litespec/helpers.go` for shared flag helpers and usage printers
- [x] Strip `main.go` to dispatch, `run()`, and `main()` only
- [x] Verify `go build ./...` and `go test ./...` still pass

## Phase 3: Add Missing Tests

- [x] Create `internal/adapter_test.go` with tests for `GetAdapter` and unknown tool ID error
- [x] Add CLI happy path tests for core commands (new, status, validate, archive, list)
- [x] Add CLI error path tests (missing name, nonexistent change, invalid flags)
- [x] Run `go test ./...` and verify all tests pass

## Phase 4: Verify

- [x] Run `go vet ./...`
- [x] Run `go test -cover ./cmd/litespec/` and verify coverage improvement
- [x] Run `go test -cover ./internal/` and verify no regression
- [x] Run `go build ./...` and verify binary builds correctly
