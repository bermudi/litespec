# evidence

## Purpose

Byte-exact evidence assembly and queue-body bookkeeping are CLI responsibilities, not agent-prose responsibilities. This capability covers the commands that produce validator-clean evidence receipts, publish them on request, and tick exactly one unit checkbox.

## Requirements

### Requirement: Receipt Emission

`litespec receipt` SHALL assemble one red-green evidence receipt for a resolved queue unit from labeled run evidence and emit numbered comment files plus the exact gh commands that post them, without invoking any gh write unless posting is explicitly requested. Resolution from a local queue file emits the numbered files without gh commands, since no issue is addressed. The command SHALL resolve the unit by exact heading and positive same-heading occurrence from the live issue body or a local queue file and SHALL refuse ambiguous, unknown, or out-of-range resolution. It SHALL refuse equal pre and post SHAs, a zero pre exit status, a nonzero post exit status, and a pre commit that is not an ancestor of the post commit, before writing any file. The assembled receipt SHALL carry the versioned evidence header with a Receipt ID identical to the validator's derivation for the same logical receipt, SHALL split oversized content only at legal boundaries using the exact continuation marker or the explicit chunk form, and SHALL parse clean through the existing evidence grammar before any file is written.

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

- **WHEN** a raw output block exceeds the GitHub comment cap
- **THEN** the split uses the explicit chunk form with repeated receipt identity and consecutive chunk numbering, or falls after a scope line, and every non-final comment ends with the exact continuation marker

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
