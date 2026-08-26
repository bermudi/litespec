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

The `litespec validate` command SHALL fetch open GitHub issues labeled `litespec` via `gh issue list` and lint each issue body as a queue. Across the entire queue body there SHALL be exactly one `Base:` line and exactly one `Branch:` line, and both SHALL occur before the first `##` heading. `Base:` SHALL contain a full 40- or 64-character hexadecimal commit ID; `Branch:` SHALL match `litespec/<kebab-change-name>`. A `##` section is treated as a unit only when its body contains a `Done means:` or `Verify:` line; prose sections SHALL be skipped. Each unit SHALL have a non-empty heading, a `Done means:` line, a `Verify:` line carrying either an inline backtick command on the same line or a fenced code block within the unit body, and a checkbox. Each unit MAY include at most one `Read first:` line and at most one `Constraints:` line; both are optional, unique, and nonempty when present. `Read first:` is context, not scope — prefer areas and rulings over long file lists. `Constraints:` states what must stay true or what is out of bounds — it never says what to edit; the worker owns the implementation path. Omit either field rather than writing a placeholder. A checked unit SHALL carry a complete evidence receipt (see Checked Unit Evidence Receipt). Missing or malformed ownership or unit elements, duplicate or empty optional fields, or a checked unit without a complete receipt SHALL produce an error. Issues without the `litespec` label SHALL NOT be scanned.

#### Scenario: Well-formed queue passes

- **WHEN** `litespec validate` scans an open issue labeled `litespec` whose body has valid `Base:` and `Branch:` ownership lines plus a valid unit
- **THEN** validation succeeds for that issue

#### Scenario: Missing ownership metadata fails

- **WHEN** a labeled queue issue omits `Base:` or `Branch:` before its first `##` heading
- **THEN** validation reports an ownership error

#### Scenario: Duplicate or malformed ownership metadata fails

- **WHEN** a queue repeats an ownership line anywhere in its body or uses a short Base or non-`litespec/` Branch
- **THEN** validation reports an ownership error

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

#### Scenario: Valid Read first and Constraints pass

- **WHEN** a unit has `Read first:` with context areas and `Constraints:` with boundary text (inline or bulleted)
- **THEN** validation succeeds for those fields

#### Scenario: Duplicate Read first or Constraints fails

- **WHEN** a unit repeats `Read first:` or `Constraints:`
- **THEN** validation reports a duplicate-field error

#### Scenario: Empty Read first or Constraints fails

- **WHEN** a unit has `Read first:` or `Constraints:` empty or with a placeholder like `-` or `none` or `n/a`
- **THEN** validation reports a nonempty error

### Requirement: Checked Unit Evidence Receipt

When a unit checkbox is checked, `litespec validate` SHALL require one verbatim red-green evidence receipt for that unit. A complete receipt MUST contain, in order: the unit's `Verify:` command quoted verbatim; `unit digest:` with 64 lowercase hexadecimal characters; `pre sha:` with a full 40- or 64-character hexadecimal commit ID; `pre exit status:` with a non-zero integer; a nonempty fenced block containing raw pre output or literal `<no output>`; `Pre-evidence scope: this command exited <status> at <sha>; nothing else is inferred.` matching the pre fields; `post sha:` with a full commit ID; `post exit status: 0`; a nonempty fenced block containing raw post output or literal `<no output>`; and `Post-evidence scope: this command exited 0 at <sha>; nothing else is inferred.` matching the post fields. The pre and post SHAs MUST differ. The `unit digest:` value MUST equal the SHA-256 digest of the unit's canonicalized contract: its heading, optional `Read first:`, `Constraints:`, and `Depends:` values, `Done means:` value, and Verify content, each length-prefixed (decimal byte length followed by `:`), line endings normalized to `\n` and trailing whitespace trimmed, absent optional fields omitted. The digest MUST NOT cover the status checkbox, Evidence content, or any other unit text. Validate SHALL recompute the expected digest from the unit's current body and report a missing, malformed, or mismatched digest as an error naming the unit and, on mismatch, both the expected and actual digests — so editing `Done means:` or `Verify:` after evidence is recorded fails validation. One canonicalization SHALL serve GitHub bodies, GitHub comments, and local queue files. Unchecked units SHALL NOT require a receipt. Local queues and issue-body receipts SHALL live in an `Evidence:` block after `Verify:` and before the checkbox. On a GitHub issue, either a comment that names the unit heading or an identity-bearing comment with the unit's exact heading and positive 1-based same-heading occurrence SHALL satisfy the check when it carries a complete receipt. A green-only legacy receipt, nonempty `Evidence:` label, or comment that only mentions `Evidence:` SHALL NOT satisfy. Validate SHALL check receipt structure only; it MUST NOT execute the recorded command, inspect Git ancestry, or interpret either fenced output.

