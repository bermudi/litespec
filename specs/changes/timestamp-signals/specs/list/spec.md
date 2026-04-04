# list

## MODIFIED Requirements

### Requirement: Enriched Change Listing

The `litespec list` command SHALL, by default, display each active change with four columns of metadata: task progress, status, born time, and relative last-touched time. Task progress SHALL be derived from the change's `tasks.md` using the existing `TaskCompletion()` function. When all tasks are complete, the status column SHALL display `✓ Complete`. When tasks exist but are incomplete, it SHALL display `<completed>/<total> tasks`. When no `tasks.md` exists or there are zero tasks, it SHALL display `No tasks`. The born time SHALL display the change's `created` timestamp from `.litespec.yaml` formatted as `YYYY-MM-DD`. The last-touched time SHALL be determined by recursively walking all files in the change directory and finding the most recent modification time, falling back to the directory's own mtime if no files are found. The relative time format for last-touched SHALL be: `just now` (< 1 min), `Xm ago` (minutes), `Xh ago` (hours), `Xd ago` (days, up to 30), or the locale date if older than 30 days.

#### Scenario: Change with completed tasks

- **WHEN** `litespec list` is run and a change has all tasks checked in `tasks.md`
- **THEN** the output shows the change name with `✓ Complete` in the status column, born date, and last-touched time

#### Scenario: Change with partial task progress

- **WHEN** `litespec list` is run and a change has 3 of 5 tasks checked
- **THEN** the output shows `3/5 tasks` in the status column, born date, and last-touched time

#### Scenario: Change with no tasks.md

- **WHEN** `litespec list` is run and a change has no `tasks.md` file
- **THEN** the output shows `No tasks` in the status column, born date, and last-touched time

#### Scenario: Column alignment

- **WHEN** `litespec list` is run with changes of varying name lengths
- **THEN** columns are left-aligned with names padded to the widest name width

#### Scenario: Default shows changes only

- **WHEN** `litespec list` is run without flags
- **THEN** only changes are listed, not specs

#### Scenario: --changes is explicit default

- **WHEN** `litespec list --changes` is run
- **THEN** the output is identical to running `litespec list` without flags

### Requirement: Enriched JSON Output

The `litespec list --json` command SHALL return structured JSON. Each change entry SHALL include `name`, `completedTasks` (int), `totalTasks` (int), `born` (ISO 8601 timestamp string from `.litespec.yaml` `created` field), `lastModified` (ISO 8601 timestamp string), and `status` (one of `no-tasks`, `complete`, or `in-progress`). The `status` field SHALL be derived: `no-tasks` when `totalTasks == 0`, `complete` when `completedTasks == totalTasks`, `in-progress` otherwise. Each spec entry SHALL include `name` and `requirementCount` (int). The top-level JSON keys SHALL be `changes` and `specs` as appropriate.

#### Scenario: JSON output for changes

- **WHEN** `litespec list --json` is run
- **THEN** each change entry has `name`, `completedTasks`, `totalTasks`, `born`, `lastModified`, and `status` fields

#### Scenario: JSON status derivation

- **WHEN** a change has 5 of 5 tasks complete
- **THEN** the JSON `status` field is `complete`

#### Scenario: JSON output for specs

- **WHEN** `litespec list --specs --json` is run
- **THEN** each spec entry has `name` and `requirementCount` fields

### Requirement: Enriched Internal Types

The `ListChanges()` function SHALL return a slice of enriched structs (not bare `[]string`) containing `Name`, `CompletedTasks` (int), `TotalTasks` (int), `Created` (time.Time), `LastModified` (time.Time). The `Created` field SHALL be populated by reading `.litespec.yaml` for each change. The `ListSpecs()` function SHALL return a slice of enriched structs containing `Name` and `RequirementCount` (int). A `GetLastModified(dir string) (time.Time, error)` helper SHALL walk a directory recursively and return the most recent file mtime, falling back to the directory's own mtime.

#### Scenario: ListChanges returns enriched data

- **WHEN** `ListChanges(root)` is called and a change has 3 of 5 tasks complete
- **THEN** the returned struct has `CompletedTasks: 3`, `TotalTasks: 5`, and `Created` populated from `.litespec.yaml`

#### Scenario: GetLastModified with nested files

- **WHEN** `GetLastModified()` is called on a directory containing files with different mtimes
- **THEN** the most recent mtime across all files is returned

#### Scenario: GetLastModified on empty directory

- **WHEN** `GetLastModified()` is called on a directory with no files
- **THEN** the directory's own mtime is returned
