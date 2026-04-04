# review

## Requirements

### Requirement: Context-Aware Review Mode

The review skill MUST detect the task completion state of the target change by reading `tasks.md` and distinguish three states: all unchecked, partially complete, and fully complete. When zero tasks are checked (including when tasks.md has no checkboxes at all), review SHALL operate in artifact review mode — evaluating proposal, specs, design, and tasks for quality, consistency, and implementation readiness without reading implementation code. When some but not all tasks are checked, review SHALL operate in implementation review mode — comparing implemented code against specs. When all tasks are checked, review SHALL operate in pre-archive review mode — reviewing both artifacts and implementation comprehensively before archiving.

#### Scenario: Artifact review on unplanned change

- **WHEN** `review` is invoked on a change where zero tasks are checked
- **THEN** the skill reviews proposal, specs, design, and tasks for quality, consistency, and gaps without reading implementation code

#### Scenario: Artifact review on empty tasks.md

- **WHEN** `review` is invoked on a change where tasks.md has no checkbox lines
- **THEN** the skill operates in artifact review mode (zero checked of zero total)

#### Scenario: Implementation review on partially implemented change

- **WHEN** `review` is invoked on a change where some but not all tasks are checked
- **THEN** the skill reviews implementation code against specs using the current behavior

#### Scenario: Pre-archive review on fully implemented change

- **WHEN** `review` is invoked on a change where all tasks are checked
- **THEN** the skill reviews both artifacts and implementation comprehensively, catching issues before archive

### Requirement: Artifact Review Dimensions

When operating in artifact review mode, review MUST evaluate planning artifacts across three dimensions: completeness (all artifacts present, specs cover the scope, scenarios are testable), consistency (proposal scope matches specs, design matches specs, tasks cover design, non-goals are respected), and readiness (scenarios describe verifiable behavior, design is concrete with file paths, tasks are phased correctly with clear boundaries). Artifact review goes beyond structural validation — it applies judgment to catch gaps that `litespec validate` cannot detect, such as vague requirements, missing edge cases, or design decisions that contradict the proposal.

#### Scenario: Proposal non-goal contradicted by spec

- **WHEN** the proposal lists something as a non-goal but a spec requirement implements it
- **THEN** the artifact review flags this as a consistency issue

#### Scenario: Vague requirement with no verifiable scenario

- **WHEN** a requirement's scenarios do not describe concrete WHEN/THEN conditions
- **THEN** the artifact review flags this as a readiness issue

#### Scenario: Design decision contradicts proposal scope

- **WHEN** design.md introduces an approach that conflicts with the proposal's stated scope
- **THEN** the artifact review flags this as a consistency issue

### Requirement: Updated Skill Description

The review skill description in `internal/paths.go` MUST be updated to reflect the context-aware behavior. The description SHALL mention artifact review (pre-implementation), implementation review (during implementation), and pre-archive review (post-implementation).

#### Scenario: Skill description mentions all modes

- **WHEN** the review skill is listed or generated
- **THEN** the description references artifact review, implementation review, and pre-archive review
