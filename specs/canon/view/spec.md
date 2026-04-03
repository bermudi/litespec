# view

## Requirements

### Requirement: Dashboard Display

The dashboard SHALL display a formatted dashboard with a title header, box-drawing separators (`══` for outer border, `──` for section underlines), and distinct sections for summary, active changes, draft changes, completed changes, specifications, and optionally a dependency graph.

#### Scenario: Basic dashboard display

- **WHEN** user runs `litespec view`
- **THEN** system displays a formatted dashboard with sections for summary, active changes, draft changes, completed changes, and specifications

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

The dashboard SHALL display a summary section with key project metrics: total specification count and requirement count combined on one line, draft change count (no tasks), active change count (tasks in progress), completed change count (all tasks done), and overall task progress as completed/total with percentage.

#### Scenario: Complete summary

- **WHEN** dashboard is rendered with specs and changes
- **THEN** system shows specifications count with requirement count, draft changes, active changes, completed changes, and task progress with percentage

#### Scenario: Empty project

- **WHEN** no specs or changes exist
- **THEN** summary shows zero counts for all metrics

### Requirement: Progress Bars

For each active change (tasks in progress), the dashboard SHALL display a Unicode progress bar using `█` for filled segments and `░` for empty segments, enclosed in brackets, followed by the completion percentage. Changes SHALL be sorted by completion percentage ascending, then alphabetically.

#### Scenario: Half-complete change

- **WHEN** a change has 1 of 2 tasks completed
- **THEN** the progress bar shows approximately half filled `[██████████████████░░] 50%`

#### Scenario: Change categorization

- **WHEN** changes exist with no tasks (draft), partial tasks (active), and all tasks complete
- **THEN** draft changes appear in a Draft Changes section with `○` bullet, active changes appear in Active Changes section with `◉` bullet and progress bar, completed changes appear in Completed Changes section with `✓` bullet

### Requirement: Specifications Section

The dashboard SHALL display specifications sorted by requirement count descending. Each spec SHALL show its name padded to 30 characters followed by the requirement count with appropriate singular/plural label.

#### Scenario: Spec display

- **WHEN** dashboard is rendered with specs
- **THEN** specs are listed with `▪` bullet, sorted by requirement count descending, with padded names and requirement counts

### Requirement: Dashboard Footer

The dashboard SHALL display a closing `══` border and a hint line directing users to `litespec list --changes` or `litespec list --specs` for detailed views.

#### Scenario: Footer display

- **WHEN** dashboard is rendered
- **THEN** output ends with `══` border and hint text about list commands
