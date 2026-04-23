# skill-generation

## ADDED Requirements

### Requirement: Skill Templates Reference Backlog

The skill templates for explore, propose, review, and grill SHALL include a prompt instructing the AI to read `specs/backlog.md` if it exists. The prompt SHALL be a single directive within each skill template, not programmatic integration. The explore skill SHALL read backlog for session context. The propose skill SHALL suggest graduating backlog items when a proposal materializes one. The review skill SHALL suggest adding deferred scope to the backlog. The grill skill SHALL reference backlog items to challenge scope boundaries.

#### Scenario: Explore skill reads backlog

- **WHEN** the explore skill template is rendered
- **THEN** it contains a directive to read `specs/backlog.md` for context on parked items

#### Scenario: Propose skill suggests graduation

- **WHEN** the propose skill template is rendered
- **THEN** it contains a directive to check if the proposal materializes a backlog item and suggest removing it

#### Scenario: Review skill suggests deferral

- **WHEN** the review skill template is rendered
- **THEN** it contains a directive to suggest adding deferred scope to `specs/backlog.md`

#### Scenario: Grill skill challenges scope

- **WHEN** the grill skill template is rendered
- **THEN** it contains a directive to read backlog and challenge scope overlaps
