# view

## ADDED Requirements

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
