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

The `litespec validate` command SHALL fetch open GitHub issues labeled `litespec` via `gh issue list` and lint each issue body as a queue. Across the entire queue body there SHALL be exactly one `Base:` line and exactly one `Branch:` line, and both SHALL occur before the first `##` heading. `Base:` SHALL contain a full 40- or 64-character hexadecimal commit ID; `Branch:` SHALL match `litespec/<kebab-change-name>`. A `##` section is treated as a unit only when its body contains a `Done means:` or `Verify:` line; prose sections SHALL be skipped. Each unit SHALL have a non-empty heading, identified `Done means:` bullet clauses, a `Scenarios:` mapping for those clause IDs, a `Verify:` line carrying either an inline backtick command on the same line or a fenced code block within the unit body, and a checkbox. Each unit MAY include at most one `Read first:`, `Constraints:`, `Depends:`, and `Boundary:` field; each present field is unique and nonempty. Filesystem, process, and network boundaries require the `Risk cases:` matrix defined below. `Read first:` is context, not scope. `Constraints:` states what must stay true or what is out of bounds; it never says what to edit. A checked unit SHALL carry a complete evidence receipt. Missing or malformed ownership, scenario mapping, boundary-risk accounting, or unit elements; duplicate or empty fields; or a checked unit without a complete receipt SHALL produce an error. Issues without the `litespec` label SHALL NOT be scanned.

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

When a unit checkbox is checked, `litespec validate` SHALL require one verbatim red-green evidence receipt for that unit. A complete receipt MUST contain, in order: a Verify command quoted verbatim; `unit digest:` with 64 lowercase hexadecimal characters; `pre sha:` with a full 40- or 64-character hexadecimal commit ID; `pre exit status:` with a non-zero integer; a nonempty fenced block containing raw pre output or literal `<no output>`; `Pre-evidence scope: this command exited <status> at <sha>; nothing else is inferred.` matching the pre fields; `post sha:` with a full commit ID; `post exit status: 0`; a nonempty fenced block containing raw post output or literal `<no output>`; and `Post-evidence scope: this command exited 0 at <sha>; nothing else is inferred.` matching the post fields. The pre and post SHAs MUST differ. The `unit digest:` value MUST equal the SHA-256 digest of the unit's canonicalized contract: its heading; optional `Read first:`, `Constraints:`, `Depends:`, and `Boundary:` values; identified `Done means:` clauses; `Scenarios:` mappings; any required `Risk cases:` entries; and Verify content, each length-prefixed (decimal byte length followed by `:`), line endings normalized to `\n` and trailing whitespace trimmed, absent optional fields omitted. The digest MUST NOT cover the status checkbox, Evidence content, append-only queue metadata, or any other unit text. Validate SHALL recompute the expected digest from the unit's current body and report a missing, malformed, or mismatched digest as an error naming the unit and, on mismatch, both the expected and actual digests — so editing `Done means:` or `Verify:` after evidence is recorded fails validation. When a receipt's declared digest differs from the current contract, validation SHALL apply the same complete receipt grammar and chunk identity checks to the receipt's own declared Verify command and digest; it SHALL accept that receipt only as a historical observation connected to the current digest by valid amendment edges. A receipt declaring the current digest MUST still match the current Verify command exactly. One canonicalization SHALL serve GitHub bodies, GitHub comments, and local queue files. Unchecked units SHALL NOT require a receipt. An initial local or issue-body receipt SHALL live in an `Evidence:` block after `Verify:` and before the checkbox; later local rebuild receipts MAY use identity-bearing records in the append-only metadata stream after all units. On GitHub, either a comment that names the unit heading or an identity-bearing comment with the unit's exact heading and positive 1-based same-heading occurrence SHALL satisfy the check when it carries a complete receipt. A green-only legacy receipt, nonempty `Evidence:` label, or comment that only mentions `Evidence:` SHALL NOT satisfy. When a single pre or post raw-output block exceeds the GitHub comment cap, the receipt MAY use consecutive `Raw output chunk:` records with `Output: pre|post`, `Chunk: <n>/<total>`, the exact `Unit occurrence:`, `Unit heading:`, and `unit digest:` identity fields, followed by closed fenced payloads. Each chunk MUST be the only raw-output chunk in its comment. Chunk totals and numbers MUST be consistent and consecutive, and validation MUST reconstruct the output by concatenating payloads in order without inserted bytes. A non-final chunk MUST end its comment with the existing literal continuation marker, and the next chunk MUST be in the immediately following comment; blank, missing-marker, intervening, dangling, orphan, duplicated, misordered, or wrong-identity continuations MUST fail. Existing field-boundary continuations remain valid. Validate SHALL check receipt structure only; it MUST NOT execute the recorded command, inspect Git ancestry, or interpret either fenced output.

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

