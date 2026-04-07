# cli-input

## Requirements

### Requirement: Change Name Validation

The `litespec new` command MUST reject change names that are empty, contain path separators (`/` or `\`), path traversal sequences (`..`), or leading/trailing whitespace. Names exceeding 100 characters SHALL be rejected to prevent filesystem issues. Reserved directory names (`canon`, `changes`, `archive`) SHALL be rejected. The command SHALL report the specific validation failure and exit with code 1.

#### Scenario: Reject empty change name

- **WHEN** `litespec new ""` is run
- **THEN** an error is printed to stderr indicating the name is empty and exit code is 1

#### Scenario: Reject path separator in name

- **WHEN** `litespec new "foo/bar"` is run
- **THEN** an error is printed to stderr indicating path separators are not allowed and exit code is 1

#### Scenario: Reject path traversal in name

- **WHEN** `litespec new "../escape"` is run
- **THEN** an error is printed to stderr indicating path traversal is not allowed and exit code is 1

#### Scenario: Reject leading or trailing whitespace

- **WHEN** `litespec new "  my-change  "` is run
- **THEN** an error is printed to stderr indicating leading/trailing whitespace is not allowed and exit code is 1

#### Scenario: Reject reserved directory name

- **WHEN** `litespec new canon` is run
- **THEN** an error is printed to stderr indicating the name is reserved and exit code is 1

#### Scenario: Reject excessively long name

- **WHEN** `litespec new` is run with a name exceeding 100 characters
- **THEN** an error is printed to stderr indicating the name is too long and exit code is 1

#### Scenario: Accept valid kebab-case name

- **WHEN** `litespec new my-feature` is run
- **THEN** the change directory is created successfully

### Requirement: Tools Flag Validation

The `litespec init` and `litespec update` commands MUST validate `--tools` values against the registered adapters. When an unknown tool ID is provided, the command SHALL list all valid tool IDs in the error message and exit with code 1.

#### Scenario: Reject unknown tool ID

- **WHEN** `litespec init --tools copilot` is run
- **THEN** an error is printed to stderr listing supported tools (e.g., "claude") and exit code is 1

#### Scenario: Accept known tool ID

- **WHEN** `litespec init --tools claude` is run
- **THEN** initialization proceeds and symlinks are created for claude

#### Scenario: Case-sensitive tool ID matching

- **WHEN** `litespec init --tools Claude` is run
- **THEN** an error is printed to stderr listing supported tools (matching is case-sensitive) and exit code is 1

### Requirement: JSON Marshal Error Reporting

All CLI commands that produce JSON output MUST check and report `MarshalJSON` errors. When marshaling fails, the command SHALL print the error to stderr and exit with code 1. JSON output MUST NOT be silently omitted.

#### Scenario: JSON marshal failure surfaces as error

- **WHEN** `litespec status --json` is run and internal MarshalJSON returns an error
- **THEN** the error is printed to stderr and exit code is 1

#### Scenario: Successful JSON output still works

- **WHEN** `litespec status --json` is run with valid data
- **THEN** valid JSON is printed to stdout and exit code is 0

### Requirement: Error Visibility in Iteration

When commands iterate over changes and encounter individual errors (e.g., unreadable metadata, corrupted files), the command SHALL collect these errors and report them as warnings rather than silently skipping the affected changes. In `--json` mode, these SHALL appear in the `warnings` array.

#### Scenario: Corrupted change metadata reported as warning

- **WHEN** `litespec status --json` is run and one change has corrupted `.litespec.yaml`
- **THEN** the output includes a warning about the corrupted change and other changes are still listed

#### Scenario: Corrupted change reported in text mode

- **WHEN** `litespec status` is run and one change has corrupted `.litespec.yaml`
- **THEN** a warning line is printed to stderr about the corrupted change and other changes are still listed

### Requirement: Adapter Registry for Tool Validation

The supported tools list in error messages MUST be derived from the `Adapters` slice in `paths.go`, not hardcoded. The `GetAdapter` function and error messages SHALL use the same source of truth.

#### Scenario: Error message reflects registered adapters

- **WHEN** `litespec init --tools unknown` is run
- **THEN** the error message lists all adapter IDs from the Adapters slice (not a hardcoded string)

### Requirement: Auto-Detection of Active Tool Adapters

The `init` and `update` commands MUST detect active tool adapters by scanning for existing adapter skill directories (e.g., `.claude/skills/`) that contain at least one symlink pointing into `.agents/skills/`. When `--tools` is not provided and active adapters are detected, the commands SHALL refresh symlinks for those adapters. When `--tools` is provided explicitly, only the listed adapters SHALL have symlinks generated; any previously active adapters not in the explicit list SHALL be ignored. The commands SHALL NOT read from any config file to determine tool adapters.

#### Scenario: Auto-detect claude adapter on update

- **WHEN** `litespec update` is run without `--tools` and `.claude/skills/` contains symlinks into `.agents/skills/`
- **THEN** the claude adapter symlinks are refreshed without requiring `--tools`

#### Scenario: Auto-detect claude adapter on init

- **WHEN** `litespec init` is run without `--tools` and `.claude/skills/` already contains symlinks into `.agents/skills/`
- **THEN** the claude adapter symlinks are refreshed without requiring `--tools`

#### Scenario: Explicit --tools overrides auto-detection on update

- **WHEN** `litespec update --tools claude` is run
- **THEN** symlinks are created only for claude regardless of what exists on disk

#### Scenario: Explicit --tools overrides auto-detection on init

- **WHEN** `litespec init --tools claude` is run and `.claude/skills/` already contains symlinks into `.agents/skills/`
- **THEN** symlinks are created only for claude

#### Scenario: No adapters detected, no --tools provided

- **WHEN** `litespec update` is run without `--tools` and no adapter skill directories contain symlinks into `.agents/skills/`
- **THEN** only `.agents/skills/` is updated, no adapter output is produced, and the command succeeds

#### Scenario: Adapter directory exists but contains no symlinks

- **WHEN** `.claude/skills/` exists but contains no symlinks pointing into `.agents/skills/`
- **THEN** the adapter is considered inactive and no adapter symlinks are generated

### Requirement: No Config File for Tool Persistence

The `init` and `update` commands MUST NOT create or read a config file for tool persistence. Tool adapter state SHALL be inferred entirely from the filesystem.

#### Scenario: init does not create config.yaml

- **WHEN** `litespec init --tools claude` is run
- **THEN** `specs/config.yaml` is not created

#### Scenario: update does not create config.yaml

- **WHEN** `litespec update --tools claude` is run
- **THEN** `specs/config.yaml` is not created or modified

#### Scenario: update works without config.yaml

- **WHEN** `litespec update` is run and `specs/config.yaml` does not exist
- **THEN** the command succeeds by auto-detecting adapters from the filesystem
