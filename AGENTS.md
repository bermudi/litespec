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
- **Design doc:** `DESIGN.md` — read it first (lean v2: GH issue is the queue)

## Core Concepts

- **Product** lives in `specs/product.md` — mental models + 2-3 flows, human + agent maintained
- **Specs** live in `specs/<feature>/spec.md` — only load-bearing contracts (SHALL/MUST + WHEN/THEN). No `canon/` — edit the file directly
- **Decisions** live in `specs/decisions/` as `NNNN-<slug>.md` — durable rulings with `spine: true` for load-bearing. Created via `touch` + `validate`, not a CLI
- **Glossary** lives in `specs/glossary.md` — ubiquitous language, curated, optional but recommended. Managed via plan skill, graceful degradation if absent
- **GH issues are the change/queue** — GH issue body holds proposal + design + queue (units with Done means + Verify). GH issue is the queue, 64k limit, no overflow. Offline fallback via `specs/queues/<name>.md` when `gh` unavailable; `--issue` required for `litespec new` to link to GH
- **Units** — one demo-able outcome per `##` with `Done means:` and `Verify:` that must fail without the outcome. Built one at a time, ticked via checkbox. No `tasks.md`
- **Two lanes** — small fix (zero ceremony, no issue required) vs new feature (plan fuzzy -> clear -> grill-me -> build -> review -> close issue)
- **Skills** are generated into `.agents/skills/` (canonical). Only three: `litespec-plan` (fuzzy/clear + grilling/codebase-design/domain-modeling), `litespec-build` (one unit at a time), `litespec-review` (adversarial). Nearly all agents discover `.agents/skills/` natively; Claude Code via symlink in `.claude/skills/`. `litespec-plan` has fuzzy mode for half-baked ideas and clear mode to nail the GH issue. Project-specific skills (`skill-creator`, `the-drill`) are tracked directly in git — NOT generated
- **Scenarios** — each requirement has named scenarios (`#### Scenario: <name>`) with WHEN/THEN format. Load-bearing requirements must have at least one scenario. Body text must contain SHALL or MUST

## Workflow

Two lanes — lean cut. Unidirectional, no backward flow.

**Small fix — zero ceremony:**
You say "fix typo" -> agent reads product + relevant spec + decisions/glossary -> edits code -> updates the one `specs/<feature>/spec.md` if it was a contract change -> done. No `new`, no issue required.

**New feature / greenfield (plan fuzzy -> clear):**
```
you: "add X" -> plan[fuzzy] (read code, ask 2-3 questions, no files — references/fuzzy.md)
          -> plan[clear] (write GH issue: proposal + design + units with Verify; also draft spec if load-bearing — references/clear.md)
          -> you: "looks good" or "grill-me" (references/grilling.md)
          -> build: one unit at a time (see unit rule)
          -> review: triage findings into lanes
          -> fix per lane
          -> close GH issue
```

- `plan[fuzzy]` is ephemeral — no files, questions/spike only
- `plan[clear]` materializes the GH issue (proposal + design + queue) + draft spec if load-bearing
- `grill-me` is a skill reference, not a CLI
- `build` implements one unit per session, satisfies Done means + Verify (Verify must fail without outcome), checks box, commits, stops
- `review` is adversarial — context-aware check of GH issue + spec vs implementation, then triages findings into lanes (see below)
- GH issue is the queue: each unit is `## <outcome>` with `Done means:` and `Verify:` and status checkbox

**Review triage — routing findings to lanes:**
Review reports findings + verdict (`PASS` | `CHANGES REQUESTED`), then routes each finding:
- **CRITICAL breaking a unit's `Done means:`/`Verify:`** → unit is not done. Uncheck its box, rebuild via `build` (loads `references/build/review-fixing.md` — scope expands, fix the pattern not just the line). Issue stays open until all units re-pass.
- **CRITICAL/WARNING outside any unit's contract** (neighboring code, help text, stale decision, drive-by) → small fix lane. No unit, no issue reopen.
- **SUGGESTION** → small fix lane, user's discretion. Not blocking.
- **"needs decision"** → create decision in `specs/decisions/` first, then route the fix per the two rules above.
- **Shape was wrong** → `plan`, not a fix.

