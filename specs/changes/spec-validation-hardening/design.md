## Architecture

This change strengthens the validation pipeline at two points: during parsing (reject malformed input early) and during validation (detect semantic conflicts). No new packages.

## Decisions

- **Reject at parse time vs validate time**: Empty names and duplicates are parse-level issues — reject them in `ParseMainSpec` and `ParseDeltaSpec` so they never reach validation. Cross-operation conflicts are semantic issues — detect them in `ValidateChange`.
- **Whole-word matching for SHALL/MUST**: Use a regex `\b(SHALL|MUST)\b` instead of `strings.Contains`. Strip fenced code blocks (```...```) and inline code (`` `...` ``) before checking.
- **RENAMED overlap fix**: Change `DetectOverlaps` to record both `OldName` (for conflict with MODIFIED/REMOVED) and `Name` (for conflict with ADDED) when the operation is `DeltaRenamed`.
- **File context for deps**: Pass the change metadata file path through to the validation error.

## File Changes

- `internal/delta.go`: Add empty name checks in `ParseMainSpec` (after line 84) and `ParseDeltaSpec` (after line 155). Add duplicate scenario name detection in `parseScenariosFromBody`. Return errors for empty/whitespace-only names.
- `internal/validate.go`: Add duplicate requirement name detection within single delta (after line 91). Add cross-operation conflict detection. Change SHALL/MUST check to use word-boundary regex after stripping code blocks. Add scenario content validation (WHEN/THEN presence). Add `File` field to dependency validation errors. Update `DetectOverlaps` call or add within-delta conflict check.
- `internal/deps.go`: In `DetectOverlaps`, when operation is `DeltaRenamed`, add a second `changeTarget` entry using `OldName` so overlaps against MODIFIED/REMOVED are caught.
- `internal/validate_test.go`: Add tests for all new validation rules.
- `internal/delta_test.go`: Add tests for empty name rejection.
- `internal/deps_test.go`: Add test for RENAMED/MODIFIED overlap detection.
