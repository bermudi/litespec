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

- **WHEN** `ValidateSkillTemplates` is called and skill "think" has no template
- **THEN** the returned list contains "think"

### Requirement: Skill Generation Tests

The `internal/skill/` package SHALL have test coverage for template registration, frontmatter marshaling, and the skill-to-skill consistency of generated output. Tests SHALL use standard Go testing patterns. The expected skill list in tests MUST include exactly the four skills: think, plan, build, review. The expected skill list MUST NOT include legacy skill IDs (explore, grill, propose, research, apply, adopt, workflow, glossary, patch, fix).

#### Scenario: Tests SHALL verify template registration

- **WHEN** `go test ./internal/skill/` is run
- **THEN** tests SHALL verify that `Get` returns non-empty content for all current skill IDs (think, plan, build, review)

### Requirement: Skill Templates Reference Backlog

The skill templates for think, plan, and review SHALL include a prompt instructing the AI to read `specs/backlog.md` if it exists. The prompt SHALL be a single directive within each skill template, not programmatic integration. The think skill SHALL read backlog for session context and SHALL read `specs/glossary.md` at session start to establish shared vocabulary, nudging the user when it encounters terms that should be defined. If no glossary exists, the think skill SHALL suggest creating one when stable terms emerge. The plan skill SHALL suggest graduating backlog items when a proposal materializes one and SHALL check whether new terms introduced in the proposal exist in the glossary, offering to update it. The review skill SHALL suggest adding deferred scope to the backlog.

#### Scenario: Think skill reads backlog

- **WHEN** the think skill template is rendered
- **THEN** it contains a directive to read `specs/backlog.md` for context on parked items

#### Scenario: Think skill reads glossary

- **WHEN** the think skill template is rendered
- **THEN** it contains a directive to read `specs/glossary.md` if it exists at session start and nudge when undefined terms are encountered

#### Scenario: Think skill degrades without glossary

- **WHEN** the think skill template is rendered
- **THEN** it contains a directive to suggest creating `specs/glossary.md` when stable terms emerge and no glossary exists

#### Scenario: Plan skill suggests graduation

- **WHEN** the plan skill template is rendered
- **THEN** it contains a directive to check if the proposal materializes a backlog item and suggest removing it

#### Scenario: Plan skill checks glossary

- **WHEN** the plan skill template is rendered
- **THEN** it contains a directive to check whether new terms are in `specs/glossary.md` and offer to update it

#### Scenario: Review skill suggests deferral

- **WHEN** the review skill template is rendered
- **THEN** it contains a directive to suggest adding deferred scope to `specs/backlog.md`

### Requirement: Think Skill Contains Workflow Routing

The think skill template SHALL include workflow phase detection: it SHALL run `litespec status <name> --json` (or equivalent) to determine the current phase and suggest the appropriate next action. When no active change exists, the think skill SHALL explain the litespec workflow and help the user decide whether to explore, grill, or create a new change.

#### Scenario: Think skill detects current phase

- **WHEN** the think skill template is rendered and an active change exists
- **THEN** it contains a directive to run `litespec status` and suggest next steps based on the current phase

#### Scenario: Think skill explains workflow when no change exists

- **WHEN** the think skill template is rendered and no active change exists
- **THEN** it contains a directive to explain the litespec workflow and help the user choose a starting point

### Requirement: Plan Skill Contains Patch Mode

The plan skill template SHALL detect patch mode from the `.litespec.yaml` file within the change directory. When mode is `patch`, the plan skill SHALL skip creation of `proposal.md`, `design.md`, and `tasks.md`, proceeding directly to delta spec creation and validation.

#### Scenario: Plan skill detects patch mode

- **WHEN** the plan skill template is rendered and the change's `.litespec.yaml` contains `mode: patch`
- **THEN** it instructs the agent to skip proposal, design, and tasks creation and proceed directly to delta spec editing

#### Scenario: Plan skill creates full artifacts in default mode

- **WHEN** the plan skill template is rendered and the change has no `mode: patch` in `.litespec.yaml`
- **THEN** it instructs the agent to create all planning artifacts (proposal, specs, design, tasks) as applicable

### Requirement: Documentation Reflects Four Skills

Project documentation files `AGENTS.md` and `DESIGN.md` SHALL reference only the four current skill IDs (think, plan, build, review) and SHALL NOT reference legacy skill IDs (explore, grill, propose, research, apply, adopt, workflow, glossary, patch, fix) as active skills.

#### Scenario: AGENTS.md references only current skills

- **WHEN** `AGENTS.md` is inspected
- **THEN** it does not list explore, grill, propose, research, apply, adopt, workflow, glossary, patch, or fix as separate skills

#### Scenario: DESIGN.md references only current skills

- **WHEN** `DESIGN.md` is inspected
- **THEN** the skill directory tree shows exactly four directories: litespec-think, litespec-plan, litespec-build, litespec-review

### Requirement: Glossary Management In Plan Skill

The `Skills` list in `internal/paths.go` MUST NOT include a standalone `glossary` skill. Glossary management SHALL be a section within the `plan` skill template (ID "plan", name "litespec-plan"). The plan skill template SHALL include instructions for the AI to read `specs/glossary.md`, propose new terms when it encounters undefined concepts, and maintain consistent formatting.

#### Scenario: No standalone glossary skill

- **WHEN** `litespec update` is run
- **THEN** `.agents/skills/litespec-glossary/SKILL.md` does not exist

#### Scenario: Glossary instructions in plan skill