#### Scenario: Unchecked unit without receipt passes

- **WHEN** a unit checkbox is unchecked and the unit has no `Evidence:` block
- **THEN** validation succeeds for that unit

#### Scenario: Checked unit with complete receipt passes

- **WHEN** a checked unit has an `Evidence:` block quoting the Verify command with distinct full pre and post SHAs, non-zero pre and zero post statuses, two nonempty fenced outputs, and matching scope lines
- **THEN** validation succeeds for that unit

#### Scenario: Prose sticker fails

- **WHEN** a checked unit has `Evidence: verified at abc123` or any other prose that lacks the required receipt fields
- **THEN** validation reports an error identifying the incomplete receipt

#### Scenario: GH comment sticker fails

- **WHEN** a checked GH issue unit has no body receipt and a comment that names the heading and contains `Evidence:` but lacks the required receipt fields
- **THEN** validation reports an error identifying the incomplete receipt

#### Scenario: GH comment complete receipt passes

- **WHEN** a checked GH issue unit has no body receipt and a comment that names the heading and carries a complete red-green receipt
- **THEN** validation succeeds for that unit

#### Scenario: Identity-bearing rebuild receipt passes

- **WHEN** a checked GH issue unit has no body receipt and a comment carries its exact heading, positive same-heading occurrence, and a complete red-green receipt
- **THEN** validation associates the receipt with that exact unit, including when another unit has the same heading

#### Scenario: Edited command, same sha, or empty fence fails

- **WHEN** a checked unit receipt omits or edits Verify, records the same pre and post SHA, or includes an empty pre or post fence
- **THEN** validation reports an error identifying the incomplete receipt

#### Scenario: Green pre or failed post cannot complete a unit

- **WHEN** a checked unit records pre exit status 0 or a non-zero post exit status
- **THEN** validation reports an error and the unit is not accepted

#### Scenario: Either verification run emits no output

- **WHEN** either receipt run emits nothing and records the literal `<no output>` in its output fence
- **THEN** validation accepts the receipt

#### Scenario: Legacy green-only receipt fails

- **WHEN** a checked unit carries the former single-sha, single-status receipt shape without pre evidence
- **THEN** validation reports an incomplete red-green receipt

#### Scenario: Receipt with matching unit digest passes

- **WHEN** a checked unit's receipt carries a `unit digest:` equal to the recomputed digest of the unit's current contract fields
- **THEN** validation succeeds for that unit

#### Scenario: Missing or malformed unit digest fails

- **WHEN** a checked unit's receipt lacks the `unit digest:` field between the Verify quote and `pre sha:`, or carries a value that is not 64 lowercase hexadecimal characters
- **THEN** validation reports an error naming the unit

#### Scenario: Edited contract after recorded evidence fails

- **WHEN** `Done means:`, `Verify:`, or any covered contract field is edited after a checked unit's receipt was posted, so the recomputed digest no longer matches the receipt's `unit digest:`
- **THEN** validation reports a mismatch error naming the unit and both the expected and actual digests

### Requirement: GitHub Rebuild Request State

