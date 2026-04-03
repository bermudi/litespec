# validate
## ADDED Requirements
### Requirement: Dependency Validation

The `ValidateChange` function SHALL validate `dependsOn` references when present in a change's `.litespec.yaml`. Each dependency name SHALL be resolved against active changes first, then archived changes. Unresolvable references SHALL produce an error. When validating multiple changes via `ValidateAll`, cycle detection SHALL run across all active changes' dependency graphs and report cycle paths as errors.

#### Scenario: Valid dependency on active change

- **WHEN** change A declares `dependsOn: [B]` and B is an active change
- **THEN** validation passes for the dependency reference

#### Scenario: Valid dependency on archived change

- **WHEN** change A declares `dependsOn: [B]` and B exists only in archive
- **THEN** validation passes for the dependency reference

#### Scenario: Invalid dependency reference

- **WHEN** change A declares `dependsOn: [nonexistent]` and no active or archived change matches
- **THEN** an error is reported: "dependency \"nonexistent\" not found"

#### Scenario: Cycle detected during bulk validation

- **WHEN** `validate --all` is run and a dependency cycle exists among active changes
- **THEN** an error is reported identifying the cycle path
