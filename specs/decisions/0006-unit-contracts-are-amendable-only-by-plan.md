---
spine: true
---

# Unit Contracts Are Amendable Only by Plan

## Status

accepted

## Context

Decision `0005-review-reopens-units` made review's reopen route append-only, but the contract a receipt satisfies was still mutable by anyone editing the queue text. Evidence receipts are content-addressed to two immutable commits and one exact command — yet validate read `Verify:` from the *current* body, so an actor could edit a unit's `Done means:` or `Verify:` after evidence was recorded and every structural check would pass against the moved goalpost. The chain of custody bound commits; nothing bound the contract text itself.

A related question: who may change a contract at all? Build owns outcomes within fixed contracts. Review judges compliance. If either could repair a contract, evidence would silently bind to whatever text stood when validation ran.

## Decision

Plan is the only actor permitted to author or alter a unit contract. A contract change is witnessed by an append-only record — GitHub comment or a block appended after all units of a local queue file — with this grammar:

```text
Amendment:
Unit occurrence: <positive 1-based occurrence>
Unit heading: <exact post-amendment heading>
Old digest: <64 lowercase hex>
New digest: <64 lowercase hex>
Reason: <one line>
```

Identity fields carry the post-amendment identity because heading is itself a contract field an amendment may rename; `Old digest:` is the only link to the superseded text and what makes a rename auditable rather than paradoxical. An amendment is an unresolved rebuild request resolved only by a later complete identity-bearing receipt whose `unit digest:` equals `New digest:`. A receipt whose digest is superseded is checked with the Verify command and digest it declares, including the repeated identity on every raw-output chunk; it is not checked against the changed current Verify command. A receipt declaring the current digest must still quote the current Verify command exactly. Observed receipt digests must form a chain over amendment edges ending at the current contract digest: two amendments between receipts bridge through their shared intermediate digest, while an edit-plus-repost with no bridging amendment is a visible validation error. This extends `0005-review-reopens-units`; where that ruling reopened units without touching the body, this ruling governs who may move the target those bodies record.

The digest is computed by `litespec digest` (canonicalized sha256 over the unit's contract fields) so the honest path is also the easy path; hand-rolled recomputation is how accidental mismatches masquerade as moved goalposts.

## Consequences

Contract edits become visible and consequential but not cryptographically prevented: comments are deletable via API, so a determined adversary can still evade the chain; the goal is narrower — unwitnessed edits invalidate receipts and block closure, witnessed edits reopen the unit honestly. Neither build nor review may repair a contract, so weak contracts found mid-flight route to plan under triage rule 3. Revisit if amendment volume shows the ceremony costs more than silent-edit risk it prevents.
