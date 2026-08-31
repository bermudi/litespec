# litespec v2 — Design (draft lean)

> Lean cut. Comment inline — this is for red pen.

## Stack

- Go 1.26.1+, `github.com/bermudi/litespec/v2`, binary `litespec`
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
You say "fix typo" -> agent reads product + relevant spec + decisions/glossary -> edits code -> updates the one `specs/<feature>/spec.md` if it was a contract change -> done. No issue required.

**New feature / greenfield (plan fuzzy -> clear):**
```
you: "add X" -> plan[fuzzy] (read code, grill by default — references/fuzzy.md loads references/grilling.md; no files)
          -> plan[clear] (start clean, create isolated branch, write GH issue + units; draft spec if load-bearing)
          -> you: "looks good" (grilling happened in fuzzy) or "grill-me" (more grilling with references/grilling.md)
          -> build: one unit at a time (see unit rule)
          -> review: triage findings into lanes
          -> fix per lane (rebuild unit via build, or small fix lane)
          -> close GH issue
```

Each queue issue owns a dedicated `litespec/<change-name>` branch. `plan[clear]` requires a clean tree, captures `Base:`, creates the branch, and records `Branch:`. Auto-loaded harness/repository instructions are review's trusted bootstrap boundary. After activation, review initially reads only the remote issue; every additional local queue, contract, implementation, and reference path is screened with its parent components before content access. Unsafe paths stop review. Unrelated work uses another branch or worktree.

Review routing is ordered: suggestion; unit violation; in-scope finding outside units; out-of-scope finding. DISPUTED is terminal non-blocking. CRITICAL/WARNING inside issue scope blocks even when no unit covers it. Evidence cross-check (red-green receipt, pre→post→HEAD ancestry, three re-runs) remains part of review. Review first constructs an independent risk inventory, then consults append-only HEAD-keyed coverage records and persists a new advisory record. A unit violation produces a rebuild request only while fewer than two rebuild cycles completed against the current unit digest. The next violation records `Re-plan required:` and plan must amend/reshape that contract before build continues. GitHub metadata lives in comments; local metadata is appended after all units in a clean metadata commit. The issue closes only when every unit is checked, every rebuild request, re-plan marker, and amendment is resolved, and review returns `PASS`.

`grill-me` is a skill reference, not a CLI. `plan` owns spec drafting in clear mode: if the feature is load-bearing, it writes/updates `specs/<feature>/spec.md` alongside the issue.

## Unit rule

One unit = one external boundary or one failure policy + one `Verify:` that would fail if it is missing. A broad demo crossing independent boundaries must split.

Every `Done means:` clause is a bullet carrying a unique bracketed ID. Required `Scenarios:` entries map every clause ID to at least one named test scenario. Filesystem, process, and network units declare `Boundary:` and a `Risk cases:` matrix covering timeout, cleanup, non-ENOENT errors, concurrency, and optional configured dependencies; each risk maps to a scenario ID or gives a concrete N/A reason. Validate checks mapping shape and references, while plan/review judge whether boundaries were omitted and tests are adequate. Optional `Read first:`, `Constraints:`, and `Depends:` retain their current meanings. All contract fields participate in the unit digest.

Evidence protocol remains one exact command: build records a meaningful non-zero run on a clean pre commit, creating at most one verifier-only commit when needed, then creates one or more implementation/fix commits. Post is the final clean commit where `Verify:` passes. Review replays Verify in a detached temporary worktree at pre, post, and a detached temporary worktree at `HEAD`; all are removed even when Verify fails. Red-green evidence does not prove that Verify targets the right behavior, so scenario mapping and adversarial review remain required. A unit may complete at most two review-requested rebuild cycles against one current contract digest. On the next unit-breaking finding, review records a digest-bound `Re-plan required:` marker instead of another rebuild request. Build refuses that marker. Plan resolves it only through an amendment from the marked digest, narrowing/renaming the unit and appending units as needed. The amendment resets rebuild counting for its new digest and remains unresolved until fresh evidence; that amendment evidence does not consume a review-requested rebuild cycle.

