## Motivation

Users migrating from OpenSpec to litespec must manually restructure their entire project — moving directories, renaming metadata files, converting date formats, and cleaning up spec files that contain extra H2 sections litespec rejects. This friction discourages adoption and creates a high barrier for teams who want to switch.

## Scope

This change introduces an `import` command that converts an OpenSpec project structure to litespec format. The command:

- Detects OpenSpec project structure (`openspec/specs/`, `openspec/changes/`, `.openspec.yaml`)
- Moves canon specs from `openspec/specs/` to `specs/canon/`
- Moves changes from `openspec/changes/` to `specs/changes/` (excluding archive)
- Renames `.openspec.yaml` to `.litespec.yaml` with date format conversion (`2026-02-21` → `2026-02-21T00:00:00Z`)
- Strips extra H2 sections from spec files that litespec's parser rejects
- Normalizes H1 titles from descriptive text to capability names
- Converts `FROM:/TO:` rename syntax to `→` arrow format
- Suggests running `litespec update` to generate skills

The command operates on a specified source directory or defaults to the current project root.

## Non-Goals

- **Bidirectional sync** — this is a one-time migration, not an ongoing sync tool
- **OpenSpec config.yaml migration** — litespec has no config file; context and rules are dropped with a warning
- **project.md / AGENTS.md migration** — these have no litespec equivalent
- **Extended metadata fields** — `provides`, `requires`, `touches`, `parent` from `.openspec.yaml` are dropped since litespec only supports `dependsOn`
- **Archive migration** — archived changes are not migrated; only active changes move over
