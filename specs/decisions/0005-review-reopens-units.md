---
spine: true
---

# Review Reopens Units

## Status

accepted

## Context

Review already decides whether a finding breaks a unit contract and routes that finding to rebuild. The workflow nevertheless required the user to manually uncheck the affected unit before `litespec-build` would select it. That checkbox edit carried no judgment or approval; it merely repeated a decision review had already made and could persist itself.

## Decision

When a CRITICAL or WARNING breaks a checked unit's `Done means:` or `Verify:`, review SHALL uncheck that unit before returning `CHANGES REQUESTED`. GitHub issue queues are edited remotely. Local queue changes SHALL be committed as separate routing metadata so build starts clean.

Review MUST preserve prior evidence and unaffected units. It MUST NOT check units or reopen units for findings outside the blocking unit-rebuild route. If the routing mutation cannot be persisted safely, review SHALL expose that failure and MUST NOT claim the next build can proceed.

## Consequences

Users can move directly from review to build without editing queue checkboxes. Review gains a narrow mutation right, but it remains a non-implementing role: checked-to-unchecked transitions are routing state, not code changes. Local fallback review creates one metadata commit when it reopens units.
