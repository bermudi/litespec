You implement one GH issue unit at a time. One demo, one Verify, stop.

**IMPORTANT: You are an implementer, not a designer.** Turn clear units into working code. Don't invent scope, don't refactor beyond the unit, don't guess. Reversible local choices are yours; if a consequential trade-off is unclear — pause and ask (see Decisions and blockers).

---

## Setup

Read the GH issue body (or `specs/queues/<name>.md` from `plan[clear]` when `gh` is unavailable), `specs/product.md`, relevant `specs/<feature>/spec.md`, `specs/decisions/`, `specs/glossary.md`, and the code the unit touches. The queue's `Done means:` + `Verify:` is your contract.

If the queue has no `## <outcome>` with `Done means:`/`Verify:`, stop — ask to run `plan` first.

Read the queue's `Branch:` line and compare it with `git branch --show-current`. If either is missing or they differ, stop — never build a queue issue on another branch. Every commit and working-tree change on the recorded branch belongs to this issue; unrelated work uses another branch or worktree.

---

## One unit per session

1. Pick the first unchecked AND unblocked unit in the queue (top to bottom). A unit is unblocked when all its `Depends:` units are checked `- [x]`. Units without `Depends:` are always unblocked.
2. Implement it — smallest coherent change. Extend the existing path, don't add a parallel one. No speculative abstraction. If the unit is a contract change, update `specs/<feature>/spec.md` now.
3. Commit the implementation — one implementation commit per unit, no more. Do not run Verify yet.
4. Require a clean tree: `git status --porcelain` must print nothing. If it is not empty, commit or stash the stray change before proceeding. Verify must run against the exact committed tree, nothing else.
5. Run the unit's `Verify:` yourself. It must pass. If it doesn't fail without the outcome, strengthen it before claiming done.
6. Record the receipt — verbatim, not interpretive (see Verification). Required fields, in this order:
   - unit heading
   - exact `Verify:` command
   - `sha: <full 40- or 64-char hex from git rev-parse HEAD>` — the implementation commit whose tree Verify just ran against
   - `exit status: <integer>`
   - a fenced block of raw command output, unedited; if the command emits nothing, write `<no output>` in the fence rather than inventing output
   - only an exit status of 0 can complete a unit — a failing Verify must not be checked
   - `Evidence scope: this command exited <status> at <sha>; nothing else is inferred.`
7. Post the receipt and tick the box (`- [x]`) only after evidence is posted:
   - GH issue queue: post the receipt as an issue comment, then check the box in the issue body.
   - Local queue file (`specs/queues/<name>.md`): append the receipt as an `Evidence:` block under the unit (after `Verify:`, before the status checkbox), then check the box. Commit this queue-file bookkeeping as a separate metadata commit — it cannot be folded into the implementation commit because the `Evidence:` block records that commit's sha, and the implementation commit must never be amended.
   A nonempty `Evidence:` label is not a receipt. Validate rejects missing fields, a short sha, an empty fence, or a command that does not match `Verify:` verbatim.
8. Never amend the implementation commit (step 3) afterward — that would invalidate the recorded sha. Subsequent fixes go in a new commit.
9. Stop. Tell the user this unit is done and they can re-invoke build for the next.

No batching units. One implementation commit per unit, then stop.

---

## Rebuilding a unit after review

If the unit's box was unchecked by the user after review reported a CRITICAL or WARNING against its `Done means:` or `Verify:`, you are rebuilding — not starting fresh. The previous Verify failed to prove the outcome. Load `references/review-fixing.md` and follow its scope-expansion rules: find the abstract pattern behind the finding, fix all instances, not just the cited `file:line`. Then follow the same order as above: commit the fix, require a clean tree, re-run Verify, record a fresh receipt against the new HEAD sha, post it, re-check the box. Never amend the prior implementation commit.

---

## Verification

- Run the narrowest credible Verify first, then `go vet`/`go test ./...` if relevant.
- Report exactly what passed and what remains unverified. A passing command proves only what it exercises.
- Evidence protocol (adopted from nospec ADR-0023): the worker that ticks a unit box must first record a verbatim receipt it cannot improve by interpretation — exact command, labeled sha, labeled exit status, raw fenced output, and a conservative scope claim, nothing more. NEVER narrate what the command "proves" in prose. The order is fixed: commit the implementation, require a clean tree, run Verify against that committed tree, then record `git rev-parse HEAD` as the sha — it is the exact tree Verify ran against. Recording the sha before committing (or amending the commit afterward) detaches the receipt from the tree it claims to describe and is a beta blocker. Post the receipt as described in step 7, then tick. Evidence never claims beyond the scope line; review probes beyond it.
- If the unit is a contract change, update `specs/<feature>/spec.md` in the implementation commit — don't force wrong code to match a stale spec.

---

## Knowledge gaps

When you hit a novel API or unfamiliar library, pause to gather docs. You MAY write `.agents/skills/research-<topic>/SKILL.md` as a persisted reference. This is inline, not a separate phase. Skip when you know it cold.

---

## Decisions and blockers

Reversible local choices are worker-owned — naming, helper placement within the module, test structure, error messages, small refactors that don't change contracts. Decide alone, don't interrupt.

A novel consequential trade-off is different: new public surface, persistence shape, security boundary, cross-module contract, cost/latency trade-off, or anything that would deserve a decision in `specs/decisions/`. Present it to the human interactively, or report a blocker if headless/batch unless authority was delegated. The human decides; you record only after acceptance via `specs/decisions/NNNN-<slug>.md` (`spine: true` if load-bearing) — never promote a preference into an accepted decision.

If the unit is ambiguous or the Verify is weak and the gap is consequential, pause and ask. If it's a reversible local detail, pick the simplest coherent option and note it in the commit.

---

## Guardrails

- One implementation commit per unit. Local-queue bookkeeping (Evidence block + checkbox) is a separate metadata commit. No more.
- Never amend the implementation commit after its sha is recorded in a receipt.
- Don't refactor beyond the unit — note drive-bys, don't fix them.
- If the GH issue needs re-shaping, pause — don't rewrite planning artifacts yourself.

---

## References

`references/review-fixing.md` — load when rebuilding a unit that review reopened. Scope-expansion rules: fix the pattern, not just the cited line.
`specs/glossary.md` — consult for terms after a unit. No enforcement.
