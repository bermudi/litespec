## Architecture

Refactor the CLI from a monolithic file to a multi-file layout. Introduce a `run() error` function pattern so `main()` becomes `func main() { if err := run(); err != nil { fmt.Fprintf(os.Stderr, "error: %v\n", err); os.Exit(1) } }`. Each command function changes from `func cmdXxx(args []string)` to `func cmdXxx(args []string) error`.

## Decisions

- **Return errors, don't exit**: Each command returns `error`. The single `main()` / `run()` handles all `os.Exit()` calls. This is the standard Go pattern for testable CLIs.
- **One file per command**: Each command gets its own file. Shared helpers (flag parsing, usage printing) stay in `main.go` or a `helpers.go`.
- **Keep internal flat for now**: The `internal/` package restructure is deferred — it's a larger refactor that touches many files and would conflict with active proposals modifying validation and archive logic.
- **Test at function level**: Call command functions directly from tests, capturing stdout/stderr via `bytes.Buffer` if needed. No subprocess testing.

## File Changes

- `cmd/litespec/main.go`: Strip down to `main()`, `run() error`, command dispatch, and shared helpers. Move command implementations out.
- `cmd/litespec/init.go`: New file — `cmdInit` implementation.
- `cmd/litespec/new.go`: New file — `cmdNew` implementation.
- `cmd/litespec/status.go`: New file — `cmdStatus` implementation.
- `cmd/litespec/validate.go`: New file — `cmdValidate` implementation.
- `cmd/litespec/list.go`: New file — `cmdList` implementation.
- `cmd/litespec/instructions.go`: New file — `cmdInstructions` implementation.
- `cmd/litespec/archive.go`: New file — `cmdArchive` implementation.
- `cmd/litespec/view.go`: New file — `cmdView` implementation.
- `cmd/litespec/update.go`: New file — `cmdUpdate` implementation.
- `cmd/litespec/completion.go`: New file — `cmdCompletion` and `cmdComplete` implementations.
- `cmd/litespec/helpers.go`: New file — shared flag helpers, usage printers.
- `cmd/litespec/main_test.go`: Refactor existing tests and add new ones for exit-free commands.
- `internal/adapter_test.go`: New file — tests for `GetAdapter`, `GenerateAdapterCommands`.
