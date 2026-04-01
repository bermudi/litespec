# validate

### Requirement: JSON Output for Validate

The `litespec validate` command MUST support a `--json` flag that returns structured JSON output containing a `valid` boolean, `errors` array, and `warnings` array. Each issue MUST include `severity`, `message`, and `file` fields.

#### Scenario: Validate with JSON flag
- **WHEN** `litespec validate --json` is run
- **THEN** output is valid JSON with `valid`, `errors`, and `warnings` fields

### Requirement: Consistent JSON Flag Convention

All litespec commands that support `--json` MUST use the same flag name and return valid JSON to stdout.

#### Scenario: JSON flag consistency
- **WHEN** any litespec command is run with `--json`
- **THEN** the output is valid JSON to stdout
