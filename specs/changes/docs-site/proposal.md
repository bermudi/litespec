# docs-site

## Motivation

litespec needs professional documentation to share with others. The current README is a single page trying to serve as both landing page and reference — it works for quick orientation but doesn't scale for someone actually learning the tool. A proper docs site makes the project presentable and discoverable.

## Scope

- Add MkDocs with Material theme as the documentation engine
- Create a `docs/` directory with structured content:
  - `index.md` — landing page: pitch, what + why
  - `concepts.md` — philosophy: what specs are, why they work, progressive rigor
  - `getting-started.md` — installation, init
  - `tutorial.md` — worked "first change" walkthrough from init to archive with real output
  - `workflow.md` — the explore→grill→... flow with named patterns (Quick Feature, Exploratory, Adopt)
  - `cli-reference.md` — every command, every flag
  - `delta-specs.md` — delta format explained with before/after merge examples
  - `project-structure.md` — what goes where and why
- Add `mkdocs.yml` at repo root configuring the Material theme and nav
- Add `pyproject.toml` with mkdocs + mkdocs-material as deps (managed via `uv`)
- Trim README.md to a brief landing page that links to the docs site
- Add GitHub Actions workflow for auto-deploying to GitHub Pages on push to main
- Docs are the source of truth — README links out, does not duplicate
- Add explicit tool compatibility guidance (currently Claude Code via symlinks, planned: Cursor, etc.)

## Non-Goals

- Auto-generating CLI reference from Go code (manual for now, can add generation later)
- Search, i18n, versioning — MkDocs Material defaults are sufficient
- Changing any litespec CLI behavior
- Removing DESIGN.md or AGENTS.md — those serve different audiences
