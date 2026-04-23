# Cross-Change Contracts — Design

## Architecture

```
                    ┌──────────────────┐
                    │  specs/glossary.md│ ← project-wide ubiquitous language
                    └──────┬───────────┘
                           │ read by
                    ┌──────▼───────┐
                    │  review skill│──── semantic cross-referencing
                    └──────────────┘     (name drift, affix variants)
```

**Review skill (semantic):** The review skill reads dependency specs/design when reviewing a change with `dependsOn`. It may also consult `specs/glossary.md` as supplementary terminology context. The AI performs the fuzzy semantic matching that code can't do well — catching `EventHandler` vs `Events`, `OutputEvent` vs `Event`, `*RPCAgent` vs `RPCAgent` — and reports name drift as WARNING findings. This is a skill prompt change — no CLI code.

## Decisions

### No structural glossary in specs

**Chosen:** The project-wide `specs/glossary.md` (from the `ubiquitous-language` change) replaces the per-spec glossary sections originally designed here.

**Why:** Maintaining glossary entries in two places (per-spec and project-wide) means the AI has to update both, which will fail often. One source of truth wins. The structural layer (Go parsing, delta operations, merge logic, validation) was all in service of per-spec glossaries — with a single project-wide file, none of that code is needed.

### Near-miss detection is the review skill's job, not validate's

**Chosen:** `validate` performs only deterministic structural checks. The review skill performs semantic name matching.

**Why:** Building normalization rules in Go (suffix stripping, plural handling, camelCase sub-token matching) is a maintenance trap. Every naming convention invents new affixes. The AI is better at "EventHandler probably means Events" than any `NormalizeTerm` function — it understands semantics, handles conventions nobody's written down, and won't false-positive on "Prevent" containing "Event."

**Example:** Change A defines method "Events". Change B's design.md references "EventHandler". The review skill flags: *WARNING: 'EventHandler' in B may refer to 'Events' from A — verify naming consistency.* Similarly, "OutputEvent" vs "Event" and "*RPCAgent" vs "RPCAgent" are caught by the AI's understanding of naming patterns.

**Tradeoff:** Semantic matching only happens during review, not on every `validate` run. Acceptable — `validate` runs frequently for quick structural feedback; review runs when you want depth.

### Review skill reads dependency artifacts directly, not glossary data from CLI

**Chosen:** The review skill prompt tells the AI to read dependency change directories directly. No new CLI command to surface dependency terms.

**Why:** The AI already reads files. The project glossary (`specs/glossary.md`) is a plain markdown file the AI can read directly. Adding a CLI command is premature — start with the skill prompt, add CLI support later if needed.

## File Changes

### `.agents/skills/litespec-review/SKILL.md`

- **Setup step**: After reading `litespec status`, check `.litespec.yaml` for `dependsOn`. If present, list dependency changes and read their artifacts (specs, design). Also read `specs/glossary.md` if it exists for supplementary terminology context.
- **All sections**: Add a "Cross-Change Consistency" dimension. When the change has `dependsOn`, cross-reference interface names, method signatures, config keys, type names against dependency specs. Name drift SHALL be reported as WARNING findings — not CRITICAL. The AI performs semantic matching that code can't do well.
- **Section C (Pre-Archive Review)**: Inherits cross-change checks from A and B — no additional changes needed.
