# Glossary

Project-wide ubiquitous language. Curated, optional but recommended. Read this before every conversation.

- **Blocking finding**: A CRITICAL or WARNING that breaks a unit's `Done means:`/`Verify:`, contradicts a durable spec/decision, or lies inside the issue-owned review scope. Only blocking findings produce `CHANGES REQUESTED`.
- **Decision**: A durable ruling in `specs/decisions/NNNN-slug.md` with `Status`/`Context`/`Decision`/`Consequences`, optional `spine: true` for load-bearing. Created via `touch` + `validate`, not a CLI.
- **GH Issue is the queue**: The GH issue body holds proposal + design + queue, plus immutable `Base: <sha>` and `Branch: litespec/<change-name>` ownership lines. Each unit is `## <outcome>` with `Done means:` and `Verify:` that must fail without the outcome. Offline fallback: `specs/queues/<name>.md`.
- **Glossary**: `specs/glossary.md` — curated terms. Managed via plan skill, graceful degradation if absent.
- **Product**: `specs/product.md` — mental models + 2-3 flows (human + agent, agent-maintained).
- **Review scope**: All safely inspectable local contracts, tracked changes, untracked regular files, and later references used by review. Every local path and parent component is screened before content access; unsafe paths stop review.
- **Trusted bootstrap**: Harness/system instructions and repository instruction files auto-loaded before `litespec-review` activates. Litespec cannot screen these inputs; its local-path guarantee begins after activation.
- **Scenario**: A named example under a requirement using `WHEN`/`THEN` format. Load-bearing requirements must have at least one scenario. Body text must contain SHALL or MUST.
- **Skill**: Generated agent instructions in `.agents/skills/<name>/SKILL.md` via `litespec update`. Only three: `litespec-plan` (fuzzy/clear + grilling/codebase-design/domain-modeling), `litespec-build` (one unit), `litespec-review` (adversarial).
- **Spec**: A load-bearing contract in `specs/<feature>/spec.md` with SHALL/MUST and WHEN/THEN scenarios. No `canon/` — edit the file directly.
- **Evidence receipt**: Verbatim record required before ticking a unit: one exact `Verify:` command plus labeled pre and post SHAs, exit statuses, nonempty fenced raw outputs, and matching conservative scope lines. Pre is non-zero for the absent outcome; post is zero for the implementation. A nonempty `Evidence:` label is not a receipt.
- **Red-green evidence**: One exact `Verify:` command recorded failing for the absent unit outcome at a clean pre commit and passing at a later clean post commit. Build may create one or more immutable implementation/fix commits after pre; post is the final clean commit where `Verify:` passes. Review reproduces both runs and checks `HEAD`; the CLI validates only receipt structure.
- **Unit**: One demo-able outcome per GH issue `##` with `Done means:` and `Verify:` that must fail without the outcome. Built one at a time, ticked via checkbox only after an evidence receipt is posted. No `tasks.md`.
