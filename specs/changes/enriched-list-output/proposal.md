## Motivation

`litespec list` currently outputs bare names — no task progress, no timestamps, no alignment. This makes it impossible to assess change health at a glance. OpenSpec's `list` command provides a richer, column-aligned view that immediately shows what's complete, what's in progress, and how stale things are. We should match that output format.

## Scope

- Enrich `litespec list` to show task progress (`3/5 tasks`, `✓ Complete`, `No tasks`), relative last-modified time (`2h ago`), and column-aligned output for changes
- Enrich `litespec list --specs` to show requirement counts (`requirements 5`) with column-aligned output
- Add `--sort recent|name` flag for changes (default: `recent`). Specs always sorted alphabetically
- Enrich JSON output: changes get `completedTasks`, `totalTasks`, `lastModified`, `status` fields; specs get `requirementCount`
- Internal refactor: `ListChanges` and `ListSpecs` return enriched structs instead of bare `[]string`

## Non-Goals

- Changes to `status`, `validate`, or other commands
- Adding new CLI commands
- Modifying how task progress is computed (reuse existing `TaskCompletion`)
- Configurable column widths or output formatting
