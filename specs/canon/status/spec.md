# status

## Requirements

### Requirement: Positional Name for Status

The `litespec status` command MUST accept an optional positional `<name>` argument instead of `--change <name>`. When provided, it shows artifact state for that specific change. When omitted, it shows all changes. The `--change` flag SHALL be removed.

#### Scenario: Status for a named change

- **WHEN** `litespec status my-feature` is run
- **THEN** artifact state for `my-feature` is shown

#### Scenario: Status with no arguments

- **WHEN** `litespec status` is run
- **THEN** all active changes are listed

#### Scenario: Status for nonexistent change

- **WHEN** `litespec status nonexistent` is run
- **THEN** an error is printed to stderr indicating the change was not found with exit code 1

### Requirement: Patch-Mode Status Display

When `litespec status <name>` (or bulk `litespec status`) encounters a patch-mode change (as defined by `IsPatchMode`), the command MUST display only the `specs` artifact line and a one-line indicator that the change is in patch mode. The proposal, design, and tasks artifacts SHALL be omitted from the per-change output (not displayed as `BLOCKED` or `READY`). In `--json` mode, patch-mode changes SHALL include a top-level `mode: "patch"` field and SHALL omit non-applicable artifacts from the `artifacts` array (or include them with `status: "n/a"`), so that downstream consumers can detect patch mode unambiguously.

#### Scenario: Patch-mode status shows only specs

- **WHEN** `litespec status my-patch` is run on a patch-mode change
- **THEN** the output displays the `specs` artifact line and an indicator like `(patch mode)`, and does not display `proposal`, `design`, or `tasks` lines

#### Scenario: Patch-mode JSON status includes mode field

- **WHEN** `litespec status my-patch --json` is run on a patch-mode change
- **THEN** the JSON output contains `"mode": "patch"` and the artifacts array reflects patch-mode classification (either by omission or `status: "n/a"`)

#### Scenario: Full-proposal status unchanged

- **WHEN** `litespec status my-feature` is run on a change with full planning artifacts
- **THEN** the output displays all four artifact lines as before, and JSON output does not contain `"mode": "patch"`

### Requirement: Patch-Mode Artifact States

The `LoadArtifactStates` function MUST recognize patch-mode changes and return only the `specs` artifact's state for them. The proposal, design, and tasks artifacts SHALL NOT appear in the returned map for patch-mode changes. `LoadChangeContext` SHALL preserve this behavior so all callers see a consistent view.

#### Scenario: LoadArtifactStates omits trio for patch mode

- **WHEN** `LoadArtifactStates(root, "my-patch")` is called on a patch-mode change with a valid delta
- **THEN** the returned map contains only the `specs` key with state `DONE`

#### Scenario: LoadArtifactStates returns full set for full-proposal change

- **WHEN** `LoadArtifactStates(root, "my-feature")` is called on a change with all four artifacts present
- **THEN** the returned map contains all four artifact IDs with their respective states
