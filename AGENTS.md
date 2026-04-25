# AGENTS.md

## Project

**litespec** — a lean, AI-native spec-driven development CLI tool written in Go.

It reimagines [OpenSpec](https://github.com/Fission-AI/OpenSpec) with stronger opinions: fewer concepts, leaner skills, unidirectional workflow, and proper dangling-delta validation.

Reference implementation lives at `reference/openspec/` for inspiration and grounding. Do not modify it.

## Project Status

This is an **active experiment**. We are learning what works by building it. Decisions made yesterday may be revised today if we find something better. Prefer trying things over planning forever.

The design emerged from a structured grilling session — question by question, branch by branch — and that spirit continues. When unsure about a direction, ask. When you see a better way, say so.

## Architecture

- **Language:** Go
- **Module:** `github.com/bermudi/litespec`
- **Binary:** `litespec`
- **Design doc:** `DESIGN.md` — read it first

## Core Concepts

- **Specs** live in `specs/canon/` — the source of truth for current capabilities
- **Changes** live in `specs/changes/<name>/` — isolated proposed modifications
- **Decisions** live in `specs/decisions/` as `NNNN-<slug>.md` — persistent architectural rulings that span changes. Created via `litespec decide`, surfaced by `list --decisions`, `validate --decisions`, and `view`. Opt-in; absence is not an error.
- **Delta specs** use ADDED/MODIFIED/REMOVED/RENAMED markers merged in strict order at archive time
- **Change dependencies** — optional `dependsOn` field in `.litespec.yaml` for prerequisite relationships between changes. Enables cycle/overlap detection, topological sorting, archive guards, and dependency graph visualization.
- **Skills** are generated into `.agents/skills/` (canonical). Nearly all AI coding agents (opencode, Cursor, Windsurf, Amazon Q, Auggie, Roo, Kilo Code, Codex, etc.) now discover `.agents/skills/` natively. Claude Code is the sole exception: it reads from `.claude/skills/`, so `--tools claude` creates symlinks there.
- **Scenarios** — each requirement has named scenarios (`#### Scenario: <name>`) with WHEN/THEN format. ADDED and MODIFIED requirements must have at least one scenario. Body text must contain SHALL or MUST.
- **Artifact-specific instructions** — `litespec instructions <artifact>` returns distinct guidance per artifact (proposal: motivation/scope/non-goals; specs: delta format + capabilities; design: architecture/decisions/file changes; tasks: phased checklist). The `template` field retains the propose workflow for context.
- **Phased tasks** — `tasks.md` organizes work into phases, applied one phase at a time
- **`view` command** — displays a dashboard with progress bars `[████░░░]`, change categories (draft/active/ready to archive), specs sorted by requirement count, and an optional dependency graph section when any change has `dependsOn`
- **Glossary** — the project's ubiquitous language lives in `specs/glossary.md`. A single, curated file defining shared terms. Read by explore, grill, and propose at session start (active — nudges when undefined terms surface). Apply references it passively. Review may consult it during cross-change review. The glossary skill manages the file. Graceful degradation if absent.

## Workflow

```
explore → grill → propose → [research →] apply → review → archive
                                          │
                                      adopt (separate path)

patch → archive  (lightweight lane for small, single-capability changes)
```

Unidirectional. No backward flow.

- **explore** and **grill** are ephemeral — no artifacts, no change directory. The AI keeps context in its window. `propose` is what materializes everything to disk.
- **propose** is the commit point. If something is wrong after proposing, start over from explore/grill.
- **patch** is a lightweight lane — `litespec patch <name> <capability>` creates a delta-only change with no planning artifacts. The delta is the contract; planning artifacts are optional scaffolding. Use for small, single-capability changes that need no design discussion.
- **research** is optional — runs after propose when external knowledge is needed. Reads artifacts from disk, identifies knowledge gaps (APIs, schemas, libraries), gathers docs, and produces research skills into `.agents/skills/research-<topic>/`. Uses skill-creator conventions for formatting. Stance is risk-scoped: skip what LLMs know cold, go deep on novel APIs/libraries. Research skills persist after archive — they accumulate as project knowledge.
- **apply** works on one phase at a time. Each phase = one agent session = one commit. Re-invoke for the next phase. Consumes research skills via natural agent discovery.
- **adopt** is a separate path — reverse-engineers specs from existing code given a file/directory path.
- **review** is context-aware AI review that adapts to change lifecycle: artifact review when no tasks are checked (evaluates planning artifacts), implementation review when some tasks are checked (runs adversarial review first, then compliance review), pre-archive review when all tasks are checked (adversarial + compliance + archive readiness + build verification). Adversarial review runs first to avoid anchoring bias — it enumerates failure scenarios from specs before reading code. No test/lint running (except build verification in pre-archive mode).
- **archive** is the commit to implemented — applying deltas to canonical specs and moving the change to the archive. Until archived, a change's deltas are tentative. Use `litespec preview <name>` to see what archive would do without making changes.

## Key Design Decisions

These came from deliberate debate. Respect the reasoning:

- **Convention over configuration** — no config files unless a concrete need arises. OpenSpec ships a stub config.yaml that nobody fills in. We skip it entirely until needed. Tool adapters are auto-detected by scanning for symlinks in adapter skill directories (e.g., `.claude/skills/`) that point into `.agents/skills/`.
- **`.agents/skills/` is canonical** — one source of truth, discovered natively by nearly every AI coding agent. `--tools claude` creates symlinks in `.claude/skills/` as the only exception (Claude Code does not read `.agents/`). No other tool-specific adapters are needed.
- **Lean skills** — minimal token usage. Each skill is focused instructions, not pages of boilerplate.
- **Dangling delta detection during `validate`** — not just at archive time. This is an improvement over OpenSpec.
- **Phase tracking derived from `tasks.md` checkboxes** — no metadata field. The first phase with unchecked tasks is the current phase.
- **Git-native workflow** — litespec manages specs. A separate harness (future work) will handle branch creation (`change/<name>`), per-phase commits, and PR creation. For now, the skills offer prompts: "Would you like a new branch?" and "Would you like a PR?"
- **CLI is a read-only context provider** — the AI never writes through the CLI. It writes artifact files directly. The CLI exists to give the AI structured data (status, instructions, validation).
- **Artifact-specific instructions** — each artifact (proposal, specs, design, tasks) gets its own instruction template via `litespec instructions <artifact>`, not a single generic template. The propose skill template is kept as a `template` field for workflow context.
- **Research skills** — produced into `.agents/skills/research-<topic>/SKILL.md` during the research phase. They are project-level agent skills containing reference documentation (API schemas, library docs, auth flows). They persist after archive as accumulated project knowledge. The apply agent discovers them naturally through skill descriptions. No CLI command needed — the research skill itself is the instructions.

## Working Conventions

- Use `stdlib` and established Go patterns
- Run `go build`, `go test`, `go vet` after changes
- Follow standard Go project layout: `cmd/`, `internal/`
- Write tests that verify behavior and system state
- No `any` equivalents — explicit types everywhere
- No comments unless absolutely necessary for non-obvious logic
- When changes affect workflow, skills, or core concepts, update `AGENTS.md` and `DESIGN.md` to match. These are living documents — if the system changes, they change too.

### Skill Generation

Skills are **not written directly** to `.agents/skills/`. The pipeline is:

1. Add a `SkillInfo` entry to `internal/paths.go` (`Skills` slice) — defines ID, name, and description
2. Create a template file in `internal/skill/<name>.go` — registers the template body via `init()`
3. Run `litespec update` — generates `.agents/skills/<name>/SKILL.md` from the `SkillInfo` metadata + registered template

Never write to `.agents/skills/` directly. Always edit the Go templates and regenerate.
