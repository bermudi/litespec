---
spine: true
---

# Repeated Rebuilds Force Contract Reshaping

## Status

accepted

## Context

Adversarial review can repeatedly discover different failures inside one broad unit while each rebuild adds only the cited regression test. Every receipt may be honest and reproducible, yet the serial review-rebuild loop still signals that the unit boundary, its scenarios, or the implementation architecture is wrong. Unlimited rebuild requests preserve a fixed contract after the evidence has shown that its shape is not guiding complete work.

The alternatives were to allow unlimited rebuilds, stop at the second finding, or let build reject a third request. Unlimited rebuilds repeat the failure. Stopping at the second finding permits only one corrective attempt. Deferring the decision to build records a route review already knows is invalid.

## Decision

A unit SHALL complete at most two review-requested rebuild cycles against one current contract digest. A cycle begins with one or more literal `Rebuild request:` records for that unit and completes when a later identity-bearing receipt resolves them. Evidence resolving an amendment is not a review-requested cycle. On the next unit-breaking review finding, review SHALL record one blocking re-plan marker instead of another rebuild request.

The marker is bound to the unit's current digest and is resolved only by a plan-authored amendment from that digest. Plan must reshape the contract, which may narrow or rename the existing unit and append additional units. The amendment remains unresolved until build supplies fresh evidence under the new contract; that evidence leaves the new digest's rebuild count at zero. A contract amendment resets the rebuild count because the reshaped digest is a new target. A marker before two completed cycles or a duplicate unresolved marker for the same digest is malformed routing metadata.

Units SHALL be shaped around one external boundary or one failure policy, not merely one broad demo. Their contract SHALL explicitly map each `Done means` clause to a named test scenario. Filesystem, process, and network boundaries SHALL account for timeout, cleanup, non-ENOENT errors, concurrency, and optional configured dependencies with a mapped scenario or a concrete N/A reason.

Review SHALL preserve an append-only coverage record keyed to the reviewed HEAD and unit identity. A fresh reviewer constructs an independent risk inventory before reading earlier coverage, then uses prior records to expand—not prove—what it exercises. Existing pre/post/HEAD evidence replay remains unchanged.

## Consequences

A third serial miss cannot return directly to build. Plan must change the contract shape, and validation must understand rebuild cycles, re-plan markers, amendment resolution, scenario mappings, and boundary-risk accounting for both GitHub and local queues. Review gains a small append-only metadata write, but findings remain ephemeral and prior coverage remains advisory. Queue contracts become more explicit, while `litespec validate` continues to validate structure rather than implementation semantics and says so in successful output.
