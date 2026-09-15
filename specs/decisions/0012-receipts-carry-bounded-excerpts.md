---
spine: true
---

# Receipts Carry Bounded Excerpts, Not Logs

## Status

accepted

## Context

Little-goblin issue #69 accumulated ~670KB across 18 comments on one queue issue; roughly 98% was verbatim test-runner output chunked into 40-58KB pieces by the decision-0008 continuation machinery. The load-bearing signal in each receipt — identity, Verify, commits, exit statuses, scope lines — is a couple of KB. Nothing consumes the archive: review replays Verify independently in detached worktrees at pre, post, and HEAD, and the full outputs already existed as `--pre-out`/`--post-out` files at assembly time. The issue comment surface paid all of the bulk and proved nothing extra.

Decision 0008 chose verbatim continuation because truncation "breaks the verbatim property review relies on when diffing posted output against replayed Verify." That property turned out to be the wrong invariant to buy at unlimited size: replay is the proof; the posted output is context. The forks were (a) keep verbatim, (b) filter lines semantically, (c) externalize logs to gists or a content-addressed store, (d) bound what the receipt carries. Verbatim at unlimited size is the archive problem itself. Semantic filtering encodes test-reporter format heuristics that rot with every reporter. Externalizing contradicts the ruling that GitHub metadata lives in comments and re-imports the mutable/fetchable dependency problems 0009 already rejected for provenance. Bounding keeps the receipt on the issue, honest about what it is.

## Decision

Newly assembled receipts declare `Protocol: evidence/v2` and carry bounded output excerpts. Each run position records the full output's byte length and SHA-256 beside one fence holding either the verbatim complete output — when it fits the excerpt budget — or a deterministic head excerpt, one literal elision marker naming the exact elided byte count, and a deterministic tail excerpt. One v2 receipt fits one comment within a fixed byte budget; it never continues across comments and never uses the chunk form. Assembly refuses visibly when identity and status fields alone exceed the budget. Excerpt budgets are fixed constants in the binary, never configuration.

The v2 Receipt ID hashes the bounded canonical fields — protocol, digest algorithm, recovery reference, routing identity, Verify, unit digest, run fields, per-run byte count, full-output SHA-256, excerpt text, and scope lines. The full output participates only through its length and hash, so the same run yields the same ID and a different run yields a different one, without carrying the log. Recovery references may name any retained receipt ID, including v1.

Receipts declaring `evidence/v1` or the legacy unversioned shape continue to validate under their retained parsers, including the decision-0008 continuation and chunk forms, for as long as records can name them (decision 0009). The verbatim ruling of 0008 is superseded for newly assembled receipts only. Validation dispatches on the declared protocol: v2 receipts must be arithmetically consistent (head plus elided plus tail equals the declared byte count), unelided fences must hash to the declared SHA-256, and a v2 receipt that continues, chunks, or exceeds the budget is a visible error. Review cross-checks replay by comparing the replayed run's byte count and SHA-256 against the declared fields, not by diffing full posted logs. Red-green discrimination still comes from replay, never from the excerpt.

## Consequences

Routine receipts stop spawning multi-comment chains; a queue issue stays readable as a queue. The elided middle of a red run can hide failing-test names from a reader who refuses to replay — accepted, because build judged the pre run live and review replays it anyway. The excerpt is context, not proof, and says so. Receipt identity becomes format-stable against output growth: appending passing tests to a suite no longer changes what the issue must hold. Every producer and consumer of the receipt grammar — assembly, validation, continuation joining, build and review skills — updates together, as 0009 requires. Local retention of full outputs is not promised by the format; anything needing the log replays the command.
