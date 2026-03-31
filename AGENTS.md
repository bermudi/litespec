# AGENTS.md

## Project

**litespec** — a lean, AI-native spec-driven development CLI tool written in Go.

It reimagines [OpenSpec](https://github.com/Fission-AI/OpenSpec) with stronger opinions: fewer concepts, leaner skills, unidirectional workflow, and proper dangling-delta validation.

## Architecture

- **Language:** Go
- **Module:** `github.com/bermudi/litespec`
- **Binary:** `litespec`
- **Design doc:** `DESIGN.md` — read it first

## Core Concepts

- **Specs** live in `specs/specs/` — the source of truth for current capabilities
- **Changes** live in `specs/changes/<name>/` — isolated proposed modifications
- **Delta specs** use ADDED/MODIFIED/REMOVED/RENAMED markers merged in strict order at archive time
- **Skills** are generated into `.agents/skills/` (canonical) with optional tool-specific adapters
- **Phased tasks** — `tasks.md` organizes work into phases, applied one phase at a time

## Workflow

```
explore → grill → propose → apply → verify → archive
                     ↑                          │
                  continue                  adopt (separate path)
```

Unidirectional. No backward flow.

## Key Design Decisions

- Convention over configuration — no config files unless needed
- `.agents/skills/` as canonical skill location, not per-tool directories
- Lean skills — minimal token usage, no boilerplate
- Dangling delta detection during `validate`, not just at archive time
- Phase tracking derived from `tasks.md` checkboxes, not metadata

## Working Conventions

- Use `stdlib` and established Go patterns
- No external dependencies unless strongly justified
- Run `go build`, `go test`, `go vet` after changes
- Follow standard Go project layout: `cmd/`, `internal/`, `pkg/`
- Write tests that verify behavior and system state
- No `any` equivalents — explicit types everywhere
