# Workflow

litespec is unidirectional with two lanes — small fix (zero ceremony) and new feature (plan fuzzy -> clear -> build -> review). GH issue is the queue — the GH issue body holds proposal + design + queue, each unit is a `## <outcome>` with `Done means:` and `Verify:` that must fail without the outcome.

## Two Lanes

**Small fix — zero ceremony:** You say "fix typo" -> agent reads `specs/product.md` + relevant `specs/<feature>/spec.md` + `specs/decisions/`/glossary -> edits code -> updates the one spec if it was a contract change -> done. No `litespec new`, no GH issue required. Direct spec edits — no ADDED/MODIFIED delta, no `canon/`.

**New feature — plan fuzzy -> clear:** `litespec-plan` in `fuzzy` mode reads code, asks 2-3 questions, does a spike, writes nothing (ephemeral, `references/fuzzy.md`). When sharp, `clear` mode writes the GH issue (proposal + design + queue with units, each with Done means + Verify — `references/clear.md`) and drafts `specs/<feature>/spec.md` if load-bearing. You say "looks good" or "grill-me" (`references/grilling.md`, codebase-design and domain-modeling when needed). Then `litespec-build` implements one unit at a time, satisfies Verify, ticks the checkbox, commits, stops. `litespec-review` is adversarial — GH issue + spec vs code. Close the GH issue when done — closed issues are disposable; durable specs/decisions/glossary stay.

## Skills

Only three generated skills in `.agents/skills/` (via `litespec update`): `litespec-plan` (fuzzy/clear + 5 references), `litespec-build` (one unit), `litespec-review` (adversarial). Offline fallback when `gh` unavailable: `specs/changes/<name>/QUEUE.md` with same `##` + Done means + Verify format.

## CLI

`litespec init` scaffolds `specs/product.md`, `specs/glossary.md`, `specs/decisions/` + skills. `litespec new <name> --issue N` links to GH issue (no folder in lean). `litespec validate` lints specs (SHALL/MUST + WHEN/THEN), decisions, and GH issue Verify shell (`bash -n`). `litespec view` shows product + feature specs + spine decisions + `gh issue list` when available. `litespec update` regenerates skills.

## Next Steps

- [Tutorial](tutorial.md) — walkthrough via GH issue queue
- [Concepts](concepts.md) — why load-bearing specs only
- [CLI Reference](cli-reference.md) — command details
