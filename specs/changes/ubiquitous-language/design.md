# Ubiquitous Language — Design

## Architecture

```
specs/glossary.md                    ← the artifact (curated, version-controlled)
        │
        ├── read by explore/grill    ← active: establishes shared vocab at session start
        ├── read by propose          ← active: checks for new terms after writing specs
        ├── read by apply            ← passive: referenced in "further reading", not enforced
        │
        ├── maintained by glossary skill  ← knows the format, proposes updates
        │
        └── published in docs/glossary.md ← human-facing documentation page
```

No CLI command. No validation. No parsing in Go. The glossary is a prose artifact — the skill reads and writes it directly.

## Decisions

### Single file, not per-spec

**Chosen:** One `specs/glossary.md` for the whole project.

**Why:** The per-spec glossary designed in `cross-change-contracts` is a contract mechanism — terms scoped to a capability, delta-merged at archive time, validated for cross-change consistency. But the glossary Pocock describes is a communication tool — the shared language for the whole project. These are different purposes, and maintaining both means the AI has to update two places, which will fail often. One source of truth wins.

**Impact on cross-change-contracts:** The per-spec glossary (phases 1-3 of that change) is removed. The review skill enhancement (phase 4) survives — it reads the project glossary instead of per-spec glossaries. The change needs to be descoped.

### No CLI command

**Chosen:** The glossary skill reads and writes `specs/glossary.md` directly. No `litespec glossary` CLI command.

**Why:** The CLI is a read-only context provider for the AI. The glossary is a markdown file the AI already knows how to read and edit. Adding a CLI command is premature — we don't know if we'll need structured access to glossary data. If we do, add it later.

### Glossary is not validated

**Chosen:** `litespec validate` does not parse or validate `specs/glossary.md`.

**Why:** The glossary is prose — term definitions written in natural language. There's no structural correctness to enforce beyond "is it valid markdown." The skill handles formatting consistency. Validation would mean parsing the glossary in Go, adding it to `ValidationResult`, and maintaining that code — all for a file that changes infrequently and has no merge semantics.

### Active in conversation skills, passive in execution skills

**Chosen:** explore and grill actively read the glossary and nudge. propose checks for new terms. apply mentions it in references but doesn't enforce. review does not enforce glossary compliance, but may consult it as supplementary context during cross-change review.

**Why:** The glossary is a communication tool. Communication happens during explore, grill, and propose — the skills where humans and AI establish shared understanding. Apply is execution — the agent should focus on coding, not terminology. Review should focus on code quality, not glossary compliance. Cluttering the reviewer's attention with glossary checks wastes their most valuable resource: focus. The one exception is cross-change review (when `dependsOn` exists): there, the glossary provides useful background terminology alongside the dependency's specs and design — but as context, not as a checklist.

### Graceful degradation when glossary doesn't exist

**Chosen:** All skills degrade gracefully. "Read `specs/glossary.md` if it exists" — if it doesn't, no error, no block. Conversation skills (explore, grill, propose) may suggest creating one when stable terms emerge. The glossary skill itself can seed the file.

**Why:** Not every project starts with a glossary. Forcing it would violate convention over configuration. The nudge-and-create pattern lets the glossary emerge naturally from real conversations rather than being imposed upfront.

## File Changes

### `specs/glossary.md` (new)

The initial glossary file. Seeded with litespec's core terms extracted from AGENTS.md and DESIGN.md. Format:

```markdown
# Glossary

Project-wide ubiquitous language. Read this before every conversation.

- **Canon**: The source-of-truth specs in `specs/canon/`. Represents what the system currently IS — accepted capabilities, not proposed changes.
- **Change**: An isolated proposed modification in `specs/changes/<name>/`. Contains planning artifacts (proposal, specs, design, tasks). Tentative until archived.
- **Delta**: A spec describing differences against canon using ADDED/MODIFIED/REMOVED/RENAMED markers. Not a standalone spec — only meaningful relative to a canonical spec.
- **Archive**: Promoting a change to implemented — merging its deltas into canon and moving the change directory to `specs/changes/archive/`.
- **Phase**: A group of related tasks in `tasks.md`. One phase = one apply session = one commit. The first phase with unchecked tasks is the current phase.
- **Skill**: Generated agent instructions in `.agents/skills/<name>/SKILL.md`. Produced from Go templates via `litespec update`, never written directly.
- **Artifact**: One of the four planning documents in a change: proposal.md, specs/, design.md, tasks.md. Created in dependency order during propose.
```

