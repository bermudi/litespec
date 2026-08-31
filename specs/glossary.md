# Glossary

Project-wide ubiquitous language. Curated, optional but recommended. Read this before every conversation.

- **Blocking finding**: A CRITICAL or WARNING that breaks a unit's `Done means:`/`Verify:`, contradicts a durable spec/decision, or lies inside the issue-owned review scope. Only blocking findings produce `CHANGES REQUESTED`.
- **Decision**: A durable ruling in `specs/decisions/NNNN-slug.md` with `Status`/`Context`/`Decision`/`Consequences`, optional `spine: true` for load-bearing. Created via `touch` + `validate`, not a CLI.
- **GH Issue is the queue**: The GH issue body holds proposal + design + queue, plus immutable `Base: <sha>` and `Branch: litespec/<change-name>` ownership lines. Each unit has identified `Done means:` clauses, named `Scenarios:`, and one failing `Verify:`; boundary units also account for standard risks. Offline fallback: `specs/queues/<name>.md`.
- **Glossary**: `specs/glossary.md` — curated terms. Managed via plan skill, graceful degradation if absent.
- **Product**: `specs/product.md` — mental models + 2-3 flows (human + agent, agent-maintained).
- **Re-plan marker**: Append-only blocking metadata recorded when another unit-breaking finding follows two completed rebuild cycles against one contract digest. Build refuses the marked contract until plan reshapes it through an amendment from that digest.
- **Review coverage record**: Append-only, HEAD-keyed account of adversarial scenarios a review exercised, did not exercise, or could not resolve. Later reviewers use it to expand an independently drafted risk inventory, never as proof.
- **Review scope**: All safely inspectable local contracts, tracked changes, untracked regular files, and later references used by review. Every local path and parent component is screened before content access; unsafe paths stop review.
- **Trusted bootstrap**: Harness/system instructions and repository instruction files auto-loaded before `litespec-review` activates. Litespec cannot screen these inputs; its local-path guarantee begins after activation.
- **Scenario**: A named example under a requirement using `WHEN`/`THEN` format. Load-bearing requirements must have at least one scenario. Body text must contain SHALL or MUST.
- **Skill**: Generated agent instructions in `.agents/skills/<name>/SKILL.md` via `litespec update`. Only three: `litespec-plan` (fuzzy/clear + grilling/codebase-design/domain-modeling), `litespec-build` (one unit), `litespec-review` (adversarial).
- **Spec**: A load-bearing contract in `specs/<feature>/spec.md` with SHALL/MUST and WHEN/THEN scenarios. No `canon/` — edit the file directly.
- **Evidence receipt**: Verbatim record required before ticking a unit: its contract `unit digest:`, one exact `Verify:` command, labeled pre and post SHAs and exit statuses, nonempty fenced raw outputs, and matching conservative scope lines. Pre is non-zero for the absent outcome; post is zero for the implementation. A nonempty `Evidence:` label is not a receipt.
- **Red-green evidence**: One exact `Verify:` command and its unit contract digest recorded with a failing absent-outcome run at a clean pre commit and a passing run at a later clean post commit. Build may create immutable implementation/fix commits after pre; review reproduces both runs and checks `HEAD`. The CLI validates receipt structure only.
- **Unit**: One external boundary or one failure policy per queue `##`, with identified `Done means:` clauses mapped to named test scenarios and one `Verify:` that fails without the outcome. Its routing identity is exact heading plus positive same-heading occurrence. Built one at a time; after two rebuild cycles against one digest, another unit-breaking finding requires contract reshaping by plan. No `tasks.md`.
