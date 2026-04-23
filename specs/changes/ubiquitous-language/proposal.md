# Ubiquitous Language

## Motivation

When an AI agent enters explore, grill, or propose mode, it has no shared vocabulary with the user beyond what it infers from scattered docs (AGENTS.md, DESIGN.md, spec files). Pocock's observation from DDD applies directly: without a ubiquitous language, the AI thinks in verbose, imprecise terms — and the implementation drifts from what the user intended.

A single, curated `specs/glossary.md` gives every agent session a common vocabulary before the conversation starts. The AI reads it, uses the right words, and nudges when it spots a concept that should be defined but isn't. This isn't a contract mechanism — it's a communication tool.

## Scope

### Artifact: `specs/glossary.md`

A single project-wide markdown file containing term definitions. Curated by humans (with AI proposals), version-controlled, evolving. Not auto-generated, not per-spec — one file for the whole project.

### Skill: `litespec-glossary`

A new agent skill that knows how to read, propose additions to, and maintain `specs/glossary.md`. Follows the existing skill generation pipeline (Go template in `internal/skill/` → `litespec update` → `.agents/skills/`).

### Skill integration

- **explore** — reads `specs/glossary.md` at start if it exists. Nudges: "This looks like a term that should live in the glossary — want me to add it?" If no glossary exists, suggests creating one when stable terms emerge.
- **grill** — reads `specs/glossary.md` at start if it exists. Same nudge behavior. If no glossary exists, may suggest creating one.
- **propose** — after writing specs, checks whether new terms were introduced that aren't in the glossary. Offers to update. If no glossary exists and the proposal introduces stable shared terms, offers to seed one.
- **apply** — passive. Mentioned in a references section as something the agent may consult for terminology. No enforcement.
- **review** — does not enforce glossary compliance. However, when reviewing a change with `dependsOn`, review may consult `specs/glossary.md` as supplementary terminology context alongside dependency specs/design.

### Docs integration

`specs/glossary.md` gets a page in the mkdocs site. The glossary is documentation for humans too.

### Per-spec glossary removal

The per-spec `## Glossary` sections designed in `cross-change-contracts` are replaced by the project-wide glossary. Maintaining two glossary locations (per-spec and project-wide) means the AI has to update both, which will fail often. One source of truth wins.

This means `cross-change-contracts` needs to be descoped: the glossary parsing, glossary delta operations, glossary merge logic, and dependency glossary loading are all removed. The review skill enhancement (cross-change semantic checking) survives — it reads the project glossary instead.

## Non-Goals

- **Not a compliance tool.** The glossary is for shared understanding, not enforcement. Review does not check glossary compliance. Apply does not require glossary awareness.
- **No auto-generation.** The glossary is curated. The AI proposes terms; the user approves. No NLP scanning of prose. Only stable, shared, or ambiguous terms belong — not every noun from a proposal.
- **No CLI command.** There is no `litespec glossary` command. The skill reads and writes the file directly. CLI support can come later if needed.
- **No structural validation.** `litespec validate` does not parse or validate `specs/glossary.md`. It's a prose artifact, not a spec.
- **No docs duplication.** The docs glossary page explains the concept and links to `specs/glossary.md` in the repo. It does not duplicate or inline the terms.
