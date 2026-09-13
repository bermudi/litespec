---
spine: true
---

# Units Require Red-Green Evidence

## Status

accepted

## Context

A unit's `Verify:` command is required to fail without its outcome, but the workflow only recorded a passing run after implementation. That left the negative claim unaudited: a vacuous command, a check aimed at behavior that already existed, or a newly authored test tailored to the implementation could be accepted without any observed failing state.

A pre run alone is insufficient when it fails because of a typo, missing command, dependency failure, or broken environment. An uncommitted pre run is also not reproducible by review. Some units introduce the test or other verifier that makes their Verify command meaningful, so the verifier cannot always run on the branch's starting commit.

## Decision

Every completed unit SHALL carry one exact Verify command with reproducible red-green evidence: a non-zero run on a clean pre-outcome commit and a zero-status run on a later clean post commit. The pre failure MUST be attributable to the absent unit outcome. Build and review SHALL reject a green pre run or an unrelated pre failure.

When Verify already runs on the starting commit, that commit is the pre commit. When the unit introduces its verifier, build MAY create at most one verifier-only commit and use it as pre. After pre, build SHALL create one or more implementation/fix commits, and post is the final clean commit where `Verify:` passes. No commit after pre may be amended, and the recorded pre commit remains immutable. Review SHALL rerun Verify at pre, post, and `HEAD` from detached temporary worktrees.

There SHALL be no green-to-green exception. Commands that only preserve an invariant or check general hygiene MAY supplement Verify, but they do not satisfy the unit contract because they do not discriminate the new outcome.

The CLI SHALL validate receipt structure only. Skills SHALL judge whether the pre failure is caused by the missing outcome and whether Verify targets the intended behavior. Red-green evidence establishes discrimination between two trees, not the correctness or completeness of the target.

## Consequences

Receipts become larger and checked legacy units require rebuilding under the new shape. Units that introduce tests may add at most one verifier-only commit, followed by one or more immutable implementation/fix commits. Review performs three Verify runs instead of one or two. In return, the negative half of the unit contract becomes reproducible and weak Verifies are exposed before a unit can complete.
