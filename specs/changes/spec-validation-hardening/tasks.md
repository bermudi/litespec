## Phase 1: Parse-Time Name Validation

- [x] Add empty name rejection in `ParseMainSpec` after extracting requirement name (return error if trimmed name is empty)
- [x] Add empty name rejection in `ParseDeltaSpec` for both regular and RENAMED requirements  
- [x] Add empty scenario name rejection in `parseScenariosFromBody`
- [x] Add tests for empty name rejection in `internal/delta_test.go`

## Phase 2: Within-Delta Duplicate Detection

- [x] Add duplicate requirement name check in `ValidateChange` — track names seen per delta file, report duplicates
- [x] Add duplicate scenario name check in `ValidateChange` — track scenario names per requirement
- [x] Add tests for duplicate detection in `internal/validate_test.go`

## Phase 3: Keyword and Content Validation

- [x] Replace `strings.Contains(req.Content, "SHALL")` with regex `\b(SHALL|MUST)\b` matching
- [x] Strip fenced code blocks (content between ``` markers) AND inline code (content between backticks) before keyword check
- [x] Add scenario content validation — check for WHEN and THEN markers in scenario body (case-sensitive substring)
- [x] Add tests for false positive rejection and content validation

## Phase 4: Cross-Operation and Overlap Fixes

- [x] Add within-delta cross-operation conflict detection in `ValidateChange` — build a map of requirement names to operations, report conflicts (including RENAMED old name vs MODIFIED/REMOVED, and RENAMED new name vs ADDED)
- [x] Fix `DetectOverlaps` in deps.go to use `OldName` for RENAMED requirements when comparing against MODIFIED/REMOVED operations
- [x] Add `File` field to dependency validation errors in `ValidateChange` (use the .litespec.yaml path)
- [x] Add tests for cross-operation conflicts and RENAMED overlap detection
