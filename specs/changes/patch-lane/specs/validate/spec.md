# validate

## ADDED Requirements

### Requirement: Optional Planning Artifacts

The `ValidateChange` function MUST treat `proposal.md`, `design.md`, and `tasks.md` as optional. Their absence SHALL NOT produce an error. The presence of at least one valid delta spec under `specs/<capability>/` remains required and SHALL produce an error if missing. This makes patch-mode changes (delta-only) valid by construction.

#### Scenario: Patch-mode change validates without planning artifacts

- **WHEN** `litespec validate <name>` is run on a change containing only `specs/<capability>/spec.md` with valid delta content
- **THEN** validation succeeds with no errors

#### Scenario: Missing delta still fails

- **WHEN** `litespec validate <name>` is run on a change with no `specs/` directory or with a `specs/` directory containing no delta files
- **THEN** validation fails with an error indicating the change has no delta spec files

#### Scenario: Full proposal change still validates

- **WHEN** `litespec validate <name>` is run on a change with `proposal.md`, `design.md`, `tasks.md`, and a valid delta
- **THEN** validation succeeds (planning artifacts pass through their content checks)

### Requirement: Proposal Content Validation

When `proposal.md` exists, the `ValidateChange` function MUST verify it contains a `## Motivation` heading (or the legacy `## Why`) and a `## Scope` heading (or the legacy `## What Changes`), and that each of those sections has at least one non-blank body line before the next heading. Missing headings or empty sections SHALL produce errors identifying the file path.

#### Scenario: Proposal with required sections passes

- **WHEN** a change's `proposal.md` contains `## Motivation` with body text and `## Scope` with body text
- **THEN** validation passes for the proposal

#### Scenario: Proposal missing motivation heading fails

- **WHEN** a change's `proposal.md` lacks both `## Motivation` and `## Why` headings
- **THEN** validation reports an error indicating the proposal is missing the motivation section

#### Scenario: Proposal missing scope heading fails

- **WHEN** a change's `proposal.md` lacks both `## Scope` and `## What Changes` headings
- **THEN** validation reports an error indicating the proposal is missing the scope section

#### Scenario: Proposal section with no body fails

- **WHEN** a change's `proposal.md` contains `## Motivation` immediately followed by another heading with no body lines between them
- **THEN** validation reports an error indicating the motivation section is empty

#### Scenario: Empty proposal file fails

- **WHEN** a change's `proposal.md` exists but is empty or whitespace-only
- **THEN** validation reports errors for missing motivation and scope sections

### Requirement: Design Content Validation

When `design.md` exists, the `ValidateChange` function MUST verify it contains at least one `## ` heading and at least three non-blank content lines outside fenced code blocks. This catches stub files without prescribing structure. Failure SHALL produce an error identifying the file path.

#### Scenario: Design with content passes

- **WHEN** a change's `design.md` contains a `## Approach` heading and several non-blank content lines
- **THEN** validation passes for the design

#### Scenario: Empty design fails

- **WHEN** a change's `design.md` exists but is empty or whitespace-only
- **THEN** validation reports an error indicating the design appears to be a stub

#### Scenario: Design with only headings fails

- **WHEN** a change's `design.md` contains only headings with no body content outside code fences
- **THEN** validation reports an error indicating the design appears to be a stub

#### Scenario: Design content inside code fences does not count

- **WHEN** a change's `design.md` contains one heading and three non-blank lines all inside a fenced code block
- **THEN** validation reports an error indicating the design appears to be a stub

### Requirement: Empty-Phase Detection in Tasks

When `tasks.md` exists, the existing phase-heading and checkbox checks MUST be extended so that every `## Phase` block contains at least one checkbox line (`- [ ]` or `- [x]`). A phase with no checkboxes SHALL produce an error identifying the file path and the phase heading.

#### Scenario: Phase with no checkboxes fails

- **WHEN** a change's `tasks.md` contains `## Phase 1` followed by another `## Phase 2` heading with no checkboxes between them
- **THEN** validation reports an error identifying the empty phase

#### Scenario: Phase with at least one checkbox passes

- **WHEN** a change's `tasks.md` contains `## Phase 1` followed by `- [ ] do thing` and then `## Phase 2`
- **THEN** validation passes for that phase
