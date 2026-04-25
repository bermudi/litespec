# archive

## Requirements

### Requirement: Delta Merge into Canon

The `litespec archive <change>` command MUST merge all delta operations (RENAMED, REMOVED, MODIFIED, ADDED) from the change's specs into the canonical spec files at `specs/canon/<capability>/spec.md`. Delta spec files within a capability directory SHALL be sorted lexicographically by filename before merging to ensure deterministic results. The merge order SHALL be RENAMED first, then REMOVED, then MODIFIED, then ADDED. If the capability does not yet exist in canon, a new directory and spec file MUST be created.

#### Scenario: Merge into existing capability

- **WHEN** `litespec archive add-rate-limit` is run and `specs/canon/auth/spec.md` already exists
- **THEN** delta operations are merged into the existing spec in the correct order

#### Scenario: Merge into new capability

- **WHEN** `litespec archive add-rate-limit` is run and `specs/canon/rate-limit/` does not exist
- **THEN** `specs/canon/rate-limit/spec.md` is created with the ADDED requirements

#### Scenario: Deterministic merge order

- **WHEN** a change has multiple delta spec files (`01-auth.md`, `02-roles.md`) in the same capability
- **THEN** they are merged in lexicographic filename order regardless of filesystem enumeration order

### Requirement: Strip Specs Subtree Before Archiving

After delta specs are merged into `specs/canon/`, the archive command MUST remove the change's `specs/` subtree before moving the change directory to `specs/changes/archive/`. The archived directory SHALL contain whichever planning artifacts existed in the source change directory plus `.litespec.yaml` if present. The `specs/` subtree MUST NOT be present in the archived directory. For full-proposal changes the archived directory typically contains `.litespec.yaml`, `proposal.md`, `design.md`, and `tasks.md`. For patch-mode changes (no planning artifacts present at archive time), the archived directory MAY contain only the merged spec history with no planning files; this is not an error.

#### Scenario: Archived full-proposal change has no specs subtree

- **WHEN** `litespec archive my-change` completes successfully on a change with `proposal.md`, `design.md`, `tasks.md`
- **THEN** the archived directory at `specs/changes/archive/<date>-my-change/` contains those planning files and does not contain a `specs/` subdirectory

#### Scenario: Archived patch-mode change has no specs subtree and no planning files

- **WHEN** `litespec archive my-patch` completes successfully on a patch-mode change (delta-only, no `proposal.md`)
- **THEN** the archived directory at `specs/changes/archive/<date>-my-patch/` does not contain a `specs/` subdirectory and is not required to contain `proposal.md`, `design.md`, or `tasks.md`

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

### Requirement: Atomic Archive Writes

The archive command MUST write merged specs to temporary files first, verify they are parseable, and only then move them to the canon directory. All capabilities in a single archive operation SHALL be treated as one unit — if any capability's write or verification fails, all canon writes SHALL be rolled back. The change directory SHALL only be archived after all canon writes succeed.

#### Scenario: Successful atomic archive

- **WHEN** `litespec archive my-change` is run and all merges succeed
- **THEN** canon specs are updated and the change directory is moved to archive

#### Scenario: Merge failure leaves canon unchanged

- **WHEN** `litespec archive my-change` is run and a merge produces an invalid spec
- **THEN** the original canon specs remain unchanged, temporary files are cleaned up, and an error is reported

#### Scenario: Write failure leaves canon unchanged

- **WHEN** `litespec archive my-change` is run and writing a canon spec fails (e.g., permissions)
- **THEN** any already-written canon specs are rolled back, the change directory remains in `changes/`, and an error is reported

### Requirement: Post-Archive Verification

After archiving a change, the command MUST verify that the archived directory exists in `specs/changes/archive/` and that the canon specs for all affected capabilities are parseable. If verification fails, an error SHALL be reported with details about what failed.

#### Scenario: Archive verification succeeds

- **WHEN** `litespec archive my-change` completes
- **THEN** the command verifies the archived directory exists and canon specs parse successfully

#### Scenario: Missing archived directory detected

- **WHEN** the rename to archive succeeds but the directory is not found (unlikely but defensive)
- **THEN** an error is reported indicating the archive directory is missing
