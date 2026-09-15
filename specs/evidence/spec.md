# evidence

## Purpose

Byte-exact evidence assembly and queue-body bookkeeping are CLI responsibilities, not agent-prose responsibilities. This capability covers the commands that produce validator-clean evidence receipts, publish them on request, and tick exactly one unit checkbox.

## Requirements

### Requirement: Receipt Emission

`litespec receipt` SHALL assemble one red-green evidence receipt for a resolved queue unit from labeled run evidence and emit numbered comment files plus the exact gh commands that post them, without invoking any gh write unless posting is explicitly requested. Resolution from a local queue file emits the numbered files without gh commands, since no issue is addressed. The command SHALL resolve the unit by exact heading and positive same-heading occurrence from the live issue body or a local queue file and SHALL refuse ambiguous, unknown, or out-of-range resolution. It SHALL refuse equal pre and post SHAs, a zero pre exit status, a nonzero post exit status, and a pre commit that is not an ancestor of the post commit, before writing any file. The assembled receipt SHALL carry the versioned evidence header with a Receipt ID identical to the validator's derivation for the same logical receipt and SHALL parse clean through the existing evidence grammar before any file is written. Newly assembled receipts SHALL declare `Protocol: evidence/v2` and conform to the Bounded Output Excerpts requirement; the exact continuation marker and explicit chunk form remain valid only for receipts declaring earlier protocols.

#### Scenario: Ambiguous heading refused

- **WHEN** the given heading matches multiple unit occurrences and no explicit occurrence is given
- **THEN** the command exits non-zero naming the ambiguity and writes no comment files

#### Scenario: Non-ancestor pre refused

- **WHEN** the pre commit is not an ancestor of the post commit
- **THEN** the command exits non-zero and writes no comment files

#### Scenario: Emission validates by construction

- **WHEN** assembly succeeds and numbered comment files are emitted
- **THEN** the emitted comments join into one complete receipt that parses clean through the existing validation grammar

#### Scenario: Oversized output splits legally

- **WHEN** a raw output block exceeds the GitHub comment cap in a receipt declaring an earlier protocol
- **THEN** the split uses the explicit chunk form with repeated receipt identity and consecutive chunk numbering, or falls after a scope line, and every non-final comment ends with the exact continuation marker

### Requirement: Bounded Output Excerpts

Newly assembled receipts SHALL declare `Protocol: evidence/v2` and carry bounded output excerpts: each run position records the full output's byte length and SHA-256 beside one fence holding either the verbatim complete output, when it fits the excerpt budget, or a deterministic head excerpt, one literal elision marker naming the exact elided byte count, and a deterministic tail excerpt. One v2 receipt SHALL fit one comment within the fixed byte budget; it SHALL NOT continue across comments or use the chunk form, and assembly SHALL refuse visibly when identity and status fields alone exceed the budget. Excerpt budgets SHALL be fixed constants, never configuration. Validation SHALL dispatch on the declared protocol: v2 receipts must be arithmetically consistent (head plus elided plus tail equals the declared byte count), unelided fences must hash to the declared SHA-256, and v2 receipts that continue, chunk, or exceed the budget SHALL error visibly. Receipts declaring `evidence/v1` or the legacy unversioned shape SHALL continue to validate under their retained parsers, including continuation and chunk forms.

#### Scenario: Oversized output emits one bounded comment

- **WHEN** a raw output exceeds the excerpt budget
- **THEN** the emitted v2 receipt is a single comment whose fence carries the head excerpt, the elision marker with the exact elided byte count, and the tail excerpt, beside the full output's byte count and SHA-256

#### Scenario: Small output stays verbatim

- **WHEN** a raw output fits the excerpt budget
- **THEN** the fence holds the complete output with no elision marker, and the declared byte count and SHA-256 match the fenced bytes

#### Scenario: v2 never continues

- **WHEN** a receipt declaring evidence/v2 ends with the continuation marker, uses the chunk form, or exceeds the byte budget
- **THEN** validation errors visibly instead of joining chunks or reconstructing output

#### Scenario: Earlier-protocol receipts keep their parser

- **WHEN** a checked unit's receipt declares evidence/v1 or the legacy unversioned shape and is valid under its own grammar
- **THEN** it validates exactly as before, including continuation and chunk forms

### Requirement: Opt-in Comment Posting

`litespec receipt` SHALL post the emitted comments itself only when posting is explicitly requested; the default invocation SHALL emit files and commands without any gh write. When posting is requested, comments SHALL be posted strictly in numbered order, and a failing gh invocation SHALL stop the chain with a visible error naming the failing command, the already posted parts, and the unposted files, without retry or reordering.

#### Scenario: Opt-in posting runs in order

- **WHEN** posting is requested and every gh invocation succeeds
- **THEN** the comments are posted in numbered order and each posted comment is reported

#### Scenario: Mid-chain failure stops visibly

- **WHEN** a gh invocation fails after earlier comments were posted
- **THEN** the command stops, names the failing command and the unposted files, and retries nothing

### Requirement: Managed Unit Checkbox Tick

`litespec issue check` SHALL tick exactly one unit checkbox in a GH issue body: it SHALL resolve the unit by exact heading and positive same-heading occurrence, write back a body differing from the fetched body by exactly that one checkbox flip plus trailing-newline normalization (the written body carries exactly one trailing newline), and verify the ownership lines are byte-unchanged after the edit. GitHub's API appends one trailing newline to every stored body, so the stored body converges to two trailing newlines and stays there; at steady state the fetch-to-fetch delta is exactly the flip. It SHALL refuse, without issuing any write, when any other body content would change, when the heading does not resolve to exactly one unchecked unit, or when gh is unavailable.

#### Scenario: Single tick preserves ownership

- **WHEN** the command resolves one unchecked unit and gh is available
- **THEN** the written body differs from the fetched body only by that checkbox flip and the ownership lines are byte-identical

#### Scenario: Refusal leaves the body untouched

- **WHEN** resolution is ambiguous, the target is already checked, or any other body byte would change
- **THEN** the command exits non-zero and issues no write

### Requirement: Digest Heading Filter

`litespec digest` SHALL accept a heading filter that prints only the digest lines whose unit heading equals the given text exactly. When the heading repeats within one queue, every occurrence SHALL be printed with its own occurrence number. A heading matching no unit SHALL exit non-zero naming the heading.

#### Scenario: Duplicate headings list every occurrence

- **WHEN** the filtered heading names two units in the same queue
- **THEN** both occurrence lines are printed, each with its own occurrence number and digest

#### Scenario: Zero matches fail visibly

- **WHEN** the filtered heading matches no unit in the queue
- **THEN** the command exits non-zero naming the heading
