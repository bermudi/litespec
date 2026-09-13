# Workflow

The full rules of the two lanes: unit shape, evidence protocol, review routing, closure. [Concepts](concepts.md) is the why; this is the how.

## Small fix — zero ceremony

Typos, bugs, one-offs, trivial refactors.

1. Tell the agent the fix.
2. It reads `specs/product.md`, the relevant spec, decisions, and glossary.
3. It edits the code.
4. If a load-bearing contract changed, it edits the one `specs/<feature>/spec.md` directly, preserving `SHALL`/`MUST` and `WHEN`/`THEN`.
5. Done. No issue, no queue.

## New feature — plan, build, review

Greenfield, API shape, CLI behavior, anything that outlasts the issue. Unidirectional: if the plan shifts, rewrite the issue (disposable), not the spec.

1. **`litespec-plan` fuzzy** — read code, grill with two or three questions, maybe spike. No files. (`references/fuzzy.md`, grilling by default.)
2. **`litespec-plan` clear** — clean tree required. Capture `Base:`, create `litespec/<change-name>`, record `Branch:`, write the labeled GH issue (proposal + design + queue), draft the spec if load-bearing. (`references/clear.md` owns the Verify rule.) Offline: same body in `specs/queues/<name>.md`.
3. **Grill-me (optional)** — adversarial shaping via `references/grilling.md`, plus codebase-design or domain-modeling when the branch applies.
4. **`litespec-build`** — one unit per session: meaningful red at a clean pre commit, one or more implementation/fix commits (immutable), green at the final clean commit where `Verify:` passes, receipt posted, box ticked, stop.
5. **`litespec-review`** — replay Verify at pre, post, and `HEAD`, each in a detached temporary worktree including a detached temporary worktree at `HEAD`; each worktree is removed even when Verify fails; adversarial probe; route findings. Red-green evidence does not prove that Verify targets the correct behavior.
6. **Merge the `Branch:`, then close the issue** — only when every box is ticked, every rebuild request / re-plan marker / amendment resolved, and review returns `PASS`.

## Units

One unit is one boundary or one failure policy — never just one broad demo. A demo crossing independent boundaries splits.

```markdown
Base: <full commit SHA>
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

Field rules:

- **`Done means:`** — one bullet per clause, each with a unique bracketed ID. The clause is the observable outcome.
- **`Scenarios:`** — every clause ID maps to at least one named test. Unknown IDs and unmapped clauses both fail validation. IDs and mappings are contract: build treats them as fixed, review judges whether the named tests actually exercise them.
- **`Verify:`** — exactly one command, and it must fail without the outcome. Fenced block preferred; an inline backtick command on the `Verify:` line also validates. Vacuous commands (`true`, `:`, `exit 0`, comment-only) fail. `bash -n` lints fenced blocks.
- **`Boundary:` / `Risk cases:`** — required when the unit touches filesystem, process, or network. `Boundary:` takes exactly one of `filesystem`, `process`, `network` (closed, case-sensitive vocabulary). `Risk cases:` accounts for all five — `timeout`, `cleanup`, `non-ENOENT errors`, `concurrency`, `optional configured dependencies` — each mapping to a scenario ID or `N/A — <concrete reason>`. No boundary, no risk block.
- **`Read first:` / `Constraints:` / `Depends:`** — optional, at most one each, nonempty when present. `Read first:` is context (areas and rulings, not a file list). `Constraints:` states what must stay true or is out of bounds — never an edit list. `Depends:` names `##` headings in the same queue; a unit is unblocked when its dependencies are checked with no unresolved requests.
- **Prose sections** (`Proposal`, `Design`, drafts) are skipped by validation — only `##` sections containing `Done means:` or `Verify:` are units. But prose is still scope: before filing, every preservation sentence in it must map onto a unit's `Done means:`/`Constraints:`, become its own regression-pin unit, or be deleted. A sentence no unit enforces is a promise nothing can test.

Before filing, plan dry-runs each Verify on the base tree: honest results are non-zero (outcome or verifier missing) or green only when every named test file actually executed — runners silently skip misnamed files and stay green. An outcome an earlier unit already delivers (or is constrained to preserve) becomes a regression pin, never a re-delivery: the named tests are the outcome, and the Verify fails while the pin is absent.

## Evidence protocol

Build proves the unit twice with one exact command, then review replays both runs plus the present.

**Build, per unit:**

