---
spine: true
---

# Review Routes Units Back to Build

## Status

accepted

## Context

Review already decides whether a finding breaks a unit contract and routes that finding to rebuild. The workflow nevertheless required the user to manually uncheck the affected unit before `litespec-build` would select it. That checkbox edit carried no judgment or approval; it merely repeated a decision review had already made.

Replacing a GitHub issue body to flip a checkbox is too broad a write for this routing decision. It risks overwriting concurrent edits and destroys the append-only history of why a checked unit became buildable again.

## Decision

When a CRITICAL or WARNING breaks a checked GitHub unit's `Done means:` or `Verify:`, review SHALL leave the issue body unchanged and post one append-only structured rebuild-request comment before returning `CHANGES REQUESTED`. The comment SHALL identify the exact unit heading and its positive 1-based occurrence among queue units with that heading.

A GitHub request remains unresolved until a later comment carries a complete evidence receipt for that same heading and occurrence. Build SHALL treat a checked unit with an unresolved request as selectable and SHALL identify the same unit in its fresh receipt. One later complete receipt resolves every earlier request for that unit.

Local queues retain the checkbox mechanism: review SHALL uncheck affected units and commit that routing metadata separately so build starts clean. Review MUST preserve prior evidence and unaffected units. It MUST NOT check units or route units for findings outside the blocking unit-rebuild route. If routing metadata cannot be persisted safely, review SHALL expose that failure and MUST NOT claim the next build can proceed.

## Consequences

Users can move directly from review to build without editing queue checkboxes. GitHub routing is append-only, avoids lost issue-body updates, and makes repeated requests and their later evidence auditable. Build selection and issue closure must account for unresolved requests even while the body checkbox remains checked. Local fallback review still creates one metadata commit when it reopens units.
