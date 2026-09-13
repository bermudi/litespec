# litespec

A lean, AI-native spec-driven development CLI written in Go.

litespec gives AI coding agents a durable contract to work against and a lean way to prove the work happened. Small fixes need no ceremony: read the relevant spec, edit the code, update the spec if the contract changed. Bigger work goes through a GitHub issue that holds the plan — proposal, design, and one verifiable unit at a time — plus three generated skills that carry the idea from fuzzy to built to reviewed.

## How it works

**Small fix — zero ceremony.** No issue, no queue. The agent reads `specs/product.md`, the relevant `specs/<feature>/spec.md`, and decisions/glossary, edits the code, and updates the spec in place if the contract changed. Done.

**New feature — plan, build, review.** The agent grills a half-baked idea in fuzzy mode (no files), nails it into a GitHub issue in clear mode (`Base:` + `Branch:` ownership, then proposal, design, and queue), implements one unit at a time with red-green evidence, and survives an adversarial review before the issue closes. Merge the branch first, then close — a closed issue leaves no work stranded.

A unit is one boundary or one failure policy: identified `Done means:` clauses mapped to named test scenarios, one `Verify:` that fails without the outcome, and a receipt that records the exact command failing at a clean pre commit and passing at a clean post commit.

## Start here

[Getting Started](getting-started.md) — install, `init`, `view`, `validate`, `update`

[Tutorial](tutorial.md) — one complete feature cycle, issue body to closed issue

[Concepts](concepts.md) — why specs survive the chat window, durable vs disposable

[Workflow](workflow.md) — the two lanes, unit shape, evidence, review routing

[CLI Reference](cli-reference.md) — every command and flag, matching `litespec --help`

[Project Structure](project-structure.md) — what lives in `specs/` vs the issue

[Glossary](glossary.md) — the ubiquitous-language workflow
