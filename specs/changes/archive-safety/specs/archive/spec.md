# archive

## MODIFIED Requirements

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

## ADDED Requirements

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
