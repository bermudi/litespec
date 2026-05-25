# Migrating to v0.19.0

v0.19.0 consolidates the skill system from 11 skills down to 4. If you're upgrading from v0.18.x or earlier, this guide covers what changed and what you need to do.

## The Short Version

```bash
litespec upgrade
litespec update          # regenerates skills with new names
```

Done. The old skill directories are cleaned up automatically.

---

## What Changed

### Skills consolidated from 6 to 4

| Before (v0.18.x) | After (v0.19.0) | What happened |
|-------------------|-----------------|---------------|
| `litespec-explore` | merged into `litespec-think` | Exploration is now a mode within think |
| `litespec-grill` | merged into `litespec-think` | Grilling is now a mode within think |
| `litespec-workflow` | merged into `litespec-think` | Workflow routing is now a mode within think |
| `litespec-propose` | merged into `litespec-plan` | Proposing is now a mode within plan |
| `litespec-adopt` | merged into `litespec-plan` | Adopting is now a mode within plan |
| `litespec-patch` | merged into `litespec-plan` | Patching is now a mode within plan |
| `litespec-glossary` | merged into `litespec-plan` | Glossary management is now part of plan |
| `litespec-apply` | merged into `litespec-build` | Implementation + review-finding fixes + research pauses |
| `litespec-fix` | merged into `litespec-build` | Fixing review findings is now a workflow within build |
| `litespec-research` | merged into `litespec-build` | Research is now an inline pause within build |
| `litespec-review` | `litespec-review` | Unchanged |

The behavior is the same — no capabilities were lost. The skills just have fewer, more focused entry points.

### Why

Eleven skills meant eleven files to maintain, eleven descriptions to keep in sync, and eleven decision points for agents figuring out which skill to invoke. In practice, explore and grill were always used together (think first, then stress-test). Propose and adopt shared most of their artifact-generation logic. Apply and fix were the same implementation cycle.

Four skills map cleanly to four stances: **think** (no artifacts), **plan** (materialize artifacts), **build** (implement code), **review** (evaluate).

### New universal flags

`--json` and `--minimal` are now available on **every** command. Previously `--json` was limited to four commands.

### New commands documented

These commands existed before v0.19.0 but weren't in the docs. They're now fully documented in the [CLI Reference](cli-reference.md):

- `litespec decide <slug>` — create architectural decision records
- `litespec patch <name> <capability>` — create a lightweight delta-only change
- `litespec preview <name>` — preview what archive would do without making changes

### New flags

| Flag | Commands | Purpose |
|------|----------|---------|
| `--decisions` | `list`, `validate` | List or validate architectural decision records |
| `--backlog` | `list` | List backlog items by section |
| `--status <state>` | `list` | Filter decisions by status (proposed/accepted/superseded) |
| `--sort number` | `list` | Sort decisions by number |
| `--type decision` | `validate` | Disambiguate name as a decision |

---

## Migration Steps

### 1. Upgrade

```bash
litespec upgrade
```

### 2. Regenerate skills

```bash
litespec update
```

This generates the four new skill directories:

```
.agents/skills/
├── litespec-think/
├── litespec-plan/
├── litespec-build/
└── litespec-review/
```

Old skill directories (`litespec-explore/`, `litespec-grill/`, `litespec-propose/`, `litespec-apply/`, `litespec-adopt/`, `litespec-fix/`, `litespec-research/`, `litespec-workflow/`, `litespec-glossary/`, `litespec-patch/`) are removed automatically. Any manually created skills (e.g., `research-<topic>/`, `the-drill/`, `skill-creator/`) are left untouched.

### 3. Refresh tool adapters

If you're using Claude Code:

```bash
litespec update --tools claude
```

This recreates the symlinks in `.claude/skills/` to point at the new skill directories.

### 4. Update custom skills or prompts

If you have custom agent prompts, system prompts, or workflow instructions that reference the old skill names, update them:

| Old reference | New reference |
|---------------|---------------|
| "use the explore skill" | "use the think skill (Exploration mode)" |
| "use the grill skill" | "use the think skill (Grilling mode)" |
| "use the propose skill" | "use the plan skill (Propose mode)" |
| "use the apply skill" | "use the build skill" |
| "use the adopt skill" | "use the plan skill (Adopt mode)" |
| "use the patch skill" | "use the plan skill (Patch mode)" |
| "use the fix skill" | "use the build skill" |
| "use the research skill" | "use the build skill (research pause)" |
| "use the glossary skill" | "use the think or plan skill (glossary management)" |
| `litespec-explore` | `litespec-think` |
| `litespec-grill` | `litespec-think` |
| `litespec-workflow` | `litespec-think` |
| `litespec-propose` | `litespec-plan` |
| `litespec-adopt` | `litespec-plan` |
| `litespec-patch` | `litespec-plan` |
| `litespec-glossary` | `litespec-plan` |
| `litespec-apply` | `litespec-build` |
| `litespec-fix` | `litespec-build` |
| `litespec-research` | `litespec-build` |

---

## Nothing Else Changed

- Spec format is unchanged
- Delta spec syntax is unchanged
- Archive behavior is unchanged
- CLI command names (other than skills) are unchanged
- Directory structure is unchanged
- Shell completions are unchanged

This is a documentation and skill-naming release. Your existing projects, specs, and changes will work exactly as before.
