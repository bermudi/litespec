# list

## ADDED Requirements

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
