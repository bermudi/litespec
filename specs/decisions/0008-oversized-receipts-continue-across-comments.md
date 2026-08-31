---
spine: true
---

# Oversized Receipts Continue Across Comments

## Status

accepted

## Context

GitHub caps issue comments at 65,536 characters. A verbatim red-green receipt carries both raw command outputs, and real test suites routinely exceed the cap — little-goblin issue #53 unit 1 produced a 58,823-byte pre output and a 49,824-byte post output, roughly 109KB with fields, impossible to post as one comment. The first instance stalled build mid-session: neither trimming (the receipt claims unedited output) nor splitting (validate required every field in one comment) was within the worker's discretion.

The alternatives were to truncate fenced output, move raw outputs off the issue, or shrink the evidence requirement. Truncation breaks the verbatim property review relies on when diffing posted output against replayed Verify. Gists or in-repo evidence files externalize queue metadata that belongs on the issue, add mutable or fetchable dependencies, and contradict the ruling that GitHub metadata lives in comments. Shrinking the receipt to fields-only discards the transcript entirely.

## Decision

A receipt that exceeds the GitHub comment cap SHALL continue into the immediately following comment rather than truncate or leave the issue. The split SHALL fall on a field boundary — after a `Pre-evidence scope:` or `Post-evidence scope:` line, never inside a fenced output block — and every full comment SHALL end with the literal line `Receipt continues in next comment (GitHub comment size limit).` Chaining across more than two comments is allowed. Fenced output SHALL remain unedited and complete; marker text inside a fence is raw output, never a marker.

`litespec validate` SHALL join a comment ending in the continuation marker with the immediately following comment before scanning, preserving comment indices and error numbering. A dangling marker (no following comment) or an interrupted continuation SHALL surface as a visible incomplete-receipt error, never as a silent pass.

## Consequences

Verbatim evidence survives real-world output sizes, and review still reads one logical receipt. Workers no longer stall on platform limits, but posting order matters: the continuation comment must be the immediately next comment, so build posts the parts back-to-back and other actors must not comment between them. Validation gains continuation-joining in comment scanning, and the build skill documents the split rule. Truncated receipts remain malformed by construction.
