# validate

## MODIFIED Requirements

### Requirement: Single Spec Validation

A `ValidateSpec(root, name)` function MUST exist in the internal package that validates a single spec by name. It is the singular counterpart to the existing `ValidateSpecs(root)` which validates all specs. It SHALL read and parse the spec at `specs/canon/<name>/spec.md`, validate its structure and requirements, and return a `*ValidationResult`.

#### Scenario: Validate existing spec

- **WHEN** `ValidateSpec(root, "auth")` is called and the spec exists
- **THEN** a ValidationResult is returned reflecting the spec's validity

#### Scenario: Validate missing spec

- **WHEN** `ValidateSpec(root, "nonexistent")` is called
- **THEN** an error is returned indicating the spec was not found
