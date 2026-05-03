# Design: Fix Skill

## Architecture

The fix skill follows the same architecture as all other litespec skills: a `SkillInfo` entry in `internal/paths.go`, a template registered in `internal/skill/fix.go` via `init()`, and automatic generation into `.agents/skills/litespec-fix/SKILL.md` via `litespec update`.

The fix skill sits between review and archive in the workflow, closing the quality feedback loop:

```
explore → grill → propose → apply → review → fix → review(verify) → archive
```

It operates on the same change as the review that produced the findings. No new change is created, no new delta specs are written. The fix skill is purely an implementation correction mechanism.

The fix skill ingests review findings as a structured seam — the output of the review phase is the input to the fix phase. Neither skill needs to know the other's internals; they communicate through the review report format (CRITICAL/WARNING/SUGGESTION with file:line and recommendation).

## Decisions

### Decision 1: Separate skill, not a mode of apply

**Chosen:** A standalone `fix` skill.

**Why:** Apply and fix have fundamentally different contracts. Apply ingests a phased task plan and builds new things. Fix ingests structured review findings and corrects existing things. Their verification standards differ (tests pass vs. findings demonstrably resolved). Their failure modes differ (incomplete implementation vs. incomplete correction + new bugs). Sharing a skill would blur these contracts and make each mode harder to reason about.

**Constraints:** Adds a 13th skill to the litespec skill list. Each skill remains focused on one job.

### Decision 2: Fix operates on review report, not tasks.md

**Chosen:** The fix skill reads the review report directly (from session context or from a review output file), not from `tasks.md`.

**Why:** Review findings are a different input shape than phased tasks. Appending findings to `tasks.md` would create a confusing blend — some tasks represent planned work, others represent corrections. It would also violate the review skill's "pure review, never write" identity (review would need to modify `tasks.md`). Review findings are backpressure, not forward tasks.

**Constraints:** The fix skill must have the review findings in its context window. The user provides these by invoking the fix skill in the same session as the review, or by pasting the review output.

### Decision 3: Verify per finding, not all at once

**Chosen:** The fix skill fixes one finding, verifies it, then moves to the next.

**Why:** This is the mechanical defense against compounding booboos. Fixing everything then testing creates a combinatorial explosion of possible interactions. Fixing one by one keeps the agent in the Smart Zone — each fix is scoped and its effect is immediately verified.

**Constraints:** May be slower than batch fixing, but correctness over speed.

### Decision 4: Escalation over silent resolution

**Chosen:** If a fix cannot be resolved (e.g., ambiguity in the finding, conflicting recommendations), the fix agent must surface that as an explicit warning in its output, not silently skip it.

**Why:** Silent resolution is indistinguishable from successful resolution to the human. An unresolvable finding that disappears is worse than one that remains flagged as "needs human judgment."

### Decision 5: Re-review after fix

**Chosen:** The fix skill instructs the agent to suggest a follow-up review after all fixes are applied.

**Why:** Fixes can introduce regressions. A lightweight re-review (ideally just verifying that previous findings are resolved) closes the quality loop. This is not enforced by the skill — it's a recommendation the human can accept or skip.

## File Changes

### `internal/paths.go` — Add fix skill entry

Add a `SkillInfo` entry to the `Skills` slice with ID `"fix"`, name `"litespec-fix"`, and description referencing review findings.

Relates to spec requirement: **Fix Skill Registration**

### `internal/skill/fix.go` — New file: fix skill template

Create a new file registering the fix skill template via `init()`. The template describes:
- Setup: load review findings and change artifacts
- Workflow: group findings by file/priority, address CRITICAL → WARNING → SUGGESTION
- Verification: fix one, verify, move on; run `litespec validate` after all fixes
- Escalation: surface unresolvable findings explicitly
- Closing: suggest re-review, commit

Relates to spec requirements: **Fix Skill Registration**, **Fix Skill Workflow**

### `internal/skill/review.go` — Update ending section

In the review template, change the ending section from instructing users to "use apply" for fixes to instead reference the fix skill.

Current text (line ~review template ending):
> If the user asks you to fix things, tell them to use apply.

Replaced with:
> If the user asks you to fix things, tell them to use the fix skill (litespec-fix).

Relates to spec requirement: **Fix Skill Handoff**

### `DESIGN.md` — Update workflow diagram

Update the workflow diagram to include `fix` between `review` and `archive`:
```
explore → grill → propose → apply → review → fix → review(verify) → archive
```

Also add fix to the Skills table.

### `.agents/skills/litespec-fix/SKILL.md` — Generated output

This file is generated by `litespec update`, not written directly. Listed for completeness — it will be the runtime skill that agents discover.
