# archive
## ADDED Requirements
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
