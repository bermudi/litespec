# validate

## Requirements

### Requirement: Spec and Decision Validation

The `litespec validate` command SHALL validate feature specs under `specs/<feature>/spec.md` and decisions under `specs/decisions/`. Each requirement body SHALL contain `SHALL` or `MUST` as a whole word outside code spans, and each requirement SHALL have at least one `#### Scenario:` with `WHEN`/`THEN` content. Decisions SHALL have `## Status` (one of `proposed`, `accepted`, `superseded`), `## Context`, `## Decision`, `## Consequences`; supersede pointers SHALL resolve to existing decisions. `ValidateAll` SHALL include specs, decisions, and skill templates in its scope. A `ValidateSpec(root, name)` and `ValidateDecision(root, slug)` function SHALL exist for single-artifact validation.

#### Scenario: Valid spec passes

- **WHEN** `litespec validate <feature>` is run on a spec with SHALL/MUST requirements and WHEN/THEN scenarios
- **THEN** validation succeeds with no errors

#### Scenario: Requirement without scenario fails

- **WHEN** `litespec validate <feature>` or `litespec validate --all` is run on a requirement with no `#### Scenario:`
- **THEN** validation reports an error and exits non-zero, regardless of `--strict`

#### Scenario: Decision missing a required section fails

- **WHEN** a decision file lacks `## Consequences`
- **THEN** validation reports an error identifying the missing section and file

#### Scenario: Validate all includes decisions

- **WHEN** `litespec validate --all` is run and `specs/decisions/` contains a malformed file
- **THEN** the decision error is included in the combined result

### Requirement: GH Issue Queue Validation

The `litespec validate` command SHALL fetch open GitHub issues labeled `litespec` via `gh issue list` and lint each issue body as a queue. A `##` section is treated as a unit only when its body contains a `Done means:` or `Verify:` line; prose sections such as `## Proposal`, `## Design`, `## Not doing`, `## Queue`, and `## Spec draft` SHALL be skipped. Each unit SHALL have a non-empty heading, a `Done means:` line, a `Verify:` line carrying either an inline backtick command on the same line or a fenced code block within the unit body, and a `- [ ]` or `- [x]` checkbox. Missing or malformed elements SHALL produce an error identifying the issue number and unit heading. Issues without the `litespec` label SHALL NOT be scanned. The `litespec` label is a hardcoded convention; no config file governs it.

#### Scenario: Well-formed queue passes

- **WHEN** `litespec validate` scans an open issue labeled `litespec` whose body has `## <outcome>` with `Done means:`, `Verify:` + fenced block, and `- [ ]`
- **THEN** validation succeeds for that issue

#### Scenario: Unit missing Done means fails

- **WHEN** a labeled issue has a `## <outcome>` unit without a `Done means:` line
- **THEN** validation reports an error naming the issue number and unit heading

#### Scenario: Verify without command or fenced block fails

- **WHEN** a unit's `Verify:` line has no inline backtick command and no fenced code block
- **THEN** validation reports an error identifying the unit

#### Scenario: Inline Verify command passes

- **WHEN** a unit's `Verify:` line carries an inline backtick command on the same line
- **THEN** validation accepts the unit without shell-linting the inline command

#### Scenario: Unlabeled issue is not scanned

- **WHEN** an open issue lacks the `litespec` label
- **THEN** validate does not attempt to parse its body

#### Scenario: Prose sections are skipped

- **WHEN** a labeled issue body contains `## Proposal`, `## Design`, and `## Spec draft` sections alongside `## <outcome>` units
- **THEN** validate lints only the units and reports no errors for the prose sections

### Requirement: Queue Unit Depends Validation

A unit MAY include a `Depends:` line listing comma-separated `##` heading references.
`validate` SHALL parse `Depends:` and check each reference matches a `##` heading in the same queue.
A dangling reference (heading not found) SHALL produce an error naming the unit and the missing dependency.
`Depends:` is optional; units without it SHALL pass.

#### Scenario: Valid Depends passes

