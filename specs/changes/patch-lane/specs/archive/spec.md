# archive

## MODIFIED Requirements

### Requirement: Strip Specs Subtree Before Archiving

After delta specs are merged into `specs/canon/`, the archive command MUST remove the change's `specs/` subtree before moving the change directory to `specs/changes/archive/`. The archived directory SHALL contain whichever planning artifacts existed in the source change directory plus `.litespec.yaml` if present. The `specs/` subtree MUST NOT be present in the archived directory. For full-proposal changes the archived directory typically contains `.litespec.yaml`, `proposal.md`, `design.md`, and `tasks.md`. For patch-mode changes (no planning artifacts present at archive time), the archived directory MAY contain only the merged spec history with no planning files; this is not an error.

#### Scenario: Archived full-proposal change has no specs subtree

- **WHEN** `litespec archive my-change` completes successfully on a change with `proposal.md`, `design.md`, `tasks.md`
- **THEN** the archived directory at `specs/changes/archive/<date>-my-change/` contains those planning files and does not contain a `specs/` subdirectory

#### Scenario: Archived patch-mode change has no specs subtree and no planning files

- **WHEN** `litespec archive my-patch` completes successfully on a patch-mode change (delta-only, no `proposal.md`)
- **THEN** the archived directory at `specs/changes/archive/<date>-my-patch/` does not contain a `specs/` subdirectory and is not required to contain `proposal.md`, `design.md`, or `tasks.md`

#### Scenario: Canon contains merged content

- **WHEN** `litespec archive my-change` completes successfully
- **THEN** `specs/canon/<capability>/spec.md` contains the merged result of all delta operations
