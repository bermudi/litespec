# Cross-Change Contracts — Tasks

## Phase 1: Types and Parsing

- [ ] Add `GlossaryEntry` struct (Name, Definition) and `DeltaGlossaryEntry` struct (Operation, Name, Definition) to `internal/types.go`
- [ ] Add `Glossary []GlossaryEntry` field to `Spec` struct
- [ ] Add `GlossaryEntries []DeltaGlossaryEntry` field to `DeltaSpec` struct
- [ ] Add `ParseGlossaryEntries(lines []string) ([]GlossaryEntry, error)` to `internal/delta.go` — parses `- **Name**: definition` format, detects duplicates
- [ ] Extend `ParseMainSpec` in `internal/delta.go` to recognize `## Glossary` after `## Requirements` (add `stateGlossary` state), parse entries via `ParseGlossaryEntries`, reject other H2 sections after Glossary
- [ ] Extend `ParseDeltaSpec` in `internal/delta.go` to recognize `## ADDED Glossary`, `## MODIFIED Glossary`, `## REMOVED Glossary` section headers (after requirement sections) and parse entries within each
- [ ] Extend `SerializeSpec` in `internal/delta.go` to emit `## Glossary` after requirements when `spec.Glossary` is non-empty
- [ ] Write tests in `internal/delta_test.go` for `ParseGlossaryEntries`: valid entries, empty input, duplicate terms, malformed entries, canonical spec with glossary, delta spec with glossary operations, round-trip serialization

## Phase 2: Delta Merge

- [ ] Extend `MergeDelta` in `internal/delta.go` to process glossary delta entries: ADDED appends, MODIFIED replaces definition, REMOVED deletes term
- [ ] Add merge error handling: duplicate ADDED term, missing REMOVED/MODIFIED target
- [ ] Write tests in `internal/delta_merge_test.go` for glossary merge: add, modify, remove, duplicate add error, missing target error

## Phase 3: Validation

- [ ] Add glossary validation to `ValidateChange`: duplicate terms within a delta, conflicting operations on same term, dangling deltas against canonical glossary
- [ ] Add `DependencyGlossary map[string][]GlossaryEntry` field to `ValidationResult` in `internal/types.go`
- [ ] Add dependency glossary loading to `ValidateChange`: when `dependsOn` is present, load dependency glossary terms (from delta specs if active, canonical specs if archived), union across spec files, warn on conflicting definitions (same term name with different definitions — "different" means not byte-for-byte identical after trimming whitespace), include in `ValidationResult.DependencyGlossary`
- [ ] Write tests in `internal/validate_test.go` for glossary validation: duplicates, conflicts, dangling deltas, dependency glossary loading, multi-capability union, conflicting definitions warning

## Phase 4: Review Skill

- [ ] Update `.agents/skills/litespec-review/SKILL.md` setup step: after reading status, check `.litespec.yaml` for `dependsOn` and read dependency artifacts
- [ ] Update `.agents/skills/litespec-review/SKILL.md` all sections: add "Cross-Change Consistency" dimension — when `dependsOn` exists, cross-reference terms and report name drift as WARNING
- [ ] Manually verify the updated review skill by reviewing the cross-change-contracts change itself — confirm cross-change consistency checks appear in the output
