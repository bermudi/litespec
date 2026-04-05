## Motivation

Users migrating from OpenSpec to litespec must manually restructure their entire project — moving directories, renaming metadata files, and converting date formats. This friction discourages adoption and creates a high barrier for teams who want to switch.

## Scope

This change introduces an `import` command that converts an OpenSpec project structure to litespec format. The command:

- Detects OpenSpec project structure (`openspec/specs/` or `openspec/changes/`)
- Moves canon specs from `openspec/specs/` to `specs/canon/` (strips " Specification" suffix from H1 titles)
- Moves changes from `openspec/changes/` to `specs/changes/` (including archive, with `specs/` subtrees stripped from archived changes; archives without `.openspec.yaml` get synthesized metadata from directory name)
- Renames `.openspec.yaml` to `.litespec.yaml` with date format conversion (handles both quoted and unquoted dates)
- Normalizes task phase labels from `## N. Name` to `## Phase N: Name`
- Warns about skipped files and directories (`config.yaml`, `project.md`, `AGENTS.md`, `explorations/`, loose files in changes directory)
- Suggests running `litespec update` to generate skills

The command accepts a `--source` flag to specify the OpenSpec project directory, defaulting to the current working directory. For imported projects, `import` replaces `init` — the imported structure already satisfies project initialization.

## Non-Goals

- **Bidirectional sync** — this is a one-time migration, not an ongoing sync tool
- **OpenSpec config.yaml migration** — litespec has no config file; context and rules are dropped with a warning
- **project.md / AGENTS.md migration** — these have no litespec equivalent
- **openspec/explorations/ migration** — this directory has no litespec equivalent and is skipped with a warning
- **Extended metadata fields** — `provides`, `requires`, `touches`, `parent` from `.openspec.yaml` are dropped since litespec only supports `dependsOn`
- **Content transformation** — spec file content is preserved; the only normalization is stripping " Specification" from H1 titles and converting task phase labels
- **Non-standard H2 sections** — if a spec contains H2 sections other than `## Purpose` or `## Requirements`, it will fail litespec validation after import and must be fixed manually
