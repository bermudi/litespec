---
spine: true
---

# Version Evidence and Recover by Appending

## Status

accepted

## Context

The historical-receipt amendment-chain ruling fixed one class of moved-goalpost failure: a receipt whose contract digest is no longer current can remain valid when an append-only amendment chain witnesses the contract revision. The evidence record itself is still an unversioned Markdown shape, and `unit digest:` still names one implicitly selected canonicalization.

Little-goblin issues #52 and #53 exposed the remaining compatibility fork. Complete receipts could be judged with a later parser, later `Verify` interpretation, or later digest implementation, while the append-only GitHub comment and local metadata lanes offered no honest repair. The parser, digest command, build/review instructions, validate spec, continuation parser, and amendment scanner must agree on how an old record is interpreted; changing only one of those files recreates the failure.

A `Supersedes:`, quarantine, or `Authorized by:` line would not solve authorization. GitHub comment text does not authenticate its claimed actor or repository authority, so accepting such a line as a state-changing control would create a fake escape hatch. Stable record identity and deterministic recovery provenance are useful without granting that authority.

## Decision

New receipts use one explicit versioned grammar. Immediately after `Evidence:` they declare:

```text
Protocol: evidence/v1
Digest algorithm: unit-contract-sha256-v1
Receipt ID: receipt-sha256-v1:<64 lowercase hex>
Recovered from: <receipt ID>  (optional)
```

The fields are ordered before the exact Verify command and the existing red-green fields. `evidence/v1` is the first explicit receipt protocol; `unit-contract-sha256-v1` is the current length-prefixed canonical unit-contract SHA-256 algorithm. Future grammar or digest changes receive new identifiers. Validate dispatches by those identifiers and retains every supported historical parser and digest algorithm for as long as receipts can name it. Unknown identifiers and partial version metadata fail visibly; they never fall back to a different interpretation.

A Receipt ID is a content-derived, stable identifier rather than an authorization claim. `receipt-sha256-v1:` hashes the canonical logical receipt with its own ID omitted, including protocol, digest algorithm, optional recovery reference, routing identity, Verify command, unit digest, run fields, scope lines, and reconstructed raw outputs, while excluding storage wrappers, checkbox state, comment boundaries, and continuation bytes. The same identity is repeated on every raw-output chunk. An old receipt with no version fields remains on the preserved legacy parser and `unit-contract-sha256-v1`; validation derives a `legacy-receipt-sha256-v1:` identifier in memory and does not rewrite the record. `evidence/legacy-v0` is only the in-memory protocol label for those receipts and pairs with `unit-contract-sha256-v1`; the separately numbered `unit-contract-sha256-v0` is a distinct retained digest algorithm that applies only when a receipt declares it explicitly.

Receipt observations carry their declared digest algorithm. If different supported algorithms produce different digests for the same current contract, that algorithm-only change is not a contract amendment. Real contract revisions still require the existing digest-linked amendment chain. A later complete versioned receipt may append `Recovered from: <receipt ID>` as provenance and independently satisfy normal evidence or rebuild-request rules; it does not erase, suppress, or reclassify the earlier record.

This ruling extends decisions `0006-unit-contracts-are-amendable-only-by-plan` and `0008-oversized-receipts-continue-across-comments`. It does not introduce a content-addressed contract snapshot store, an issue-body transaction protocol, a closure attestation, or a state-changing supersession/quarantine control. Those require a later authenticated provenance/closure mechanism (for example a repository-owned controller or signed authority); an `Authorized by:` string alone is never sufficient.

## Consequences

Old complete unversioned receipts remain readable after parser and digest evolution, and new receipts tell both validate and future tools exactly which interpretation to use. Algorithm-only evolution no longer fabricates a contract amendment, while actual contract edits remain auditable through the existing chain. Append-only recovery can add a complete, independently checked record without mutating history.

Receipt IDs make continuation, duplicate, and recovery relationships deterministic, but they are not signatures and do not authorize a lifecycle decision. Truly malformed or unauthenticated historical records can still block validation until a later authenticated recovery/quarantine design exists. Every producer and consumer of the receipt grammar must update together; a supported historical implementation must not be removed while records can reference it.
