## Architecture

This is a deletion-heavy change with one new concept: `DetectActiveAdapters`. The function lives in `internal/adapter.go` alongside the existing `GenerateAdapterCommands`. The overall data flow changes from:

```
init/update → read config.yaml → get toolIDs → generate adapters → write config.yaml
```

to:

```
init/update → detect active adapters from filesystem → get toolIDs → generate adapters
               (or use explicit --tools if provided)
```

No new files, no new packages. The change removes more code than it adds.

## Decisions

**Auto-detect by scanning adapter skill dirs for symlinks into `.agents/skills/`.**
- Chosen because it uses the filesystem as the source of truth — the symlinks themselves encode which adapters are active.
- `DetectActiveAdapters` iterates over the `Adapters` slice from `paths.go`, checking each adapter's `SkillsDir` for at least one symlink whose resolved target is inside `.agents/skills/`. It returns a `[]string` of adapter IDs (matching `ToolAdapter.ID`).
- Alternative considered: store tools list in a dotfile inside `.agents/`. Rejected because it's still a config file by another name.
- Constraint: detection relies on symlinks pointing at `.agents/skills/`. If a user manually creates non-symlink files in `.claude/skills/`, they won't be detected as adapter content. This is acceptable — adapter dirs are managed by litespec.

**Detection checks for at least one symlink, not exact skill count.**
- Chosen for robustness — the adapter is active if its dir has any symlinks, regardless of whether the skill list has changed since last generation.
- `update` will regenerate all symlinks for detected adapters regardless, so an incomplete set just means the next update fixes it.

**Delete `ProjectConfig` type entirely.**
- Chosen because no other code references it beyond `saveToolIDs`, `ReadProjectConfig`, and `WriteProjectConfig`.
- If a future need for project-level config arises, it can be re-introduced with a fresh design rather than carrying this stub.

## File Changes

- **`internal/config.go`** — DELETE. Contains `ReadProjectConfig`, `WriteProjectConfig`, `ConfigPath`. No longer needed.
- **`internal/config_test.go`** — DELETE. Tests for the deleted config functions.
- **`internal/types.go`** — MODIFY. Remove `ProjectConfig` struct.
- **`internal/paths.go`** — MODIFY. Remove `ConfigFileName` constant.
- **`internal/adapter.go`** — MODIFY. Add `DetectActiveAdapters(root string) []string` function that scans each registered adapter's skill directory for symlinks into `.agents/skills/`.
- **`cmd/litespec/helpers.go`** — MODIFY. Remove `saveToolIDs` function.
- **`cmd/litespec/init.go`** — MODIFY. Replace config-based fallback with `DetectActiveAdapters`. Remove config write after adapter generation.
- **`cmd/litespec/update.go`** — MODIFY. Same changes as init.go.
- **`specs/config.yaml`** — DELETE. The file that started all this.
- **`cmd/litespec/main_test.go`** — MODIFY. Update tests that reference config reading/writing, `ReadProjectConfig`, or `saveToolIDs`.
- **`docs/cli-reference.md`** — MODIFY. Remove references to config persistence.
- **`docs/getting-started.md`** — MODIFY. Update to explain auto-detection behavior.
- **`docs/project-structure.md`** — MODIFY. Remove `specs/config.yaml` from project structure references.
