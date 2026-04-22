# preview

## ADDED Requirements

### Requirement: Preview Command

The CLI MUST provide a `litespec preview <change-name>` command. The command SHALL accept exactly one positional argument: the name of an active change. The command MUST NOT accept multiple change names. The command SHALL error with a clear message when the change name is empty, the change directory does not exist, or the change has already been archived.

#### Scenario: Preview an active change

- **WHEN** `litespec preview my-change` is run and `specs/changes/my-change/` exists with delta specs
- **THEN** a structural summary is printed to stdout

#### Scenario: Empty change name rejected

- **WHEN** `litespec preview` is run with no positional argument
- **THEN** an error is printed to stderr indicating a change name is required

#### Scenario: Non-existent change rejected

- **WHEN** `litespec preview missing-change` is run
- **THEN** an error is printed to stderr indicating the change was not found

#### Scenario: Archived change rejected

- **WHEN** `litespec preview archived-change` is run and the change exists only in `specs/changes/archive/`
- **THEN** an error is printed to stderr indicating archived changes cannot be previewed

### Requirement: Structural Summary Output

For each capability affected by the change, the preview output MUST display the capability name and its status (`NEW SPEC` if the capability does not exist in canon, otherwise `MODIFIED`). Under each capability, the output SHALL list every operation with its type (`ADDED`, `MODIFIED`, `REMOVED`, `RENAMED`) and the requirement name. Operations MUST be grouped by capability and ordered according to the canonical merge sequence: RENAMED first, then REMOVED, then MODIFIED, then ADDED.

#### Scenario: New capability summary

- **WHEN** `litespec preview add-auth` affects a capability not yet in canon
- **THEN** the capability is shown as `NEW SPEC` with all requirements listed as `ADDED`

#### Scenario: Modified capability summary

- **WHEN** `litespec preview add-auth` affects an existing capability with ADDED and MODIFIED requirements
- **THEN** the capability is shown as `MODIFIED` with operations listed in merge order

#### Scenario: Multiple capabilities

- **WHEN** a change affects three capabilities
- **THEN** each capability appears as a separate section in the output

### Requirement: Operation Counts Footer

The preview output MUST conclude with a summary line showing: the number of capabilities affected, the number of requirements added, the number modified, the number removed, and the number renamed.

#### Scenario: Mixed operations footer

- **WHEN** a change adds 3 requirements, modifies 2, removes 1, and renames 1 across 2 capabilities
- **THEN** the footer reads `2 capabilities affected • 3 added • 2 modified • 1 removed • 1 renamed`

#### Scenario: ADDED-only footer

- **WHEN** a change adds 5 requirements across 1 new capability
- **THEN** the footer reads `1 capability affected • 5 added • 0 modified • 0 removed • 0 renamed`

### Requirement: JSON Output

The `litespec preview` command MUST support a `--json` flag. When provided, the command SHALL output a single JSON object to stdout containing a `capabilities` array and a `totals` object. Each capability entry SHALL include `name`, `isNew`, and `operations` (an array of objects with `type` and `requirement` fields). The `totals` object SHALL include `capabilities`, `added`, `modified`, `removed`, and `renamed` as integers.

#### Scenario: JSON output for mixed change

- **WHEN** `litespec preview my-change --json` is run
- **THEN** valid JSON is emitted with `capabilities` and `totals` fields

#### Scenario: JSON output for empty change

- **WHEN** `litespec preview my-change --json` is run and the change has no delta specs
- **THEN** JSON with empty `capabilities` array and all-zero totals is emitted

### Requirement: No Side Effects

The preview command MUST NOT write to the canonical specs directory, the changes directory, the archive directory, or any temporary file that persists after the command exits. The command SHALL only read existing files and compute the merged result in memory.

#### Scenario: Canon unchanged after preview

- **WHEN** `litespec preview my-change` is run
- **THEN** no file in `specs/canon/` is modified

#### Scenario: Change directory unchanged after preview

- **WHEN** `litespec preview my-change` is run
- **THEN** the change directory remains in `specs/changes/my-change/` with all files intact

### Requirement: Empty Change Handling

When a change exists but contains no delta specs, the preview command SHALL emit the message "No changes to preview" (or the JSON equivalent) and exit with code 0.

#### Scenario: Change with no deltas

- **WHEN** `litespec preview my-change` is run and the change directory has no `specs/` subdirectory
- **THEN** the output indicates no changes to preview and exit code is 0

### Requirement: Merge Failure Reporting

If `PrepareArchiveWrites` or `MergeDelta` returns an error during preview, the command SHALL print the error message to stderr and exit with code 1. The error MUST NOT be silently swallowed.

#### Scenario: Conflicting operations detected

- **WHEN** a change contains conflicting delta operations that fail during merge
- **THEN** the error is printed to stderr and exit code is 1
