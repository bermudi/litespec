# project-structure

## Requirements

### Requirement: Per-Command File Organization

The CLI entry point MUST be split into separate files, one per command, following the pattern `cmd/litespec/<command>.go`. The `main.go` file SHALL contain only the `main()` function, the command dispatcher, and shared helpers. Each command file SHALL define its command function at package level. The `archive.go` command file SHALL remain — the archive CLI command is unaffected by removal of the archive skill.

#### Scenario: Command files exist

- **WHEN** the `cmd/litespec/` directory is listed
- **THEN** files like `init.go`, `new.go`, `status.go`, `validate.go`, `list.go`, `instructions.go`, `archive.go`, `view.go`, `update.go`, `completion.go` exist alongside `main.go`

#### Scenario: main.go contains only dispatch and helpers

- **WHEN** `main.go` is inspected
- **THEN** it contains `main()`, `printUsage()`, shared flag helpers, and the command switch — no command implementation logic

### Requirement: Exit-Free Command Functions

Command functions MUST NOT call `os.Exit()` directly. Instead, each command function SHALL return an error. The `main()` function SHALL handle error reporting and exit code determination in a single location. This enables command functions to be tested without process termination.

#### Scenario: Command returns error instead of exiting

- **WHEN** a command encounters an error (e.g., change not found)
- **THEN** the command function returns a non-nil error and `main()` prints it to stderr and exits with code 1

#### Scenario: Successful command returns nil error

- **WHEN** a command completes successfully
- **THEN** it returns nil and `main()` exits with code 0

### Requirement: Adapter Test Coverage

The `internal/adapter.go` file MUST have corresponding test coverage in `internal/adapter_test.go`. Tests SHALL verify adapter lookup, symlink generation, and error cases for unknown tool IDs.

#### Scenario: Adapter tests pass

- **WHEN** `go test ./internal/ -run Adapter` is run
- **THEN** tests for `GetAdapter`, `GenerateAdapterCommands`, and unknown tool ID error pass

### Requirement: CLI Command Test Coverage

Each command MUST have at least one test covering its happy path and one test covering its primary error path. Tests SHALL invoke command functions directly (not via `os/exec`) using the refactored exit-free signatures.

#### Scenario: Happy path test for new command

- **WHEN** `TestCmdNew` is run with a valid name
- **THEN** the change directory is created and no error is returned

#### Scenario: Error path test for new command

- **WHEN** `TestCmdNew` is run with an empty name
- **THEN** an error is returned indicating the name is invalid

#### Scenario: Happy path test for status command

- **WHEN** `TestCmdStatus` is run with an existing change
- **THEN** status information is returned without error
