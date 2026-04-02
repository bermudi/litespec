# docs-site

## Requirements

### Requirement: Documentation Engine

The project SHALL use MkDocs with the Material theme as its documentation engine, configured via `mkdocs.yml` at the repo root and Python dependencies managed through `uv` with a `pyproject.toml`.

#### Scenario: Local docs preview

- **WHEN** a developer runs `uv run mkdocs serve`
- **THEN** a local preview of the docs site is available with hot-reload

#### Scenario: Build for deployment

- **WHEN** a developer runs `uv run mkdocs build`
- **THEN** a static site is generated in the `site/` directory

### Requirement: Documentation Pages

The `docs/` directory SHALL contain the following markdown pages: `index.md` (landing page), `concepts.md` (philosophy and why spec-driven dev matters), `getting-started.md` (installation, init), `tutorial.md` (worked "first change" walkthrough with real output from init to archive), `workflow.md` (the explore→grill→... flow with named patterns like "Quick Feature", "Exploratory", and "Adopt"), `cli-reference.md` (every command and flag), `delta-specs.md` (delta format explained with before/after merge examples), and `project-structure.md` (directory layout explained).

#### Scenario: Complete page set

- **WHEN** the docs site is built
- **THEN** all nine pages are accessible via the navigation and render correctly

### Requirement: Tutorial Walkthrough

The `tutorial.md` page SHALL contain a complete worked example from `litespec init` through `litespec archive`, showing what the AI says and does at each step. The example SHALL include actual file contents at each stage (proposal, specs, design, tasks) and describe the state of the repository before and after each command.

#### Scenario: Tutorial covers full workflow

- **WHEN** a new user reads the tutorial
- **THEN** they understand exactly what happens from init to archive and have a mental model of the full workflow

### Requirement: Concepts Page

The `concepts.md` page SHALL explain what a spec IS vs ISN'T, what makes a good requirement and scenario, progressive rigor, and WHY spec-driven development works. It SHALL include examples of good vs bad specs.

#### Scenario: Concepts convince the reader

- **WHEN** a skeptical reader visits the concepts page
- **THEN** they understand the rationale behind spec-driven development and when it applies

### Requirement: Delta Specs Worked Example

The `delta-specs.md` page SHALL include a before/after example showing how a canonical spec changes after applying ADDED/MODIFIED/REMOVED/RENAMED delta operations at archive time. The example SHALL be concrete and show the exact transformation.

#### Scenario: Delta merge is clear

- **WHEN** a user reads the delta-specs page
- **THEN** they understand how delta operations merge into the canonical spec

### Requirement: Tool Compatibility

The documentation SHALL explicitly list which AI tools are supported by litespec and how they integrate (currently Claude Code via symlinks, planned: Cursor, etc.). This SHALL be documented in either `getting-started.md` or a dedicated section in `index.md`.

#### Scenario: Tool support is clear

- **WHEN** a user wants to know if litespec works with their AI tool
- **THEN** they can quickly find this information in the documentation

### Requirement: README as Landing Link

The `README.md` SHALL be trimmed to a brief summary with a prominent link to the docs site, removing the detailed command reference and workflow sections that now live in the docs.

#### Scenario: README links to docs

- **WHEN** a visitor reads the README on GitHub
- **THEN** they see a short description of litespec and a clear link to the full documentation site

### Requirement: GitHub Pages Deployment

A GitHub Actions workflow SHALL auto-deploy the docs site to GitHub Pages on every push to the `main` branch.

#### Scenario: Auto-deploy on push

- **WHEN** a commit is pushed to `main`
- **THEN** the GitHub Actions workflow builds the MkDocs site and deploys it to GitHub Pages
