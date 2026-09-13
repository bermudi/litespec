---
spine: true
---

# Merge Before Close

## Status

accepted

## Context

A `litespec/*` branch vanished from the remote with its queue issue closed. Forensics showed the work had in fact squash-merged (PR #14 into `v2`, twelve days before `v2` reached main) — but nothing in the workflow would have stopped the opposite outcome: a closed issue while live work sat only on a deletable branch. The closure rule listed checked units, resolved routing metadata, and review PASS, and said nothing about the branch. Merge was customary but unenforceable, and the obvious convenience road — close now, merge whenever — stays open by default.

The alternatives were to keep merge customary, or to enforce mergedness with a git ancestry check. Customary merge is exactly the gap: it holds until one forgetful afternoon. An ancestry check cannot work: squash merges rewrite commits, so a landed branch tip is never an ancestor of main, and it would refuse genuine landings.

## Decision

A GH issue SHALL close only after its `Branch:` is merged: merge first, then close. A closed issue leaves no work stranded on a branch.

The merged PR is the test, not git ancestry. A merge into a release branch counts once its PR lands — the branch is merged even while main waits for the release.

## Consequences

Review gains one closure condition alongside checked units, resolved metadata, and PASS; the review spec, skill template, DESIGN.md, and AGENTS.md state it. There is deliberately no CLI enforcement: `validate` checks structure rather than implementation semantics or integration state, and there is no close command to gate. The rule is procedural and squash-safe by construction. Issues stay open as the honest pending-integration signal until the branch merges.
