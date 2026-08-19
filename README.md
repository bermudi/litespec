# litespec

A lean, AI-native spec-driven development CLI.

`go 1.26+` | [Design Doc](DESIGN.md) | [Docs](https://bermudi.github.io/litespec/) | Inspired by [OpenSpec](https://github.com/Fission-AI/OpenSpec)

---

litespec gives AI coding agents structured workflows that keep your codebase aligned with your specifications. It's a reimagining of OpenSpec with stronger opinions: fewer concepts, leaner skills, unidirectional flow. GH issue is the queue — the GH issue body is proposal + design + queue (64k limit, no overflow). No `specs/changes/` — closed issues are deleted, durable specs live in `specs/<feature>/spec.md`.

## Quick Start

```bash
# Install
go install github.com/bermudi/litespec/cmd/litespec@latest

# Initialize — scaffolds specs/product.md, specs/glossary.md, specs/decisions/ + .agents/skills/
litespec init

# Small fix — zero ceremony (no issue)
# agent reads product + spec + decisions -> edits code -> updates spec if contract -> done

# New feature — link to GH issue
litespec new add-feature --issue 42   # GH issue holds proposal + design + queue
litespec validate
litespec view                         # product + specs + open litespec GH issues
litespec update                       # regenerate 3 skills
```

## Two Lanes

**Small fix — zero ceremony:** You say "fix typo" -> agent reads product + relevant spec + decisions/glossary -> edits code -> updates the one `specs/<feature>/spec.md` if contract change -> done. No `new`, no issue required.

**New feature — plan fuzzy -> clear:** `plan[fuzzy]` (read code, ask 2-3 questions, no files) -> `plan[clear]` (write GH issue: proposal + design + units with Done means + Verify; also draft spec if load-bearing) -> you: "looks good" or "grill-me" -> `build` one unit at a time (Verify must fail without outcome) -> `review` adversarial -> close GH issue.

Each unit is one demo-able outcome `## <name>` with `Done means:` and `Verify:`.

## What Makes litespec Different

- **GH issue is the queue** — proposal + design + queue live in the GH issue body, not `specs/changes/` or `QUEUE.md`. Offline fallback via `specs/queues/<name>.md` when `gh` is unavailable; `--issue` required for `litespec new` to link to GH.
- **Convention over configuration** — zero config files. All defaults.
- **3 skills only** — `litespec-plan` (fuzzy/clear + grilling/codebase-design/domain-modeling), `litespec-build` (one unit), `litespec-review` (adversarial). Progressive disclosure via `references/`, lean tokens.
- **Two lanes** — small fix (zero ceremony) vs new feature (plan fuzzy -> clear -> build -> review).
- **Direct spec edits** — no ADDED/MODIFIED delta flow. Edit `specs/<feature>/spec.md` directly. SHALL/MUST + WHEN/THEN.

## Documentation

**[Full Documentation → https://bermudi.github.io/litespec/](https://bermudi.github.io/litespec/)**

- [Getting Started](https://bermudi.github.io/litespec/getting-started/) — Installation and setup
- [Tutorial](https://bermudi.github.io/litespec/tutorial/) — Walkthrough via GH issue queue
- [CLI Reference](https://bermudi.github.io/litespec/cli-reference/) — `init`, `new --issue`, `validate`, `view`, `update`
- [Workflow](https://bermudi.github.io/litespec/workflow/) — Two lanes, GH issue queue, unit rule
- [Concepts](https://bermudi.github.io/litespec/concepts/) — Load-bearing specs, decisions, glossary

## Contributing

```bash
git clone https://github.com/bermudi/litespec.git
cd litespec
go build ./cmd/litespec
./litespec update    # generate skills into .agents/skills/
```

Skills are generated from Go templates in `internal/skill/` — not tracked in git. Run `litespec update` after cloning.

## Status

This is an active experiment. Decisions made yesterday may be revised today if we find something better.
