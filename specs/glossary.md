# Glossary

Project-wide ubiquitous language. Curated, optional but recommended. Read this before every conversation.

- **Blocking finding**: A CRITICAL or WARNING that breaks a unit's `Done means:`/`Verify:`, contradicts a durable spec/decision, or lies inside the issue-owned review scope. Only blocking findings produce `CHANGES REQUESTED`.
- **Decision**: A durable ruling in `specs/decisions/NNNN-slug.md` with `Status`/`Context`/`Decision`/`Consequences`, optional `spine: true` for load-bearing. Created via `touch` + `validate`, not a CLI.
- **GH Issue is the queue**: The GH issue body holds proposal + design + queue, plus immutable `Base: <sha>` and `Branch: litespec/<change-name>` ownership lines. Each unit is `## <outcome>` with `Done means:` and `Verify:` that must fail without the outcome. Offline fallback: `specs/queues/<name>.md`.
- **Glossary**: `specs/glossary.md` — curated terms. Managed via plan skill, graceful degradation if absent.
- **Product**: `specs/product.md` — mental models + 2-3 flows (human + agent, agent-maintained).
- **Review scope**: All safely inspectable tracked changes and untracked regular files from a queue issue's `Base:` to its dedicated `Branch:` working tree. Paths are screened before content access; unsafe paths stop review.
- **Scenario**: A named example under a requirement using `WHEN`/`THEN` format. Load-bearing requirements must have at least one scenario. Body text must contain SHALL or MUST.
- **Skill**: Generated agent instructions in `.agents/skills/<name>/SKILL.md` via `litespec update`. Only three: `litespec-plan` (fuzzy/clear + grilling/codebase-design/domain-modeling), `litespec-build` (one unit), `litespec-review` (adversarial).
- **Spec**: A load-bearing contract in `specs/<feature>/spec.md` with SHALL/MUST and WHEN/THEN scenarios. No `canon/` — edit the file directly.
- **Unit**: One demo-able outcome per GH issue `##` with `Done means:` and `Verify:` that must fail without the outcome. Built one at a time, ticked via checkbox. No `tasks.md`.
