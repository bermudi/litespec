---
spine: true
---

# Receipts Quote the Validator's Verify Extraction

## Status

accepted

## Context

Two Verify-extraction functions coexisted. `unitVerifyCommand` is what validate's receipt cross-check compares against and what `unitContractDigest` binds, so it is what digests and validation agree on. `locateVerifyCommand` is fence-aware and tolerant of prose between the `Verify:` label and its fence, and `ResolveQueueUnit` used it — so `litespec receipt` quoted a Verify command that validate's evidence grammar did not. On adversarial but structurally valid queue shapes (a `Verify:` line inside an earlier fenced Constraints block; prose between the label and its fence) the two extractors diverge and receipts came out that validate rejects, contradicting the "validates by construction" property receipts are supposed to have.

## Decision

Canonical Verify-extraction semantics are the validator-bound extractor: `unitVerifyCommand` is the single source every consumer quoting a Verify command must use, and `ResolveQueueUnit` (hence `litespec receipt`) mirrors validate by construction. When the two disagreed, the validator wins; whatever `unitVerifyCommand` yields is canonical, garbage or not. Improving extraction quality on edge shapes is deliberately out of scope here: it would alter contract digests mid-beta, and contract changes are plan territory per decision `0006-unit-contracts-are-amendable-only-by-plan`. `locateVerifyCommand` remains in validate only for its structural presence checks, never as a quote source.

## Consequences

Receipts can no longer diverge from what validate cross-checks, so any extraction weakness (naive fence handling, prose intolerance) is now uniformly visible in digests and receipts instead of hiding behind a friendlier quote. The empty extraction is a visible refusal in `litespec receipt` rather than a rejected receipt later. Fixing the underlying extraction semantics requires a plan-authored change that owns the digest churn and any amendment chain it triggers.
