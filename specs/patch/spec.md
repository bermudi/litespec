# patch

## Requirements

### Requirement: Patch Command Scaffold

The `litespec patch <name> <capability>` command MUST scaffold a delta-only change. It SHALL create the change directory at `specs/changes/<name>/` and a delta spec stub at `specs/changes/<name>/specs/<capability>/spec.md`. The stub SHALL contain a top-level `# <capability>` heading and an `## ADDED Requirements` section as a starting point. The command MUST NOT create `proposal.md`, `design.md`, or `tasks.md`. The command SHALL write a `.litespec.yaml` with `mode: patch` to declare the change's lane. The command SHALL print the path of the created stub and a brief next-step hint pointing at archive.

#### Scenario: Patch creates change with delta stub only

- **WHEN** `litespec patch add-foo-flag bar` is run in a litespec project where `add-foo-flag` does not exist
- **THEN** the directory `specs/changes/add-foo-flag/specs/bar/` is created with a `spec.md` file containing `# bar` and `## ADDED Requirements`, and `.litespec.yaml` contains `mode: patch`, and no `proposal.md`, `design.md`, or `tasks.md` exists in `specs/changes/add-foo-flag/`

#### Scenario: Patch refuses existing change

- **WHEN** `litespec patch foo bar` is run and `specs/changes/foo/` already exists
- **THEN** an error is printed to stderr indicating the change already exists, and no files are created

#### Scenario: Patch requires both arguments

- **WHEN** `litespec patch foo` is run without a capability argument
- **THEN** an error is printed to stderr indicating both name and capability are required, with exit code 1

#### Scenario: Patch validates name and capability

- **WHEN** `litespec patch <invalid> <invalid>` is run with names containing slashes or invalid characters
- **THEN** an error is printed to stderr indicating the names are invalid, with exit code 1

### Requirement: Patch-Mode Change Detection

The internal change-loading logic MUST recognize a patch-mode change via the `mode` field in `.litespec.yaml`. When `mode` is set to `patch`, the change is in patch mode. When the field is absent or any other value, the change is in full-proposal mode. Patch-mode detection SHALL be exposed via an `IsPatchMode(root, name string) bool` function in the internal package. This function is the single source of truth for patch-mode classification across status, view, and other consumers.

#### Scenario: Detect patch-mode change

- **WHEN** `IsPatchMode(root, "add-foo-flag")` is called and the change's `.litespec.yaml` contains `mode: patch`
- **THEN** the function returns true

#### Scenario: Full proposal change is not patch-mode

- **WHEN** `IsPatchMode(root, "big-feature")` is called and the change's `.litespec.yaml` has no `mode` field
- **THEN** the function returns false

#### Scenario: Empty change is not patch-mode

- **WHEN** `IsPatchMode(root, "stub")` is called and the change has no `.litespec.yaml`
- **THEN** the function returns false

### Requirement: Patch Lane Skill

A `litespec-patch` skill MUST be registered in the `Skills` list and have a corresponding template registered. The skill SHALL describe when to use the patch lane (small, single-capability changes; new flags; minor behavioral tweaks) and when not to (multi-capability changes, REMOVED requirements, anything needing design discussion → use `propose`). The skill SHALL instruct the AI to write the delta, implement, and archive without producing `proposal.md`, `design.md`, or `tasks.md`.

#### Scenario: Patch skill exists in registry

- **WHEN** the `Skills` list is inspected
- **THEN** an entry with ID `litespec-patch` exists with a non-empty description

#### Scenario: Patch skill template is registered

- **WHEN** `litespec update` is run in a litespec project
- **THEN** `.agents/skills/litespec-patch/SKILL.md` is generated with the registered template

### Requirement: Workflow Skill References Patch Lane

The `litespec-workflow` skill template SHALL document the patch lane as a sibling workflow to the propose flow. It SHALL describe the patch flow as `patch → archive` and explain when to choose patch over propose.

#### Scenario: Workflow skill mentions patch

- **WHEN** the generated `.agents/skills/litespec-workflow/SKILL.md` is inspected
- **THEN** it contains a section describing the patch lane and when to choose it
