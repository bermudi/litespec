# cli-input

## ADDED Requirements

### Requirement: Auto-Detection of Active Tool Adapters

The `init` and `update` commands MUST detect active tool adapters by scanning for existing adapter skill directories (e.g., `.claude/skills/`) that contain at least one symlink pointing into `.agents/skills/`. When `--tools` is not provided and active adapters are detected, the commands SHALL refresh symlinks for those adapters. When `--tools` is provided explicitly, only the listed adapters SHALL have symlinks generated; any previously active adapters not in the explicit list SHALL be ignored. The commands SHALL NOT read from any config file to determine tool adapters.

#### Scenario: Auto-detect claude adapter on update

- **WHEN** `litespec update` is run without `--tools` and `.claude/skills/` contains symlinks into `.agents/skills/`
- **THEN** the claude adapter symlinks are refreshed without requiring `--tools`

#### Scenario: Auto-detect claude adapter on init

- **WHEN** `litespec init` is run without `--tools` and `.claude/skills/` already contains symlinks into `.agents/skills/`
- **THEN** the claude adapter symlinks are refreshed without requiring `--tools`

#### Scenario: Explicit --tools overrides auto-detection on update

- **WHEN** `litespec update --tools claude` is run
- **THEN** symlinks are created only for claude regardless of what exists on disk

#### Scenario: Explicit --tools overrides auto-detection on init

- **WHEN** `litespec init --tools claude` is run and `.claude/skills/` already contains symlinks into `.agents/skills/`
- **THEN** symlinks are created only for claude

#### Scenario: No adapters detected, no --tools provided

- **WHEN** `litespec update` is run without `--tools` and no adapter skill directories contain symlinks into `.agents/skills/`
- **THEN** only `.agents/skills/` is updated, no adapter output is produced, and the command succeeds

#### Scenario: Adapter directory exists but contains no symlinks

- **WHEN** `.claude/skills/` exists but contains no symlinks pointing into `.agents/skills/`
- **THEN** the adapter is considered inactive and no adapter symlinks are generated

### Requirement: No Config File for Tool Persistence

The `init` and `update` commands MUST NOT create or read a config file for tool persistence. Tool adapter state SHALL be inferred entirely from the filesystem.

#### Scenario: init does not create config.yaml

- **WHEN** `litespec init --tools claude` is run
- **THEN** `specs/config.yaml` is not created

#### Scenario: update does not create config.yaml

- **WHEN** `litespec update --tools claude` is run
- **THEN** `specs/config.yaml` is not created or modified

#### Scenario: update works without config.yaml

- **WHEN** `litespec update` is run and `specs/config.yaml` does not exist
- **THEN** the command succeeds by auto-detecting adapters from the filesystem
