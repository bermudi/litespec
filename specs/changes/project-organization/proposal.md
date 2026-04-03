## Motivation

The CLI entry point is a 1274-line monolith with 11+ direct `os.Exit()` calls, making individual commands untestable. The `internal/` package mixes spec parsing, validation, dependency resolution, and JSON serialization in a single flat namespace. Test coverage reflects this: 76% for `internal/` but only 3% for the CLI layer. As the project grows, this structure makes it harder to reason about changes and verify behavior.

## Scope

- Split `cmd/litespec/main.go` into per-command files
- Refactor `os.Exit()` calls into a single return-based error pattern
- Extract testable command functions that return errors instead of calling `os.Exit`
- Add adapter tests
- Increase CLI test coverage for core command paths

## Non-Goals

- Restructuring the `internal/` package into subpackages (too large a refactor for one change; defer to future work)
- Adding `context.Context` support (separate change)
- Adding a Makefile or build automation (separate change)
- Migrating to a CLI framework (cobra, etc.)
