# review

## ADDED Requirements

### Requirement: Cross-Change Dependency Awareness

When reviewing a change that declares `dependsOn`, the review skill MUST read the dependency's specs and design artifacts in addition to the change's own artifacts. The review SHALL cross-reference interface names, method signatures, config keys, type names, and glossary terms across the dependency boundary. Mismatches between the reviewed change's references and the dependency's exported terms SHALL be reported as CRITICAL findings. This applies to all three review modes (artifact, implementation, pre-archive).

#### Scenario: Artifact review with dependency

- **WHEN** artifact review is invoked on a change that depends on another change
- **THEN** the review reads the dependency's specs and design, and cross-references shared terms

#### Scenario: Mismatched interface name across dependency

- **WHEN** change B depends on change A, A's spec defines an interface named "RPCAgent", and B's spec references "*RPCAgent" (pointer variant)
- **THEN** the review reports a CRITICAL finding about the name mismatch

#### Scenario: Consistent references across dependency

- **WHEN** change B depends on change A, and all of B's references to A's interfaces match exactly
- **THEN** no cross-change findings are reported

#### Scenario: Dependency has no specs

- **WHEN** change B depends on archived change A and A's specs have been merged into canon
- **THEN** the review reads the canonical specs for cross-referencing instead
