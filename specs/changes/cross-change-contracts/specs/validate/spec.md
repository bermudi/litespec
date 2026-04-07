# validate

## ADDED Requirements

### Requirement: Dependency Glossary Availability

When validating a change that declares `dependsOn`, the `ValidateChange` function MUST load glossary terms from each dependency's specs (delta specs if active, canonical specs if archived) and include them in the validation result. Terms from multiple spec files within a dependency SHALL be unioned (deduplicated by name, with a warning on conflicting definitions for the same term). This structured data enables downstream consumers (e.g., the review skill) to perform semantic cross-referencing. Terms not referenced in the current change's prose SHALL NOT produce warnings — semantic matching is the review skill's responsibility.

#### Scenario: Active dependency glossary loaded

- **WHEN** change B depends on active change A, and A's delta specs export glossary terms "RPCAgent" and "Event"
- **THEN** validation loads those terms and includes them in the result for downstream use

#### Scenario: Archived dependency glossary loaded

- **WHEN** change B depends on archived change A, and A's glossary terms are in the canonical spec
- **THEN** validation reads glossary terms from the canonical spec and includes them in the result

#### Scenario: Dependency has no glossary

- **WHEN** change B depends on change A and A's specs contain no glossary section
- **THEN** no glossary terms are loaded for that dependency and no error is emitted

#### Scenario: Multi-capability glossary union

- **WHEN** change B depends on change A, and A has glossary entries in three separate spec files
- **THEN** validation unions all entries across those files, deduplicates by name, and warns on conflicting definitions for the same term name

### Requirement: Glossary Duplicate Detection in Delta

The `ValidateChange` function MUST detect when the same glossary term name appears in more than one operation section within a single delta spec. A term that is ADDED then MODIFIED, or MODIFIED then REMOVED, in the same delta is an error — operations must be sequenced across separate delta files or reconciled into a single operation.

#### Scenario: Duplicate glossary term in single delta

- **WHEN** a delta spec ADDS glossary term "RPCAgent" twice
- **THEN** validation reports an error: duplicate glossary term "RPCAgent"

#### Scenario: Conflicting glossary operations

- **WHEN** a delta spec ADDS glossary term "RPCAgent" and also MODIFIES glossary term "RPCAgent"
- **THEN** validation reports an error about conflicting glossary operations on "RPCAgent"

### Requirement: Glossary Dangling Delta Detection

MODIFIED and REMOVED glossary operations SHALL be checked against the canonical spec's glossary. If the target term does not exist in the canonical glossary, an error SHALL be emitted — the same dangling delta detection already applied to requirements.

#### Scenario: Remove nonexistent glossary term

- **WHEN** a delta REMOVES glossary term "RPCAgent" but the canonical spec has no such term
- **THEN** validation reports an error: glossary term "RPCAgent" not found in canonical spec

#### Scenario: Modify nonexistent glossary term

- **WHEN** a delta MODIFIES glossary term "RPCAgent" but the canonical spec has no such term
- **THEN** validation reports an error: glossary term "RPCAgent" not found in canonical spec
