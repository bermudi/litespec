# Workflow

litespec is unidirectional and has two lanes. GH issue is the queue — proposal + design + queue live in the GH issue body. Durable contracts live in `specs/<feature>/spec.md` and `specs/decisions/`. Disposable work lives in a closed GH issue.

## Two Lanes

### Small fix — zero ceremony

For typos, bugs, one-offs, and trivial refactors.

1. Tell the agent the fix.
2. The agent reads `specs/product.md`, the relevant `specs/<feature>/spec.md`, `specs/decisions/`, and `specs/glossary.md`.
3. The agent edits the code.
4. If the fix changed a load-bearing contract, the agent edits the one `specs/<feature>/spec.md` directly. Preserve `SHALL`/`MUST` and `WHEN`/`THEN`.
5. Done.

No GH issue. No folder.

### New feature — plan fuzzy to clear

For greenfield, API shape, CLI behavior, or anything that will outlast the issue.

1. `litespec-plan` in **fuzzy** mode: read code, ask 2–3 questions, maybe spike, write no files. Use `references/fuzzy.md`.
2. `litespec-plan` in **clear** mode: require a clean tree, capture `Base:`, create `litespec/<change-name>`, and record `Branch:` in the labeled GH issue body before its proposal, design, and units. If `gh` is unavailable, write the same body to `specs/queues/<name>.md`. Draft a spec if load-bearing.
3. **grill-me** (optional): adversarial shaping. Use `references/grilling.md`. Pull in `codebase-design` or `domain-modeling` when needed.
4. `litespec-build`: establish a meaningful red Verify at a clean pre commit, implement one unit, require green at the implementation post commit, then post the complete receipt before checking the box.
5. `litespec-review`: replay Verify at pre, post, and `HEAD`, verify the recorded branch, review tracked and untracked issue-owned work, then route findings.
6. Close the GH issue only when all units are checked and review returns `PASS`.

Unidirectional. If the plan shifts, rewrite the GH issue (disposable), not the durable spec.

## Units

A unit is one demo-able outcome with a `Verify:` that must fail without the outcome.

In the GH issue body:

```markdown
Base: <full commit ID>
Branch: litespec/show-dependency-graph

## Queue

## Show dependency graph in `view`
Done means: `litespec view` displays arrows between dependent changes
Verify: `go test ./...` and `litespec view | grep "->"` returns a non-empty line
- [ ] pending

## Validate specs from `view`
Done means: the dashboard lists each `specs/<feature>/spec.md` with its requirement count
Verify: `litespec view | grep "Specifications"` shows the spec name and count
- [ ] pending
```

- Build one unit per session.
- `Verify:` must be a concrete command or assertion. If it would pass without the outcome, the unit is too big.
- Before implementation, run the exact Verify on a clean pre commit. It must fail because the outcome is absent. If the verifier is introduced by the unit, create one verifier-only commit first.
- Create exactly one implementation commit, then run the same Verify on that clean post commit and require exit status 0.
- Tick the checkbox only after posting one receipt with the exact command plus full pre/post SHAs, statuses, fenced raw outputs, and matching scope lines. Never amend either evidence commit. A nonempty `Evidence:` label is not enough.
- Put unrelated work on another branch or worktree.

Review checks pre→post→`HEAD` ancestry and replays the exact Verify in detached temporary worktrees at pre and post, then at `HEAD`; temporary worktrees are removed even after failures and the current worktree is never checked out to evidence commits. Red-green evidence does not prove that Verify targets the correct behavior, so review still probes it adversarially. Findings then route in order: suggestions are non-blocking; unit violations rebuild the unit; CRITICAL/WARNING inside issue scope blocks as a direct fix or new parent unit; findings outside issue scope route without blocking. Auto-loaded instructions are trusted bootstrap inputs. After skill activation, only the remote issue is read initially; every additional local queue, contract, implementation, or reference path is screened before content access.

## When to Write a Spec

A spec is a durable, load-bearing contract in `specs/<feature>/spec.md`.

Write one when:

- You are defining a CLI command, API shape, file format, or public interface.
- Being wrong six months from now would break downstream work.
- The contract needs `SHALL`/`MUST` and `WHEN`/`THEN` scenarios.

Skip one when:

- It is a one-off, trivial refactor, prototype, or internal-only detail.
- It will not outlast the GH issue.
- The change is purely cosmetic or a bug fix with no contract change.

For small fixes, edit the existing spec in place. For new features, `litespec-plan` drafts the spec during `clear` mode.

## Durable vs Disposable

| Durable (keep) | Disposable (close/delete) |
|---|---|
| `specs/product.md` | GH issue body and comments |
| `specs/<feature>/spec.md` | |
| `specs/decisions/NNNN-<slug>.md` (`spine: true` for load-bearing) | |
| `specs/glossary.md` | |

Rule of thumb: if being stale would mislead the next reader, keep it. Otherwise delete it after merge.

## Skills and Adapters

litespec generates three skills into `.agents/skills/`:

| Skill | Purpose |
|---|---|
| `litespec-plan` | Fuzzy/clear planning, grilling, codebase-design, domain-modeling |
| `litespec-build` | Implement one unit at a time, satisfy `Done means:` and `Verify:` |
| `litespec-review` | Adversarial review of GH issue + spec vs implementation |

- `.agents/skills/` is canonical. Nearly every AI coding agent discovers it natively.
- `litespec init --tools claude` or `litespec update --tools claude` creates symlinks in `.claude/skills/` for Claude Code.
- Run `litespec update` after changing templates or pulling a new version.
- Project-specific skills are tracked in git directly, not generated.

## See Also

- [Concepts](concepts.md)
- [CLI Reference](cli-reference.md)
- [Getting Started](getting-started.md)