#### Scenario: Chunked oversized raw output passes

- **WHEN** a checked GitHub unit uses consecutive identity-bearing raw-output chunks for an oversized pre or post output, with every payload fenced and every continuation comment immediately following the previous one
- **THEN** validation accepts the receipt and reconstructs each raw output by byte-for-byte payload concatenation without truncation or inserted bytes

#### Scenario: Interrupted or duplicated raw-output chunk fails

- **WHEN** a raw-output chunk continuation is separated by a blank or other comment, or repeats or skips a chunk number
- **THEN** validation reports an incomplete evidence receipt instead of accepting a partial or reordered transcript

#### Scenario: Receipt with matching unit digest passes

- **WHEN** a checked unit's receipt carries a `unit digest:` equal to the recomputed digest of the unit's current contract fields
- **THEN** validation succeeds for that unit

#### Scenario: Missing or malformed unit digest fails

- **WHEN** a checked unit's receipt lacks the `unit digest:` field between the Verify quote and `pre sha:`, or carries a value that is not 64 lowercase hexadecimal characters
- **THEN** validation reports an error naming the unit

#### Scenario: Edited contract after recorded evidence fails

- **WHEN** `Done means:`, `Verify:`, or any covered contract field is edited after a checked unit's receipt was posted, so the recomputed digest no longer matches the receipt's `unit digest:`
- **THEN** validation reports a mismatch error naming the unit and both the expected and actual digests

#### Scenario: Amended receipt may use its superseded Verify command

- **WHEN** a complete receipt declares a superseded digest and quotes its own historical Verify command, and a valid same-heading or renamed amendment chain connects that digest to the current contract
- **THEN** validation accepts the receipt as a chained observation while still requiring complete current-digest evidence

#### Scenario: Current digest does not excuse an edited Verify command

- **WHEN** a complete receipt declares the current contract digest but quotes a Verify command different from the current unit's Verify
- **THEN** validation reports an incomplete receipt instead of treating the command edit as stale evidence

#### Scenario: Renamed stale chunks repeat their top-level identity

- **WHEN** a superseded receipt uses raw-output chunks after a heading rename
- **THEN** every chunk repeats the parsed top-level occurrence and heading exactly, and a missing or mismatched identity fails validation

### Requirement: Queue Rebuild Request State

`litespec validate` SHALL associate each structured rebuild request with exactly one unit by exact heading and positive 1-based occurrence among units with that heading. The exact grammar is `Rebuild request:` followed by `Unit occurrence: <positive integer>` and `Unit heading: <exact heading>` on consecutive lines with no other content. A later complete identity-bearing evidence receipt SHALL resolve every earlier request for that identity. An unresolved request, malformed request, malformed identity-bearing receipt, or identity that does not identify exactly one queue unit MUST produce a validation error. The same grammar SHALL apply to GitHub comments and the local queue metadata stream.

#### Scenario: Later receipt resolves repeated requests

- **WHEN** two rebuild requests name the same unit identity and a later complete receipt names that identity
- **THEN** both requests are resolved

#### Scenario: Ambiguous rebuild metadata fails

- **WHEN** rebuild metadata is malformed or its heading and occurrence do not identify exactly one unit
- **THEN** validation reports an error rather than guessing

### Requirement: Witnessed Contract Amendments and Digest Chains

`litespec validate` SHALL parse structured contract amendment records using one grammar for GitHub comments and local queue files. A complete amendment record MUST contain, in order: the exact line `Amendment:`; `Unit occurrence:` naming the unit's positive 1-based same-heading occurrence; `Unit heading:` naming the unit's exact post-amendment heading; `Old digest:` with 64 lowercase hexadecimal characters identifying the superseded contract; `New digest:` with 64 lowercase hexadecimal characters identifying the amended contract; and `Reason:` with one nonempty line. Identity fields SHALL always carry the post-amendment identity because the heading itself is a contract field an amendment may rename; `Old digest:` is the only link to the superseded contract text. An amendment record is append-only routing metadata: on GitHub it SHALL be a comment with the issue body untouched, and locally it SHALL be a block appended after all units of `specs/queues/<name>.md`; no actor other than `litespec-plan` MAY author or alter a unit contract.