Do not invent units for non-unit findings. Do not reopen the issue for small fixes. The issue closes when all its units pass.

## Key Design Decisions

These came from deliberate debate. Respect the reasoning:

- **Convention over configuration** — no config files unless a concrete need arises. OpenSpec ships a stub config.yaml that nobody fills in. We skip it entirely until needed. Tool adapters are auto-detected by scanning for symlinks in adapter skill directories (e.g., `.claude/skills/`) that point into `.agents/skills/`
- **`.agents/skills/` is canonical** — one source of truth, discovered natively by nearly every AI coding agent. `--tools claude` creates symlinks in `.claude/skills/` as the only exception (Claude Code does not read `.agents/`). No other tool-specific adapters are needed
- **Lean skills** — minimal token usage. Each skill is focused instructions, not pages of boilerplate. 3 skills only: plan (fuzzy/clear), build (one unit), review (adversarial). Progressive disclosure via `references/` — detail lives there only when branch applies
- **GH issue is the queue/change** — proposal + design + queue live in the issue body, not `QUEUE.md`. Keeps what doesn't rot, drops ceremony. Small fix needs no issue at all
- **`litespec` label marks queue issues** — hardcoded convention, no config. `validate` scans open issues with this label; `view` filters to it. `plan[clear]` instructs adding the label when creating the issue
- **Local queue fallback** — `specs/queues/<name>.md` mirrors the GH issue 1:1 when `gh` is unavailable. `<name>` comes from `litespec new <name> --issue N`. Handles multi-feature changes
- **Unit = demo + failing Verify** — one thing you can demo + one Verify that would fail if missing. `build` must satisfy Verify before claiming done
- **Direct spec edits** — no ADDED/MODIFIED delta flow. Edit `specs/<feature>/spec.md` directly. Preserve SHALL/MUST + WHEN/THEN but drop delta merge complexity
- **No `canon/`, `backlog.md` in lean** — GH issues is the backlog, `specs/<feature>/spec.md` is the durable spec when needed
- **Validate structure, not semantics** — the CLI validates structural contracts (syntax, references, spec format). Do not encode heuristic checks that compensate for model limitations. If a model gap bites repeatedly, fix the prompt in the relevant skill — that scales with model capability
- **Review triages into existing lanes** — no fix skill, no 4th concept. Review routes each finding structurally: does it cite a unit's `Done means:`/`Verify:`? If yes → rebuild that unit (uncheck box, `build` with scope expansion). If no → small fix lane. Findings are ephemeral — they route to existing tracking (unit checkboxes) or get fixed immediately. No finding tracker, no `tasks.md`

## Working Conventions

- Use `stdlib` and established Go patterns
- Run `go build`, `go test`, `go vet` after changes
- Follow standard Go project layout: `cmd/`, `internal/`
- Write tests that verify behavior and system state
- No `any` equivalents — explicit types everywhere
- No comments unless absolutely necessary for non-obvious logic
- When changes affect workflow, skills, or core concepts, update `AGENTS.md` and `DESIGN.md` to match. These are living documents — if the system changes, they change too
- When adding or modifying CLI commands/flags, update `internal/commandspec.go` — the completion system is auto-generated from the `CommandSpecs` registry. The registry is the single source of truth; `internal/completion.go` derives all completions from it

### Skill Generation

**Domain skills** (`litespec-plan`, `litespec-build`, `litespec-review`) are generated by the Go binary. The pipeline is:

1. Add a `SkillInfo` entry to `internal/paths.go` (`Skills` slice) — defines ID, name, and description (plan is fuzzy/clear, build is one unit, review is adversarial)
2. Create a template file in `internal/skill/templates/<id>.md` — embedded via `embed.FS`
3. Run `litespec update` — generates `.agents/skills/<name>/SKILL.md` from the `SkillInfo` metadata + registered template. Skills can also include resource files in `internal/skill/templates/references/<id>/`, which are generated into `.agents/skills/<name>/references/`. `litespec-plan` has 5 references: fuzzy, clear, grilling, codebase-design, domain-modeling

**Project-specific skills** (`skill-creator`, `the-drill`) live in `.agents/skills/` as tracked git files. They are NOT in the `Skills` slice and NOT generated by `litespec update`. Edit them directly
