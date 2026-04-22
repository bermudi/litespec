# validate

## ADDED Requirements

### Requirement: Decision Validation

The `litespec validate` command SHALL validate decisions when `--decisions` is specified or when `--all` is used. Validation SHALL parse each file in `specs/decisions/`, check required section presence, verify status is a valid enum, detect duplicate numbers, detect duplicate slugs, and verify all supersede pointers resolve. `ValidateAll` SHALL include decisions in its scope. A `ValidateDecision(root, slug)` function SHALL exist for validating a single decision by slug.

#### Scenario: Validate all includes decisions

- **WHEN** `litespec validate --all` is run and `specs/decisions/` contains malformed files
- **THEN** errors for the decisions are included in the combined result

#### Scenario: Validate only decisions

- **WHEN** `litespec validate --decisions` is run
- **THEN** only decision files are validated; changes and specs are skipped

#### Scenario: Duplicate number detected

- **WHEN** two files `0003-foo.md` and `0003-bar.md` both exist in `specs/decisions/`
- **THEN** validation reports an error identifying the duplicate number

#### Scenario: Positional name resolves to decision

- **WHEN** `litespec validate 0003-foo` is run and `0003-foo` matches a decision slug (and no change or spec)
- **THEN** only that decision is validated

#### Scenario: Ambiguous name across decision and change

- **WHEN** `litespec validate foo` is run and `foo` is both a change and a decision slug suffix
- **THEN** validation reports an ambiguity error suggesting `--type decision`

### Requirement: Type Disambiguation Includes Decision

The `--type` flag accepted by `litespec validate` SHALL accept `decision` as a valid value in addition to `change` and `spec`. When `--type decision` is supplied, the positional name SHALL be resolved against decision slugs (matching either the full `NNNN-slug` name or the slug portion alone).

#### Scenario: Explicit decision type

- **WHEN** `litespec validate beta-tools --type decision` is run
- **THEN** the decision whose slug matches `beta-tools` is validated