`litespec validate` SHALL associate each structured GitHub rebuild request with exactly one unit by exact heading and positive 1-based occurrence among units with that heading. A later complete identity-bearing evidence receipt SHALL resolve every earlier request for that identity. An unresolved request, malformed request, malformed identity-bearing receipt, or identity that does not identify exactly one queue unit MUST produce a validation error.

#### Scenario: Later receipt resolves repeated requests

- **WHEN** two rebuild requests name the same unit identity and a later complete receipt names that identity
- **THEN** both requests are resolved

#### Scenario: Ambiguous rebuild metadata fails

- **WHEN** rebuild metadata is malformed or its heading and occurrence do not identify exactly one unit
- **THEN** validation reports an error rather than guessing

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

For each unit's `Verify:` fenced code block, validate SHALL run `bash -n` on the block contents and report any syntax error as a validation error identifying the issue number, unit heading, and shell error text. An inline `Verify:` command (backtick command on the `Verify:` line extracted as first→last `` ` `` span) SHALL be accepted as non-empty content. If `bash` is not on `PATH`, validate SHALL emit a warning per fenced block (not an error) and check only that the block is non-empty. Validate SHALL reject obviously vacuous Verify commands — a fenced block or inline span that is comment-only or a single `true`, `:`, or `exit 0` optionally followed by `;` and/or `# comment` — reporting `Verify command is obviously vacuous; assert the unit outcome` (a blank/whitespace-only block is reported as `Verify block is empty`). When both an inline span and a fenced block are present on the same unit, only the fenced block SHALL be linted. Lint does not claim to understand whether an otherwise valid command discriminates the outcome. Validate SHALL NOT execute the Verify command.

#### Scenario: Valid shell passes

- **WHEN** a Verify fenced block contains syntactically valid bash
- **THEN** `bash -n` succeeds and validation passes for that block

#### Scenario: Shell syntax error fails

- **WHEN** a Verify fenced block contains an unclosed quote
- **THEN** validation reports an error with the `bash -n` output and the unit heading

#### Scenario: bash absent degrades to warning

- **WHEN** `bash` is not on `PATH` and a Verify block is non-empty
- **THEN** validation emits a warning, not an error, for that block

#### Scenario: Vacuous fenced Verify fails

- **WHEN** a Verify fenced block is `true`, `:`, `exit 0`, or comment-only
- **THEN** validation reports `Verify command is obviously vacuous; assert the unit outcome`

#### Scenario: Vacuous inline Verify fails

- **WHEN** a unit's `Verify:` line carries an inline backtick span `true` (or `:`, `exit 0`) and no fenced block
- **THEN** validation reports `Verify command is obviously vacuous; assert the unit outcome`

#### Scenario: Non-vacuous Verify passes

- **WHEN** a Verify block or inline span is `go test ./...` or `echo ok` or multi-line, or `returns true` prose-contaminated span `true` and `go test` (first→last extraction)
- **THEN** validation does not report vacuous

#### Scenario: Both spellings present — fence wins

- **WHEN** a unit has both an inline `true` vacuous span and a fenced `go test` block, or vice versa with fenced `true` and inline `go test`
- **THEN** only the fenced block is linted; the first case passes, the second fails for the fence

### Requirement: Local Queue Fallback

When `gh` is not on `PATH` or `gh issue list` fails (e.g. no GitHub remote is configured), validate SHALL auto-discover files at `specs/queues/<name>.md` and apply the same unit format and Verify shell lint rules as for GH issue bodies. `<name>` is the change name chosen during `plan[clear]`. The `--queue <path>` flag SHALL validate a single local queue file. The `--issue N` flag SHALL fetch and validate a single GH issue by number. When `gh` is available and `gh issue list` succeeds, both GH issues labeled `litespec` and local `specs/queues/*.md` files SHALL be validated. A `gh issue list` failure SHALL produce a warning and skip GH queue validation without failing the command.

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
