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

A receipt that exceeds the GitHub comment cap SHALL continue into the immediately following comment rather than truncate or leave the issue. The ordinary split SHALL fall on a field boundary — after a `Pre-evidence scope:` or `Post-evidence scope:` line — and every full comment SHALL end with the literal line `Receipt continues in next comment (GitHub comment size limit).` Chaining across more than two comments is allowed.

When one pre or post raw-output block is itself too large for one comment, the output position SHALL use explicit chunk records. Each record is exactly the fields `Raw output chunk:`, `Output: pre|post`, `Chunk: <n>/<total>`, `Unit occurrence: <positive integer>`, `Unit heading: <exact heading>`, and `unit digest: <digest>`, followed by one closed fenced payload. The first record occupies the output position in the receipt; each later record is in the immediately following comment and repeats the identity. Chunk numbers SHALL be consecutive from 1 through one fixed total, and validation SHALL reconstruct the raw output by concatenating payloads in order without inserted bytes. Every non-final chunk comment uses the literal continuation marker. Fenced output SHALL remain unedited; choose a delimiter longer than any delimiter line in its payload. Marker text inside a fence is raw output, never a marker.

`litespec validate` SHALL join a comment ending in the continuation marker with the literal immediately following comment before scanning, preserving comment indices and error numbering. A blank or other intervening comment interrupts the chain; it MUST NOT be skipped. Dangling, interrupted, duplicated, misordered, or wrong-identity continuations SHALL surface as visible incomplete-receipt errors, never as a silent pass.

## Consequences

Verbatim evidence survives real-world output sizes, and review still reads one logical receipt. Workers no longer stall on platform limits, but posting order matters: the continuation comment must be the immediately next comment, so build posts the parts back-to-back and other actors must not comment between them. Validation gains continuation-joining in comment scanning, and the build skill documents the split rule. Truncated receipts remain malformed by construction.
