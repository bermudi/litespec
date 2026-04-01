## ADDED Requirements

### Requirement: JSON Output for Validate
The `litespec validate` command MUST support a `--json` flag that returns structured JSON output containing a `valid` boolean, `errors` array, and `warnings` array. Each issue MUST include `severity`, `message`, and `file` fields.

### Requirement: Consistent JSON Flag Convention
All litespec commands that support `--json` MUST use the same flag name and return valid JSON to stdout.
