# openspec-import

## Requirements

### Requirement: OpenSpec Project Detection

The `litespec import` command SHALL detect an OpenSpec project by the presence of an `openspec/specs/` directory or an `openspec/changes/` directory. If neither is found, the command SHALL print an error indicating no OpenSpec project was detected and exit with code 1.

#### Scenario: Detect valid OpenSpec project

- **WHEN** `litespec import` is run in a directory containing `openspec/specs/`
- **THEN** the command proceeds with import

#### Scenario: Reject non-OpenSpec project

- **WHEN** `litespec import` is run in a directory without `openspec/specs/` or `openspec/changes/`
- **THEN** an error is printed indicating no OpenSpec project was found and exit code is 1

### Requirement: Canon Spec Migration

The import command SHALL copy all capability specs from `openspec/specs/<capability>/spec.md` to `specs/canon/<capability>/spec.md`. The H1 title SHALL be normalized by removing a trailing " Specification" suffix if present (e.g., `# cli-init Specification` becomes `# cli-init`).

#### Scenario: Migrate canon specs

- **WHEN** `openspec/specs/` contains capability directories with `spec.md` files
- **THEN** each `spec.md` is copied to `specs/canon/<capability>/spec.md`

#### Scenario: Strip Specification suffix from H1

- **WHEN** a spec file has H1 title `# cli-init Specification`
- **THEN** the migrated file has H1 title `# cli-init`

#### Scenario: Preserve H1 without suffix

- **WHEN** a spec file has H1 title `# dietary-plans`
- **THEN** the migrated file preserves the H1 as-is

### Requirement: Change Migration

The import command SHALL copy all active changes from `openspec/changes/<name>/` to `specs/changes/<name>/`. Each change's metadata file SHALL be renamed from `.openspec.yaml` to `.litespec.yaml`. Only directory entries SHALL be processed; loose files at the root of `openspec/changes/` SHALL be skipped with a warning.

#### Scenario: Migrate active changes

- **WHEN** `litespec import` is run with changes in `openspec/changes/`
- **THEN** each change directory is copied to `specs/changes/`

#### Scenario: Skip loose files in changes directory

- **WHEN** `openspec/changes/IMPLEMENTATION_ORDER.md` exists as a loose file
- **THEN** a warning is printed and the file is skipped

### Requirement: Archive Migration with Metadata Synthesis

The import command SHALL copy archived changes from `openspec/changes/archive/<name>/` to `specs/changes/archive/<name>/`. Any `specs/` subdirectory within an archived change SHALL be stripped during migration. If an archived change lacks `.openspec.yaml`, the command SHALL synthesize a `.litespec.yaml` by extracting the date prefix from the directory name (e.g., `2026-04-01-change-name` yields `created: 2026-04-01T00:00:00Z`).

#### Scenario: Migrate archived change with specs stripped

- **WHEN** `openspec/changes/archive/old-change/` contains a `specs/` subdirectory
- **THEN** the change is copied to `specs/changes/archive/old-change/` without the `specs/` subdirectory and a summary is printed

#### Scenario: Migrate archived change without specs

- **WHEN** `openspec/changes/archive/old-change/` does not contain a `specs/` subdirectory
- **THEN** the change is copied to `specs/changes/archive/old-change/` as-is

#### Scenario: Synthesize metadata from directory name

- **WHEN** `openspec/changes/archive/2026-04-01-my-change/` lacks `.openspec.yaml`
- **THEN** `.litespec.yaml` is created with `created: 2026-04-01T00:00:00Z` and `schema: spec-driven`

#### Scenario: Use existing metadata when present

- **WHEN** `openspec/changes/archive/2026-04-01-my-change/` contains `.openspec.yaml`
- **THEN** the metadata is converted normally and no synthesis occurs

### Requirement: Metadata Format Conversion

