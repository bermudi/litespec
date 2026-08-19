# litespec v2 — Design (draft lean)

> Lean cut. Comment inline — this is for red pen.

## Stack

- Go, `github.com/bermudi/litespec`, binary `litespec`
- No config file. Convention over config.

## Directory

```
project/
├── specs/
│   ├── product.md          # mental models + 2-3 flows (human + agent, agent-maintained)
│   ├── glossary.md         # optional, curated terms (from mattpocock)
│   ├── decisions/
│   │   └── NNNN-slug.md    # durable rulings, see below
│   └── <feature>/
│       └── spec.md         # only load-bearing features/contracts (optional)
└── .agents/skills/
    ├── litespec-plan/      # fuzzy + clear, see Skills
    ├── litespec-build/
    └── litespec-review/
```

Notes:
- No `canon/`. `specs/<feature>/spec.md` is the durable spec when needed.
- No `backlog.md`. GH issues is the backlog.
- No `QUEUE.md` in v2 lean. GH issue body is proposal + design + queue. Local queue fallback is `specs/queues/<name>.md` when `gh` is unavailable.

## What lives forever vs what gets deleted

**Durable (curated, small):**
- `product.md` — what we are, what we aren't, how we think, 2-3 flows
- `specs/<feature>/spec.md` — promises that break things if wrong (CLI, API shape, file formats)
- `decisions/` — why we chose a hard path
- `glossary.md` — shared words

**Disposable (deleted after merge):**
- GH issues (closed) and issue comments

Test: if being stale would mislead a new person/agent, keep it. Else delete after merge.

## Two lanes

**Small fix — zero ceremony:**
You say "fix typo" -> agent reads product + relevant spec + decisions/glossary -> edits code -> updates the one `specs/<feature>/spec.md` if it was a contract change -> done. No `new`, no issue required.

**New feature / greenfield (plan fuzzy -> clear):**
```
you: "add X" -> plan[fuzzy] (read code, grill by default — references/fuzzy.md loads references/grilling.md; no files)
          -> plan[clear] (write GH issue: proposal + design + units with Verify; also draft spec if load-bearing — references/clear.md)
          -> you: "looks good" (grilling happened in fuzzy) or "grill-me" (more grilling with references/grilling.md)
          -> build: one unit at a time (see unit rule)
          -> review: triage findings into lanes
          -> fix per lane (rebuild unit via build, or small fix lane)
          -> close GH issue
```

Review triages findings structurally: does the finding break a unit's `Done means:`/`Verify:`? If yes → uncheck box, rebuild via `build` (scope expands — fix the pattern, not just the line). If no and it is trivial → small fix lane. If no and it needs real work outside any existing unit's contract → draft a new unit and create a GH sub-issue via `gh issue create --parent <N> --label litespec`, or write to `specs/queues/<parent-name>-review.md` if `gh` is unavailable. No fix skill, no finding tracker — findings route to existing tracking (unit checkboxes), get fixed immediately, or spawn a sub-issue.

`grill-me` is a skill reference, not a CLI. `plan` owns spec drafting in clear mode: if the feature is load-bearing, it writes/updates `specs/<feature>/spec.md` alongside the issue.

## Unit rule

One unit = one thing you can demo + one `Verify:` that would fail if it's missing.

In GH issue body:
```markdown
## Show graph for 2 changes
Done means: `litespec view` shows arrows between deps
Verify: `go test ./...` and view output contains "->"
- [ ] pending
```

Agent ticks by checking the box. No `tasks.md`.

## Spec format (only for load-bearing)

```markdown
# <feature>

## Requirements

### Requirement: <name>
Body must contain SHALL or MUST.

#### Scenario: <short name>
- **WHEN** <condition>
- **THEN** <outcome>
```

ADDED/MODIFIED not needed — edit the file directly. Small fix edits it, greenfield creates it.

## Decisions

File: `specs/decisions/NNNN-slug.md`

```markdown
# Title

## Status
proposed | accepted | superseded

## Context
## Decision
## Consequences
```

Optional frontmatter: `spine: true` for load-bearing. Created by agent via skill `references/adr.md`, not a CLI. `validate` checks sections, `view` stars spine.

No `litespec decide` command. `touch` + `validate` is enough.

## Glossary

`specs/glossary.md`, optional but highly recommended. Curated. Skill nudges when a new core term appears. No gate.


## Skills

3 generated skills, lean, directive, progressive disclosure. `think+plan` merged — fuzzy vs clear are modes of `plan`.

| Skill | When |
|-------|------|
| plan | turn intent into bounded GH issue (+ spec). Fuzzy vs clear are references |
| build | implement one unit, satisfy Verify |
| review | adversarial check, triage findings into lanes |

`litespec-plan` references (load only when branch applies, distilled from AgenticWiki — no links, no theory):
- `references/fuzzy.md` — half-baked idea; grill by default with `references/grilling.md`, research/spike, no files yet
- `references/clear.md` — sharp idea: write GH issue body (proposal+design+queue) + `specs/<feature>/spec.md` if load-bearing (owns the Verify rule)
- `references/grilling.md` — default fuzzy process; also `grill-me` or unresolved branches
- `references/codebase-design.md` — thin vertical slice, reuse existing path, smallest coherent change (distilled from tracer bullets / vertical slices / infrastructure blindness / over-engineering)
- `references/domain-modeling.md` — new ubiquitous term -> glossary

`litespec-build` references:
- `references/build/review-fixing.md` — rebuilding a unit that review reopened. Scope expands: fix the pattern, not just the cited line

`litespec-review` references:
- `references/review/adversarial-review.md` — constructing adversarial scenarios for interaction bugs, state transitions, wiring gaps

No alias for `think`. Add if dogfooding shows we miss it. Detail lives in `references/` only when that branch applies — borrow grill/domain ideas from mattpocock/skills on our terms.

Generated via `litespec update` from `internal/skill/templates/` (embed.FS). `.agents/skills/` is canonical.

## CLI (minimal)

| Command | Purpose |
|---------|---------|
| `litespec init` | scaffold `specs/` + skills |
| `litespec new <name> [--issue N]` | link to GH issue (no folder in lean) |
| `litespec validate [--decisions] [--issue N] [--queue <path>]` | lint specs + decisions + GH issue queue (labeled litespec) + local specs/queues/ fallback + Verify shell (bash -n) |
| `litespec view` | product + features + open `litespec` GH issues (via `gh` if present) + decisions (spine starred) |
| `litespec update` | regenerate skills |

No `patch`, `archive`, `decide`, `preview`, `import` until needed. Add when pain appears.

## GH issues as the change

GH issue is the queue — the GH issue body is proposal + design + queue (64k limit, no overflow design needed).

- GH issue body is proposal + design + queue. 64k limit — no overflow design needed.
- The `litespec` label marks queue issues. `validate` scans open issues with this label; `view` filters to it.
- `litespec new <name> --issue N` links to a GH issue; when `gh` is unavailable, `specs/queues/<name>.md` is the local fallback.
- `view` auto-detects `gh` + GitHub remote. No config flag.

## Resolved for v2 lean

- Local queue fallback at `specs/queues/<name>.md` when `gh` unavailable — mirrors the GH issue 1:1, handles multi-feature changes
- Product flows: list explicitly in `product.md` (models + flows as lists).
- `litespec new` starts as link only (`--issue N`), not auto-create. Add `gh issue create` later if needed.
