# Concepts

litespec is built for AI-native, spec-driven development. These concepts are the foundation. For the full design rationale, see [`DESIGN.md`](https://github.com/bermudi/litespec/blob/main/DESIGN.md).

## Why spec-driven development for AI agents

AI coding agents are great at implementation, but context is lost between sessions. A spec is a durable contract that survives the chat window. It states **what** the system must do (requirements with `SHALL` or `MUST`) and **when** it matters (named `WHEN`/`THEN` scenarios), so the next agent can verify the same thing the first agent built.

Specs are not roadmaps, design documents, or hand-wavy intentions. They are small, load-bearing contracts for the parts of the system that break if they are wrong: CLI shapes, API surfaces, file formats, and core workflows.

## The two-lane workflow

Work is either a small fix or a new feature. The lanes differ in ceremony, not rigor.

### Small fix — zero ceremony

"Fix the typo" or "rename this function." The agent reads `specs/product.md`, the relevant `specs/<feature>/spec.md`, and `specs/decisions/` / `specs/glossary.md`; edits the code; and updates the spec in place if the contract changed. No GitHub issue, no queue.

### New feature — plan, then build

A fuzzy idea becomes a clear issue, then a demo-able unit at a time:

1. **Plan fuzzy** — `litespec-plan` reads the codebase, asks two or three questions, does a quick spike, and writes nothing. Ephemeral.
2. **Plan clear** — `litespec-plan` writes the GitHub issue body: proposal, design, and queue of units. It also drafts `specs/<feature>/spec.md` if the feature is load-bearing.
3. **Grill-me** — stress-test the plan with `litespec-plan`, pulling in codebase-design and domain-modeling references when needed.
4. **Build one unit** — `litespec-build` records the exact `Verify:` failing for the absent outcome at a clean pre commit, implements one `## <outcome>`, records the same command passing at the implementation post commit, posts the receipt, ticks the checkbox, and stops.
5. **Review** — `litespec-review` replays Verify at pre, post, and `HEAD` in detached temporary worktrees with guaranteed cleanup, then adversarially checks the issue + spec against the implementation. Red-green evidence does not prove the command targets the right behavior.
6. **Close the issue** — the issue is disposable. Durable specs, decisions, and glossary stay.

Unidirectional. One unit per session.

## GitHub issue as the queue

The GitHub issue body is the change: proposal, design, and queue. Each queue item is a unit:

```markdown
## Show graph for 2 changes
Done means: `litespec view` shows arrows between deps
Verify: `go test ./...` and view output contains "->"
- [ ] pending
```

`plan[clear]` requires a clean tree, records `Base:`, creates `litespec/<change-name>`, and records `Branch:` in the labeled GH issue. If `gh` is unavailable, it writes the same body to `specs/queues/<name>.md`. All work on that branch belongs to the issue.

The 64 KiB issue limit is enough because the queue contains only units, not full designs. Durable design and reasoning live in `specs/product.md`, `specs/<feature>/spec.md`, and `specs/decisions/`.

## Durable specs vs. ephemeral issues

**Durable (curated, small):**

- `specs/product.md` — mental models, flows, and what the project is and isn't.
- `specs/<feature>/spec.md` — load-bearing contracts. Edited directly.
- `specs/decisions/NNNN-<slug>.md` — durable rulings. `spine: true` marks load-bearing ones.
- `specs/glossary.md` — ubiquitous language.

**Disposable (closed after the work ships):**

- GitHub issues and their comments.

Rule of thumb from `DESIGN.md`: if being stale would mislead a new person or agent, keep it. Otherwise it goes in the issue and is closed when done.

## Direct spec edits

There is no staging file for spec changes. Small fixes and new features edit `specs/<feature>/spec.md` directly. If a requirement changes, change the file. If a feature is load-bearing, create the file.

A spec uses this format:

```markdown
# <feature>

## Requirements

### Requirement: <name>
Body must contain SHALL or MUST.

#### Scenario: <short name>
- **WHEN** <condition>
- **THEN** <outcome>
```

Each load-bearing requirement has a `SHALL` or `MUST` body and at least one named `WHEN`/`THEN` scenario.

Good specs make an observable promise; bad specs leave the implementation to guess:

```markdown
Bad: The dashboard should be fast and intuitive.

Good:
### Requirement: Queue-only dashboard
`litespec view` SHALL list only open issues labeled `litespec`.

#### Scenario: Unlabeled issue
- **WHEN** the repository has an unrelated open issue
- **THEN** `litespec view` does not list it
```

## Convention over configuration

litespec has no config file. Conventions are enough:

- `specs/` is where durable docs live.
- `.agents/skills/` is the canonical skill directory.
- `.claude/skills/` is a symlink for Claude Code.
- `litespec view` auto-detects `gh` and the GitHub remote.
- Tool adapters are discovered by scanning symlinked skill directories.

Adapters are added only when a concrete need appears.

## Three lean skills

AI skills are generated by `litespec update` into `.agents/skills/`. They are short, directive, and progressive: details live in `references/` and are loaded only when the branch applies.

| Skill | What it does |
|-------|--------------|
| `litespec-plan` | Fuzzy exploration, clear issue writing, grilling, codebase design, and domain modeling. |
| `litespec-build` | One unit at a time. Record red at pre and green at post, post the receipt, tick the box, stop. |
| `litespec-review` | Replay pre/post/HEAD evidence, then adversarially review the issue + spec against the implementation. |

These are the only generated skills. Project-specific skills are tracked directly in `.agents/skills/` and are not generated by the CLI.

## The bottom line

litespec gives AI agents a durable, verifiable contract to work against and a lean workflow that keeps the contract aligned with the code. The GitHub issue carries the temporary plan; the `specs/` directory carries the durable truth.
