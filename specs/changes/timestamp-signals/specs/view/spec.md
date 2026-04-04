# view

## MODIFIED Requirements

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
