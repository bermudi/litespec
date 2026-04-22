# litespec — Design Document

AI-native spec-driven development tool. A leaner, opinionated reimagining of [OpenSpec](https://github.com/Fission-AI/OpenSpec).

## Stack

- **Language:** Go
- **Module:** `github.com/bermudi/litespec`
- **Binary:** `litespec`

## Directory Structure

```
project/
├── specs/
│   ├── canon/                    # source of truth (accepted capabilities)
│   │   └── <capability>/
│   │       └── spec.md
│   └── changes/                  # active changes
│       ├── <name>/
│       │   ├── .litespec.yaml    # metadata (schema + timestamp)
│       │   ├── proposal.md       # why & what
│       │   ├── design.md         # how (technical decisions)
│       │   ├── tasks.md          # phased implementation checklist
│       │   └── specs/            # delta specs
│       │       └── <capability>/
│       │           └── spec.md
│       └── archive/              # implemented changes (YYYY-MM-DD-<name>/)
│           └── <date>-<name>/
│               ├── .litespec.yaml
│               ├── proposal.md
│               ├── design.md
│               └── tasks.md      # planning artifacts
├── decisions/                    # architectural decision records (optional)
│   └── NNNN-<slug>.md
└── .agents/skills/               # generated skills (canonical)
    ├── litespec-explore/
    ├── litespec-grill/
    ├── litespec-propose/
    ├── litespec-research/
    ├── litespec-review/
    ├── litespec-apply/
    ├── litespec-adopt/
    └── research-<topic>/         # research skills (optional, produced by research phase)
```

## Workflow

Unidirectional flow:

```
explore → grill → propose → [research →] apply → review → archive
                                          │
                                      adopt (separate path)
```

No backward flow. If something is wrong after propose, start over from explore/grill. Research is optional — skip it when the change doesn't involve external dependencies.

## Skills

| Skill | Type | Behavior |
|-------|------|----------|
| `explore` | Ephemeral | Thinking mode. No artifacts, no change dir. Conversational. |
| `grill` | Ephemeral | Relentless Q&A on the explored idea. No artifacts. Resolves every branch of the design tree before proceeding. |
| `propose` | Materializes | Creates change dir + proposal + specs + design + tasks (all at once). This is the commit point. |
| `research` | Knowledge-gathering | Reads artifacts, identifies knowledge gaps, gathers docs/APIs/schemas, produces research skills into `.agents/skills/research-<topic>/`. Uses skill-creator conventions for formatting. |
| `apply` | Phase-based | Implements tasks per phase in `tasks.md`. One phase per invocation. AI focuses on one area without doing the whole implementation at once. Consumes research skills via natural discovery. |
| `review` | AI review | Context-aware review that adapts to change lifecycle: artifact review (0 tasks checked — evaluates planning artifacts for quality, consistency, readiness), implementation review (some tasks checked — compares code against specs), pre-archive review (all tasks checked — reviews both artifacts and code comprehensively before archiving). Pure AI review — no test/lint running. |
| `adopt` | Reverse-engineer | Takes a file/directory path. Generates a change proposal with specs from existing code. For code that has no spec yet. |

## Tasks (Phased)

Tasks are organized into phases for focused implementation:

```markdown
## Phase 1: Foundation
- [ ] Set up database schema
- [ ] Create migration

## Phase 2: Core Logic
- [ ] Implement auth service
- [ ] Add middleware

## Phase 3: Integration
- [ ] Wire up routes
- [ ] Add error handling
```

Phase tracking is derived from `tasks.md` — no metadata field. The first phase with unchecked tasks is the current phase.

## Delta Spec System

Full delta operations with semantic merging:

| Operation | Syntax | Merge Behavior |
|-----------|--------|----------------|
| `ADDED` | `## ADDED Requirements` | Append to end of main spec |
| `MODIFIED` | `## MODIFIED Requirements` | Replace matching requirement by header |
| `REMOVED` | `## REMOVED Requirements` | Delete from main spec |
| `RENAMED` | `## RENAMED Requirements` | Change section header, preserve content |

Applied in strict order at archive time:
1. `RENAMED` — establishes correct headers for subsequent operations
2. `REMOVED` — eliminates requirements before modifications
3. `MODIFIED` — updates remaining requirements
4. `ADDED` — appends new requirements

### Glossary Operations

Glossary entries support the same delta operations as requirements:

| Operation | Syntax | Merge Behavior |
|-----------|--------|----------------|
| `ADDED` | `## ADDED Glossary` | Append to canon glossary |
| `MODIFIED` | `## MODIFIED Glossary` | Replace definition of existing term |
| `REMOVED` | `## REMOVED Glossary` | Delete term from canon glossary |

ADDED terms must not already exist in canon. MODIFIED/REMOVED terms must exist in canon.

### Improvement over OpenSpec: Dangling Delta Detection

`validate` catches dangling deltas — MODIFIED/REMOVED operations referencing requirements that don't exist in the target spec. OpenSpec only fails on these at archive time. litespec catches them during validation.

### Canonical Spec Format

Canonical specs (`specs/canon/<capability>/spec.md`) use this structure:

```markdown
# <capability>

## Purpose               ← optional

## Requirements          ← required

### Requirement: <name>
<body text — must contain SHALL or MUST>

#### Scenario: <short name>
- **WHEN** <condition>
- **THEN** <expected outcome>
```

- `## Purpose` is optional prose before requirements. If present, `SerializeSpec` emits it.
- `## Requirements` is required — all `### Requirement:` blocks must appear inside it.
- No other H2 sections are permitted between H1 and `## Requirements`.

### Scenarios

Each requirement has named scenarios (`#### Scenario: <name>`) with WHEN/THEN format. Scenarios describe expected behavior — the format is opaque text, not parsed structurally.

Rules:
- ADDED and MODIFIED requirements must have at least one scenario
- ADDED and MODIFIED requirement body text must contain `SHALL` or `MUST`
- REMOVED requirements are name-only — no body or scenarios
- RENAMED requirements preserve content and scenarios under the new name

### Glossary (optional)

Canonical specs MAY include a `## Glossary` H2 section after `## Requirements`:

```markdown
## Glossary

- **TermName**: definition text
```

- Entries use `- **TermName**: definition` format
- Term names must be non-empty and unique within a spec
- No other H2 sections are permitted after `## Glossary`

## Artifact Dependency Graph

```
proposal ──────► specs ──┐
     │                   ├──► tasks
     └──► design ────────┘
```

States:
- **BLOCKED** — dependencies not yet satisfied
- **READY** — all dependencies exist, artifact does not
- **DONE** — artifact file exists on disk

## Tool Integration

- **Canonical location:** `.agents/skills/` — SKILL.md with YAML frontmatter
- **Thin adapter layer:** Optional generation of tool-specific commands via `litespec init --tools claude,cursor,...`
- Skills are lean — minimal token usage, no boilerplate

## Configuration

Convention over configuration. No config file. All defaults baked in. If a need arises later, add it then.

Tool adapters are auto-detected by scanning for symlinks in adapter skill directories (e.g., `.claude/skills/`) that point into `.agents/skills/`.

## CLI Commands

| Command | Purpose |
|---------|---------|
| `litespec init [--tools ...]` | Scaffold `specs/` dir + generate skills (+ optional tool-specific commands) |
| `litespec new <name>` | Create a new change directory with `.litespec.yaml` metadata |
| `litespec validate [<name>] [--all|--changes|--specs] [--type change|spec] [--strict]` | Validate artifact structure, delta syntax, dangling deltas, dependency cycles/overlaps |
| `litespec status [<name>]` | Show artifact graph state (BLOCKED/READY/DONE) |
| `litespec instructions <artifact>` | Return artifact-specific instructions for AI to create an artifact |
| `litespec list [--specs|--changes] [--sort name|recent|deps]` | List specs or changes (deps sort uses topological order) |
| `litespec view` | Display dashboard overview with progress bars, specs, changes (draft/active/ready-to-archive), and dependency graph |
| `litespec update [--tools ...]` | Regenerate skills and adapter symlinks |
| `litespec archive <change> [--allow-incomplete]` | Apply deltas to canon + move to archive (marks change as implemented; errors if unarchived dependencies exist) |
| `litespec completion <shell>` | Print shell completion script (bash, zsh, fish) |
| `litespec __complete <words...>` | Hidden backend for dynamic shell completions |
| `litespec upgrade` | Check for latest version and upgrade via `go install` |
| `litespec import --source <dir>` | Import an OpenSpec project to litespec format |

## Archive Behavior

Archiving a change **promotes it to implemented**: deltas are merged into canonical specs (the source of truth), and the change directory is moved to the archive. Until a change is archived, its deltas are tentative — not part of the canonical spec.

`litespec archive <change>` performs these steps in order:

1. **Validate** — run `ValidateChange` (artifacts exist, delta syntax valid, no dangling deltas)
2. **Check dependencies** — error if the change has unarchived dependencies; warn with `--allow-incomplete`
3. **Check tasks** — all checkboxes must be checked, unless `--allow-incomplete`
4. **Merge deltas** — apply RENAMED→REMOVED→MODIFIED→ADDED into `specs/canon/<capability>/spec.md`
5. **Move** — relocate the change directory to `specs/changes/archive/<YYYY-MM-DD>-<name>/`

The archive operation is transactional: the change is moved to the archive first, then canonical specs are written atomically. If the write fails, the change is restored from archive.

## Change Metadata

Each change directory contains `.litespec.yaml`:

```yaml
schema: spec-driven
created: "2026-03-31T10:30:00Z"
dependsOn:          # optional — list of change names this change depends on
  - parent-change
```

Minimal. No phase tracking — derived from `tasks.md`. The `dependsOn` field is optional and establishes prerequisite relationships between changes.

## Change Dependencies

Changes can declare optional `dependsOn` relationships in `.litespec.yaml`. This enables:

- **Cycle detection** — `validate --changes` and `validate --all` detect circular dependencies
- **Overlap detection** — validates that changes sharing a dependency don't modify the same capability requirements
- **Topological sorting** — `list --changes --sort deps` orders changes by dependency (level-by-level BFS, alphabetical within each level); falls back to alphabetical on cycles
- **Archive guard** — `archive` errors when the change being archived has unarchived dependencies; warns with `--allow-incomplete`
- **Dependency graph** — `view` renders a tree-style DAG with box-drawing characters when any active change has `dependsOn`

Resolution checks active changes first, then archived changes. Active takes priority on name collision. Archived change names are extracted by stripping the date prefix.

### Dependency Glossary Loading

When validating a change with `dependsOn`, `validate` loads glossary terms from each dependency's specs (delta specs if active, canonical specs if archived). Terms are unioned across spec files, deduplicated by name, with warnings on conflicting definitions. Loaded glossary terms are included in the `ValidationResult` for downstream consumers (e.g., the review skill performs semantic cross-referencing on these terms).

## Glossary

**Glossary** — An optional section in canonical and delta specs that exports term definitions for downstream consumers. Enables cross-change consistency checking when a change declares `dependsOn`.
