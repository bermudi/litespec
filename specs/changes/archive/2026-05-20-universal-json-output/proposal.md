# Universal `--json` and `--minimal` Output

## Motivation

litespec's `--json` flag is inconsistent. It exists on `status`, `validate`, `list`, `instructions`, `preview`, `view`, `new`, and `patch` — but is missing from `init`, `archive`, `decide`, `update`, `upgrade`, and `import`. Agents consuming litespec output get structured JSON from some commands and unparseable prose from others. This makes scripting fragile — consumers need to handle two entirely different output modes depending on which command they run.

Meanwhile, the existing text output is optimized for humans reading terminals. It includes decorative characters (progress bars, box-drawing, separators, tips). When an LLM reads it as context, most of those tokens are noise. There's no "just the facts" mode that strips prose and emits only actionable signal.

## Scope

**Capability affected:** `cli-input` (new requirements for `--json`/`--minimal` flag handling)

### What's included

1. **Universal `--json` on every command that produces output**, except `completion` (pure shell script) and `__complete` (internal completion backend). Specifically adding `--json` to:
   - `init` — returns what was created (dirs, skills, adapters)
   - `archive` — returns the archive result (validated specs, change moved)
   - `decide` — returns the created decision (number, slug, file path)
   - `update` — returns what was updated (skills, adapters)
   - `upgrade` — returns current/latest version, whether upgrade happened
   - `import` — returns import stats (canon specs, changes, archives, warnings)

2. **`--minimal` flag** on every command that also has `--json`. Outputs only actionable signal — no prose, no tips, no decorative formatting. Designed for LLM context windows. For commands that already have `--json`, `--minimal` outputs a subset of the JSON fields (only the action-relevant ones). For text output, `--minimal` strips everything except the core data.

3. **Register both flags in `commandspec.go`** so shell completion discovers them.

4. **Update help text** for every affected command.

### What's NOT included

- Changing `completion` or `__complete` — they're not data commands
- Switching from JSON to TOON or any other serialization format
- Changing the structure of existing JSON output (backward compatible additions only)
- Any changes to skill templates
