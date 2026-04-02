# status

## Requirements

### Requirement: Positional Name for Status

The `litespec status` command MUST accept an optional positional `<name>` argument instead of `--change <name>`. When provided, it shows artifact state for that specific change. When omitted, it shows all changes. The `--change` flag SHALL be removed.

#### Scenario: Status for a named change

- **WHEN** `litespec status my-feature` is run
- **THEN** artifact state for `my-feature` is shown

#### Scenario: Status with no arguments

- **WHEN** `litespec status` is run
- **THEN** all active changes are listed

#### Scenario: Status for nonexistent change

- **WHEN** `litespec status nonexistent` is run
- **THEN** an error is printed to stderr indicating the change was not found with exit code 1
