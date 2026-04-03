## Motivation

The CLI silently swallows errors in several critical paths: JSON marshal failures produce empty output, invalid change names can create dangerous directories, and unsupported `--tools` values fail late with confusing messages. These rough edges erode trust in the tool's output, especially when consumed programmatically via `--json`.

## Scope

- Validate change names at creation time (reject empty, path separators, traversal sequences)
- Validate `--tools` flag values against registered adapters
- Surface JSON marshal errors instead of silently ignoring them
- Report skipped changes during status/list instead of silently omitting them
- Harmonize error message format across all commands (all errors to stderr with consistent prefix)
- Derive supported tools list from the Adapters registry instead of hardcoding

## Non-Goals

- Restructuring the CLI into separate command files (separate proposal)
- Adding `context.Context` support (separate proposal)
- Changing the flag parsing approach (no cobra/urfave migration)