- **WHEN** a unit has `Depends: <existing unit heading>`
- **THEN** validation succeeds

#### Scenario: Dangling Depends fails

- **WHEN** a unit has `Depends: <non-existent heading>`
- **THEN** validation reports an error naming the unit and the missing dependency

#### Scenario: No Depends passes

- **WHEN** a unit has no `Depends:` line
- **THEN** validation succeeds

#### Scenario: Multiple Depends all valid passes

- **WHEN** a unit has `Depends: <heading1>, <heading2>` and both headings exist as units
- **THEN** validation succeeds

### Requirement: Verify Shell Syntax Lint

For each unit's `Verify:` fenced code block, validate SHALL run `bash -n` on the block contents and report any syntax error as a validation error identifying the issue number, unit heading, and shell error text. An inline `Verify:` command (backtick command on the `Verify:` line) SHALL be accepted as non-empty content and is not shell-linted, since inline Verify lines may mix commands with prose code references. If `bash` is not on `PATH`, validate SHALL emit a warning per fenced block (not an error) and check only that the block is non-empty. Validate SHALL NOT execute the Verify command.

#### Scenario: Valid shell passes

- **WHEN** a Verify fenced block contains syntactically valid bash
- **THEN** `bash -n` succeeds and validation passes for that block

#### Scenario: Shell syntax error fails

- **WHEN** a Verify fenced block contains an unclosed quote
- **THEN** validation reports an error with the `bash -n` output and the unit heading

#### Scenario: bash absent degrades to warning

- **WHEN** `bash` is not on `PATH` and a Verify block is non-empty
- **THEN** validation emits a warning, not an error, for that block

### Requirement: Local Queue Fallback

When `gh` is not on `PATH` or `gh issue list` fails (e.g. no GitHub remote is configured), validate SHALL auto-discover files at `specs/queues/<name>.md` and apply the same unit format and Verify shell lint rules as for GH issue bodies. `<name>` mirrors the change name supplied to `litespec new <name> --issue N`. The `--queue <path>` flag SHALL validate a single local queue file. The `--issue N` flag SHALL fetch and validate a single GH issue by number. When `gh` is available and `gh issue list` succeeds, both GH issues labeled `litespec` and local `specs/queues/*.md` files SHALL be validated. A `gh issue list` failure SHALL produce a warning and skip GH queue validation without failing the command.

#### Scenario: Local queue validated when gh absent

- **WHEN** `gh` is absent and `specs/queues/add-auth.md` exists with a well-formed queue
- **THEN** validate lints that file and reports its units

#### Scenario: --queue flag validates one file

- **WHEN** `litespec validate --queue specs/queues/add-auth.md` is run
- **THEN** only that file is validated as a queue

#### Scenario: --issue flag fetches one issue

- **WHEN** `litespec validate --issue 42` is run and `gh` is available
- **THEN** only issue #42 is fetched and validated

#### Scenario: Both gh and local queues validated together

- **WHEN** `gh` is available and both labeled issues and `specs/queues/*.md` files exist
- **THEN** validate lints both sources and merges results

#### Scenario: gh issue list failure warns and skips

- **WHEN** `gh` is on `PATH` but `gh issue list` fails (e.g. no GitHub remote)
- **THEN** validate emits a warning and skips GH queue validation without failing the command

### Requirement: Offline Graceful Degradation

When `gh` is not on `PATH` or `gh issue list` fails AND no `specs/queues/` directory exists, validate SHALL emit a single warning that the queue was not checked and continue validating specs and decisions. The command's exit status SHALL reflect only the specs and decisions that were validated. Under `--strict`, the absence of any queue source SHALL NOT itself be an error.

#### Scenario: No gh and no queues directory warns once

- **WHEN** `gh` is absent and `specs/queues/` does not exist
- **THEN** validate emits one warning and proceeds to validate specs and decisions

#### Scenario: Specs still validated offline

- **WHEN** `gh` is absent, no `specs/queues/` exists, and a spec has an error
- **THEN** validate reports the spec error and exits non-zero
