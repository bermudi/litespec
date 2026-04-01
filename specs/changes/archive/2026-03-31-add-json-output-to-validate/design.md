# Design: Add JSON Output to Validate

## Technical Approach
Add a `ValidationResultJSON` type in `internal/json.go` with proper json tags, and a builder function `BuildValidationResultJSON` that converts the existing `ValidationResult` into the JSON-friendly type. Update `cmdValidate` in main.go to accept `--json` and output accordingly.

## Architecture Decisions

### Decision: Reuse existing ValidationResult
The internal `ValidationResult` uses `Severity` type which marshals awkwardly. Rather than changing it, we create a parallel JSON type with clean string fields.

**Rationale:** Avoids touching existing validation logic. Keeps the JSON layer separate.

### Decision: Single flat structure
No nested objects — errors and warnings are arrays of objects with severity, message, and file fields.

**Rationale:** Simpler for AI agents to parse. Matches the convention of other JSON endpoints.

## File Changes
- `internal/json.go` (modified) — add `ValidationResultJSON` type + builder
- `cmd/litespec/main.go` (modified) — add `--json` to validate command
