# validate

## MODIFIED Requirements

### Requirement: JSON Output for Validate

The `litespec validate` command MUST support a `--json` flag that returns structured JSON output containing a `valid` boolean, `errors` array, and `warnings` array. Each issue MUST include `severity`, `message`, and `file` fields. This applies to all validate modes: positional name, bulk flags, and default (no arguments).

#### Scenario: Validate single change with JSON flag

- **WHEN** `litespec validate <change-name> --json` is run
- **THEN** output is valid JSON with `valid`, `errors`, and `warnings` fields

#### Scenario: Validate all with JSON flag

- **WHEN** `litespec validate --all --json` is run
- **THEN** output is valid JSON with `valid`, `errors`, and `warnings` fields covering all changes and specs

#### Scenario: Validate specs with JSON flag

- **WHEN** `litespec validate --specs --json` is run
- **THEN** output is valid JSON with `valid`, `errors`, and `warnings` fields covering only specs

#### Scenario: Validate changes with JSON flag

- **WHEN** `litespec validate --changes --json` is run
- **THEN** output is valid JSON with `valid`, `errors`, and `warnings` fields covering only changes

### Requirement: Consistent JSON Flag Convention

All litespec commands that support `--json` MUST use the same flag name and return valid JSON to stdout. Each command's JSON output MUST reflect its current argument surface (positional names, bulk flags, etc.).

#### Scenario: JSON flag consistency across commands

- **WHEN** any litespec command is run with `--json`
- **THEN** the output is valid JSON to stdout

## ADDED Requirements

### Requirement: Positional Name Argument

The `litespec validate` command MUST accept an optional positional `<name>` argument. When provided, the command SHALL auto-detect whether the name refers to a change or a spec by checking both `ListChanges()` and `ListSpecs()`. The `--change` flag SHALL be removed.

#### Scenario: Validate a named change

- **WHEN** `litespec validate my-feature` is run and `my-feature` exists as a change
- **THEN** only that change is validated

#### Scenario: Validate a named spec

- **WHEN** `litespec validate auth` is run and `auth` exists as a spec
- **THEN** only that spec is validated

#### Scenario: Unknown name

- **WHEN** `litespec validate nonexistent` is run and the name matches neither a change nor a spec
- **THEN** an error is printed to stderr with exit code 1

#### Scenario: No name and no flags

- **WHEN** `litespec validate` is run with no arguments
- **THEN** it behaves identically to `--all` (validates all changes and specs)

### Requirement: Type Disambiguation

When a positional name matches both a change and a spec, the command MUST report an ambiguity error. The user SHALL use `--type change` or `--type spec` to disambiguate. The `--type` flag MUST only be used with a positional name. Using `--type` without a positional name or with bulk flags SHALL produce an error.

#### Scenario: Ambiguous name without --type

- **WHEN** `litespec validate shared-name` is run and `shared-name` exists as both a change and a spec
- **THEN** an error is printed to stderr indicating the name is ambiguous and suggesting `--type`

#### Scenario: Ambiguous name with --type change

- **WHEN** `litespec validate shared-name --type change` is run
- **THEN** only the change is validated

#### Scenario: Ambiguous name with --type spec

- **WHEN** `litespec validate shared-name --type spec` is run
- **THEN** only the spec is validated

#### Scenario: --type without positional name

- **WHEN** `litespec validate --type change` is run with no positional name
- **THEN** an error is printed indicating `--type` requires a positional name

#### Scenario: --type with bulk flag

- **WHEN** `litespec validate --all --type change` is run
- **THEN** an error is printed indicating `--type` cannot be used with bulk flags

### Requirement: Bulk Validation Flags

The `litespec validate` command MUST support `--all`, `--changes`, and `--specs` flags. `--all` validates all changes and all specs. `--changes` validates all changes only. `--specs` validates all specs only. The flags MAY be combined — combining `--changes` and `--specs` is equivalent to `--all`. These flags are mutually exclusive with the positional `<name>` argument.

#### Scenario: Validate all

- **WHEN** `litespec validate --all` is run
- **THEN** all changes and all specs are validated

#### Scenario: Validate all changes

- **WHEN** `litespec validate --changes` is run
- **THEN** only changes are validated

#### Scenario: Validate all specs

- **WHEN** `litespec validate --specs` is run
- **THEN** only specs are validated

#### Scenario: Combined --changes and --specs

- **WHEN** `litespec validate --changes --specs` is run
- **THEN** all changes and all specs are validated, equivalent to `--all`

#### Scenario: Bulk flag with positional name

- **WHEN** `litespec validate my-change --all` is run
- **THEN** an error is printed indicating the positional name and bulk flags are mutually exclusive

### Requirement: Single Spec Validation

A `ValidateSpec(root, name)` function MUST exist in the internal package that validates a single spec by name. It is the singular counterpart to the existing `ValidateSpecs(root)` which validates all specs. It SHALL read and parse the spec at `specs/specs/<name>/spec.md`, validate its structure and requirements, and return a `*ValidationResult`.

#### Scenario: Validate existing spec

- **WHEN** `ValidateSpec(root, "auth")` is called and the spec exists
- **THEN** a ValidationResult is returned reflecting the spec's validity

#### Scenario: Validate missing spec

- **WHEN** `ValidateSpec(root, "nonexistent")` is called
- **THEN** an error is returned indicating the spec was not found