- **WHEN** the plan skill template is rendered
- **THEN** it contains instructions for managing `specs/glossary.md` including reading, proposing terms, and formatting

#### Scenario: Glossary section handles missing file

- **WHEN** the plan skill is invoked and `specs/glossary.md` does not exist
- **THEN** the skill offers to create and seed the glossary file

### Requirement: Build Skill Contains Fix Workflow

The build skill template SHALL include a structured workflow for addressing review findings. The workflow MUST include: loading the review report and change artifacts, grouping findings by file and priority, addressing CRITICAL findings first followed by WARNING then SUGGESTION, verifying each fix individually before moving to the next, running `litespec validate <name>` after all fixes to confirm no structural regressions, and committing only after all findings are resolved. The skill SHALL escalate unresolvable findings as a new warning rather than silently dropping them.

#### Scenario: Build skill addresses findings in priority order

- **WHEN** the build skill template is rendered
- **THEN** it instructs the agent to address CRITICAL findings before WARNING before SUGGESTION

#### Scenario: Build skill verifies per finding

- **WHEN** the build skill template is rendered
- **THEN** it instructs the agent to verify each fix individually before proceeding to the next finding

#### Scenario: Build skill validates after all fixes

- **WHEN** the build skill template is rendered
- **THEN** it instructs the agent to run `litespec validate <name>` after all fixes are applied

#### Scenario: Build skill escalates unresolvable findings

- **WHEN** the build skill template is rendered
- **THEN** it instructs the agent to surface unresolvable findings as an explicit warning rather than silently dropping them

### Requirement: Build Skill Contains Research Pause

The build skill template SHALL include a pause condition: when the agent encounters a knowledge gap during implementation (novel APIs, unfamiliar libraries, non-obvious authentication flows), it SHALL gather the relevant documentation and MAY produce a research skill file at `.agents/skills/research-<topic>/SKILL.md` for future reference. The research skill file, if produced, SHALL use the skill-creator format conventions and persist after archive as accumulated project knowledge. This replaces the former standalone research skill as an inline step within the build workflow.

#### Scenario: Agent encounters knowledge gap during implementation

- **WHEN** the build skill is active and the agent needs documentation for an unfamiliar API
- **THEN** the agent pauses implementation, gathers docs, and optionally produces a research skill file

#### Scenario: Research skill persists after archive

- **WHEN** a research skill file was produced during build
- **THEN** the file remains in `.agents/skills/research-<topic>/` after the change is archived

### Requirement: Build Skill References Glossary

The build skill template SHALL include a passive reference to `specs/glossary.md` in a references section. The agent MAY consult the glossary for terminology after completing a phase. No enforcement, no nudge — purely optional context.

#### Scenario: Build skill references glossary

- **WHEN** the build skill template is rendered
- **THEN** it contains a reference to `specs/glossary.md` as optional terminology context, without enforcement directives

### Requirement: Four Skill Registration

The `Skills` list in `internal/paths.go` MUST contain exactly four entries with the following IDs and names:

| ID | Name | Description (concise) |
|----|------|----------------------|
| `think` | `litespec-think` | Explore ideas and stress-test plans for litespec changes. |
| `plan` | `litespec-plan` | Create or update litespec change proposals and patches. |
| `build` | `litespec-build` | Implement litespec changes, fix review findings, and research knowledge gaps. |
| `review` | `litespec-review` | Adversarial review of litespec artifacts or implementation. |

A corresponding Go template MUST be registered in `internal/skill/` for each ID via `init()`. Legacy skill IDs (explore, grill, propose, research, apply, adopt, workflow, glossary, patch, fix) SHALL NOT appear in the `Skills` list. Legacy template Go files (explore.go, grill.go, propose.go, research.go, apply.go, adopt.go, workflow.go, glossary.go, patch.go, fix.go) SHALL be removed from `internal/skill/`.

#### Scenario: Exactly four skills registered

- **WHEN** the `Skills` variable is inspected
- **THEN** it contains exactly four entries with IDs "think", "plan", "build", "review"

#### Scenario: No legacy skill IDs

- **WHEN** the `Skills` variable is inspected
- **THEN** none of the following IDs appear: explore, grill, propose, research, apply, adopt, workflow, glossary, patch, fix

#### Scenario: All four templates registered

- **WHEN** `Get("think")`, `Get("plan")`, `Get("build")`, `Get("review")` are called
- **THEN** non-empty template content is returned for each

#### Scenario: Skills are generated by update

- **WHEN** `litespec update` is run
- **THEN** `.agents/skills/litespec-think/SKILL.md`, `.agents/skills/litespec-plan/SKILL.md`, `.agents/skills/litespec-build/SKILL.md`, and `.agents/skills/litespec-review/SKILL.md` are generated

#### Scenario: Legacy skill directories are removed

- **WHEN** `litespec update` is run
- **THEN** directories for legacy skills (litespec-explore, litespec-grill, litespec-propose, litespec-research, litespec-apply, litespec-adopt, litespec-workflow, litespec-glossary, litespec-patch, litespec-fix) are removed from `.agents/skills/` if they exist

#### Scenario: Adapter symlinks for legacy skills are removed

- **WHEN** `litespec update` is run
- **THEN** symlinks in adapter skill directories (e.g., `.claude/skills/`) that reference legacy skill IDs are removed

#### Scenario: Legacy template Go files are removed

- **WHEN** the change is implemented
- **THEN** none of the following files exist in `internal/skill/`: explore.go, grill.go, propose.go, research.go, apply.go, adopt.go, workflow.go, glossary.go, patch.go, fix.go
