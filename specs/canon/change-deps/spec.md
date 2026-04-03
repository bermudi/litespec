# change-deps

## Requirements

### Requirement: Dependency Metadata Field

The `ChangeMeta` struct SHALL support an optional `dependsOn` field containing a list of change names. When absent, the change has no dependencies. The field SHALL be read from and written to `.litespec.yaml` alongside existing `schema` and `created` fields.

#### Scenario: Metadata with dependencies

- **WHEN** a change's `.litespec.yaml` contains `dependsOn: [add-user-auth]`
- **THEN** the parsed `ChangeMeta` has `DependsOn` equal to `["add-user-auth"]`

#### Scenario: Metadata without dependencies

- **WHEN** a change's `.litespec.yaml` has no `dependsOn` field
- **THEN** the parsed `ChangeMeta` has `DependsOn` equal to `nil`

### Requirement: Dependency Resolution

The system SHALL resolve `dependsOn` references against both active changes and archived changes. Active changes take priority over archived changes when the same base name exists in both locations. A dependency is considered satisfied when the referenced change exists in the archive (meaning its deltas have already been merged into canon). An active dependency is not yet satisfied — the depended-upon change has not yet landed.

#### Scenario: Resolve to active change

- **WHEN** a change depends on "add-auth" and an active change named "add-auth" exists
- **THEN** the dependency resolves to the active change

#### Scenario: Resolve to archived change

- **WHEN** a change depends on "add-auth" and no active change exists but an archived "2026-04-01-add-auth" exists
- **THEN** the dependency resolves to the archived change

#### Scenario: Active takes priority over archived

- **WHEN** a change depends on "add-auth" and both an active "add-auth" and archived "2026-04-01-add-auth" exist
- **THEN** the dependency resolves to the active change

#### Scenario: Missing dependency

- **WHEN** a change depends on "nonexistent" and no active or archived change with that name exists
- **THEN** validation reports an error indicating the dependency target was not found

### Requirement: Cycle Detection

The system SHALL detect dependency cycles across active changes. When a cycle is detected (A depends on B, B depends on C, C depends on A), validation SHALL report an error listing the cycle path. Cycle detection SHALL run during `validate --all` and `validate --changes` modes (where all active changes are validated together). Single-change validation does not require cycle detection since it only examines one change's metadata.

#### Scenario: Simple cycle

- **WHEN** change A depends on B and change B depends on A
- **THEN** validation reports an error: "dependency cycle detected: A -> B -> A"

#### Scenario: Longer cycle

- **WHEN** change A depends on B, B depends on C, and C depends on A
- **THEN** validation reports an error identifying the full cycle path

#### Scenario: No cycle with valid chain

- **WHEN** change A depends on B and B depends on C (archived)
- **THEN** no cycle error is reported

### Requirement: Overlap Detection

When validating multiple active changes (`validate --all` or `validate --changes`), the system SHALL detect when two or more active changes target the same canonical spec with delta operations on overlapping requirements. An overlap warning SHALL be emitted for each pair of overlapping changes. The warning SHALL be suppressed when a `dependsOn` edge already exists between the two changes in either direction.

#### Scenario: Overlapping MODIFIED operations

- **WHEN** change A and change B both have MODIFIED requirements targeting the same canonical spec and no `dependsOn` edge between them
- **THEN** a warning is emitted indicating the overlap

#### Scenario: Overlap suppressed by dependency

- **WHEN** change A depends on change B and both target the same canonical spec
- **THEN** no overlap warning is emitted for the A-B pair

#### Scenario: No overlap across unrelated capabilities

- **WHEN** change A targets "validate" and change B targets "list"
- **THEN** no overlap warning is emitted
