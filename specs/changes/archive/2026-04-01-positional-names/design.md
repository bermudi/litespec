# Design: Positional Names

## Architecture

Three commands change their argument surface. No new packages. No new dependencies. The change is confined to argument parsing in `cmd/litespec/main.go` and one new function in `internal/validate.go`.

```
Before:
  litespec validate [--change <name>] [--all] [--strict] [--json]
  litespec status [--change <name>] [--json]
  litespec instructions <artifact> [--change <name>] [--json]

After:
  litespec validate [<name>] [--all] [--changes] [--specs] [--type change|spec] [--strict] [--json]
  litespec status [<name>] [--json]
  litespec instructions <artifact> [--json]
```

## Decisions

1. **Positional names are optional** — `validate` and `status` default to showing everything when no name is given. This preserves backward compatibility with `litespec validate` (no args) meaning "validate all".

2. **Auto-detection via list intersection** — `cmdValidate` calls both `ListChanges()` and `ListSpecs()`, checks membership, and resolves type. If both match and no `--type` is given, it errors with a clear message.

3. **`ValidateSpec(root, name)` is new** — currently only `ValidateSpecs(root)` (plural, validates all) exists. The singular form validates one spec by name, mirroring `ValidateChange(root, name)`. It reads `specs/specs/<name>/spec.md`, parses it, and checks for errors.

4. **`--change` is fully removed** — no backward compat alias. The flag was only in the CLI for a few versions and the user base is the developer. Clean break.

5. **`instructions` JSON output changes shape** — without `--change`, the JSON no longer includes `changeName`, `changeDir`, `schemaName`, `dependencies`, or `unlocks`. It returns `artifactId`, `description`, `instruction`, and `template`. The `outputPath` becomes the conventional path pattern (e.g. `proposal.md`) rather than a resolved absolute path. This is a breaking change acknowledged in the proposal.

6. **Bulk flags are combinable** — `--changes` and `--specs` can be used together (equivalent to `--all`). All bulk flags are mutually exclusive with a positional name. `--type` requires a positional name and is mutually exclusive with bulk flags.

7. **Separate capability specs** — validate, status, and instructions each get their own delta spec directory rather than stuffing all requirements into the validate spec.

## File Changes

| File | Change |
|------|--------|
| `cmd/litespec/main.go` | Rewrite `cmdValidate()`, `cmdStatus()`, `cmdInstructions()`. Update `printUsage()`. Remove all `--change` parsing. |
| `internal/validate.go` | Add `ValidateSpec(root, name string) (*ValidationResult, error)`. |
| `internal/json.go` | Add `BuildArtifactInstructionsStandaloneJSON(artifactID string)`. Remove change-specific fields from artifact instructions flow. |
| `DESIGN.md` | Update CLI commands table to reflect new argument surfaces. |
| `specs/specs/validate/spec.md` | Updated via delta merge at archive time. |
| `specs/specs/status/spec.md` | New capability spec created at archive time. |
| `specs/specs/instructions/spec.md` | New capability spec created at archive time. |
