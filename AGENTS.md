# AGENTS.md

## Project

**litespec** — a lean, AI-native spec-driven development CLI tool written in Go.

It reimagines [OpenSpec](https://github.com/Fission-AI/OpenSpec) with stronger opinions: fewer concepts, leaner skills, unidirectional workflow, and strict structural validation.

Reference implementation lives at `reference/openspec/` for inspiration and grounding. Do not modify it.

## Project Status

This is an **active experiment**. We are learning what works by building it. Decisions made yesterday may be revised today if we find something better. Prefer trying things over planning forever.

The design emerged from a structured grilling session — question by question, branch by branch — and that spirit continues. When unsure about a direction, ask. When you see a better way, say so.

## Architecture

- **Language:** Go
- **Module:** `github.com/bermudi/litespec/v2`
- **Binary:** `litespec`
- **Design doc:** `DESIGN.md` — read it first (lean v2: GH issue is the queue)

## Core Concepts

- **Product** lives in `specs/product.md` — mental models + 2-3 flows, human + agent maintained
- **Specs** live in `specs/<feature>/spec.md` — only load-bearing contracts (SHALL/MUST + WHEN/THEN). No `canon/` — edit the file directly
- **Decisions** live in `specs/decisions/` as `NNNN-<slug>.md` — durable rulings with `spine: true` for load-bearing. Created via `touch` + `validate`, not a CLI
- **Glossary** lives in `specs/glossary.md` — ubiquitous language, curated, optional but recommended. Managed via plan skill, graceful degradation if absent
- **GH issues are the change/queue** — GH issue body holds proposal + design + queue (units with Done means + Verify + optional Read first / Constraints / Depends) + immutable `Base: <sha>` and `Branch: litespec/<change-name>` ownership lines. `plan[clear]` starts clean and creates the dedicated branch; all work on it belongs to the issue. Offline fallback is `specs/queues/<name>.md`.
- **Units** — one demo-able outcome per `##` with `Done means:` and `Verify:` that must fail without the outcome, plus optional `Read first:` (context, not scope — areas/rulings, not file lists) and `Constraints:` (boundaries: what must stay true or out of bounds — never what to edit; omit rather than placeholder) and `Depends:`. Checked units must carry a complete red-green evidence receipt scoped to that unit (GH issue: body or comments; local queue: `Evidence:` block after `Verify:`, before the checkbox). Unchecked units are unaffected. One exact Verify is recorded at two immutable clean commits: a non-zero pre run because the outcome is absent, then a zero-status post run at the final clean commit where `Verify:` passes. Each run has a full SHA, integer status, fenced raw output, and matching conservative scope line. A nonempty `Evidence:` label is not enough. Built one at a time, ticked via checkbox only after the receipt is posted. No `tasks.md`
- **Two lanes** — small fix (zero ceremony, no issue required) vs new feature (plan fuzzy (grill by default) -> clear -> build -> review -> close issue)
- **Skills** are generated into `.agents/skills/` (canonical). Only three: `litespec-plan` (fuzzy/grill/clear + codebase-design/domain-modeling), `litespec-build` (one unit at a time), `litespec-review` (adversarial). Nearly all agents discover `.agents/skills/` natively; Claude Code via symlink in `.claude/skills/`. `litespec-plan` has fuzzy mode (grill by default) for half-baked ideas and clear mode to nail the GH issue. Project-specific skills (`the-drill`) are tracked directly in git — NOT generated
- **Scenarios** — each requirement has named scenarios (`#### Scenario: <name>`) with WHEN/THEN format. Load-bearing requirements must have at least one scenario. Body text must contain SHALL or MUST

## Workflow

Two lanes — lean cut. Unidirectional, no backward flow.

**Small fix — zero ceremony:**
You say "fix typo" -> agent reads product + relevant spec + decisions/glossary -> edits code -> updates the one `specs/<feature>/spec.md` if it was a contract change -> done. No issue required.

**New feature / greenfield (plan fuzzy -> clear):**
```
you: "add X" -> plan[fuzzy] (read code, grill by default — references/fuzzy.md loads references/grilling.md; no files)
          -> plan[clear] (start clean, create isolated branch, write GH issue: proposal + design + units with Verify; also draft spec if load-bearing)
          -> you: "looks good" (grilling happened in fuzzy) or "grill-me" (more grilling with references/grilling.md)
          -> build: one unit at a time (see unit rule)
          -> review: triage findings into lanes
          -> fix per lane
          -> close GH issue
```

- `plan[fuzzy]` is ephemeral — no files, grill/spike only
- `plan[clear]` requires a clean tree, creates `litespec/<change-name>`, and materializes the GH issue (Base + Branch + proposal + design + queue) + draft spec if load-bearing
- `grill-me` is a skill reference, not a CLI. `references/fuzzy.md` loads `references/grilling.md` by default
- `build` implements one unit per session: from a clean tree it runs the exact `Verify:` before implementation and requires a meaningful non-zero pre run. If the verifier does not exist, it may create at most one verifier-only commit. It then creates one or more implementation/fix commits, never amends them, and records post at the final clean commit where `Verify:` passes. It posts both runs as a verbatim receipt, ticks the box, and never amends pre. Local-queue bookkeeping is a separate metadata commit. Stops
- `review` is adversarial — context-aware check of GH issue + spec vs implementation, cross-checks every checked unit's pre/post receipt and ancestry, and replays Verify using a detached temporary worktree at pre, another at post, and a detached temporary worktree at `HEAD`; all are removed even when Verify fails. It then triages findings into lanes (DISPUTED is terminal non-blocking). Red-green discrimination does not prove that Verify targets the right behavior; review still judges that. (See below.)
- GH issue is the queue: each unit is `## <outcome>` with `Done means:` and `Verify:` and status checkbox, plus optional `Read first:` / `Constraints:` (both unique/nonempty, omit rather than placeholder) and `Depends:` and a complete evidence receipt when checked (body or comments for GH, `Evidence:` block for local)

**Review triage — first matching rule wins (DISPUTED is terminal non-blocking, never routes):**
Review's auto-loaded harness/system instructions, `AGENTS.md`, and review `SKILL.md` are a trusted bootstrap boundary. After activation, review reads only the remote GH issue before screening every additional local path—including a queue fallback, specs, decisions, tracked/untracked work, and later references. Unsafe paths stop review without a verdict. DISPUTED disposition (adversarial candidate rejected by cited authority) never blocks and never routes.
1. **SUGGESTION** → non-blocking small fix lane.
2. **CRITICAL or WARNING breaking a unit contract** → blocking rebuild via `build`; WARNINGs route here too.
3. **CRITICAL or WARNING inside review scope, outside units** → blocking direct fix if trivial, append a blocking unit to the parent if non-trivial, or `plan` if the shape is wrong.
4. **CRITICAL or WARNING outside review scope and units** → non-blocking small fix, or draft for a later `plan[clear]` that creates its own queue/branch.

`needs decision` is reported before applying the matching route and does not alter blocking status. The issue closes only when every unit is checked **and** review returns `PASS`.

## Key Design Decisions

These came from deliberate debate. Respect the reasoning:

- **Convention over configuration** — no config files unless a concrete need arises. OpenSpec ships a stub config.yaml that nobody fills in. We skip it entirely until needed. Tool adapters are auto-detected by scanning for symlinks in adapter skill directories (e.g., `.claude/skills/`) that point into `.agents/skills/`
- **`.agents/skills/` is canonical** — one source of truth, discovered natively by nearly every AI coding agent. `--tools claude` creates symlinks in `.claude/skills/` as the only exception (Claude Code does not read `.agents/`). No other tool-specific adapters are needed
- **Lean skills** — minimal token usage. Each skill is focused instructions, not pages of boilerplate. 3 skills only: plan (fuzzy/grill/clear), build (one unit), review (adversarial). Progressive disclosure via `references/` — detail lives there only when branch applies
- **GH issue is the queue/change** — proposal + design + queue live in the issue body, not `QUEUE.md`. Keeps what doesn't rot, drops ceremony. Small fix needs no issue at all
- **`litespec` label marks queue issues** — hardcoded convention, no config. `validate` scans open issues with this label; `view` filters to it. `plan[clear]` instructs adding the label when creating the issue
- **Local queue fallback** — `specs/queues/<name>.md` mirrors the GH issue 1:1 when `gh` is unavailable. `<name>` is the change name chosen during `plan[clear]`. Handles multi-feature changes
- **Unit = demo + failing Verify** — one thing you can demo + one Verify that would fail if missing. `build` must satisfy Verify before claiming done
- **Direct spec edits** — no ADDED/MODIFIED delta flow. Edit `specs/<feature>/spec.md` directly. Preserve SHALL/MUST + WHEN/THEN but drop delta merge complexity
- **No `canon/`, `backlog.md` in lean** — GH issues is the backlog, `specs/<feature>/spec.md` is the durable spec when needed
- **Validate structure, not semantics** — the CLI validates structural contracts (syntax, references, spec format). Do not encode heuristic checks that compensate for model limitations. If a model gap bites repeatedly, fix the prompt in the relevant skill — that scales with model capability
- **Review routes exhaustively** — suggestion; unit violation; in-scope outside-unit finding; out-of-scope finding. In-scope CRITICAL/WARNING blocks even without a unit; non-trivial work becomes a unit on the existing parent branch. Out-of-scope work routes to a later `plan[clear]` and its own branch. No finding tracker, no `tasks.md`
- **Queue issues own isolated branches** — `plan[clear]` starts from a clean tree, records `Base:` and `Branch:`, and creates `litespec/<change-name>`. All branch work belongs to the issue; unrelated work uses another branch/worktree. Review checks ownership and includes tracked plus untracked files. See decision `0002-queue-issues-own-isolated-branches`

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

1. Add a `SkillInfo` entry to `internal/paths.go` (`Skills` slice) — defines ID, name, and description (plan is fuzzy/grill/clear, build is one unit, review is adversarial)
2. Create a template file in `internal/skill/templates/<id>.md` — embedded via `embed.FS`
3. Run `litespec update` — generates `.agents/skills/<name>/SKILL.md` from the `SkillInfo` metadata + registered template. Skills can also include resource files in `internal/skill/templates/references/<id>/`, which are generated into `.agents/skills/<name>/references/`. `litespec-plan` has 5 references: fuzzy, clear, grilling, codebase-design, domain-modeling

**Project-specific skills** (`the-drill`) live in `.agents/skills/` as tracked git files. They are NOT in the `Skills` slice and NOT generated by `litespec update`. Edit them directly
