## Motivation

`specs/config.yaml` exists solely to persist the `--tools claude` flag between `init` and `update`. It's a 2-line file that directly contradicts the project's stated principle: "convention over configuration — zero config files." OpenSpec ships a stub config.yaml that nobody fills in, and litespec was designed to skip this pattern. Yet here it is.

The only data stored is `tools: [claude]`, which is used to re-generate `.claude/skills/` symlinks on `update`. But the symlinks themselves already encode this information — if `.claude/skills/` exists and contains symlinks pointing at `.agents/skills/`, the tool adapter is active. The filesystem *is* the config.

## Scope

- **Remove** `specs/config.yaml` from the project and stop generating it
- **Remove** `internal/config.go`, `internal/config_test.go`, `ConfigFileName`, `ConfigPath`, `ReadProjectConfig`, `WriteProjectConfig`, `ProjectConfig` type, and `saveToolIDs` helper
- **Replace** config-based tool discovery with filesystem auto-detection: `init` and `update` detect existing adapter symlinks and refresh them when `--tools` is not explicitly provided
- **Update** `init` and `update` to detect active adapters by scanning for existing adapter skill directories (e.g., `.claude/skills/`) that contain symlinks into `.agents/skills/`
- **Update** tests that reference config reading/writing
- **Update** docs (cli-reference, getting-started, project-structure) to remove config references

## Non-Goals

- Adding new tool adapters (only claude exists today)
- Changing the adapter symlink mechanism itself (still symlinks into `.agents/skills/`)
- Modifying how `--tools` works when explicitly provided (validation, symlink creation unchanged)
- Removing the `--tools` flag (still needed for first-time setup or adding new adapters)
