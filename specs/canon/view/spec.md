# view

## Requirements

### Requirement: Dashboard Display

The dashboard SHALL display a formatted dashboard with a title header, box-drawing separators (`══` for outer border, `──` for section underlines), and distinct sections for summary, active changes, draft changes, completed changes, specifications, and optionally a dependency graph. Each change entry in the active, draft, and completed sections SHALL display the change name, born date (formatted as `YYYY-MM-DD` from `.litespec.yaml` `created` field), and relative last-touched time (filesystem-derived mtime). The born and last-touched times SHALL be shown after the change name in parentheses, e.g. `◉ add-auth  (born 2026-04-01, touched 3d ago)`.

#### Scenario: Basic dashboard display

- **WHEN** user runs `litespec view`
- **THEN** system displays a formatted dashboard with sections for summary, active changes, draft changes, completed changes, and specifications, with born and last-touched timestamps per change

#### Scenario: No specs directory

- **WHEN** user runs `litespec view` in a directory without a specs directory
- **THEN** system displays error message indicating no specs directory was found

#### Scenario: Active change with timestamps

- **WHEN** user runs `litespec view` and an active change was created on 2026-04-01 and last modified 2 hours ago
- **THEN** the change line shows progress bar and `(born 2026-04-01, touched 2h ago)`

### Requirement: Dependency Graph Section

When any active change has a `dependsOn` field, the dashboard SHALL display a dependency graph section showing the DAG of active changes. The graph SHALL use tree-style indentation with box-drawing characters. Changes with no dependencies SHALL appear as roots. Changes with no `dependsOn` field and no dependents SHALL be listed separately as unrelated changes. Each node in the graph SHALL show the change name with born and last-touched timestamps.

#### Scenario: Simple dependency chain

- **WHEN** change A has no dependencies and change B depends on A
- **THEN** the graph section shows A as a root with B as its child, both with timestamps

#### Scenario: Multiple roots

- **WHEN** change A and change C have no dependencies and change B depends on A
- **THEN** the graph shows A (with child B) and C as separate roots, all with timestamps

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

### Requirement: Decisions Section

When `specs/decisions/` exists and contains at least one decision, the dashboard SHALL display a Decisions section between the Specifications section and the Dependency Graph section (or before the footer if no dependency graph is shown). Decisions SHALL be listed in two groups: active decisions (status `proposed` or `accepted`) shown individually with number, slug, and status, followed by a single summary line `superseded: N` if any superseded decisions exist. Active decisions SHALL be sorted by number ascending. The Decisions section SHALL be omitted entirely when no decisions exist.

#### Scenario: Active decisions displayed

- **WHEN** `litespec view` is run and three accepted decisions exist
- **THEN** the Decisions section lists each with number, slug, and status

#### Scenario: Superseded decisions summarized

- **WHEN** `litespec view` is run and two decisions are `accepted` and three are `superseded`
- **THEN** the Decisions section lists the two active decisions and shows `superseded: 3`

#### Scenario: No decisions omits section

- **WHEN** `litespec view` is run and `specs/decisions/` is absent or empty
- **THEN** the Decisions section does not appear in the dashboard

### Requirement: Summary Includes Decision Count

The summary section of the dashboard SHALL include a decision count when any decisions exist. The count SHALL be formatted as `Decisions: <active>/<total>` where `active` excludes superseded entries.

#### Scenario: Summary with decisions

- **WHEN** `litespec view` is run with 4 accepted and 2 superseded decisions
- **THEN** the summary line shows `Decisions: 4/6`

#### Scenario: Summary without decisions

- **WHEN** `litespec view` is run and no decisions exist
- **THEN** the summary omits the decisions line entirely
