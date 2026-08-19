# decisions

## Requirements

### Requirement: Decisions Directory

The project SHALL support an optional `specs/decisions/` directory containing architectural decision records. Each decision SHALL be a single markdown file named `NNNN-<kebab-name>.md` where `NNNN` is a zero-padded four-digit sequence number (e.g., `0001-single-shared-workspace.md`). The directory's absence SHALL NOT be an error — decisions are opt-in. Decision files SHALL NOT be moved, renamed, or deleted during archive or any other litespec operation.

#### Scenario: Decisions directory is optional

- **WHEN** a project has no `specs/decisions/` directory
- **THEN** all litespec commands operate normally with no error or warning

#### Scenario: Decision file naming convention

- **WHEN** a decision file is created with name `0003-beta-tools-session-bound.md`
- **THEN** the number is parsed as `3` and the slug is `beta-tools-session-bound`

#### Scenario: Decisions survive archive

- **WHEN** a change citing a decision slug is archived
- **THEN** the decision file remains in `specs/decisions/` unchanged

### Requirement: Decision File Structure

A decision file MUST contain an H1 title, a `## Status` section with a single-word status value, a `## Context` section, a `## Decision` section, and a `## Consequences` section. The status value SHALL be one of `proposed`, `accepted`, or `superseded`. Decision files MAY include an optional `## Supersedes` or `## Superseded-By` section containing one or more decision slug references (as markdown list items or inline). Decision files MUST NOT contain `## Requirements` or other reserved spec-format headers — they are narrative documents, not specs.

#### Scenario: Valid decision structure

- **WHEN** a decision file contains `# Title`, `## Status`, `## Context`, `## Decision`, `## Consequences`
- **THEN** parsing succeeds and all four required sections are populated

#### Scenario: Missing required section

- **WHEN** a decision file lacks `## Consequences`
- **THEN** parsing returns an error identifying the missing section

#### Scenario: Unrecognized status value

- **WHEN** a decision file declares status `draft`
- **THEN** parsing returns an error listing the allowed values

#### Scenario: Supersede pointers parsed

- **WHEN** a decision file contains `## Superseded-By` with item `0007-new-workspace-model`
- **THEN** the parsed decision exposes `SupersededBy = ["0007-new-workspace-model"]`

### Requirement: Supersede Linking

When a decision declares `## Superseded-By: <target-slug>`, the validator SHALL check that the target decision exists. When a decision declares `## Supersedes: <target-slug>`, the validator SHALL check that the target decision exists and has status `superseded`. A decision with status `superseded` SHALL have a `## Superseded-By` pointer to a non-superseded decision. These checks produce validation errors, not warnings.

#### Scenario: Supersede pointer resolves

- **WHEN** decision `0005` declares `## Supersedes: 0002-old-model` and file `0002-old-model.md` exists with status `superseded`
- **THEN** validation passes

#### Scenario: Supersede pointer dangling

- **WHEN** decision `0005` declares `## Supersedes: nonexistent`
- **THEN** validation reports an error identifying the dangling pointer

#### Scenario: Superseded decision lacks forward pointer

- **WHEN** decision `0002-old-model` has status `superseded` but no `## Superseded-By` section
- **THEN** validation reports an error indicating every superseded decision must point forward
