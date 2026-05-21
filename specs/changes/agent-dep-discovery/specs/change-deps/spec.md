# change-deps

## MODIFIED Requirements

### Requirement: Dependency Metadata Field

The `ChangeMeta` struct SHALL support an optional `dependsOn` field containing a list of change names. When absent, the change has no dependencies. The field SHALL be read from and written to `.litespec.yaml` alongside existing `schema` and `created` fields.

The `new` command SHALL append a YAML comment line `# dependsOn: []` after the `schema` and `created` fields in the generated `.litespec.yaml`. This comment serves as a discovery hint — agents reading the file learn the field exists without needing external documentation. The comment MUST NOT be valid YAML (no uncommented empty list) to avoid producing a non-nil `DependsOn` on parse.

#### Scenario: Metadata with dependencies

- **WHEN** a change's `.litespec.yaml` contains `dependsOn: [add-user-auth]`
- **THEN** the parsed `ChangeMeta` has `DependsOn` equal to `["add-user-auth"]`

#### Scenario: Metadata without dependencies

- **WHEN** a change's `.litespec.yaml` has no `dependsOn` field
- **THEN** the parsed `ChangeMeta` has `DependsOn` equal to `nil`

#### Scenario: Generated file contains dependency hint

- **WHEN** a new change is created via `litespec new <name>`
- **THEN** the `.litespec.yaml` file contains a `# dependsOn: []` comment line after the `created` field

## ADDED Requirements

### Requirement: Dependency Discovery in Agent Workflow

The `status --json` command SHALL include a `dependsOn` field mirroring the change's declared dependencies. The propose skill SHALL instruct agents to check for related active changes and declare `dependsOn` in `.litespec.yaml` when the new change builds on or must follow another change.

#### Scenario: Status output reflects dependencies

- **WHEN** a change has `dependsOn: [add-auth]` in its `.litespec.yaml`
- **THEN** `litespec status <name> --json` includes `"dependsOn": ["add-auth"]` in the response

#### Scenario: Status output with no dependencies

- **WHEN** a change has no `dependsOn` field
- **THEN** `litespec status <name> --json` omits the `dependsOn` key from the response

#### Scenario: Agent declares dependency during propose

- **WHEN** an agent creates a change that modifies capabilities also touched by an active change "add-auth"
- **THEN** the agent sets `dependsOn: [add-auth]` in `.litespec.yaml` after running `litespec new`
