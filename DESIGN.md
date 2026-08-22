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

Review routing is ordered: suggestion; unit violation; in-scope finding outside units; out-of-scope finding. DISPUTED is terminal non-blocking (never routes, never blocks, generates no unit) and sits outside the blocking chain. CRITICAL/WARNING inside issue scope blocks even when no unit covers it. Out-of-scope findings route without blocking. Evidence cross-check (checked units: red-green receipt, command verbatim, pre→post→HEAD ancestry, three re-runs) and DISPUTED citation bar (rejecting authority required) are part of review. The issue closes only when every unit is checked and review returns `PASS`.

`grill-me` is a skill reference, not a CLI. `plan` owns spec drafting in clear mode: if the feature is load-bearing, it writes/updates `specs/<feature>/spec.md` alongside the issue.

## Unit rule

One unit = one thing you can demo + one `Verify:` that would fail if it's missing.

Optional per-unit fields: `Read first:` (context, not scope — areas and rulings, not file lists) and `Constraints:` (boundaries — what must stay true or is out of bounds, never what to edit). Both are unique and nonempty when present; omit rather than placeholder. `Depends:` lists other unit headings. The worker owns the implementation path.

Evidence protocol: before implementation, build runs the exact `Verify:` on a clean pre commit and requires a non-zero failure caused by the absent outcome. If the verifier does not exist yet, build may create one verifier-only commit and use it as pre. Build then creates exactly one implementation commit and requires the same command to exit 0 there. Neither evidence commit may be amended. The receipt records the command once, then `pre sha:`, non-zero `pre exit status:`, fenced raw output, matching `Pre-evidence scope:`, followed by `post sha:`, `post exit status: 0`, fenced raw output, and matching `Post-evidence scope:`. GH issue evidence lives as a comment or body `Evidence:` block; local queue evidence is an `Evidence:` block after `Verify:` and before the checkbox, committed separately. Validate checks structure only. Review checks pre→post→`HEAD` ancestry, runs Verify in a detached temporary worktree at pre and requires the missing-outcome failure, runs it in another detached temporary worktree at post and requires success with the outcome, then requires success at `HEAD`. Worktrees are removed even on failure; the current review worktree is never checked out to an evidence SHA. Red-green evidence shows discrimination between two trees; it does not prove that Verify targets the correct behavior, so adversarial review remains required.

In GH issue body — immutable `Base: <sha>` and `Branch: litespec/<change-name>` ownership lines above the units, then:
```markdown
## Show graph for 2 changes
Read first: litespec view, specs/product.md
Constraints: API remains backward compatible; no new config files
Depends: Show graph for 1 change
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
| `litespec init [--tools <ids>]` | scaffold `specs/` + skills |
| `litespec validate [--all\|--specs\|--decisions\|--issue <N>\|--queue <path>] [--type T]` | lint specs + decisions + GH issue queue (labeled litespec) + local specs/queues/ fallback + Verify shell (bash -n) |
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
