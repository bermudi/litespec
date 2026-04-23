# Architectural Decision Records

## Motivation

Litespec has no home for locked architectural reasoning that spans multiple changes. Three existing artifact types fall short:

- **Canon specs** describe *what* a capability does, not *why it was built that way*.
- **`design.md`** lives inside one change, gets moved to `specs/changes/archive/<date>-<name>/` on archive, and becomes hard to cite from subsequent changes. It also conflates "how to implement this change" with "standing architectural rulings that should outlive this change".
- **Research skills** hold external knowledge (API docs, schemas), not internal rulings.

On a real, non-trivial project (little-goblin, a Telegram-hosted AI agent), an architectural review produced roughly a dozen locked decisions — *single shared workspace*, *β tools bound at creation*, *fire-every-tool-filter-at-buffer*, *strict skill isolation*, *recursion cap of 3*, and so on. None of these fit neatly into any existing litespec shape, so they end up in ad-hoc files like `progress.md` or `AGENTS.md`, outside the spec system. When future changes need to cite them, the citation target does not exist.

This change introduces **decisions** as a first-class artifact: persistent, numbered, cross-referenceable rulings that survive archive, are authored outside the change workflow, and are surfaced by `list`, `view`, and `validate`.

## Scope

### New artifact type: decisions

- A new directory `specs/decisions/` containing one file per decision: `NNNN-kebab-name.md`.
- Each decision is a standalone markdown file with a small required structure: title (H1), status (*proposed* | *accepted* | *superseded*), context, decision, consequences. Optional `supersedes:` / `superseded-by:` links to other decision slugs.
- Decisions are **not scoped to a single change**. They are project-level rulings. A change's `design.md` may cite one or more decision slugs; a decision may be referenced by zero or many changes.
- Decisions **never archive**. They live in `specs/decisions/` permanently. When a decision is replaced, the old file stays on disk with `status: superseded` and a `superseded-by:` pointer.

### CLI surface

- `litespec decide <name>` — create a new decision file with next available number, scaffolded structure, status *proposed*.
- `litespec list --decisions [--status <state>] [--sort number|recent|name]` — list decisions with number, title, status.
- `litespec validate --decisions` (and inclusion in `--all`) — validates decision file structure: required sections present, status is a valid value, supersede links resolve, no duplicate numbers, no orphan *superseded-by* (target exists).
- `litespec view` — renders a **Decisions** section alongside Specs and Changes, grouped by status (active decisions first, superseded collapsed or counted).

### Cross-referencing

- No structural linking requirement — decisions are cited by slug (e.g., "per `0003-beta-tools-session-bound`") in prose inside `design.md`, canon spec `## Purpose` sections, or other decisions. Validation does not require citations; it only checks that *declared* `supersedes:` / `superseded-by:` pointers resolve.

### Skill updates

- `grill` and `propose` skills gain a short note: when a locked architectural ruling emerges that is broader than the current change, suggest creating a decision via `litespec decide` rather than burying it in `design.md`.
- `review` skill, when reviewing `design.md`, may flag language that looks like a standing ruling ("we will always…", "all changes must…") and suggest promoting it to a decision. This is advisory, not enforced.

## Non-Goals

- **Not ADR tooling in the full MADR / Nygard sense.** We do not generate templates for "options considered", "pros/cons matrices", or decision logs with reviewers. A decision file is prose with a small shape, not a form.
- **Not tied to changes.** Decisions are not created from within a change directory and have no `dependsOn`, no tasks, no deltas. They are their own artifact type.
- **Not auto-extracted from design.md.** We do not parse `design.md` to lift rulings into decisions. Promotion is explicit and author-driven.
- **Not required.** A project can use litespec without creating any decisions. The `decisions/` directory is optional; its absence is not an error.
- **Not versioned inside a single file.** Revising a decision means creating a new numbered file with `supersedes: <old-slug>` and marking the old one *superseded*. No in-file history, no git-style diffs. Git handles history.
- **No generated human docs.** We do not render `docs/decisions.md` or similar synthesis. `view` and `list --decisions` are the surface. Downstream tools can consume the directory if they want.
- **No change to the canonical spec format.** Decisions live alongside canon and changes, not inside them. Canon specs gain nothing new structurally.
