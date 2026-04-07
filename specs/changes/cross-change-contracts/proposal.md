# Cross-Change Contracts

## Motivation

When change B declares `dependsOn: [A]`, litespec validates the dependency exists and checks for cycles and overlaps — but it never verifies that B's references to A's interfaces are consistent. Both changes describe shared concepts (interface names, method signatures, config keys, type names) in prose, and prose drifts.

In a recent review of two dependent changes, 6 of 10 findings were cross-change consistency failures: `EventHandler` vs `Events()`, `OutputEvent` vs `Event`, `*RPCAgent` vs `RPCAgent`. These are exactly the kind of errors a tool should catch, but litespec's dependency system is a guard rail (ordering, cycles, overlaps), not a contract verifier.

Two complementary fixes:

1. **Glossary** — a lightweight term registry in each change that exports named terms. When change B depends on change A, `validate` checks that B's term references resolve against A's exports. No IDL, no schema — just name agreement.

2. **Cross-change-aware review skill** — when reviewing a change with dependencies, the review skill reads the dependency's specs and design too, cross-referencing interface names, method signatures, and type names across the boundary.

## Scope

### Glossary (structural layer)

- New optional `## Glossary` section in delta spec files, after requirements
- Each glossary entry is a term with a short definition: `- **TermName**: description`
- `validate` parses glossary entries and performs structural checks (duplicates, conflicting operations, dangling deltas)
- When change B `dependsOn` change A, `validate` loads A's glossary terms (from delta specs if active, canonical specs if archived) and makes them available as structured data
- Glossary terms from archived changes persist in the canonical spec after merge
- No semantic matching in `validate` — no near-miss detection, no normalization, no fuzzy comparison

### Review skill enhancement (semantic layer)

- When a change has `dependsOn`, the review skill reads dependency specs and design artifacts
- Cross-references interface names, method signatures, config keys, type names, and glossary terms across the boundary
- The AI performs semantic matching (e.g., `EventHandler` vs `Events`, `*RPCAgent` vs `RPCAgent`, `OutputEvent` vs `Event`)
- Reports name drift as WARNING findings in the review report

### Canonical spec format

- `## Glossary` becomes a permitted optional H2 section after `## Requirements` in canonical specs
- `SerializeSpec` emits glossary if present
- Delta merge handles glossary entries with the same ADDED/MODIFIED/REMOVED operations as requirements

## Non-Goals

- **Not an IDL or type system.** Glossary entries are prose terms with definitions, not typed schemas. We don't validate method signatures structurally — just name agreement.
- **No automatic term extraction.** Authors write glossary entries explicitly. No NLP or heuristic scanning of prose to discover terms.
- **No cross-change validation without `dependsOn`.** Glossary checks only trigger between changes with declared dependency edges. Unrelated changes are not cross-checked.
- **No near-miss detection in `validate`.** Semantic name matching (EventHandler vs Events, OutputEvent vs Event) is the review skill's job, not the validator's. Building normalization rules in Go is a maintenance trap — the AI handles this better.
- **No glossary in the `view` command dashboard.** Term counts and glossary display in `view` are out of scope. Can be added later if valuable.
- **No breaking change to existing specs.** Glossary is optional. Existing specs without glossary sections remain valid.
