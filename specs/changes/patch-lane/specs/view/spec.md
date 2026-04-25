# view

## ADDED Requirements

### Requirement: Patch-Mode Changes In Dashboard

The `litespec view` dashboard MUST display patch-mode changes (as defined by `IsPatchMode`) in a distinct category, separate from draft, active, and completed full-proposal changes. Patch-mode changes SHALL appear in a section labeled "Patch Changes" with a `◆` bullet (or another distinct marker), each line showing the change name, born date, and last-touched relative time. Progress bars SHALL NOT be displayed for patch-mode changes (they have no tasks). The summary section MUST include a "Patch Changes" count alongside draft, active, and completed counts. Patch-mode changes SHALL NOT be miscategorized as draft.

#### Scenario: Patch change appears in Patch Changes section

- **WHEN** `litespec view` is run and an active patch-mode change exists
- **THEN** the dashboard contains a "Patch Changes" section listing the change with a `◆` bullet, born date, and touched time, and the change does not appear in the Draft, Active, or Completed sections

#### Scenario: Summary includes patch count

- **WHEN** `litespec view` is run and one patch-mode and two full-proposal changes exist
- **THEN** the summary line(s) include a count of patch changes alongside other change-category counts

#### Scenario: No patch changes omits section

- **WHEN** `litespec view` is run and no patch-mode changes exist
- **THEN** the "Patch Changes" section is omitted entirely from the dashboard
