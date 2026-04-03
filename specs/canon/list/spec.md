# list

## Requirements

### Requirement: Enriched Change Listing

The `litespec list` command SHALL, by default, display each active change with three columns of metadata: task progress, status, and relative last-modified time. Task progress SHALL be derived from the change's `tasks.md` using the existing `TaskCompletion()` function. When all tasks are complete, the status column SHALL display `✓ Complete`. When tasks exist but are incomplete, it SHALL display `<completed>/<total> tasks`. When no `tasks.md` exists or there are zero tasks, it SHALL display `No tasks`. The last-modified time SHALL be determined by recursively walking all files in the change directory and finding the most recent modification time, falling back to the directory's own mtime if no files are found. The relative time format SHALL be: `just now` (< 1 min), `Xm ago` (minutes), `Xh ago` (hours), `Xd ago` (days, up to 30), or the locale date if older than 30 days.

#### Scenario: Change with completed tasks

- **WHEN** `litespec list` is run and a change has all tasks checked in `tasks.md`
- **THEN** the output shows the change name with `✓ Complete` in the status column

#### Scenario: Change with partial task progress

- **WHEN** `litespec list` is run and a change has 3 of 5 tasks checked
- **THEN** the output shows `3/5 tasks` in the status column

#### Scenario: Change with no tasks.md

- **WHEN** `litespec list` is run and a change has no `tasks.md` file
- **THEN** the output shows `No tasks` in the status column

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

The `litespec list --json` command SHALL return structured JSON. Each change entry SHALL include `name`, `completedTasks` (int), `totalTasks` (int), `lastModified` (ISO 8601 timestamp string), and `status` (one of `no-tasks`, `complete`, or `in-progress`). The `status` field SHALL be derived: `no-tasks` when `totalTasks == 0`, `complete` when `completedTasks == totalTasks`, `in-progress` otherwise. Each spec entry SHALL include `name` and `requirementCount` (int). The top-level JSON keys SHALL be `changes` and `specs` as appropriate.

#### Scenario: JSON output for changes

- **WHEN** `litespec list --json` is run
- **THEN** each change entry has `name`, `completedTasks`, `totalTasks`, `lastModified`, and `status` fields

#### Scenario: JSON status derivation

- **WHEN** a change has 5 of 5 tasks complete
- **THEN** the JSON `status` field is `complete`

#### Scenario: JSON output for specs

- **WHEN** `litespec list --specs --json` is run
- **THEN** each spec entry has `name` and `requirementCount` fields

### Requirement: Enriched Internal Types

The `ListChanges()` function SHALL return a slice of enriched structs (not bare `[]string`) containing `Name`, `CompletedTasks` (int), `TotalTasks` (int), `LastModified` (time.Time). The `ListSpecs()` function SHALL return a slice of enriched structs containing `Name` and `RequirementCount` (int). A `GetLastModified(dir string) (time.Time, error)` helper SHALL walk a directory recursively and return the most recent file mtime, falling back to the directory's own mtime.

#### Scenario: ListChanges returns enriched data

- **WHEN** `ListChanges(root)` is called and a change has 3 of 5 tasks complete
- **THEN** the returned struct has `CompletedTasks: 3` and `TotalTasks: 5`

#### Scenario: GetLastModified with nested files

- **WHEN** `GetLastModified()` is called on a directory containing files with different mtimes
- **THEN** the most recent mtime across all files is returned

#### Scenario: GetLastModified on empty directory

- **WHEN** `GetLastModified()` is called on a directory with no files
- **THEN** the directory's own mtime is returned
