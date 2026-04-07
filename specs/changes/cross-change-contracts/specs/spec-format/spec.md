# spec-format

## ADDED Requirements

### Requirement: Glossary Section in Canonical Specs

Canonical spec files MAY include a `## Glossary` H2 section after `## Requirements`. The glossary SHALL contain term entries formatted as markdown list items: `- **TermName**: definition`. Term names MUST be non-empty and unique within a single spec. The parser SHALL extract glossary entries into a `[]GlossaryEntry` field on the `Spec` struct. If no `## Glossary` section is present, the field SHALL be nil. No other H2 sections are permitted after `## Requirements`.

#### Scenario: Spec with glossary

- **WHEN** a canonical spec contains `## Glossary` after `## Requirements` with entries `- **RPCAgent**: the agent interface` and `- **Event**: output event type`
- **THEN** the parser extracts two `GlossaryEntry` values with names "RPCAgent" and "Event" and their definitions

#### Scenario: Spec without glossary

- **WHEN** a canonical spec has no `## Glossary` section
- **THEN** the parsed `Spec` has a nil `Glossary` field and no error is returned

#### Scenario: Duplicate glossary term

- **WHEN** a canonical spec contains two glossary entries with the same term name
- **THEN** the parser returns an error indicating the duplicate term

#### Scenario: Invalid H2 after Requirements

- **WHEN** a canonical spec contains an H2 section other than `## Glossary` after `## Requirements`
- **THEN** the parser returns an error

### Requirement: Glossary Section in Delta Specs

Delta spec files MAY include `## ADDED Glossary`, `## MODIFIED Glossary`, and `## REMOVED Glossary` sections. Glossary operation sections SHALL appear after requirement operation sections (e.g., `## ADDED Requirements` comes before `## ADDED Glossary`). ADDED and MODIFIED entries SHALL use `- **TermName**: definition` format. REMOVED entries SHALL use `- **TermName**` format (name only, definition omitted). ADDED entries append to the canonical glossary. MODIFIED entries replace the definition of an existing term. REMOVED entries delete a term by name. The delta parser SHALL extract these into a `[]DeltaGlossaryEntry` field on the `DeltaSpec` struct.

#### Scenario: Delta adds glossary terms

- **WHEN** a delta spec contains `## ADDED Glossary` with `- **RPCAgent**: the agent interface`
- **THEN** the parser extracts a `DeltaGlossaryEntry` with operation ADDED, name "RPCAgent", and the definition

#### Scenario: Delta removes glossary term

- **WHEN** a delta spec contains `## REMOVED Glossary` with `- **RPCAgent**`
- **THEN** the parser extracts a `DeltaGlossaryEntry` with operation REMOVED, name "RPCAgent", and empty definition

#### Scenario: Delta with no glossary sections

- **WHEN** a delta spec has no glossary sections
- **THEN** the parsed `DeltaSpec` has a nil `GlossaryEntries` field

### Requirement: Serializer Emits Glossary

The `SerializeSpec` function MUST emit the `## Glossary` section after `## Requirements` if the spec has a non-empty `Glossary` field. Each entry SHALL be serialized as `- **TermName**: definition` on its own line.

#### Scenario: Serialize spec with glossary

- **WHEN** `SerializeSpec` is called on a spec with glossary entries
- **THEN** the output contains `## Glossary` after requirements, with each entry as a list item

#### Scenario: Serialize spec without glossary

- **WHEN** `SerializeSpec` is called on a spec with nil glossary
- **THEN** no `## Glossary` section appears in the output

#### Scenario: Round-trip preserves glossary

- **WHEN** a spec with glossary is parsed, serialized, and re-parsed
- **THEN** the glossary entries are identical

### Requirement: Delta Merge Handles Glossary

The `MergeDelta` function MUST apply glossary delta operations during archive merge. ADDED entries SHALL be appended to the canonical spec's glossary. MODIFIED entries SHALL replace the definition of a matching term. REMOVED entries SHALL delete the matching term. Duplicate ADDED terms (already present in canon) SHALL produce an error. REMOVED/MODIFIED terms not found in canon SHALL produce an error.

#### Scenario: Merge adds glossary term

- **WHEN** a delta ADDS glossary term "RPCAgent" and canon has no such term
- **THEN** the merged spec includes the "RPCAgent" glossary entry

#### Scenario: Merge removes glossary term

- **WHEN** a delta REMOVES glossary term "RPCAgent" and canon has "RPCAgent"
- **THEN** the merged spec no longer includes the "RPCAgent" glossary entry

#### Scenario: Merge modifies glossary term

- **WHEN** a delta MODIFIES glossary term "RPCAgent" with a new definition and canon has "RPCAgent"
- **THEN** the merged spec has "RPCAgent" with the new definition

#### Scenario: Add duplicate glossary term fails

- **WHEN** a delta ADDS glossary term "RPCAgent" and canon already has "RPCAgent"
- **THEN** the merge returns an error
