# Proposal: adopt-list-backlog

## Motivation

The `litespec list --backlog` command and its backing parser (`ParseBacklogItems`) were implemented directly without going through the spec workflow. This adopt change reverse-engineers specs from the existing code to establish a baseline for future modifications.

## Scope

Adopt the following implemented behavior:

- **Backlog item parser** (`internal/backlog.go`) — `ParseBacklog`, `ParseBacklogItems`, `normalizeBacklogSection`, `extractBacklogTitle`, and the `BacklogItem`/`BacklogSummary` structs
- **List command backlog flag** (`cmd/litespec/list.go`) — `--backlog` flag handling, mutual exclusivity, text and JSON output
- **JSON type** (`internal/json.go`) — `BacklogItemJSON`
- **Command spec** (`internal/commandspec.go`) — `--backlog` flag registration
- **Help text** (`cmd/litespec/helpers.go`, `cmd/litespec/main.go`) — usage strings

## Files Analyzed

- `internal/backlog.go`
- `internal/json.go`
- `cmd/litespec/list.go`
- `internal/commandspec.go`
- `cmd/litespec/helpers.go`
- `cmd/litespec/main.go`
- `internal/backlog_test.go`

## Capabilities Discovered

1. **backlog-item-parser** — line-by-line parsing of `specs/backlog.md`, extracting structured items with section keys and titles
2. **list-backlog-flag** — CLI surface for surfacing backlog items via `litespec list --backlog`, including mutual exclusivity, text/JSON output, and shell completion

## Non-Goals

- Modifying the existing `ParseBacklog` (summary counts) behavior — that predates this change and already has coverage via `view`
- Adding CLI-level integration tests for the `list` command — that's a broader test infrastructure gap, not specific to `--backlog`
