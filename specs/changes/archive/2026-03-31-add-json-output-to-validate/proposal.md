# Proposal: Add JSON Output to Validate

## Motivation
The `validate` command currently only outputs human-readable text. AI agents consuming litespec need structured JSON output to programmatically assess change health and make decisions about archiving.

## Scope
- Add `--json` flag to `litespec validate`
- Return structured JSON with errors, warnings, and valid flag
- Support all existing validate modes (`--change`, `--all`, `--strict`)

## Approach
Reuse the existing `ValidationResult` type and marshal it to JSON. The `--json` flag is already the convention established by `list`, `status`, and `instructions`.

## Non-goals
- Changing the validation logic itself
- Adding new validation rules