A valid amendment SHALL constitute an unresolved rebuild request for its identity: a checked unit MAY be selected by `litespec-build` while it stands even though its checkbox remains checked, and it SHALL resolve only when a later complete identity-bearing evidence receipt carries a `unit digest:` equal to the amendment's `New digest:`. For each unit identity, the last valid amendment's `New digest:` SHALL equal the unit's current contract digest. The distinct digests observed for that identity across receipts, together with the current contract digest, MUST form a chain: every transition between consecutive distinct observed digests, and from the last observed digest to the current contract digest, SHALL be connected by a path over amendment edges (`Old digest:` → `New digest:`), where two amendments between consecutive observations are connected by their shared intermediate digest. A silent contract edit followed by a fresh receipt with no bridging amendment SHALL therefore produce a validation error naming the unit and the disconnected digests, because a valid historical receipt whose digest no longer matches the current contract participates as an observation rather than a structural failure only when an amendment claims its digest as provenance. An unresolvable identity-bearing receipt whose declared digest no amendment records SHALL remain a boundary error. A malformed amendment record, an identity that does not identify exactly one queue unit, or a terminal `New digest:` that does not match the current contract digest SHALL produce a visible validation error rather than a guess. Validate MUST NOT execute commands or consult Git history to evaluate chains.

#### Scenario: Amendment makes a checked unit selectable again

- **WHEN** a checked unit's contract changed, a valid amendment comment names its post-amendment identity with matching Old and New digests, and no later complete receipt resolves it
- **THEN** validation reports the unresolved request so build can select the checked unit while no existing queue text was modified

#### Scenario: Later receipt carrying the New digest resolves the amendment

- **WHEN** a later complete identity-bearing evidence receipt declares the `unit digest:` equal to the amendment's `New digest`
- **THEN** the amendment-request state resolves without editing the issue body or any prior comment

#### Scenario: Heading rename accepted through post-amendment identity

- **WHEN** an amendment renames a unit heading, names the post-amendment heading and occurrence, and records the pre-rename digest as `Old digest`
- **THEN** validation accepts the rename provenance and treats receipts recorded under the superseded heading as chained observations via `Old digest`

#### Scenario: Silent edit plus fresh receipt fails the chain

- **WHEN** a contract field changed without any amendment, old and new receipts both exist, and the transition between their distinct digests is not bridged by an amendment edge
- **THEN** validation reports a broken-chain error naming the unit and the disconnected digests

#### Scenario: Successive heading renames follow digest-linked identities

- **WHEN** valid amendments witness `Old/X` → `Middle/Y` → `Final/Z`, the current queue contains only `Final/Z`, and receipts use each revision's exact occurrence, heading, digest, and Verify command
- **THEN** validation accepts the unique occurrence-preserving digest path while rejecting an intermediate receipt whose identity does not exactly match its recorded post-amendment heading

#### Scenario: Two amendments between receipts bridge the chain

- **WHEN** two valid amendments witness X→Y and Y→Z between a receipt at X and a later receipt at Z
- **THEN** validation accepts the chain because the transition path crosses the shared intermediate digest Y

#### Scenario: Malformed amendment fails visibly

- **WHEN** a comment or local block begins `Amendment:` but omits a required field, misorders fields, or leaves `Reason:` empty
- **THEN** validation reports a malformed-amendment error instead of guessing at intent

#### Scenario: Terminal New digest mismatch fails

- **WHEN** the last valid amendment for a unit identity declares a `New digest:` that does not equal the unit's current contract digest
- **THEN** validation reports the mismatch so plan cannot re-anchor history onto a fabricated contract

#### Scenario: Local queue amendment block parses by the same grammar

- **WHEN** a block identical to the GitHub amendment grammar is appended after the units of a local queue file
- **THEN** validation parses it, reports its unresolved request against the named unit, and ignores it as unit-body content

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

### Requirement: Unit Scenario Mapping

Every queue unit SHALL express `Done means:` as one or more nonempty bullet clauses carrying unique bracketed IDs, followed by a `Scenarios:` block that maps every clause ID to at least one nonempty named test scenario. Every scenario mapping MUST reference an existing clause ID, and every clause ID MUST appear in the mapping. `Scenarios:` SHALL be unique and nonempty. The mapping SHALL participate in the unit contract digest. Validate SHALL check the identifiers and one-to-one coverage structure but SHALL NOT inspect test source or claim that a named test exercises the clause.

#### Scenario: Every clause has a named test scenario

- **WHEN** a unit has `Done means:` clauses `[timeout]` and `[cleanup]` and its `Scenarios:` block maps both IDs to named tests
- **THEN** queue validation accepts the scenario mapping and includes it in the unit digest

#### Scenario: Unmapped or unknown clause fails

