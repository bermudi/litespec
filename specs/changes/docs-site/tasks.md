## Phase 1: Infrastructure

- [x] Create `pyproject.toml` with mkdocs and mkdocs-material dependencies
- [x] Create `mkdocs.yml` with site config, Material theme, and nav structure
- [x] Verify `uv run mkdocs build` produces a working site

## Phase 2: Content

- [ ] Create `docs/index.md` — landing page with pitch and key ideas
- [ ] Create `docs/concepts.md` — philosophy: what specs are, why they work, progressive rigor
- [ ] Create `docs/getting-started.md` — installation, init
- [ ] Create `docs/tutorial.md` — worked "first change" walkthrough from init to archive with real output
- [ ] Create `docs/workflow.md` — the full workflow with named patterns (Quick Feature, Exploratory, Adopt)
- [ ] Create `docs/cli-reference.md` — every command and flag documented
- [ ] Create `docs/delta-specs.md` — delta format with before/after merge example
- [ ] Create `docs/project-structure.md` — directory layout explained

## Phase 3: Deployment

- [ ] Create `.github/workflows/docs.yml` for GitHub Pages auto-deploy
- [ ] Trim `README.md` to brief summary with link to docs site
- [ ] Verify full build and local preview work correctly
