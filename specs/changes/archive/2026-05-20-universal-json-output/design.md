# Design: Universal `--json` and `--minimal` Output

## Architecture

The change extends the existing `--json` pattern (used by `status`, `validate`, `list`, `instructions`, `preview`, `view`, `new`, `patch`) to the remaining data-producing commands. It also adds a new `--minimal` flag that works alongside `--json` or alone.

The pattern is already established:
1. Parse `--json` flag from args
2. Register in `checkUnknownFlags`
3. Execute command logic
4. If `asJSON`, marshal structured output and print; otherwise print text
5. Register the flag in `CommandSpecs` for shell completion

`--minimal` follows the same pattern: parse, pass through, and branch at output time.

## Decisions

### `--minimal` is a filter, not a format

`--minimal` doesn't change the data model — it filters the output. This means:
- `--minimal --json` produces a strict subset of `--json` fields
- `--minimal` alone produces terse text (no decorative formatting)
- Implementation: each command's JSON output path branches on a `minimal` bool

### Errors go to stderr only when pre-output; in-json otherwise

Current behavior: commands print errors to stderr. `--json` mode already reports errors in the JSON structure for `validate`. The same pattern applies: when `--json` is set, structural errors (validation failures, warnings) go into the JSON. Only catastrophic failures (can't find project root, flag parse error) go to stderr.

### Types live next to the command that owns them

Single-command JSON types go in `cmd/litespec/` as local structs — same pattern as `viewJSON` and friends in `view.go`. Only types shared across packages (e.g., `ValidationResultJSON` which is constructed in `internal/`) belong in `internal/json.go`. This change adds no new types to `internal/json.go`. The new JSON types for `init`, `archive`, `decide`, `update`, `upgrade`, and `import` are all defined in their respective command files.

### No generic minimal helper

Each command constructs its own minimal output using a local struct or anonymous struct literal. No reflection-based field stripping. Type-safe, grep-friendly, breaks at compile time if fields change.

### `--minimal` field selection per command

| Command | Full JSON fields | Minimal JSON fields |
|---|---|---|
| `validate` | `valid`, `errors`, `warnings`, `summary` | `valid`, `errors` |
| `status` | `changeName`, `schemaName`, `isComplete`, `mode`, `artifacts` | `changeName`, `isComplete`, `artifacts[].id`, `artifacts[].status` |
| `list` | `changes[]` with all fields | `changes[].name`, `changes[].status` |
| `view` | `summary`, `changes`, `specs`, `decisions`, `graph` | `summary` only |
| `instructions` | `artifactId`, `description`, `instruction`, `outputPath` | `artifactId`, `instruction` |
| `preview` | `capabilities[]`, `totals` | `totals` only |
| `init` | `initialized`, `directories`, `skills`, `adapters` | `initialized` |
| `archive` | `change`, `capabilities[]`, `archivedPath` | `archived: true`, `capabilities[]` |
| `decide` | `number`, `slug`, `title`, `filePath` | `number`, `slug`, `filePath` |
| `update` | `skillsUpdated`, `adapters[]` | `updated: true` |
| `upgrade` | `previousVersion`, `newVersion`, `upgraded`, `hint` | `previousVersion`, `newVersion`, `upgraded` |
| `import` | `canonSpecs`, `activeChanges`, `archives`, `warnings[]`, `skippedFiles` | `imported: true`, `canonSpecs`, `activeChanges`, `archives` |

For text-only `--minimal`, each command emits the tersest readable form: names only, one per line, no headers.

## File Changes

### `internal/commandspec.go`
- Add `--json` flag to `init`, `archive`, `decide`, `update`, `upgrade`, `import` command specs
- Add `--minimal` flag to every command that has `--json`

### `internal/json.go`
- No changes. New types go in `cmd/litespec/` command files.

### `cmd/litespec/helpers.go`
- Add `--minimal` to the `jsonFlag` constant or add a `minimalFlag` constant
- Add helper `parseOutputFlags(args) (asJSON, asMinimal bool)`
- Update help text functions for all affected commands

### `cmd/litespec/init.go`
- Define local `initResultJSON` and `initMinimalJSON` structs
- Parse `--json` and `--minimal` flags
- When `--json`: emit `initResultJSON`
- When `--minimal`: emit terse text

### `cmd/litespec/archive.go`
- Define local `archiveResultJSON` and `archiveMinimalJSON` structs
- Parse `--json` and `--minimal` flags
- When `--json`: emit `archiveResultJSON` after successful archive
- When `--minimal`: emit terse text

### `cmd/litespec/decide.go`
- Define local `decideResultJSON` struct
- Parse `--json` and `--minimal` flags
- When `--json`: emit `decideResultJSON`
- When `--minimal`: emit terse text

### `cmd/litespec/update.go`
- Define local `updateResultJSON` struct
- Parse `--json` and `--minimal` flags
- When `--json`: emit `updateResultJSON`
- When `--minimal`: emit terse text

### `cmd/litespec/upgrade.go`
- Define local `upgradeResultJSON` struct
- Parse `--json` and `--minimal` flags
- When `--json`: emit `upgradeResultJSON`
- When `--minimal`: emit terse text

### `cmd/litespec/import.go`
- Define local `importResultJSON` struct
- Parse `--json` and `--minimal` flags
- When `--json`: emit `importResultJSON`
- When `--minimal`: emit terse text

### `cmd/litespec/status.go`, `list.go`, `validate.go`, `instructions.go`, `view.go`, `preview.go`, `new.go`, `patch.go`
- Parse `--minimal` flag (already have `--json`)
- Branch at output time: if `--minimal`, emit subset

### `cmd/litespec/main_test.go`
- CLI tests for `--json` on each newly-enabled command
- CLI tests for `--minimal` on representative commands
- CLI tests for `--minimal --json` on representative commands
