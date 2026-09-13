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

The `docs/` directory SHALL contain the following markdown pages: `index.md` (landing page), `concepts.md` (philosophy and why spec-driven dev matters), `getting-started.md` (installation, init), `tutorial.md` (worked feature walkthrough from plan to review), `workflow.md` (the two-lane flow: small fix vs new feature), `cli-reference.md` (every command and flag for the 6 v2 commands), `project-structure.md` (directory layout explained), and `glossary.md` (explains what the ubiquitous language is, how litespec uses it, how to maintain it, and links to `specs/glossary.md` as the living source of truth — does not duplicate or inline terms).

#### Scenario: Complete page set

- **WHEN** the docs site is built
- **THEN** all pages including the glossary page are accessible via the navigation and render correctly

#### Scenario: Glossary docs page

- **WHEN** a user navigates to the glossary page
- **THEN** they see an explanation of the ubiquitous language concept, how litespec uses it, and a link to `specs/glossary.md` as the source of truth

### Requirement: Tutorial Walkthrough

The `tutorial.md` page SHALL contain a complete worked example of the new-feature lane: `litespec init` → plan[fuzzy] → plan[clear] (write GH issue) → build (implement one unit) → review → close issue. The example SHALL include actual GH issue body format, example spec.md contents, and command output at each stage.

#### Scenario: Tutorial covers full workflow

- **WHEN** a new user reads the tutorial
- **THEN** they understand exactly what happens from init to closing the GH issue and have a mental model of the full v2 workflow

### Requirement: Concepts Page

The `concepts.md` page SHALL explain what a spec IS vs ISN'T, what makes a good requirement and scenario, the two-lane workflow, GH issue as queue, direct spec edits, and WHY spec-driven development works for AI agents. It SHALL include examples of good vs bad specs.

#### Scenario: Concepts convince the reader

- **WHEN** a skeptical reader visits the concepts page
- **THEN** they understand the rationale behind spec-driven development and when it applies

### Requirement: CLI Reference Completeness

The `cli-reference.md` page SHALL document all 6 v2 commands (`init`, `validate`, `view`, `update`, `upgrade`, `completion`) with usage, flags, and examples. It SHALL NOT reference removed commands (`new`, `list`, `status`, `instructions`, `import`, `preview`, `archive`, `patch`).

#### Scenario: CLI reference matches binary

- **WHEN** a user runs `litespec --help` and compares to the CLI reference page
- **THEN** every command and flag listed in the help output is documented on the page

### Requirement: Tool Compatibility

The documentation SHALL explicitly list which AI tools are supported by litespec and how they integrate (currently Claude Code via symlinks). This SHALL be documented in either `getting-started.md` or a dedicated section in `index.md`.

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
