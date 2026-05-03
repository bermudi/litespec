# skill-generation

## Requirements

### Requirement: Template Missing Error

The `GenerateSkills` function MUST return an error when a skill in the `Skills` list has no registered template. The error SHALL use the format `fmt.Errorf("skill %s: template not registered", skillID)`. If the `Skills` list is empty, the function SHALL succeed without error.

#### Scenario: Skill with missing template produces error

- **WHEN** `GenerateSkills` is called and a skill in the `Skills` list has no registered template
- **THEN** an error is returned in the format "skill <id>: template not registered"

#### Scenario: All skills have templates

- **WHEN** `GenerateSkills` is called and all skills have registered templates
- **THEN** all skill files are generated in `.agents/skills/` without errors

#### Scenario: Empty skills list succeeds

- **WHEN** `GenerateSkills` is called with an empty `Skills` list
- **THEN** no error is returned and no files are generated

### Requirement: Adapter Template Missing Error

The `GenerateAdapterCommands` function MUST return an error when a skill in the `Skills` list has no registered template. The error SHALL use the format `fmt.Errorf("skill %s: template not registered for adapter %s", skillID, toolID)`.

#### Scenario: Adapter generation with missing template

- **WHEN** `GenerateAdapterCommands` is called and a skill has no registered template
- **THEN** an error is returned in the format "skill <id>: template not registered for adapter <tool>"

#### Scenario: Adapter generation succeeds for all skills

- **WHEN** `GenerateAdapterCommands` is called and all skills have templates
- **THEN** symlinks are created for all skills in the adapter skills directory

### Requirement: Template Registration Validation

A `ValidateSkillTemplates` function MUST exist that checks every skill in the `Skills` list has a non-empty registered template. It SHALL return a slice of skill IDs that are missing templates (empty slice, not nil, when all are valid). This function MAY be called during `litespec validate` to catch registration issues early.

#### Scenario: All templates registered

- **WHEN** `ValidateSkillTemplates` is called and all skills have templates
- **THEN** an empty list is returned

#### Scenario: Missing template detected

- **WHEN** `ValidateSkillTemplates` is called and skill "explore" has no template
- **THEN** the returned list contains "explore"

### Requirement: Skill Generation Tests

The `internal/skill/` package SHALL have test coverage for template registration, frontmatter marshaling, and the skill-to-skill consistency of generated output. Tests SHALL use standard Go testing patterns. The expected skill list in tests MUST NOT include `continue`, MUST NOT include `archive`, and MUST use `review` instead of `verify`. The expected skill list MUST include `patch` alongside the other litespec-* skills.

#### Scenario: Tests SHALL verify template registration

- **WHEN** `go test ./internal/skill/` is run
- **THEN** tests SHALL verify that `Get` returns non-empty content for all known skill IDs (explore, grill, propose, review, apply, adopt, glossary, patch)

### Requirement: Skill Templates Reference Backlog

The skill templates for explore, propose, review, and grill SHALL include a prompt instructing the AI to read `specs/backlog.md` if it exists. The prompt SHALL be a single directive within each skill template, not programmatic integration. The explore skill SHALL read backlog for session context and SHALL read `specs/glossary.md` at session start to establish shared vocabulary, nudging the user when it encounters terms that should be defined. If no glossary exists, the explore skill SHALL suggest creating one when stable terms emerge. The propose skill SHALL suggest graduating backlog items when a proposal materializes one and SHALL check whether new terms introduced in the proposal exist in the glossary, offering to update it. The review skill SHALL suggest adding deferred scope to the backlog. The grill skill SHALL reference backlog items to challenge scope boundaries and SHALL read `specs/glossary.md` at session start, nudging when new terms emerge during the grilling process.

#### Scenario: Explore skill reads backlog

- **WHEN** the explore skill template is rendered
- **THEN** it contains a directive to read `specs/backlog.md` for context on parked items

#### Scenario: Explore skill reads glossary

- **WHEN** the explore skill template is rendered
- **THEN** it contains a directive to read `specs/glossary.md` if it exists at session start and nudge when undefined terms are encountered

#### Scenario: Explore skill degrades without glossary

- **WHEN** the explore skill template is rendered
- **THEN** it contains a directive to suggest creating `specs/glossary.md` when stable terms emerge and no glossary exists

#### Scenario: Propose skill suggests graduation

- **WHEN** the propose skill template is rendered
- **THEN** it contains a directive to check if the proposal materializes a backlog item and suggest removing it

#### Scenario: Propose skill checks glossary

- **WHEN** the propose skill template is rendered
- **THEN** it contains a directive to check whether new terms are in `specs/glossary.md` and offer to update it

#### Scenario: Review skill suggests deferral

- **WHEN** the review skill template is rendered
- **THEN** it contains a directive to suggest adding deferred scope to `specs/backlog.md`

#### Scenario: Grill skill challenges scope

- **WHEN** the grill skill template is rendered
- **THEN** it contains a directive to read backlog and challenge scope overlaps

#### Scenario: Grill skill reads glossary

- **WHEN** the grill skill template is rendered
- **THEN** it contains a directive to read `specs/glossary.md` if it exists at session start and nudge when new terms emerge

### Requirement: Glossary Skill Template

The `Skills` list in `internal/paths.go` MUST include a `glossary` skill entry with ID "glossary", name "litespec-glossary", and a description indicating it manages the project's ubiquitous language. A corresponding Go template MUST be registered in `internal/skill/glossary.go` via `init()`. The template SHALL instruct the AI to read `specs/glossary.md`, propose new terms when it encounters undefined concepts, and maintain consistent formatting.

#### Scenario: Glossary skill is generated

- **WHEN** `litespec update` is run
- **THEN** `.agents/skills/litespec-glossary/SKILL.md` is generated with valid frontmatter and glossary management instructions

#### Scenario: Glossary skill appears in skill list

- **WHEN** `litespec list --json` is run or `Skills` is inspected
- **THEN** a skill with ID "glossary" and name "litespec-glossary" is present

#### Scenario: Glossary skill handles missing glossary file

- **WHEN** the glossary skill is invoked and `specs/glossary.md` does not exist
- **THEN** the skill offers to create and seed the glossary file

### Requirement: Apply Skill Glossary Reference

The apply skill template SHALL include a passive reference to `specs/glossary.md` in a references section. The agent MAY consult the glossary for terminology after completing a phase. No enforcement, no nudge — purely optional context.

#### Scenario: Apply skill references glossary

- **WHEN** the apply skill template is rendered
- **THEN** it contains a reference to `specs/glossary.md` as optional terminology context, without enforcement directives

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
