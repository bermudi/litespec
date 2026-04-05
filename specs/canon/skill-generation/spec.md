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

The `internal/skill/` package SHALL have test coverage for template registration, frontmatter marshaling, and the skill-to-skill consistency of generated output. Tests SHALL use standard Go testing patterns. The expected skill list in tests MUST NOT include `continue` and MUST use `review` instead of `verify`.

#### Scenario: Tests SHALL verify template registration

- **WHEN** `go test ./internal/skill/` is run
- **THEN** tests SHALL verify that `Get` returns non-empty content for all known skill IDs (explore, grill, propose, review, apply, adopt, archive)

#### Scenario: Tests SHALL verify frontmatter format

- **WHEN** `go test ./internal/skill/` is run
- **THEN** tests SHALL verify that generated skill files contain valid YAML frontmatter with name and description fields
