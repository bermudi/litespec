# view
## ADDED Requirements
### Requirement: Dashboard Display

The system SHALL provide a `view` command that displays a dashboard overview of specs, changes, and their dependency relationships.

#### Scenario: Basic dashboard display

- **WHEN** user runs `litespec view`
- **THEN** system displays a formatted dashboard with sections for summary, active changes, and specifications

#### Scenario: No specs directory

- **WHEN** user runs `litespec view` in a directory without a specs directory
- **THEN** system displays error message indicating no specs directory was found

### Requirement: Dependency Graph Section

When any active change has a `dependsOn` field, the dashboard SHALL display a dependency graph section showing the DAG of active changes. The graph SHALL use tree-style indentation with box-drawing characters. Changes with no dependencies SHALL appear as roots. Changes with no `dependsOn` field and no dependents SHALL be listed separately as unrelated changes.

#### Scenario: Simple dependency chain

- **WHEN** change A has no dependencies and change B depends on A
- **THEN** the graph section shows A as a root with B as its child

#### Scenario: Multiple roots

- **WHEN** change A and change C have no dependencies and change B depends on A
- **THEN** the graph shows A (with child B) and C as separate roots

#### Scenario: No dependencies at all

- **WHEN** no active change has `dependsOn` set
- **THEN** the dependency graph section is omitted entirely

### Requirement: Summary Section

The dashboard SHALL display a summary section with key project metrics: total specification count, total requirement count across all specs, active change count, draft change count (no tasks), and overall task completion percentage.

#### Scenario: Complete summary

- **WHEN** dashboard is rendered with specs and changes
- **THEN** system shows total specs, total requirements, active changes, draft changes, and task completion percentage

#### Scenario: Empty project

- **WHEN** no specs or changes exist
- **THEN** summary shows zero counts for all metrics
