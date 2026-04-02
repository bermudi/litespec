# Proposal: Positional Names

## Motivation

Three commands (`validate`, `status`, `instructions`) use `--change <name>` to target a specific change. This is verbose and inconsistent with the reference OpenSpec implementation, which uses positional arguments. It also creates a friction point for shell completions — tab-completing a change name requires knowing you're after a flag, not a positional argument.

Additionally, `litespec instructions` requires `--change` even though artifact instruction templates are static and change-independent. The AI only needs the guidance, not computed context. The `--change` flag was artificial overhead.

## Scope

- **`validate <name>`** — accept a positional name, auto-detect whether it's a change or spec by checking both lists. Add `--changes`, `--specs`, `--all` bulk flags. Add `--type change|spec` for disambiguation when a name collides across both namespaces. Remove `--change` flag.
- **`status <name>`** — accept a positional name instead of `--change <name>`. Remove `--change` flag.
- **`instructions <artifact>`** — remove `--change` requirement. The command returns static artifact creation guidance. Keep `--json` for structured output. Valid artifacts remain: `proposal`, `specs`, `design`, `tasks`.
- **`validate` internal** — add a new `ValidateSpec(root, name)` function to validate a single spec by name (currently only `ValidateSpecs` validates all specs).
- **Delta specs** — modify the existing `validate` spec with MODIFIED/ADDED requirements. Create new `status` and `instructions` capability specs.

## Ordering

This change MUST be archived before `shell-completions`. The shell-completions change references `--change` flag positions that this proposal removes. After archiving `positional-names`, the `shell-completions` change must be updated to reference positional arguments instead.

## Breaking Changes

The `instructions --json` output format changes. The fields `changeName`, `changeDir`, `schemaName`, `dependencies`, and `unlocks` are removed. The new output contains `artifactId`, `description`, `instruction`, `template`, and `outputPath`.

## Non-Goals

- Renaming the `specs/specs/` directory structure (separate concern).
- Changing `list`, `archive`, `new`, `init`, or `update` commands.
- Adding interactive selectors (no TTY detection, no inquirer-style prompts).
- Changing the validate command's existing JSON output structure (the shape remains the same, just covers more modes).
