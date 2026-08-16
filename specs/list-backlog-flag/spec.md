# list-backlog-flag

## Requirements

### Requirement: --backlog flag

The `list` command SHALL accept a `--backlog` flag that displays individual backlog items instead of changes. When specified, the command reads `specs/backlog.md` via `ParseBacklogItems` and outputs items grouped by their section.

#### Scenario: Backlog with items

- **WHEN** the user runs `litespec list --backlog` and backlog.md contains items
- **THEN** output shows a "Backlog:" header, followed by items grouped under human-readable section labels (Deferred, Open Questions, Future, Other), each item prefixed with `▪`

#### Scenario: Empty or missing backlog

- **WHEN** the user runs `litespec list --backlog` and backlog.md is missing or has no items
- **THEN** output shows "Backlog:" followed by "  (none)"

### Requirement: mutual exclusivity

The `--backlog` flag SHALL be mutually exclusive with `--specs`, `--decisions`, and `--changes`. Combining `--backlog` with any of these flags produces an error.

#### Scenario: Backlog with specs

- **WHEN** the user runs `litespec list --backlog --specs`
- **THEN** the command returns an error: "--backlog is mutually exclusive with --specs and --decisions"

#### Scenario: Backlog with changes

- **WHEN** the user runs `litespec list --backlog --changes`
- **THEN** the command returns an error: "--backlog and --changes are mutually exclusive"

#### Scenario: Backlog with decisions

- **WHEN** the user runs `litespec list --backlog --decisions`
- **THEN** the command returns an error: "--backlog is mutually exclusive with --specs and --decisions"

### Requirement: JSON output

When `--backlog --json` is specified, the command SHALL output a JSON object with a `backlog` array. Each element contains `section` (the hyphenated section key: "deferred", "open-questions", "future", "other") and `title` (the extracted item title). The array is ordered by file position.

#### Scenario: JSON output with items

- **WHEN** the user runs `litespec list --backlog --json` and backlog.md has items
- **THEN** output is a JSON object `{ "backlog": [{ "section": "...", "title": "..." }, ...] }`

#### Scenario: JSON output empty

- **WHEN** the user runs `litespec list --backlog --json` and backlog.md is missing
- **THEN** output is a JSON object `{}` (the `backlog` key is omitted via `omitempty`)

### Requirement: shell completion registration

The `--backlog` flag SHALL be registered in the command spec registry (`CommandSpecs`) so that shell completion systems discover it.

#### Scenario: Completion for list command

- **WHEN** the user types `litespec list --` and triggers shell completion
- **THEN** `--backlog` appears in the completions alongside `--specs`, `--changes`, `--decisions`, etc.
