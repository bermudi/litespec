# review

## Requirements

### Requirement: Adversarial Review of Issue and Spec vs Implementation

The `litespec-review` skill SHALL perform a context-aware, adversarial review by reading the GH issue body (its proposal, design, and units with `Done means:` and `Verify:`), any load-bearing `specs/<feature>/spec.md`, and relevant `specs/decisions/`, then comparing the stated intent and durable contracts against the implementation. The review SHALL probe for interaction bugs, state transitions, wiring gaps, and contract violations, not only syntax or surface compliance.

#### Scenario: Review with GH issue and spec

- **WHEN** `litespec-review` is invoked for a change with a GH issue and a load-bearing spec
- **THEN** it reads the issue and spec and probes the implementation against them

#### Scenario: Review for small fix without issue

- **WHEN** `litespec-review` is invoked on a small fix with no GH issue
- **THEN** it reads the relevant `specs/<feature>/spec.md` and the changed code before reporting findings

### Requirement: Findings and Verdict

The `litespec-review` skill SHALL report each finding with a **Severity** (`CRITICAL`, `WARNING`, or `SUGGESTION`), a **Location** (`file:line` or unit), **Evidence** (excerpt or observation), and a **Fix direction** (one unambiguous instruction). The review SHALL conclude with a verdict of `PASS` or `CHANGES REQUESTED`.

#### Scenario: Pass verdict

- **WHEN** all units satisfy their `Done means:` and `Verify:` contracts and no CRITICAL or WARNING findings remain
- **THEN** the review returns `PASS`

#### Scenario: Changes requested verdict

- **WHEN** the review identifies a CRITICAL or WARNING finding
- **THEN** it returns `CHANGES REQUESTED`

### Requirement: Triage into Lanes

For each finding, `litespec-review` SHALL first determine whether the finding cites a unit's `Done means:` or `Verify:`, then route it into exactly one of the following lanes:
- A finding that is **CRITICAL** and breaks a unit's `Done means:` or `Verify:` SHALL be routed to `litespec-build`: the unit's checkbox is unchecked, `references/build/review-fixing.md` is loaded, and the unit is rebuilt with expanded scope; the GH issue stays open until all units re-pass.
- A finding that is **CRITICAL** or **WARNING** and lies outside any unit's contract (neighboring code, help text, stale decision, drive-by) SHALL be routed to the small fix lane; no unit is created and the issue is not reopened.
- A **SUGGESTION** SHALL be routed to the small fix lane at the user's discretion and is not blocking.
- A finding that is `"needs decision"` SHALL first create a decision in `specs/decisions/`, then be routed per the rules above.
- A finding where the unit's outcome or shape is wrong SHALL be routed to `litespec-plan`, not a fix.
- A non-trivial finding outside any unit's contract that needs real implementation work but is not a shape problem SHALL be routed to a new unit: draft `## <outcome>`, `Done means:`, `Verify:`, and `Depends:` if it blocks on existing units; create a GH sub-issue via `gh issue create --parent <N> --label litespec` or, if `gh` is unavailable, write the unit to `specs/queues/<parent-name>-review.md`.

#### Scenario: Critical finding breaks a unit contract

- **WHEN** a CRITICAL finding cites a unit's `Done means:` or `Verify:` violation with direct evidence
- **THEN** the finding routes to `litespec-build`, the unit checkbox is unchecked, and the issue stays open

#### Scenario: Warning outside unit contract is a small fix

- **WHEN** a WARNING finding concerns neighboring code or a stale decision and does not cite a unit contract
- **THEN** it routes to the small fix lane and does not reopen the issue

#### Scenario: Finding needs a decision first

- **WHEN** a finding requires a durable architectural ruling before it can be fixed
- **THEN** it is reported as `"needs decision"` and routed to `specs/decisions/` before any fix

#### Scenario: Shape was wrong

- **WHEN** a finding shows the implemented outcome does not match the intended shape of the unit
- **THEN** the finding is routed to `litespec-plan`, not a fix

#### Scenario: Non-trivial finding outside all units

- **WHEN** a finding needs a unit's worth of work and does not break an existing unit's contract
- **THEN** the review drafts a new unit and creates a GH sub-issue or local queue entry, but does not implement it

### Requirement: Pure Review Role

The `litespec-review` skill MUST NOT write code, modify files, or implement fixes. It SHALL report findings and route them. Creating a GH sub-issue or a `specs/queues/<parent-name>-review.md` entry is routing, not implementation.

#### Scenario: Review does not edit code

- **WHEN** `litespec-review` runs
- **THEN** it does not edit source files, check or uncheck checkboxes, or commit changes

#### Scenario: Sub-issue creation is routing

- **WHEN** review creates a GH sub-issue or local queue entry for a new unit
- **THEN** it only drafts the unit in the issue or queue body and does not implement the code

### Requirement: Adversarial Scenario Reference

When the change contains stateful code paths, `litespec-review` SHALL load `references/review/adversarial-review.md` and construct worst-case scenarios from the specs before tracing implementation code. For each scenario it SHALL trace the code path, tag handling as `Handled`, `Missing`, or `Uncertain`, and report any confirmed or inferred gap as a finding.

#### Scenario: Adversarial review for stateful code

- **WHEN** the change contains state transitions, multi-entity operations, or concurrent access
- **THEN** review loads `references/review/adversarial-review.md`, enumerates adversarial scenarios, and checks them against the implementation

#### Scenario: Adversarial finding breaks a unit

- **WHEN** an adversarial scenario confirms an interaction bug that breaks a unit's `Done means:` or `Verify:`
- **THEN** the finding is CRITICAL and routes to `litespec-build` with the unit checkbox unchecked

### Requirement: No Unit for Trivial Findings

`litespec-review` SHALL NOT invent new units for trivial findings; those SHALL be routed to the small fix lane. A new unit SHALL be drafted only when a finding needs a unit's worth of work and does not break an existing unit's contract.

#### Scenario: Typo in help text

- **WHEN** review finds a typo in help text or other trivial issue outside any unit contract
- **THEN** it routes the finding to the small fix lane and does not create a unit

#### Scenario: Missing contract for new work

- **WHEN** a finding requires real implementation work that no existing unit covers
- **THEN** the review may draft a new unit and route it to a sub-issue or queue entry

### Requirement: Issue Closure Condition

A GH issue SHALL remain open until all of its units pass their `Done means:` and `Verify:` contracts; it SHALL close when all units pass. Small fix findings outside any unit contract SHALL NOT reopen the issue.

#### Scenario: All units pass

- **WHEN** all unit checkboxes are checked and `litespec-review` returns `PASS`
- **THEN** the GH issue may be closed

#### Scenario: Small fix does not reopen issue

- **WHEN** a small fix is applied for a WARNING or SUGGESTION outside any unit contract
- **THEN** the issue is not reopened

### Requirement: No Persistent Finding Tracker

`litespec-review` findings SHALL be ephemeral. A finding SHALL route to an existing unit checkbox, be fixed immediately in the appropriate lane, or become a new queue issue. The review skill SHALL NOT maintain a finding tracker, task list, or persistent finding artifact.

#### Scenario: Findings route or become queue issues

- **WHEN** `litespec-review` finishes
- **THEN** its findings are either resolved via `litespec-build`/small fix or recorded as a GH sub-issue or local queue entry; no finding tracker or task list is created

#### Scenario: Re-review reads current state

- **WHEN** `litespec-review` is re-run after fixes
- **THEN** it evaluates the current state from the issue, spec, and code, not from a previous finding log
