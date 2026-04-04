# openspec-import

## ADDED Requirements

### Requirement: OpenSpec Project Detection

The `litespec import` command SHALL detect an OpenSpec project by the presence of an `openspec/specs/` directory or an `openspec/changes/` directory. If neither is found, the command SHALL print an error indicating no OpenSpec project was detected and exit with code 1.

#### Scenario: Detect valid OpenSpec project

- **WHEN** `litespec import` is run in a directory containing `openspec/specs/`
- **THEN** the command proceeds with import

#### Scenario: Reject non-OpenSpec project

- **WHEN** `litespec import` is run in a directory without `openspec/specs/` or `openspec/changes/`
- **THEN** an error is printed indicating no OpenSpec project was found and exit code is 1

### Requirement: Canon Spec Migration

The import command SHALL move all capability specs from `openspec/specs/<capability>/spec.md` to `specs/canon/<capability>/spec.md`. During migration, the command SHALL normalize H1 titles from descriptive text to kebab-case capability names and strip any H2 sections other than `Purpose` and `Requirements` that appear before the `Requirements` section.

#### Scenario: Migrate spec with descriptive H1 title

- **WHEN** a spec file has `# CLI Archive Command Specification`
- **THEN** the migrated file has `# cli-archive` as the H1

#### Scenario: Strip extra H2 sections from spec

- **WHEN** a spec file contains `## Why These Decisions` or `## Core Principles`
- **THEN** those sections are removed from the migrated file

#### Scenario: Preserve Purpose and Requirements sections

- **WHEN** a spec file contains `## Purpose` and `## Requirements`
- **THEN** those sections are preserved in the migrated file

### Requirement: Change Migration

The import command SHALL move all active changes from `openspec/changes/<name>/` to `specs/changes/<name>/`. Archive directories at `openspec/changes/archive/` SHALL be skipped and a warning printed. Each change's metadata file SHALL be renamed from `.openspec.yaml` to `.litespec.yaml`.

#### Scenario: Migrate active changes

- **WHEN** `litespec import` is run with changes in `openspec/changes/`
- **THEN** each change directory is copied to `specs/changes/`

#### Scenario: Skip archive directory

- **WHEN** `openspec/changes/archive/` exists
- **THEN** the archive directory is skipped and a warning is printed

### Requirement: Metadata Format Conversion

When converting `.openspec.yaml` to `.litespec.yaml`, the command SHALL convert date-only `created` values (e.g., `2026-02-21`) to ISO 8601 format (e.g., `2026-02-21T00:00:00Z`). The `schema` field SHALL be preserved. Any fields not recognized by litespec (`provides`, `requires`, `touches`, `parent`) SHALL be dropped with a warning. The `dependsOn` field SHALL be preserved if present.

#### Scenario: Convert date format

- **WHEN** `.openspec.yaml` contains `created: 2026-02-21`
- **THEN** `.litespec.yaml` contains `created: 2026-02-21T00:00:00Z`

#### Scenario: Drop unsupported metadata fields

- **WHEN** `.openspec.yaml` contains `provides` or `touches` fields
- **THEN** those fields are omitted from `.litespec.yaml` and a warning is printed

### Requirement: Delta Spec Rename Format Conversion

The import command SHALL convert OpenSpec's `FROM:/TO:` bullet-style rename syntax to litespec's `→` arrow format in delta spec files. A requirement block with `FROM: ### Requirement: Old Name` and `TO: ### Requirement: New Name` SHALL become `### Requirement: Old Name → New Name`.

#### Scenario: Convert FROM/TO rename to arrow format

- **WHEN** a delta spec contains `FROM: ### Requirement: Old Name` and `TO: ### Requirement: New Name`
- **THEN** the migrated spec contains `### Requirement: Old Name → New Name`

### Requirement: Import Dry Run Mode

The `litespec import` command SHALL support a `--dry-run` flag that prints what would be migrated without modifying any files. The output SHALL list all specs, changes, and transformations that would occur.

#### Scenario: Dry run lists planned migrations

- **WHEN** `litespec import --dry-run` is run
- **THEN** a summary of specs and changes to be migrated is printed without writing any files

### Requirement: Import Target Directory

The `litespec import` command SHALL support a `--source` flag to specify the OpenSpec project root directory. When omitted, it SHALL default to the current working directory.

#### Scenario: Import from specified source directory

- **WHEN** `litespec import --source /path/to/old-project` is run
- **THEN** the command reads from the specified directory and writes to the current project

### Requirement: Config and Context File Warnings

When an OpenSpec project contains `openspec/config.yaml`, `openspec/project.md`, or `openspec/AGENTS.md`, the import command SHALL print warnings that these files have no litespec equivalent and will not be migrated. The command SHALL NOT fail due to their presence.

#### Scenario: Warn about config.yaml

- **WHEN** `openspec/config.yaml` exists
- **THEN** a warning is printed that config.yaml will not be migrated

#### Scenario: Warn about project.md

- **WHEN** `openspec/project.md` exists
- **THEN** a warning is printed that project.md will not be migrated

#### Scenario: Warn about AGENTS.md

- **WHEN** `openspec/AGENTS.md` exists
- **THEN** a warning is printed that AGENTS.md will not be migrated

### Requirement: Conflict Detection

If a spec or change already exists at the litespec target path, the import command SHALL report a conflict and skip that item unless a `--force` flag is provided. With `--force`, existing files SHALL be overwritten.

#### Scenario: Detect conflicting spec

- **WHEN** `specs/canon/cli-input/spec.md` already exists and the import would write the same path
- **THEN** a conflict warning is printed and the spec is skipped

#### Scenario: Force overwrite existing files

- **WHEN** `litespec import --force` is run with existing target files
- **THEN** existing files are overwritten without conflict warnings

### Requirement: Post-Import Suggestion

After a successful import, the command SHALL print a message suggesting the user run `litespec update` to generate skills and adapters.

#### Scenario: Suggest update after import

- **WHEN** `litespec import` completes successfully
- **THEN** a message is printed suggesting `litespec update` to generate skills
