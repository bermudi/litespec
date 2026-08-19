# skill-generation

## Requirements

### Requirement: Skill Registry Contains Exactly Three Generated Skills

The `Skills` slice in `internal/paths.go` SHALL contain exactly three `SkillInfo` entries with ID `plan` and name `litespec-plan`, ID `build` and name `litespec-build`, and ID `review` and name `litespec-review`. The `Skills` slice SHALL NOT contain a `think` skill, and it SHALL NOT contain the legacy skill IDs `explore`, `grill`, `propose`, `research`, `apply`, `adopt`, `workflow`, `glossary`, `patch`, or `fix`.

#### Scenario: Exactly three generated skills

- **WHEN** the `Skills` variable is inspected
- **THEN** it contains exactly three entries with IDs `plan`, `build`, and `review` and names `litespec-plan`, `litespec-build`, and `litespec-review`

#### Scenario: No legacy or think skill IDs

- **WHEN** the `Skills` variable is inspected
- **THEN** it does not contain the IDs `think`, `explore`, `grill`, `propose`, `research`, `apply`, `adopt`, `workflow`, `glossary`, `patch`, or `fix`

### Requirement: Skill Templates Are Loaded and Validated

Each skill ID in the `Skills` slice SHALL have a corresponding template file at `internal/skill/templates/<id>.md` that is loaded from the `embed.FS` and registered by `loadTemplates()` during package `init()`. The `GenerateSkills` function MUST return an error formatted `skill %s: template not registered` when a skill has no registered template. An empty `Skills` slice SHALL cause `GenerateSkills` to succeed without generating files.

#### Scenario: Missing skill template returns error

- **WHEN** `GenerateSkills` is called and a skill in the `Skills` list has no registered template
- **THEN** it returns an error in the format `skill <id>: template not registered`

#### Scenario: All skill templates registered

- **WHEN** `GenerateSkills` is called and all skills have registered templates
- **THEN** `.agents/skills/<name>/SKILL.md` is generated for each skill without errors

#### Scenario: Empty skills list succeeds

- **WHEN** the `Skills` slice is empty and `GenerateSkills` is called
- **THEN** no error is returned and no files are generated

### Requirement: Adapter Commands Generate Symlinks

The `Adapters` slice in `internal/paths.go` SHALL contain exactly one `ToolAdapter` with ID `claude` and skills directory `.claude/skills`. The `GenerateAdapterCommands` function MUST return an error formatted `skill %s: template not registered for adapter %s` when a skill lacks a registered template. When all skills are registered, `GenerateAdapterCommands` SHALL create a symlink in the adapter's skills directory for each generated skill, with the link pointing to the corresponding directory under `.agents/skills/`. Both `litespec init --tools claude` and `litespec update --tools claude` SHALL trigger adapter symlink generation.

#### Scenario: Adapter generation with missing template

- **WHEN** `GenerateAdapterCommands` is called for `claude` and a skill has no registered template
- **THEN** it returns an error in the format `skill <id>: template not registered for adapter claude`

#### Scenario: Adapter symlinks created on init and update

- **WHEN** `litespec init --tools claude` or `litespec update --tools claude` is run
- **THEN** `.claude/skills/litespec-plan`, `.claude/skills/litespec-build`, and `.claude/skills/litespec-review` are symlinks pointing to `.agents/skills/litespec-*`

### Requirement: Reference Files Are Generated

A generated skill MAY include reference files under `internal/skill/templates/references/<id>/`. The `GenerateSkills` function SHALL copy each such reference file into `.agents/skills/<name>/references/<file>.md`. Local `references/...` paths in a generated `SKILL.md` SHALL resolve to files in that skill's generated references directory. The `litespec-plan` skill SHALL have the five references `fuzzy.md`, `clear.md`, `grilling.md`, `codebase-design.md`, and `domain-modeling.md`. The `litespec-build` skill SHALL have the reference `review-fixing.md`. The `litespec-review` skill SHALL have the reference `adversarial-review.md`.

#### Scenario: Plan references generated

- **WHEN** `litespec update` is run
- **THEN** `.agents/skills/litespec-plan/references/` contains `fuzzy.md`, `clear.md`, `grilling.md`, `codebase-design.md`, and `domain-modeling.md`

#### Scenario: Build and review references generated

- **WHEN** `litespec update` is run
- **THEN** `.agents/skills/litespec-build/references/review-fixing.md` and `.agents/skills/litespec-review/references/adversarial-review.md` exist

#### Scenario: Generated skill references resolve

- **WHEN** all skills are generated
- **THEN** every local `references/...` path in each generated `SKILL.md` resolves to a generated reference file

### Requirement: litespec update Generates Canonical Skills

