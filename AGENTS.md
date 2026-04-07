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
- **Delta specs** use ADDED/MODIFIED/REMOVED/RENAMED markers merged in strict order at archive time
- **Change dependencies** — optional `dependsOn` field in `.litespec.yaml` for prerequisite relationships between changes. Enables cycle/overlap detection, topological sorting, archive guards, and dependency graph visualization.
- **Skills** are generated into `.agents/skills/` (canonical). This is the only skill directory we target — nearly all AI coding agents (opencode, Cursor, Windsurf, Amazon Q, Auggie, Roo, Kilo Code, Codex, etc.) now discover `.agents/skills/` natively. Claude Code is the sole exception: it reads from `.claude/skills/`, so `--tools claude` creates symlinks there.
- **Scenarios** — each requirement has named scenarios (`#### Scenario: <name>`) with WHEN/THEN format. ADDED and MODIFIED requirements must have at least one scenario. Body text must contain SHALL or MUST.
- **Artifact-specific instructions** — `litespec instructions <artifact>` returns distinct guidance per artifact (proposal: motivation/scope/non-goals; specs: delta format + capabilities; design: architecture/decisions/file changes; tasks: phased checklist). The `template` field retains the propose workflow for context.
- **Phased tasks** — `tasks.md` organizes work into phases, applied one phase at a time
- **`view` command** — displays a dashboard with progress bars `[████░░░]`, change categories (draft/active/completed), specs sorted by requirement count, and an optional dependency graph section when any change has `dependsOn`

## Workflow

```
explore → grill → propose → review → apply → review → archive
                                          │
                                      adopt (separate path)
```

Unidirectional. No backward flow.

- **explore** and **grill** are ephemeral — no artifacts, no change directory. The AI keeps context in its window. `propose` is what materializes everything to disk.
- **propose** is the commit point. If something is wrong after proposing, start over from explore/grill.
- **apply** works on one phase at a time. Each phase = one agent session = one commit. Re-invoke for the next phase.
- **adopt** is a separate path — reverse-engineers specs from existing code given a file/directory path.
- **review** is context-aware AI review: artifact review when no tasks are checked (evaluates planning artifacts), implementation review when some tasks are checked (code vs specs), pre-archive review when all tasks are checked (both artifacts and code). No test/lint running.

## Key Design Decisions

These came from deliberate debate. Respect the reasoning:

- **Convention over configuration** — no config files unless a concrete need arises. OpenSpec ships a stub config.yaml that nobody fills in. We skip it entirely until needed.
- **`.agents/skills/` is canonical** — one source of truth, discovered natively by nearly every AI coding agent. `--tools claude` creates symlinks in `.claude/skills/` as the only exception (Claude Code does not read `.agents/`). No other tool-specific adapters are needed.
- **Lean skills** — minimal token usage. Each skill is focused instructions, not pages of boilerplate.
- **Dangling delta detection during `validate`** — not just at archive time. This is an improvement over OpenSpec.
- **Phase tracking derived from `tasks.md` checkboxes** — no metadata field. The first phase with unchecked tasks is the current phase.
- **Git-native workflow** — litespec manages specs. A separate harness (future work) will handle branch creation (`change/<name>`), per-phase commits, and PR creation. For now, the skills offer prompts: "Would you like a new branch?" and "Would you like a PR?"
- **CLI is a read-only context provider** — the AI never writes through the CLI. It writes artifact files directly. The CLI exists to give the AI structured data (status, instructions, validation).
- **Artifact-specific instructions** — each artifact (proposal, specs, design, tasks) gets its own instruction template via `litespec instructions <artifact>`, not a single generic template. The propose skill template is kept as a `template` field for workflow context.

## Working Conventions

- Use `stdlib` and established Go patterns
- Run `go build`, `go test`, `go vet` after changes
- Follow standard Go project layout: `cmd/`, `internal/`
- Write tests that verify behavior and system state
- No `any` equivalents — explicit types everywhere
- No comments unless absolutely necessary for non-obvious logic
- When changes affect workflow, skills, or core concepts, update `AGENTS.md` and `DESIGN.md` to match. These are living documents — if the system changes, they change too.
