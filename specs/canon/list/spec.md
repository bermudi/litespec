# list

## Requirements

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

### Requirement: Enriched Spec Listing

The `litespec list --specs` command SHALL display each canonical spec with its name and requirement count. The requirement count SHALL be derived by parsing each `specs/canon/<name>/spec.md` with `ParseMainSpec()` and counting `len(spec.Requirements)`. If parsing fails, the count SHALL be 0. Specs SHALL always be sorted alphabetically by name. The output format SHALL be column-aligned with names padded to the widest name width.

#### Scenario: Spec with requirements

- **WHEN** `litespec list --specs` is run and a spec has 5 requirements
- **THEN** the output shows a table with `Name` and `Requirements` column headers, and the spec row shows `5` in the Requirements column

#### Scenario: Spec with parse failure

- **WHEN** `litespec list --specs` is run and a spec's `spec.md` cannot be parsed
- **THEN** the output shows `requirements 0` next to the spec name

### Requirement: Sort Flag

The `litespec list` command SHALL support a `--sort` flag accepting `recent` (default), `name`, or `deps`. When `--sort recent`, changes SHALL be ordered by last-modified time descending (most recent first). When `--sort name`, changes SHALL be ordered alphabetically ascending. When `--sort deps`, changes SHALL be ordered by topological sort of their dependency graph: changes with no dependencies first, then changes whose dependencies are already listed, with lexicographic tie-breaking at each level. Changes with no `dependsOn` field are treated as roots. The `--sort` flag SHALL only apply to changes — specs are always sorted alphabetically. The `--sort` flag with `--specs` only SHALL have no effect.

#### Scenario: Default sort is recent

- **WHEN** `litespec list` is run with no `--sort` flag
- **THEN** changes are ordered by most recently modified first

#### Scenario: Sort by name

- **WHEN** `litespec list --sort name` is run
- **THEN** changes are ordered alphabetically by name

#### Scenario: Sort by dependency order

- **WHEN** `litespec list --sort deps` is run and change B depends on change A
- **THEN** change A appears before change B in the output

#### Scenario: Sort deps with unrelated changes

- **WHEN** `litespec list --sort deps` is run and change B depends on A, while C has no dependencies
- **THEN** A and C appear before B, with A and C ordered alphabetically

#### Scenario: Sort deps with no dependencies

- **WHEN** `litespec list --sort deps` is run and no change has `dependsOn`
- **THEN** changes are sorted alphabetically as a fallback

#### Scenario: Sort deps with cycles

- **WHEN** `litespec list --sort deps` is run and a dependency cycle exists among active changes
- **THEN** all changes are sorted alphabetically as a fallback and a warning is printed to stderr

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

### Requirement: Decision Listing

The `litespec list --decisions` command SHALL display each decision in `specs/decisions/` with four columns: number, slug, status, and title. Decisions SHALL be sorted by number ascending by default. A `--status <state>` flag SHALL filter to decisions matching the given status (`proposed`, `accepted`, or `superseded`). A `--sort <mode>` flag SHALL accept `number` (default), `recent` (by file mtime descending), or `name` (alphabetical by slug). The `--decisions` flag SHALL be mutually exclusive with `--changes` and `--specs`.

#### Scenario: Default decision listing

- **WHEN** `litespec list --decisions` is run with three decisions numbered 0001, 0002, 0003
- **THEN** decisions are listed in number order with number, slug, status, and title columns

#### Scenario: Filter by status

- **WHEN** `litespec list --decisions --status superseded` is run
- **THEN** only decisions with status `superseded` are shown

#### Scenario: Sort by recent

- **WHEN** `litespec list --decisions --sort recent` is run
- **THEN** decisions are ordered by file modification time descending

#### Scenario: Mutually exclusive with changes

- **WHEN** `litespec list --decisions --changes` is run
- **THEN** an error is printed indicating the flags are mutually exclusive

### Requirement: Decision JSON Output

The `litespec list --decisions --json` command SHALL return a JSON object with a top-level `decisions` key containing an array. Each entry SHALL include `number` (int), `slug` (string), `title` (string), `status` (string), `supersedes` (array of slugs), `supersededBy` (array of slugs), and `lastModified` (ISO 8601 timestamp).

#### Scenario: Decision JSON fields

- **WHEN** `litespec list --decisions --json` is run
- **THEN** each decision entry contains `number`, `slug`, `title`, `status`, `supersedes`, `supersededBy`, and `lastModified`