### `internal/paths.go`

Add a new `SkillInfo` entry to the `Skills` slice:

```go
{
    ID:          "glossary",
    Name:        "litespec-glossary",
    Description: "Manage the project's ubiquitous language in specs/glossary.md. Use when the user wants to review, update, or seed the project glossary, or says \"glossary\".",
},
```

### `internal/skill/glossary.go` (new)

New template file. Registers the glossary skill template via `init()`. The template instructs the agent to:

1. Read `specs/glossary.md`
2. Understand the current terms
3. Propose additions/modifications when asked or when encountering undefined terms
4. Maintain the `- **Term**: definition` format consistently
5. Include the "not-that" — what a term explicitly does NOT mean, where disambiguation matters

### `internal/skill/explore.go`

Add glossary awareness to the explore template:

- In the "Litespec Awareness" section, add: read `specs/glossary.md` if it exists to establish shared vocabulary
- Add nudge behavior: when a concept surfaces during exploration that seems foundational but isn't in the glossary, offer to add it

### `internal/skill/grill.go`

Add glossary awareness to the grill template:

- At session start, read `specs/glossary.md` to speak the same language
- During grilling, when a new term crystallizes from the discussion, nudge: "This looks like a term for the glossary — want me to add it?"

### `internal/skill/propose.go`

Add glossary check to the propose template:

- After writing specs, check whether new terms were introduced that aren't in `specs/glossary.md`
- Offer to update the glossary with the new terms

### `internal/skill/apply.go`

Add passive glossary reference:

- In a references/further reading section at the end, mention `specs/glossary.md` as available context for terminology
- No enforcement, no nudge — the agent may check it at the end of a phase if it wants to

### `docs/glossary.md` (new)

New docs page explaining:
- What the ubiquitous language is and why it matters (cite DDD/Pocock briefly)
- How litespec uses it (which skills read it, the nudge behavior)
- How to maintain it (the glossary skill, manual edits)
- Links to `specs/glossary.md` in the repo as the living source of truth (does not duplicate or inline terms)

### `mkdocs.yml`

Add `Glossary: glossary.md` to the nav section.

### `internal/skill/skill_test.go`

Update the expected skill ID list to include "glossary" so template registration tests pass.

## Impact on cross-change-contracts

The `cross-change-contracts` change needs to be descoped. Specifically:

**Removed (phases 1-3):**
- `GlossaryEntry` and `DeltaGlossaryEntry` structs
- `ParseGlossaryEntries` function
- Glossary parsing in `ParseMainSpec` and `ParseDeltaSpec`
- Glossary serialization in `SerializeSpec`
- Glossary merge logic in `MergeDelta`
- Glossary validation in `ValidateChange`
- `DependencyGlossary` field in `ValidationResult`
- Dependency glossary loading
- All associated tests
- Delta spec changes for glossary operations
- Validate spec changes for glossary

**Survives (phase 4, adapted):**
- Review skill cross-change dependency awareness — but reads `specs/glossary.md` for term context instead of per-spec glossary entries

**cross-change-contracts artifact updates:**
- Descope `proposal.md` — remove glossary structural layer from scope, keep review skill enhancement
- Descope `design.md` — remove glossary architecture/decisions/file changes, adapt review to read project glossary
- Descope `tasks.md` — remove phases 1-3, keep phase 4 adapted
- Descope `specs/spec-format/spec.md` — remove all glossary requirements
- Descope `specs/validate/spec.md` — remove all glossary validation requirements
- Descope `specs/review/spec.md` — adapt to reference project glossary as supplementary context, name drift stays WARNING not CRITICAL

**DESIGN.md and AGENTS.md updates:**
- Remove `## Glossary` from canonical spec format
- Remove glossary delta operations
- Remove dependency glossary loading
- Add `specs/glossary.md` as a project artifact
- Update glossary section in DESIGN.md to reflect project-wide glossary
