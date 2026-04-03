# cli-input

## ADDED Requirements

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