The `litespec update` command SHALL invoke `GenerateSkills` to write each generated skill to `.agents/skills/<name>/SKILL.md` from the skill's `SkillInfo` metadata and its registered template, with a YAML frontmatter block containing `name` and `description`. `.agents/skills/` SHALL be the canonical skills directory. `litespec update` SHALL also remove stale `litespec-*` directories under `.agents/skills/` that are not in the active `Skills` list.

#### Scenario: Update generates canonical skill files

- **WHEN** `litespec update` is run
- **THEN** `.agents/skills/litespec-plan/SKILL.md`, `.agents/skills/litespec-build/SKILL.md`, and `.agents/skills/litespec-review/SKILL.md` are generated with frontmatter and template content

#### Scenario: Update removes stale legacy directories

- **WHEN** `litespec update` is run and `.agents/skills/` contains a stale `litespec-explore` directory
- **THEN** that directory is removed

### Requirement: Skill Generation Does Not Follow Symlinks

Before generating skills, `GenerateSkills` SHALL reject symlinks in `.agents/skills/`, its skill directories, generated files, resource files, and every parent directory in those paths. It SHALL return an error identifying the refused symlink and SHALL NOT write through it.

#### Scenario: Symlinked generated file is refused

- **WHEN** `.agents/skills/litespec-plan/SKILL.md` is a symlink to a file outside the project
- **THEN** `litespec update` returns an error refusing to generate through the symlink and leaves the target file unchanged

### Requirement: Adapter Auto-Detection Scans Symlinks

The `DetectActiveAdapters` function SHALL scan each configured adapter skills directory for symlinks whose resolved target lies inside `.agents/skills/`. It SHALL return the IDs of adapters for which at least one such symlink exists. The `litespec update` command SHALL use `DetectActiveAdapters` when no `--tools` flag is provided.

#### Scenario: Active claude adapter detected

- **WHEN** `.claude/skills/` contains a symlink whose target resolves inside `.agents/skills/`
- **THEN** `DetectActiveAdapters` returns `claude` and `litespec update` recreates the symlinks

#### Scenario: No adapter detected

- **WHEN** no adapter skills directory contains a symlink pointing into `.agents/skills/`
- **THEN** `DetectActiveAdapters` returns an empty list and `litespec update` creates no adapter symlinks

### Requirement: Stale Skill Directories Are Detected

The `CheckStaleSkills` function SHALL detect directories under `.agents/skills/` whose names begin with `litespec-` but are not in the active `Skills` list, and it SHALL return a warning message naming the stale directories and instructing the user to run `litespec update`. It SHALL ignore directories that do not begin with `litespec-`, such as `skill-creator`, `the-drill`, or `research-vision`.

#### Scenario: Legacy litespec directories detected

- **WHEN** `.agents/skills/` contains `litespec-explore`, `litespec-grill`, or `litespec-propose`
- **THEN** `CheckStaleSkills` returns a message naming those directories and containing `litespec update`

#### Scenario: Non-litespec directories ignored

- **WHEN** `.agents/skills/` contains `skill-creator`, `the-drill`, or `research-vision`
- **THEN** `CheckStaleSkills` returns an empty string

#### Scenario: Current litespec skills not reported

- **WHEN** `.agents/skills/` contains `litespec-plan`, `litespec-build`, and `litespec-review`
- **THEN** `CheckStaleSkills` returns an empty string

### Requirement: Project-Specific Skills Are Not Generated

Project-specific skills, such as `the-drill`, SHALL live directly in `.agents/skills/` as tracked git files. They SHALL NOT appear in the `Skills` slice, and `litespec update` SHALL NOT generate or overwrite them. The `CheckStaleSkills` function SHALL ignore them.

#### Scenario: Project-specific skill preserved

- **WHEN** `litespec update` is run and `.agents/skills/the-drill/` is a tracked git directory
- **THEN** `litespec update` does not overwrite or remove `the-drill`

#### Scenario: Project-specific skill not in Skills slice

- **WHEN** the `Skills` variable is inspected
- **THEN** it does not contain `the-drill`

### Requirement: ValidateSkillTemplates Reports Missing Templates

The `ValidateSkillTemplates` function SHALL exist and SHALL return a slice of skill IDs from the supplied list that have no registered template. It SHALL return an empty slice, not `nil`, when every supplied skill ID has a registered template.

#### Scenario: All templates valid

- **WHEN** `ValidateSkillTemplates` is called with the current skill IDs and all templates are registered
- **THEN** it returns an empty, non-nil slice

#### Scenario: Missing template detected

- **WHEN** `ValidateSkillTemplates` is called and a skill ID has no registered template
- **THEN** the returned slice contains that skill ID
