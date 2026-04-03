# validate

## MODIFIED Requirements

### Requirement: Dependency Validation

The `ValidateChange` function SHALL validate `dependsOn` references when present in a change's `.litespec.yaml`. Each dependency name SHALL be resolved against active changes first, then archived changes. Unresolvable references SHALL produce an error that includes the source file path. When validating multiple changes via `ValidateAll`, cycle detection SHALL run across all active changes' dependency graphs and report cycle paths as errors.

#### Scenario: Valid dependency on active change

- **WHEN** change A declares `dependsOn: [B]` and B is an active change
- **THEN** validation passes for the dependency reference

#### Scenario: Valid dependency on archived change

- **WHEN** change A declares `dependsOn: [B]` and B exists only in archive
- **THEN** validation passes for the dependency reference

#### Scenario: Invalid dependency reference

- **WHEN** change A declares `dependsOn: [nonexistent]` and no active or archived change matches
- **THEN** an error is reported with the `.litespec.yaml` file path: "dependency \"nonexistent\" not found"

#### Scenario: Cycle detected during bulk validation

- **WHEN** `validate --all` is run and a dependency cycle exists among active changes
- **THEN** an error is reported identifying the cycle path

## ADDED Requirements

### Requirement: Empty Name Rejection

The validation system MUST reject requirement and scenario names that are empty or contain only whitespace. This applies to both canonical specs and delta specs. An empty name SHALL produce an error indicating the file and the nature of the empty name.

#### Scenario: Empty requirement name in delta spec

- **WHEN** a delta spec contains `### Requirement:` with no name
- **THEN** validation reports an error: "empty requirement name" with the file path

#### Scenario: Whitespace-only requirement name

- **WHEN** a delta spec contains `### Requirement:   ` with only spaces
- **THEN** validation reports an error: "empty requirement name" with the file path

#### Scenario: Empty scenario name

- **WHEN** a requirement contains `#### Scenario:` with no name
- **THEN** validation reports an error: "empty scenario name in requirement <name>" with the file path

#### Scenario: Whitespace-only scenario name

- **WHEN** a requirement contains `#### Scenario:   ` with only spaces
- **THEN** validation reports an error: "empty scenario name in requirement <name>" with the file path

### Requirement: Duplicate Name Detection

The validation system MUST detect duplicate requirement names within a single delta spec file and duplicate scenario names within a single requirement. Duplicates SHALL produce an error identifying both the original and duplicate name.

#### Scenario: Duplicate requirement names in single delta

- **WHEN** a delta spec file contains two ADDED requirements both named "Login"
- **THEN** validation reports an error: "duplicate requirement name \"Login\"" with the file path

#### Scenario: Duplicate scenario names in single requirement

- **WHEN** a requirement contains two scenarios both named "happy path"
- **THEN** validation reports an error: "duplicate scenario name \"happy path\" in requirement <name>" with the file path

### Requirement: Scenario Content Validation

ADDED and MODIFIED requirements MUST have at least one scenario whose content contains both `WHEN` and `THEN` markers as plain text (bold formatting is not required). Markers MAY appear in any order within the scenario body. Scenarios with empty content SHALL produce an error. The marker check SHALL use case-sensitive substring matching.

#### Scenario: Scenario without WHEN/THEN content

- **WHEN** an ADDED requirement has a scenario with empty body
- **THEN** validation reports an error indicating the scenario must contain WHEN and THEN

#### Scenario: Scenario with valid WHEN/THEN content

- **WHEN** an ADDED requirement has a scenario with "WHEN ... THEN ..."
- **THEN** validation passes for that scenario

### Requirement: Whole-Word Keyword Matching

The SHALL/MUST keyword check in requirement content MUST match whole words only. Keywords appearing inside fenced code blocks (```...```), inline code (`` `...` ``), or as substrings of other words SHALL NOT satisfy the requirement. The check SHALL strip code spans before applying word boundary detection.

#### Scenario: SHALL inside code block not accepted

- **WHEN** an ADDED requirement's only "SHALL" appears inside a fenced code block
- **THEN** validation reports an error that the requirement must contain SHALL or MUST

#### Scenario: SHALL inside inline code not accepted

- **WHEN** an ADDED requirement's only "SHALL" appears inside backtick inline code
- **THEN** validation reports an error that the requirement must contain SHALL or MUST

#### Scenario: SHALL as whole word accepted

- **WHEN** an ADDED requirement contains "The system SHALL do X" outside code blocks
- **THEN** validation passes for the keyword check

#### Scenario: SHALL as substring not accepted

- **WHEN** an ADDED requirement contains "MARSHALL" but no standalone "SHALL"
- **THEN** validation reports an error that the requirement must contain SHALL or MUST

### Requirement: Cross-Operation Conflict Detection

The validation system MUST detect conflicting operations on the same requirement within a single delta spec. If a requirement appears in more than one operation section (e.g., both MODIFIED and REMOVED), an error SHALL be reported. Additionally, RENAMED operations SHALL be checked for conflicts using both the old name (against MODIFIED/REMOVED) and the new name (against ADDED).

#### Scenario: MODIFIED and REMOVED on same requirement

- **WHEN** a delta spec MODIFIES requirement "Login" and also REMOVES requirement "Login"
- **THEN** validation reports an error: "conflicting operations on requirement \"Login\""

#### Scenario: RENAMED old name conflicts with MODIFIED

- **WHEN** a delta spec RENAMES "Login"→"Auth" and also MODIFIES "Login"
- **THEN** validation reports an error about conflicting operations on "Login"

#### Scenario: RENAMED new name conflicts with ADDED

- **WHEN** a delta spec RENAMES "Login"→"Auth" and also ADDS "Auth"
- **THEN** validation reports an error about conflicting operations on "Auth"

### Requirement: RENAMED Overlap Uses Old Name

The `DetectOverlaps` function in `deps.go` MUST use the RENAMED requirement's `OldName` field when checking for overlaps with MODIFIED/REMOVED operations in other changes. This ensures that if change A renames "Login"→"Auth" and change B modifies "Login", the overlap is detected.

#### Scenario: RENAMED overlaps with MODIFIED in another change

- **WHEN** change A RENAMES "Login"→"Auth" and change B MODIFIES "Login" in the same capability
- **THEN** a warning is reported about the overlap on "Login"

#### Scenario: RENAMED does not overlap with its new name

- **WHEN** change A RENAMES "Login"→"Auth" and change B MODIFIES "Auth" in the same capability
- **THEN** no overlap warning is reported (B modifies the new name, which is valid after A archives)
