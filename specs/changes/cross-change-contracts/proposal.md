# Cross-Change Contracts

## Motivation

When change B declares `dependsOn: [A]`, litespec validates the dependency exists and checks for cycles and overlaps — but it never verifies that B's references to A's interfaces are consistent. Both changes describe shared concepts (interface names, method signatures, config keys, type names) in prose, and prose drifts.

In a recent review of two dependent changes, 6 of 10 findings were cross-change consistency failures: `EventHandler` vs `Events()`, `OutputEvent` vs `Event`, `*RPCAgent` vs `RPCAgent`. These are exactly the kind of errors a tool should catch, but litespec's dependency system is a guard rail (ordering, cycles, overlaps), not a contract verifier.

A cross-change-aware review skill addresses this: when reviewing a change with dependencies, the review skill reads the dependency's specs and design too, cross-referencing interface names, method signatures, and type names across the boundary.

## Scope

### Review skill enhancement (semantic layer)

- When a change has `dependsOn`, the review skill reads dependency specs and design artifacts
- Cross-references interface names, method signatures, config keys, type names across the boundary
- May consult `specs/glossary.md` as supplementary terminology context alongside dependency specs/design
- The AI performs semantic matching (e.g., `EventHandler` vs `Events`, `*RPCAgent` vs `RPCAgent`, `OutputEvent` vs `Event`)
- Reports name drift as WARNING findings in the review report

## Non-Goals

- **No glossary structural layer.** Per-spec glossary sections, glossary parsing in Go, glossary delta operations, glossary merge logic, and glossary validation are handled by the `ubiquitous-language` change instead — as a project-wide `specs/glossary.md` file, not per-spec sections.
- **Not an IDL or type system.** Cross-change checks are semantic, not structural. We don't validate method signatures in Go code — the AI catches naming inconsistencies.
- **No automatic term extraction.** Authors write glossary entries explicitly. No NLP or heuristic scanning.
- **No cross-change validation without `dependsOn`.** Cross-referencing only triggers between changes with declared dependency edges.
- **No near-miss detection in `validate`.** Semantic name matching is the review skill's job, not the validator's.