- **WHEN** a `Done means:` clause has no scenario mapping or a scenario mapping names an ID absent from `Done means:`
- **THEN** queue validation reports the unmatched ID and exits non-zero

### Requirement: External Boundary Risk Accounting

A queue unit MAY declare one `Boundary:` value. A filesystem, process, or network boundary MUST include one `Risk cases:` block accounting for `timeout`, `cleanup`, `non-ENOENT errors`, `concurrency`, and `optional configured dependencies`. Each risk SHALL map to a scenario ID declared by that unit or use `N/A — <nonempty reason>`. `Boundary:` and `Risk cases:` SHALL participate in the unit contract digest. Validate SHALL enforce the shape and references but SHALL NOT judge whether a boundary was omitted, an N/A reason is true, or a mapped test is behaviorally adequate.

#### Scenario: Boundary risks map to scenarios or reasons

- **WHEN** a process unit maps timeout and cleanup to scenario IDs and marks the other required risks N/A with nonempty reasons
- **THEN** queue validation accepts the risk accounting and includes it in the unit digest

#### Scenario: Boundary risk is missing

- **WHEN** a filesystem, process, or network unit omits a required risk or gives a blank N/A reason
- **THEN** queue validation reports the missing or empty risk entry and exits non-zero

#### Scenario: Unknown boundary value fails

- **WHEN** a unit declares `Boundary:` with a value other than exactly `filesystem`, `process`, or `network` — such as `Filesystem`, `Process`, `Network`, `database`, or any other unknown word
- **THEN** queue validation reports that the boundary value is outside the closed, case-sensitive vocabulary and exits non-zero

#### Scenario: Omitted boundary stays valid

- **WHEN** a unit declares no `Boundary:` field at all
- **THEN** queue validation accepts the unit without requiring boundary vocabulary or risk accounting

### Requirement: Re-plan Marker State

`litespec validate` SHALL scan routing metadata oldest to newest and count completed review-requested rebuild cycles per unit contract digest. A cycle begins with one or more literal `Rebuild request:` records and completes when a later complete identity-bearing evidence receipt resolves them. An amendment and the receipt resolving it MUST NOT count as a review-requested rebuild cycle. After two completed cycles for the current digest, another rebuild request MUST be invalid. The valid route is exactly `Re-plan required:`, `Unit occurrence: <positive integer>`, `Unit heading: <exact heading>`, `Unit digest: <current 64 lowercase hex digest>`, and `Reason: <nonempty one-line reason>` on consecutive lines. A marker before two completed cycles or a second unresolved marker for the same identity and digest MUST be invalid. A marker SHALL remain unresolved until a plan-authored amendment has an `Old digest:` equal to the marker digest. The amendment's normal fresh-evidence requirement then applies, while the new digest starts with zero completed cycles. GitHub comments and the append-only metadata stream after local queue units SHALL use the same request, receipt, marker, and amendment grammar.

#### Scenario: Third rebuild request is rejected

- **WHEN** two rebuild request-to-receipt cycles completed against one unit digest and another rebuild request is recorded
- **THEN** validation reports that the unit requires re-planning instead of another rebuild

#### Scenario: Plan amendment resolves marker and resets count

- **WHEN** a re-plan marker names the current digest and a later valid amendment starts at that digest
- **THEN** the marker resolves, the amendment remains selectable until fresh evidence arrives, and rebuild counting restarts at zero for the new digest

#### Scenario: Amendment evidence is not a review rebuild

- **WHEN** fresh evidence resolves an amendment without resolving any literal `Rebuild request:` record
- **THEN** the current digest still has zero completed review-requested rebuild cycles

#### Scenario: Premature or duplicate marker fails

- **WHEN** a re-plan marker appears before two completed cycles or another unresolved marker already exists for the same identity and digest
- **THEN** validation reports malformed routing state and exits non-zero

### Requirement: Validation Scope Output

A successful `litespec validate` invocation SHALL state `structure ok; implementation semantics not verified` in human-readable output. Minimal output MUST carry equivalent `structure-ok` and `semantics-unverified` signals. Full and minimal JSON SHALL preserve the structural `valid` boolean and include machine-readable fields identifying the validation scope as structure and implementation semantics as unverified. Counts MAY follow the scope statement. The command MUST NOT describe successful structural validation with an unqualified `ok`.

#### Scenario: Text success names its limit

- **WHEN** validation finds no structural errors in text mode
- **THEN** output begins with `structure ok; implementation semantics not verified`

#### Scenario: JSON success names its limit

- **WHEN** validation finds no structural errors in JSON or minimal JSON mode
- **THEN** the response reports structural validity and explicitly reports that implementation semantics were not verified
