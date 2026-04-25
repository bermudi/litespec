# skill-generation

## MODIFIED Requirements

### Requirement: Skill Generation Tests

The `internal/skill/` package SHALL have test coverage for template registration, frontmatter marshaling, and the skill-to-skill consistency of generated output. Tests SHALL use standard Go testing patterns. The expected skill list in tests MUST NOT include `continue`, MUST NOT include `archive`, and MUST use `review` instead of `verify`. The expected skill list MUST include `patch` alongside the other litespec-* skills.

#### Scenario: Tests SHALL verify template registration

- **WHEN** `go test ./internal/skill/` is run
- **THEN** tests SHALL verify that `Get` returns non-empty content for all known skill IDs (explore, grill, propose, review, apply, adopt, glossary, patch)