Review coverage is append-only metadata keyed to reviewed HEAD and unit identity. A fresh reviewer drafts risks independently before reading earlier records, then uses them to expand unexercised paths. Coverage is advisory: it does not satisfy receipts, suppress current investigation, or prove correctness. GitHub stores metadata in comments; local queues use an append-only metadata stream after all units.

In a queue body — immutable `Base:` and `Branch:` lines above the units, then:
```markdown
## Terminate timed-out probe process trees
Boundary: process
Done means:
- [tree-timeout] A timed-out probe terminates its descendants
Scenarios:
- [tree-timeout] TestProbeTimeoutTerminatesDescendants
Risk cases:
- timeout: [tree-timeout]
- cleanup: [tree-timeout]
- non-ENOENT errors: N/A — no filesystem lookup
- concurrency: N/A — each probe owns its process tree
- optional configured dependencies: N/A — probe is mandatory once selected
Verify: `go test ./... -run TestProbeTimeoutTerminatesDescendants`
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

Optional frontmatter: `spine: true` for load-bearing. `validate` checks sections, `view` stars spine.

No `litespec decide` command. `touch` + `validate` is enough.

### When to write one

The bar is high, on purpose. A decision closes off a road someone will reasonably propose again — the signal is explaining *why not* rather than *how*. Both must hold:

1. **Real contention.** Someone argued for the other road, or demonstrably will — it's the obvious default. A road closed by argument during grilling counts by construction.
2. **No better home.** Reasoning that fits at the line that would change is a comment there, not a decision. Decisions hold reasoning that spans files, or argues against a road with no single line to attach to.

When in doubt, don't — a decisions directory that reads like a changelog has stopped being useful. `Context` records what was actually measured; `Consequences` state what would justify revisiting.

### Layering

`AGENTS.md` carries the compressed rule; the decision file carries the full reasoning, once. Neither restates the other — the split is the design, not duplication to clean up.

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
- `references/build/review-fixing.md` — rebuilding a unit that review routed back to build. Scope expands: fix the pattern, not just the cited line

`litespec-review` references:
- `references/review/adversarial-review.md` — constructing adversarial scenarios for interaction bugs, state transitions, wiring gaps

No alias for `think`. Add if dogfooding shows we miss it. Detail lives in `references/` only when that branch applies — borrow grill/domain ideas from mattpocock/skills on our terms.

Generated via `litespec update` from `internal/skill/templates/` (embed.FS). `.agents/skills/` is canonical.

## CLI (minimal)

| Command | Purpose |
|---------|---------|
| `litespec init [--tools <ids>]` | scaffold `specs/` + skills |
| `litespec validate [--all\|--specs\|--decisions\|--issue <N>\|--queue <path>] [--type T]` | lint structural contracts and report `structure ok; implementation semantics not verified` on success |
| `litespec view` | product + features + open `litespec` GH issues (via `gh` if present) + decisions (spine starred) |
| `litespec update [--tools <ids>]` | regenerate skills and adapters |
| `litespec upgrade` | check for and install the latest version via `go install` |
| `litespec completion <shell>` | generate shell completion script (bash, zsh, fish) |

## GH issues as the change

GH issue is the queue — the GH issue body is proposal + design + queue (64k limit, no overflow design needed).

- GH issue body is proposal + design + queue. 64k limit — no overflow design needed.
- The `litespec` label marks queue issues. `validate` scans open issues with this label; `view` filters to it.
- `plan[clear]` creates the labeled GH issue; when `gh` is unavailable it writes `specs/queues/<name>.md`.
- `view` auto-detects `gh` + GitHub remote. No config flag.

## Resolved for v2 lean

- Local queue fallback at `specs/queues/<name>.md` when `gh` unavailable — mirrors the GH issue 1:1, handles multi-feature changes
- Product flows: list explicitly in `product.md` (models + flows as lists).
