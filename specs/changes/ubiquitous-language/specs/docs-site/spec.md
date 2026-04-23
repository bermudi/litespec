# docs-site

## MODIFIED Requirements

### Requirement: Documentation Pages

The `docs/` directory SHALL contain the following markdown pages: `index.md` (landing page), `concepts.md` (philosophy and why spec-driven dev matters), `getting-started.md` (installation, init), `tutorial.md` (worked "first change" walkthrough with real output from init to archive), `workflow.md` (the explore→grill→... flow with named patterns like "Quick Feature", "Exploratory", and "Adopt"), `cli-reference.md` (every command and flag), `delta-specs.md` (delta format explained with before/after merge examples), `project-structure.md` (directory layout explained), and `glossary.md` (explains what the ubiquitous language is, how litespec uses it, how to maintain it, and links to `specs/glossary.md` as the living source of truth — does not duplicate or inline terms).

#### Scenario: Complete page set

- **WHEN** the docs site is built
- **THEN** all pages including the glossary page are accessible via the navigation and render correctly

#### Scenario: Glossary docs page

- **WHEN** a user navigates to the glossary page
- **THEN** they see an explanation of the ubiquitous language concept, how litespec uses it, and a link to `specs/glossary.md` as the source of truth
