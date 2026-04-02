# spec-format

## ADDED Requirements

### Requirement: Requirements Section Wrapper

Canonical spec files MUST use a `## Requirements` H2 section header to contain all `### Requirement:` blocks. The parser SHALL reject specs that place `### Requirement:` headings outside of a `## Requirements` section. `## Requirements` SHALL appear after an optional `## Purpose` section and before any `### Requirement:` blocks.

#### Scenario: Valid spec with Requirements wrapper

- **WHEN** a canonical spec contains `## Requirements` followed by `### Requirement:` blocks
- **THEN** the parser successfully extracts all requirements

#### Scenario: Spec missing Requirements wrapper

- **WHEN** a canonical spec contains `### Requirement:` blocks without a preceding `## Requirements` header
- **THEN** the parser returns an error indicating `## Requirements` section is missing

#### Scenario: Requirement heading before Requirements wrapper

- **WHEN** a canonical spec contains `### Requirement:` before encountering `## Requirements`
- **THEN** the parser returns an error indicating requirements must be inside the `## Requirements` section

### Requirement: Optional Purpose Section

Canonical spec files MAY include a `## Purpose` H2 section before the `## Requirements` section. `## Purpose` is the sole permitted H2 section between the H1 capability heading and `## Requirements`. Any other H2 section before `## Requirements` SHALL cause a parse error.

#### Scenario: Spec with Purpose and Requirements

- **WHEN** a canonical spec contains `## Purpose` followed by `## Requirements` followed by requirements
- **THEN** the parser successfully extracts all requirements and captures the purpose text

#### Scenario: Spec with unsupported H2 before Requirements

- **WHEN** a canonical spec contains a H2 section other than `## Purpose` before `## Requirements`
- **THEN** the parser returns an error indicating only `## Purpose` is permitted before `## Requirements`

#### Scenario: Spec with Requirements only, no Purpose

- **WHEN** a canonical spec contains `## Requirements` with no preceding `## Purpose`
- **THEN** the parser successfully extracts all requirements with no error

### Requirement: Serializer Emits Requirements Wrapper

The `SerializeSpec` function MUST emit the `## Requirements` H2 section header before all requirement blocks. If the spec has a non-empty `Purpose` field, the serializer SHALL emit `## Purpose` before `## Requirements`.

#### Scenario: Serialize spec with purpose

- **WHEN** `SerializeSpec` is called on a spec with purpose text and requirements
- **THEN** the output contains `## Purpose` followed by the purpose text followed by `## Requirements` followed by requirement blocks

#### Scenario: Serialize spec without purpose

- **WHEN** `SerializeSpec` is called on a spec with no purpose text
- **THEN** the output contains `## Requirements` followed by requirement blocks with no `## Purpose` section

#### Scenario: Round-trip preserves structure

- **WHEN** a spec is parsed via `ParseMainSpec` and then serialized via `SerializeSpec`
- **THEN** re-parsing the serialized output produces identical capability, purpose, and requirements
