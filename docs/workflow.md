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
2. `litespec-plan` in **clear** mode: write the labeled GH issue body with `## Proposal`, `## Design`, and `## Queue` (units with `Done means:` and `Verify:`). If `gh` is unavailable, write the same body to `specs/queues/<name>.md` (`<name>` is the change name chosen in this step). Draft `specs/<feature>/spec.md` if the feature is load-bearing. Use `references/clear.md`.
3. **grill-me** (optional): adversarial shaping. Use `references/grilling.md`. Pull in `codebase-design` or `domain-modeling` when needed.
4. `litespec-build`: implement one unit at a time. Each unit must satisfy `Done means:` and `Verify:` before the box is checked.
5. `litespec-review`: adversarial check of the queue + spec against the implementation.
6. Close the GH issue when all units are done.

Unidirectional. If the plan shifts, rewrite the GH issue (disposable), not the durable spec.

## Units

A unit is one demo-able outcome with a `Verify:` that must fail without the outcome.

In the GH issue body:

```markdown
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
- Tick the checkbox when `Verify:` passes, then commit and stop.

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
