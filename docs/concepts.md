# Concepts

Why litespec exists, what its pieces are, and how they fit together. For the full design rationale, see [`DESIGN.md`](https://github.com/bermudi/litespec/blob/main/DESIGN.md).

## Why specs survive the chat window

AI coding agents are good at implementation and bad at memory. Context evaporates between sessions: the first agent knew *why* the rate limit is per-IP, the next one guesses. A spec is a durable contract that outlives the conversation — it states **what** the system must do (requirements with `SHALL` or `MUST`) and **when** it matters (named `WHEN`/`THEN` scenarios), so the next agent verifies the same thing the first one built.

Specs are not roadmaps, design documents, or intentions. They are small, load-bearing promises about the parts that break when wrong: CLI shapes, API surfaces, file formats, core workflows. If being stale would mislead the next reader, it belongs in a spec. Otherwise it lives in the issue and disappears when the issue closes.

## The two lanes

Work is either a small fix or a new feature. The lanes differ in ceremony, not rigor.

**Small fix — zero ceremony.** "Fix the typo." The agent reads `specs/product.md`, the relevant spec, decisions, and glossary; edits the code; updates the spec in place if the contract changed. No issue, no queue.

**New feature — plan, build, review.** A fuzzy idea becomes a clear issue, then one verifiable unit at a time:

1. **Plan fuzzy** — `litespec-plan` reads the codebase, grills you with two or three questions, maybe runs a tiny spike. Writes nothing. Ephemeral.
2. **Plan clear** — `litespec-plan` writes the GitHub issue: `Base:` + `Branch:` ownership, proposal, design, queue of units. Drafts `specs/<feature>/spec.md` if the feature is load-bearing.
3. **Build one unit** — `litespec-build` records the exact `Verify:` failing at a clean pre commit, implements the unit in one or more implementation/fix commits, records the same command passing at the final clean commit where `Verify:` passes, posts the receipt, ticks the box, stops. One unit per session.
4. **Review** — `litespec-review` replays Verify at pre, post, and `HEAD`, each in a detached temporary worktree including a detached temporary worktree at `HEAD`, with each removed even when Verify fails. Then it adversarially checks issue + spec against implementation. Red-green evidence does not prove that Verify targets the correct behavior — it only proves Verify distinguishes the two trees. That judgment is the review.
5. **Close** — merge the issue's `Branch:` first, then close the issue. The spec stays; the issue is disposable.

One direction only. If the plan shifts, rewrite the issue, not the durable spec.

## The issue is the queue

The GitHub issue body is the change: proposal, design, and queue, headed by immutable ownership lines:

```markdown
Base: <full commit SHA at plan time>
Branch: litespec/show-dependency-graph

## Proposal
...

## Design
...

## Show dependency graph in `view`
Boundary: process
Done means:
- [graph] `litespec view` shows arrows between dependent changes
Scenarios:
- [graph] TestViewShowsDependencyArrows
Risk cases:
- timeout: N/A — view is a local read with no deadline
- cleanup: N/A — view creates no temp state
- non-ENOENT errors: N/A — no filesystem lookup beyond specs/
- concurrency: N/A — single read-only invocation
- optional configured dependencies: N/A — view requires no optional services
Verify: `go test ./internal -run TestViewShowsDependencyArrows`
- [ ] pending
```

One unit is one boundary or one failure policy — not one demo. Every bracketed `Done means:` clause maps to a named test scenario; filesystem, process, and network units account for the five standard risks per scenario or reasoned N/A. One exact `Verify:` gates the unit, and it must fail without the outcome. `plan[clear]` creates the labeled issue from a clean tree on a dedicated `litespec/<change-name>` branch; all work on that branch belongs to the issue. When `gh` is unavailable, `specs/queues/<name>.md` is the same body as a local file.

## Evidence, not prose

A checked unit carries a receipt, not a claim. The receipt quotes the exact `Verify:`, the unit's contract digest (from `litespec digest`), labeled pre/post SHAs and exit statuses, both raw outputs in unedited fences, and scope lines that say what the runs show and nothing more. New receipts open with `Protocol: evidence/v1`, `Digest algorithm: unit-contract-sha256-v1`, and a content-derived `Receipt ID:`. The CLI assembles them (`litespec receipt`), the validator structural-checks them, build posts them, review replays them — and review still asks whether the command tests the right thing. `litespec issue check` ticks the box afterward, exactly one flip, ownership lines byte-unchanged; hand-editing issue bodies is retired.

Digest transitions are witnessed, not silent. Only `litespec-plan` may change a unit contract, and it does so with an append-only `Amendment:` record chaining the old digest to the new one. After two review-requested rebuilds against one digest, the next unit-breaking finding stops routing to build and requires plan to reshape the contract. See [Workflow](workflow.md) for the full protocol.

## Durable vs disposable

| Keep (curated, small) | Close (disposable) |
|---|---|
| `specs/product.md` — mental models + 2–3 flows | GH issue body and comments |
| `specs/<feature>/spec.md` — load-bearing contracts, edited directly | |
| `specs/decisions/NNNN-<slug>.md` — durable rulings (`spine: true` when load-bearing) | |
| `specs/glossary.md` — shared words, curated | |

## What a good spec looks like

One spec per load-bearing feature, edited in place — no staging, no delta flow:

```markdown
# <feature>

## Requirements

### Requirement: <name>
Body must contain SHALL or MUST.

#### Scenario: <short name>
- **WHEN** <condition>
- **THEN** <outcome>
```

Good specs make an observable promise; bad ones leave the implementation guessing:

```markdown
Bad: The dashboard should be fast and intuitive.

Good:
### Requirement: Queue-only dashboard
`litespec view` SHALL list only open issues labeled `litespec`.

#### Scenario: Unlabeled issue
- **WHEN** the repository has an unrelated open issue
- **THEN** `litespec view` does not list it
```

Decisions close off a road someone will reasonably propose again — the signal is *why not*, not *how*. The bar is high on purpose: real contention plus no better home (a comment at the line that would change beats a decision file). `touch` + `validate` is enough; there is no decide command.

## Convention over configuration

No config file. `specs/` holds durable docs, `.agents/skills/` is the canonical skill directory (`.claude/skills/` symlinks exist only because Claude Code doesn't read `.agents/`), the `litespec` label marks queue issues, `view` auto-detects `gh`. Adapters are discovered by scanning for symlinks — added only when a concrete tool needs one.

## Three lean skills

Generated by `litespec update` into `.agents/skills/`. Short and directive; detail lives in `references/` and loads only when the branch applies.

| Skill | What it does |
|-------|--------------|
| `litespec-plan` | Fuzzy grilling, clear issue writing, codebase design, domain modeling, glossary |
| `litespec-build` | One unit: red at pre, green at post, receipt, tick, stop |
| `litespec-review` | Replay pre/post/HEAD, adversarial probe, route findings |

Project-specific skills (like `the-drill`) live directly in `.agents/skills/` as tracked files — never generated, never overwritten.