When converting `.openspec.yaml` to `.litespec.yaml`, the command SHALL convert date-only `created` values (e.g., `2026-02-21`) to ISO 8601 format (e.g., `2026-02-21T00:00:00Z`). Both quoted and unquoted date strings SHALL be handled. The `schema` field SHALL be preserved. Any fields not recognized by litespec (`provides`, `requires`, `touches`, `parent`) SHALL be dropped with a warning. The `dependsOn` field SHALL be preserved if present.

#### Scenario: Convert unquoted date format

- **WHEN** `.openspec.yaml` contains `created: 2026-02-21`
- **THEN** `.litespec.yaml` contains `created: 2026-02-21T00:00:00Z`

#### Scenario: Convert quoted date format

- **WHEN** `.openspec.yaml` contains `created: "2026-02-21"`
- **THEN** `.litespec.yaml` contains `created: 2026-02-21T00:00:00Z`

#### Scenario: Drop unsupported metadata fields

- **WHEN** `.openspec.yaml` contains `provides` or `touches` fields
- **THEN** those fields are omitted from `.litespec.yaml` and a warning is printed

### Requirement: Import Dry Run Mode

The `litespec import` command SHALL support a `--dry-run` flag that prints what would be migrated without modifying any files. The output SHALL list all specs, changes, and archive entries that would be migrated.

#### Scenario: Dry run lists planned migrations

- **WHEN** `litespec import --dry-run` is run
- **THEN** a summary of specs, changes, and archives to be migrated is printed without writing any files

### Requirement: Import Target Directory

The `litespec import` command SHALL accept a `--source` flag to specify the OpenSpec project directory. When omitted, it SHALL default to the current working directory.

#### Scenario: Import from specified source directory

- **WHEN** `litespec import --source /path/to/old-project` is run
- **THEN** the command reads from the specified directory and writes to the current working directory

#### Scenario: Import from current working directory

- **WHEN** `litespec import` is run without `--source`
- **THEN** the command reads from the current working directory

### Requirement: Skipped Directory and File Warnings

When the import encounters `openspec/config.yaml`, `openspec/project.md`, `openspec/AGENTS.md`, or `openspec/explorations/`, the command SHALL print warnings that these files/directories have no litespec equivalent and will not be migrated. The command SHALL NOT fail due to their presence.

#### Scenario: Warn about config.yaml

- **WHEN** `openspec/config.yaml` exists
- **THEN** a warning is printed that config.yaml will not be migrated

#### Scenario: Warn about project.md

- **WHEN** `openspec/project.md` exists
- **THEN** a warning is printed that project.md will not be migrated

#### Scenario: Warn about AGENTS.md

- **WHEN** `openspec/AGENTS.md` exists
- **THEN** a warning is printed that AGENTS.md will not be migrated

#### Scenario: Warn about explorations directory

- **WHEN** `openspec/explorations/` exists
- **THEN** a warning is printed that the explorations directory will not be migrated

### Requirement: Conflict Detection

If a spec or change already exists at the litespec target path, the import command SHALL report a conflict and skip that item unless a `--force` flag is provided. With `--force`, existing files SHALL be overwritten.

#### Scenario: Detect conflicting spec

- **WHEN** `specs/canon/cli-input/spec.md` already exists and the import would write the same path
- **THEN** a conflict warning is printed and the spec is skipped

#### Scenario: Force overwrite existing files

- **WHEN** `litespec import --force` is run with existing target files
- **THEN** existing files are overwritten without conflict warnings

### Requirement: Task Phase Label Normalization

The import command SHALL convert OpenSpec task phase labels from `## N. Name` format to litespec's `## Phase N: Name` format in imported `tasks.md` files.

#### Scenario: Convert numeric phase label

- **WHEN** a `tasks.md` file contains `## 1. Core logic`
- **THEN** the migrated file contains `## Phase 1: Core logic`

### Requirement: Post-Import Suggestion

After a successful import, the command SHALL print a message suggesting the user run `litespec update` to generate skills. The message SHALL note that `import` replaces `init` for imported projects — no separate initialization is needed.

#### Scenario: Suggest update after import

- **WHEN** `litespec import` completes successfully
- **THEN** a message is printed suggesting `litespec update` and noting that `import` replaces `init` for imported projects
