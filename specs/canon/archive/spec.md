# archive

## Requirements

### Requirement: Delta Merge into Canon

The `litespec archive <change>` command MUST merge all delta operations (RENAMED, REMOVED, MODIFIED, ADDED) from the change's specs into the canonical spec files at `specs/canon/<capability>/spec.md`. The merge order SHALL be RENAMED first, then REMOVED, then MODIFIED, then ADDED. If the capability does not yet exist in canon, a new directory and spec file MUST be created.

#### Scenario: Merge into existing capability

- **WHEN** `litespec archive add-rate-limit` is run and `specs/canon/auth/spec.md` already exists
- **THEN** delta operations are merged into the existing spec in the correct order

#### Scenario: Merge into new capability

- **WHEN** `litespec archive add-rate-limit` is run and `specs/canon/rate-limit/` does not exist
- **THEN** `specs/canon/rate-limit/spec.md` is created with the ADDED requirements

### Requirement: Strip Specs Subtree Before Archiving

After delta specs are merged into `specs/canon/`, the archive command MUST remove the change's `specs/` subtree before moving the change directory to `specs/changes/archive/`. The archived directory SHALL contain only planning artifacts (`.litespec.yaml`, `proposal.md`, `design.md`, `tasks.md`). The `specs/` subtree MUST NOT be present in the archived directory.

#### Scenario: Archived change has no specs subtree

- **WHEN** `litespec archive my-change` completes successfully
- **THEN** the archived directory at `specs/changes/archive/<date>-my-change/` does not contain a `specs/` subdirectory

#### Scenario: Canon contains merged content

- **WHEN** `litespec archive my-change` completes successfully
- **THEN** `specs/canon/<capability>/spec.md` contains the merged result of all delta operations

### Requirement: Canon Directory Naming

The canonical specs directory SHALL be named `canon/` and located at `<root>/specs/canon/`. The internal constant SHALL be `CanonDirName = "canon"` and the path function SHALL be `CanonPath(root)`. No code or path SHALL reference `specs/specs/`.

#### Scenario: Init creates canon directory

- **WHEN** `litespec init` is run in a new project
- **THEN** the directory `specs/canon/` is created

#### Scenario: Canon path function returns correct path

- **WHEN** `CanonPath("/project")` is called
- **THEN** the result is `/project/specs/canon`

### Requirement: Dependency Archive Guard

The `litespec archive` command SHALL check whether the change being archived has any active dependencies — i.e., other active changes that declare `dependsOn` referencing the change being archived. If an active dependency is found, the command SHALL block with an error indicating which changes depend on the current one. This check SHALL be bypassed with the `--allow-incomplete` flag.

#### Scenario: Archive change with no dependents

- **WHEN** `litespec archive add-auth` is run and no other change depends on add-auth
- **THEN** archiving proceeds normally

#### Scenario: Archive change with active dependent blocked

- **WHEN** `litespec archive add-auth` is run and change add-rate-limiting depends on add-auth
- **THEN** an error is reported: "change \"add-rate-limiting\" depends on \"add-auth\"; archive it first or use --allow-incomplete"

#### Scenario: Bypass with --allow-incomplete

- **WHEN** `litespec archive add-auth --allow-incomplete` is run and an active dependent exists
- **THEN** archiving proceeds with a warning about the active dependent
