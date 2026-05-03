# skill-generation

## ADDED Requirements

### Requirement: Fix Skill Registration

The `Skills` list in `internal/paths.go` MUST include a `fix` skill entry with ID "fix", name "litespec-fix", and a description indicating it addresses review findings. A corresponding Go template MUST be registered in `internal/skill/fix.go` via `init()`. The template SHALL instruct the AI to ingest review findings, address them in priority order (CRITICAL → WARNING → SUGGESTION), verify each fix individually, and commit when all are resolved.

#### Scenario: Fix skill is generated

- **WHEN** `litespec update` is run
- **THEN** `.agents/skills/litespec-fix/SKILL.md` is generated with valid frontmatter and fix workflow instructions

#### Scenario: Fix skill appears in skill list

- **WHEN** `litespec list --json` is run or `Skills` is inspected
- **THEN** a skill with ID "fix" and name "litespec-fix" is present

#### Scenario: Fix template is registered

- **WHEN** `Get("fix")` is called
- **THEN** non-empty template content is returned

### Requirement: Fix Skill Workflow

The fix skill template SHALL describe a structured workflow for addressing review findings. The workflow MUST include: loading the review report and change artifacts, grouping findings by file and priority, addressing CRITICAL findings first followed by WARNING then SUGGESTION, verifying each fix individually before moving to the next, running `litespec validate <name>` after all fixes to confirm no structural regressions, and committing only after all findings are resolved. The skill SHALL escalate unresolvable findings as a new warning rather than silently dropping them.

#### Scenario: Fix skill addresses findings in priority order

- **WHEN** the fix skill template is rendered
- **THEN** it instructs the agent to address CRITICAL findings before WARNING before SUGGESTION

#### Scenario: Fix skill verifies per finding

- **WHEN** the fix skill template is rendered
- **THEN** it instructs the agent to verify each fix individually before proceeding to the next finding

#### Scenario: Fix skill validates after all fixes

- **WHEN** the fix skill template is rendered
- **THEN** it instructs the agent to run `litespec validate <name>` after all fixes are applied

#### Scenario: Fix skill escalates unresolvable findings

- **WHEN** the fix skill template is rendered
- **THEN** it instructs the agent to surface unresolvable findings as an explicit warning rather than silently dropping them
