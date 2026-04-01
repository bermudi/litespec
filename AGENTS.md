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

- **Specs** live in `specs/specs/` — the source of truth for current capabilities
- **Changes** live in `specs/changes/<name>/` — isolated proposed modifications
- **Delta specs** use ADDED/MODIFIED/REMOVED/RENAMED markers merged in strict order at archive time
- **Skills** are generated into `.agents/skills/` (canonical). Claude Code gets symlinks via `--tools claude`.
- **Phased tasks** — `tasks.md` organizes work into phases, applied one phase at a time

## Workflow

```
explore → grill → propose → apply → verify → archive
                     ↑                          │
                  continue                  adopt (separate path)
```

Unidirectional. No backward flow.

- **explore** and **grill** are ephemeral — no artifacts, no change directory. The AI keeps context in its window. `propose` is what materializes everything to disk.
- **propose** is the commit point. If something is wrong after proposing, start over from explore/grill.
- **apply** works on one phase at a time. Each phase = one agent session = one commit. Re-invoke for the next phase.
- **adopt** is a separate path — reverse-engineers specs from existing code given a file/directory path.
- **verify** is pure AI review of code vs specs. No test/lint running.

## Key Design Decisions

These came from deliberate debate. Respect the reasoning:

- **Convention over configuration** — no config files unless a concrete need arises. OpenSpec ships a stub config.yaml that nobody fills in. We skip it entirely until needed.
- **`.agents/skills/` is canonical** — one source of truth. `--tools claude` creates symlinks in `.claude/skills/` for Claude Code.
- **Lean skills** — minimal token usage. Each skill is focused instructions, not pages of boilerplate.
- **Dangling delta detection during `validate`** — not just at archive time. This is an improvement over OpenSpec.
- **Phase tracking derived from `tasks.md` checkboxes** — no metadata field. The first phase with unchecked tasks is the current phase.
- **Git-native workflow** — litespec manages specs. A separate harness (future work) will handle branch creation (`change/<name>`), per-phase commits, and PR creation. For now, the skills offer prompts: "Would you like a new branch?" and "Would you like a PR?"
- **CLI is a read-only context provider** — the AI never writes through the CLI. It writes artifact files directly. The CLI exists to give the AI structured data (status, instructions, validation).

## Working Conventions

- Use `stdlib` and established Go patterns
- No external dependencies unless strongly justified (yaml.v3 is the only one so far)
- Run `go build`, `go test`, `go vet` after changes
- Follow standard Go project layout: `cmd/`, `internal/`
- Write tests that verify behavior and system state
- No `any` equivalents — explicit types everywhere
- No comments unless absolutely necessary for non-obvious logic

## What's Next

Things we know we want but haven't built yet:

- Git-native workflow integration (branch per change, phase commits, PR creation)
- Tests — the codebase is green but has zero test coverage
- Skill template refinement based on real usage