1. Clean tree, with at most one verifier-only commit before it when the unit introduces its own verifier. Run the exact `Verify:` on the clean starting commit before implementation — it must exit non-zero *because the outcome is absent*. Unrelated failures (typo, missing dependency, broken environment) stop the unit.
2. If the verifier doesn't exist yet, commit just the verifier first (at most one verifier-only commit; loud not-implemented stubs keep gates green while pre stays red) and use that as pre. Otherwise the starting commit is pre.
3. Implement in one or more implementation/fix commits. Never amend pre or any implementation commit.
4. Clean tree. Run the same `Verify:` — exit 0 with the outcome present. That final clean commit is post.
5. Assemble the receipt with `litespec receipt` (see [CLI Reference](cli-reference.md)): exact command, `unit digest:` from `litespec digest`, labeled pre/post SHAs and statuses, both raw outputs in unedited fences (`<no output>` if empty), matching scope lines, opening with `Protocol: evidence/v1`, `Digest algorithm: unit-contract-sha256-v1`, and a content-derived `Receipt ID:`. Pre must ancestor post.
6. Post it — GitHub comment, or `Evidence:` block under the unit for local queues — then tick the box with `litespec issue check` (GitHub) or a separate metadata commit (local). A prose `Evidence:` label is not a receipt. Then stop; one unit per session.

Oversized receipts split, never truncate: field-boundary splits after a scope line, or identity-bearing `Raw output chunk:` records for a single giant output — every non-final comment ending with the literal continuation marker. `validate` joins only the literal next comment; anything else is an incomplete-receipt error.

**Review, per checked unit:** a detached temporary worktree at pre (must fail for the absent outcome), a detached temporary worktree at post (must pass), and a detached temporary worktree at `HEAD` (must still pass) — each worktree removed even when Verify fails, evidence SHAs never checked out in the working tree. Red-green evidence does not prove that Verify targets the correct behavior; the adversarial probe decides whether it tests the right thing. Superseded receipts validate against their own declared protocol, command, and digest; current-digest receipts must match current Verify exactly.

**Rebuilds:** a checked unit with an unresolved rebuild request is selectable again; its fresh receipt carries the same heading + occurrence and resolves all earlier requests for that identity. After two completed rebuild cycles against one digest, the next unit-breaking finding records `Re-plan required:` instead — build refuses the marked contract until plan reshapes it through an `Amendment:` (old digest → new digest, witnessed append-only), which resets the count and stays unresolved until fresh evidence lands on the new digest. Only plan authors or alters contracts; silent edits followed by fresh receipts fail the digest chain.

## Review routing

Every finding carries Severity (`CRITICAL` / `WARNING` / `SUGGESTION`), Location, Evidence, and one unambiguous Fix direction. Severity is confidence it's wrong; scope decides whether this issue owns it. First matching rule wins — and `DISPUTED` (an adversarial candidate that cited authority explicitly rejects) is terminal: never blocks, never routes.

1. **SUGGESTION** → small-fix lane, non-blocking, user's discretion.
2. **CRITICAL/WARNING breaking a unit contract** → blocking. Fewer than two rebuild cycles against the current digest: rebuild request → `litespec-build`. Two cycles done: re-plan marker → `litespec-plan`. WARNINGs follow the same threshold.
3. **CRITICAL/WARNING in scope, outside units** → blocking. Trivial: direct fix on the issue branch. Non-trivial and well-shaped: append an unchecked unit to this issue, build it here. Wrong shape: `litespec-plan`.
4. **CRITICAL/WARNING outside scope and units** → non-blocking. Trivial: small-fix lane. Non-trivial: draft for a later `plan[clear]` with its own queue and branch. Wrong shape: `litespec-plan`.

A finding needing a durable ruling reports `needs decision: <question>` first — it doesn't change blocking status. Review never writes code, ticks boxes, edits issue bodies for routing, or removes evidence; rebuild requests, re-plan markers, coverage records, and parent-unit appends are the only permitted routing mutations. Review also posts a HEAD-keyed coverage record per review — advisory only, never proof; each reviewer drafts risks independently first, then uses prior records to hunt gaps.

## Closure

The issue closes only when every unit checkbox is checked, every rebuild request is resolved, review returns `PASS`, and the issue's `Branch:` is merged (decision 0008) — merge first, then close. Every re-plan marker and amendment must also be resolved (every observed digest chains to the current contract). Non-blocking routed findings never prevent closure — they already have their own lane.

## Skills and adapters

Three skills, generated into `.agents/skills/` by `litespec update`:

| Skill | Purpose |
|---|---|
| `litespec-plan` | Fuzzy grilling, clear issue writing, codebase design, domain modeling, glossary |
| `litespec-build` | One unit: red, green, receipt, tick, stop |
| `litespec-review` | Replay, adversarial probe, route |

`.agents/skills/` is canonical. `litespec init --tools claude` symlinks into `.claude/skills/` for Claude Code; later `update` runs auto-detect. Project-specific skills live in `.agents/skills/` as tracked files — never generated, never overwritten.

## See also

- [Concepts](concepts.md) — the why behind all of this
- [Tutorial](tutorial.md) — the cycle end to end
- [CLI Reference](cli-reference.md) — the exact commands
