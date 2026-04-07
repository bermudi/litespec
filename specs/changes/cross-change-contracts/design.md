# Cross-Change Contracts — Design

## Architecture

Two independent layers that reinforce each other:

```
                    ┌──────────────┐
                    │  Delta Spec  │
                    │  (glossary)  │
                    └──────┬───────┘
                           │ parsed by
                    ┌──────▼───────┐
                    │   validate   │──── structural checks only
                    └──────┬───────┘     (duplicates, conflicts,
                           │              dangling deltas, loads
                           │              dependency terms)
                    ┌──────▼───────┐
                    │  review skill│──── semantic cross-referencing
                    └──────────────┘     (name drift, affix variants)
```

**Layer 1 — Glossary (structural):** Extends the spec format with an optional `## Glossary` section. The parser, serializer, delta merge, and validator all become glossary-aware. `validate` performs deterministic checks (duplicates, conflicting operations, dangling deltas) and loads dependency glossary terms as structured data — but does no semantic matching.

**Layer 2 — Review skill (semantic):** The review skill reads dependency specs/design/glossary when reviewing a change with `dependsOn`. The AI performs the fuzzy semantic matching that code can't do well — catching `EventHandler` vs `Events`, `OutputEvent` vs `Event`, `*RPCAgent` vs `RPCAgent` — and reports name drift as WARNING findings. This is a skill prompt change — no CLI code.

## Decisions

### Glossary lives in specs, not a separate artifact

**Chosen:** Glossary entries are a section within `spec.md` (canonical) and delta `spec.md` files.

**Why:** A glossary is metadata about a spec's requirements — what terms they export for downstream consumers. Adding a separate `glossary.yaml` would be a new artifact type, new dependency edges, new status tracking. Embedding it in the spec keeps the artifact count at four and follows the existing pattern (Purpose is also an optional section in spec.md).

**Tradeoff:** Glossary entries are per-capability, not per-change. A change that exports terms across multiple capabilities will have glossary entries scattered across multiple spec files. This matches the existing model — specs are organized by capability.

### Near-miss detection is the review skill's job, not validate's

**Chosen:** `validate` performs only deterministic structural checks. The review skill performs semantic name matching.

**Why:** Building normalization rules in Go (suffix stripping, plural handling, camelCase sub-token matching) is a maintenance trap. Every naming convention invents new affixes. The AI is better at "EventHandler probably means Events" than any `NormalizeTerm` function — it understands semantics, handles conventions nobody's written down, and won't false-positive on "Prevent" containing "Event."

**Example:** Change A exports glossary term "Events" (a method). Change B's design.md references "EventHandler". The review skill flags: *WARNING: 'EventHandler' in B may refer to 'Events' from A — verify naming consistency.* Similarly, "OutputEvent" vs "Event" and "*RPCAgent" vs "RPCAgent" are caught by the AI's understanding of naming patterns, not by regex.

**Tradeoff:** Semantic matching only happens during review, not on every `validate` run. Acceptable — `validate` runs frequently for quick structural feedback; review runs when you want depth.

### Review skill reads dependency artifacts, not glossary data from CLI

**Chosen:** The review skill prompt tells the AI to read dependency change directories directly. No new CLI command to surface dependency terms.

**Why:** The AI already reads files. Adding a `litespec glossary <change>` command would be convenient but premature — we don't know if the review skill's cross-referencing will work well enough to justify dedicated tooling. Start with the skill prompt, add CLI support later if needed.

### Glossary format uses bold term names

**Chosen:** `- **TermName**: definition text` format.

**Why:** Parseable with a simple regex (`^\s*-\s*\*\*(.+?)\*\*:\s*(.+)$`), visually readable in markdown, consistent with how terms are already referenced in spec prose. The bold makes terms grep-able across specs.

## File Changes

### `internal/types.go`

- Add `GlossaryEntry` struct with `Name` and `Definition` string fields
- Add `Glossary []GlossaryEntry` field to `Spec` struct
- Add `DeltaGlossaryEntry` struct with `Operation DeltaOperation`, `Name`, and `Definition` fields
- Add `GlossaryEntries []DeltaGlossaryEntry` field to `DeltaSpec` struct

### `internal/delta.go`

- **`ParseGlossaryEntries(lines []string) ([]GlossaryEntry, error)`** — shared parsing logic for both canonical and delta specs. Parses `- **Name**: definition` format, detects duplicates.
- **`ParseMainSpec`**: Add `stateGlossary` after `stateRequirements`. When `## Glossary` is encountered in `stateRequirements`, transition to `stateGlossory`. Parse entries via `ParseGlossaryEntries`. Any other H2 in `stateGlossory` is an error.
- **`ParseDeltaSpec`**: Recognize `## ADDED Glossary`, `## MODIFIED Glossary`, `## REMOVED Glossary` as valid section headers (after requirement sections). Parse entries within each section.
- **`SerializeSpec`**: After emitting requirements, emit `## Glossary` with entries if `spec.Glossary` is non-empty.
- **`MergeDelta`**: After processing requirements, process glossary entries: ADDED appends, MODIFIED replaces definition, REMOVED deletes. Same error semantics as requirements (duplicate add, missing target).

### `internal/validate.go`

- **`ValidateChange`**: Add glossary-specific validation within delta specs: duplicate terms, conflicting operations, dangling deltas against canonical glossary.
- **Dependency glossary loading**: After existing dependency validation, if the change has `dependsOn`, for each dependency load its delta specs (active) or canonical specs (archived), extract glossary terms, union across spec files (dedup by name, warn on conflicting definitions), and include the result in `ValidationResult` for downstream consumers (e.g., the review skill).
- Add `DependencyGlossary map[string][]GlossaryEntry` field to `ValidationResult` to carry loaded dependency terms.

### `.agents/skills/litespec-review/SKILL.md`

- **Setup step**: After reading `litespec status`, check `.litespec.yaml` for `dependsOn`. If present, list dependency changes and read their artifacts (specs, design, glossary).
- **All sections**: Add a "Cross-Change Consistency" dimension. When the change has `dependsOn`, cross-reference interface names, method signatures, config keys, type names, and glossary terms against dependency specs. Name drift (e.g., `EventHandler` when the dependency exports `Events`, `*RPCAgent` vs `RPCAgent`, `OutputEvent` vs `Event`) SHALL be reported as WARNING findings — not CRITICAL. The AI performs semantic matching that code can't do well.
- **Section C (Pre-Archive Review)**: Inherits cross-change checks from A and B — no additional changes needed.
