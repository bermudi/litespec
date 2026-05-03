# review

## ADDED Requirements

### Requirement: Fix Skill Handoff

The review skill template SHALL direct users to the fix skill when findings need to be addressed. The ending section of the review template MUST reference `litespec-fix` as the appropriate skill for resolving review findings, replacing the current instruction to "use apply." The review skill SHALL remain pure review — it MUST NOT write code, modify files, or implement fixes.

#### Scenario: Review directs users to fix skill

- **WHEN** the review skill template is rendered
- **THEN** the ending section instructs the user to use the fix skill for addressing review findings, not apply

#### Scenario: Review skill remains pure review

- **WHEN** the review skill template is rendered
- **THEN** it states that review mode is pure review and must never write code or implement fixes
