# project-structure

## MODIFIED Requirements

### Requirement: Per-Command File Organization

The CLI entry point MUST be split into separate files, one per command, following the pattern `cmd/litespec/<command>.go`. The `main.go` file SHALL contain only the `main()` function, the command dispatcher, and shared helpers. Each command file SHALL define its command function at package level. The `archive.go` command file SHALL remain — the archive CLI command is unaffected by removal of the archive skill.

#### Scenario: Command files exist

- **WHEN** the `cmd/litespec/` directory is listed
- **THEN** files like `init.go`, `new.go`, `status.go`, `validate.go`, `list.go`, `instructions.go`, `archive.go`, `view.go`, `update.go`, `completion.go` exist alongside `main.go`

#### Scenario: main.go contains only dispatch and helpers

- **WHEN** `main.go` is inspected
- **THEN** it contains `main()`, `printUsage()`, shared flag helpers, and the command switch — no command implementation logic
