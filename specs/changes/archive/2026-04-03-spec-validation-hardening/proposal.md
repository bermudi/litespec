## Motivation

The validation system has gaps that allow malformed specs to pass: empty requirement/scenario names slip through, SHALL/MUST keywords are matched inside code blocks producing false positives, and cross-operation conflicts within a single delta go undetected. These gaps mean invalid specs reach archive time and fail with confusing merge errors instead of being caught early at validation time.

## Scope

- Reject empty requirement and scenario names at parse time
- Detect duplicate requirement names within a single delta spec
- Detect duplicate scenario names within a single requirement
- Validate scenario content contains WHEN/THEN markers (not just count)
- Match SHALL/MUST as whole words, not substrings
- Detect cross-operation conflicts within a single delta (e.g., RENAMED + MODIFIED on same name)
- Fix RENAMED overlap detection to use OldName for matching against other operations
- Add file context to dependency validation errors

## Non-Goals

- Cross-delta RENAME chain validation (separate proposal if needed)
- Changing the delta spec format or parser behavior
- Adding new delta operations
