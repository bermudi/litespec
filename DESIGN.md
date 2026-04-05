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
│       └── archive/              # completed changes (YYYY-MM-DD-<name>/)
│           └── <date>-<name>/
│               ├── .litespec.yaml
│               ├── proposal.md
│               ├── design.md
│               └── tasks.md      # planning artifacts only — no specs/ subtree
└── .agents/skills/               # generated skills (canonical)
    ├── litespec-explore/
    ├── litespec-grill/
    ├── litespec-propose/
    ├── litespec-review/
    ├── litespec-apply/
    └── litespec-adopt/
```

## Workflow

Unidirectional flow:

```
explore → grill → propose → review → apply → review → archive
                                          │
                                      adopt (separate path)
```

No backward flow. If something is wrong after propose, start over from explore/grill.

## Skills

| Skill | Type | Behavior |
|-------|------|----------|
| `explore` | Ephemeral | Thinking mode. No artifacts, no change dir. Conversational. |
| `grill` | Ephemeral | Relentless Q&A on the explored idea. No artifacts. Resolves every branch of the design tree before proceeding. |
| `propose` | Materializes | Creates change dir + proposal + specs + design + tasks (all at once). This is the commit point. |
| `apply` | Phase-based | Implements tasks per phase in `tasks.md`. One phase per invocation. AI focuses on one area without doing the whole implementation at once. |
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

## CLI Commands

| Command | Purpose |
|---------|---------|
| `litespec init [--tools ...]` | Scaffold `specs/` dir + generate skills (+ optional tool-specific commands) |
| `litespec new <name>` | Create a new change directory with `.litespec.yaml` metadata |
| `litespec validate [<name>] [--all\|--changes\|--specs] [--type change\|spec] [--strict]` | Validate artifact structure, delta syntax, dangling deltas, dependency cycles/overlaps |
| `litespec status [<name>]` | Show artifact graph state (BLOCKED/READY/DONE) |
| `litespec instructions <artifact>` | Return artifact-specific instructions for AI to create an artifact |
| `litespec list [--specs\|--changes] [--sort name\|recent\|deps]` | List specs or changes (deps sort uses topological order) |
| `litespec view` | Display dashboard overview with progress bars, specs, changes, and dependency graph |
| `litespec update [--tools ...]` | Regenerate skills and adapter symlinks |
| `litespec archive <change> [--allow-incomplete]` | Apply deltas + move to archive (warns if other changes depend on this one) |
| `litespec completion <shell>` | Print shell completion script (bash, zsh, fish) |
| `litespec __complete <words...>` | Hidden backend for dynamic shell completions |

## Archive Behavior

`litespec archive <change>` performs these steps in order:

1. **Validate** — run `ValidateChange` (artifacts exist, delta syntax valid, no dangling deltas)
2. **Check tasks** — all checkboxes must be checked, unless `--allow-incomplete`
3. **Merge deltas** — apply RENAMED→REMOVED→MODIFIED→ADDED into `specs/canon/<capability>/spec.md`
4. **Strip specs/ subtree** — remove the change's `specs/` directory before archiving
5. **Move** — relocate the change directory to `specs/changes/archive/<YYYY-MM-DD>-<name>/`

The archived directory MUST contain only planning artifacts (`.litespec.yaml`, `proposal.md`, `design.md`, `tasks.md`). The `specs/` subtree MUST NOT be present — its contents have already been merged into the canonical `specs/canon/` source of truth.

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
- **Archive guard** — `archive` warns when other active changes depend on the change being archived; errors unless `--allow-incomplete`
- **Dependency graph** — `view` renders a tree-style DAG with box-drawing characters when any active change has `dependsOn`

Resolution checks active changes first, then archived changes. Active takes priority on name collision. Archived change names are extracted by stripping the date prefix.
