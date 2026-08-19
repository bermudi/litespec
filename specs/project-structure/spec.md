# project-structure

## Requirements

### Requirement: CLI Entry Point Split

The CLI entry point MUST be split into separate source files, one per command, under `cmd/litespec/`. The command files `init.go`, `new.go`, `validate.go`, `view.go`, `update.go`, `upgrade.go`, and `completion.go` SHALL each define one command function at package level. Shared rendering, flag parsing, and project-root helpers SHALL live in `render.go`, `helpers.go`, and `project.go` respectively. The `main.go` file SHALL contain only `main()`, the command dispatcher, and shared helpers.

#### Scenario: Command files exist

- **WHEN** the `cmd/litespec/` directory is listed
- **THEN** `main.go`, `init.go`, `new.go`, `validate.go`, `view.go`, `update.go`, `upgrade.go`, `completion.go`, `render.go`, `helpers.go`, `project.go`, and `main_test.go` are present

#### Scenario: main.go dispatches without command logic

- **WHEN** `main.go` is inspected
- **THEN** it contains `main()`, `run()`, `printUsage()`, the command switch, and shared helpers — no command-specific implementation logic

### Requirement: Exit-Free Command Functions

Command functions MUST NOT call `os.Exit()` directly. Each command function SHALL return an `error`. The `main()` function SHALL handle all error reporting and exit code determination in a single location. This enables command functions to be tested directly without terminating the process.

#### Scenario: Error path returns an error

- **WHEN** a command function encounters an error (e.g., `cmdNew` is called without a name)
- **THEN** it returns a non-nil error and `main()` prints it to stderr and exits with code 1

#### Scenario: Happy path returns nil

- **WHEN** a command function completes successfully
- **THEN** it returns `nil` and `main()` exits with code 0

### Requirement: Command Specification Registry

`internal/commandspec.go` SHALL contain a `CommandSpecs` registry that is the single source of truth for supported commands, their flags, positional arguments, and completion metadata. `internal/completion.go` SHALL derive all command and flag completions from this registry. When adding or modifying CLI commands or flags, the `CommandSpecs` registry MUST be updated.

#### Scenario: Completions derive from the registry

- **WHEN** the completion system is asked for commands and flags
- **THEN** it returns the entries registered in `CommandSpecs`, including hidden `__complete`

#### Scenario: New flag requires registry update

- **WHEN** a command adds or renames a flag
- **THEN** `CommandSpecs` is updated and `internal/completion.go` reflects the change without separate hardcoding

### Requirement: Standard Go Project Layout

The project SHALL follow the standard Go layout: the `litespec` binary source lives under `cmd/litespec/`, reusable packages live under `internal/`, and there is no top-level application code outside those roots. After changes to Go source or project structure, `go build`, `go test`, and `go vet` SHALL pass.

#### Scenario: Build and vet pass

- **WHEN** `go build ./...` and `go vet ./...` are run from the repository root
- **THEN** both complete with no errors

### Requirement: Adapter Test Coverage

`internal/adapter.go` MUST have corresponding test coverage in `internal/adapter_test.go`. Tests SHALL verify adapter lookup (`GetAdapter`), symlink generation and cleanup (`GenerateAdapterCommands` and `cleanStaleSymlinks`), active adapter detection (`DetectActiveAdapters`), and error cases for unknown tool IDs.

#### Scenario: Adapter tests pass

- **WHEN** `go test ./internal/ -run Adapter` is run
- **THEN** tests for `GetAdapter`, `GenerateAdapterCommands`, and unknown tool ID errors pass

#### Scenario: Unknown tool returns an error

- **WHEN** `GenerateAdapterCommands` is called with a tool ID that is not in `Adapters`
- **THEN** it returns an error naming the unknown tool and no symlink is created

### Requirement: Command Test Coverage

Each command in `cmd/litespec/` MUST have happy-path and error-path test coverage in `cmd/litespec/main_test.go`. Tests SHALL invoke the command functions directly (not via `os/exec`) using their exit-free `func([]string) error` signatures.

#### Scenario: New command has happy and error paths

- **WHEN** `TestCmdNewDirect_HappyPath` and `TestCmdNewDirect_MissingName` are run
- **THEN** the happy path links the change to a GH issue and the error path returns a missing-name error

#### Scenario: Init command has happy and error paths

- **WHEN** `TestCmdInitDirect_HappyPath` and `TestCmdInitDirect_UnknownTool` are run
- **THEN** the happy path initializes the project and the error path rejects an unknown tool

#### Scenario: Validate command has happy and error paths

- **WHEN** `TestCmdValidateDirect_AllJSON` and `TestCmdValidateDirect_InvalidType` are run
- **THEN** the happy path validates all specs and the error path rejects an invalid `--type` value

#### Scenario: Every command function is tested directly

- **WHEN** `cmdInit`, `cmdNew`, `cmdValidate`, `cmdView`, `cmdUpdate`, `cmdUpgrade`, and `cmdCompletion` are called from `cmd/litespec/main_test.go`
- **THEN** each command has at least one happy-path and one error-path direct test
