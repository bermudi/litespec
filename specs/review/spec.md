# review

## Requirements

### Requirement: Trusted Review Bootstrap

The `litespec-review` safety contract SHALL treat harness/system instructions and repository instruction files auto-loaded by the harness to activate review—including applicable `AGENTS.md` files and the selected review `SKILL.md`—as trusted bootstrap inputs outside litespec's screening guarantee. Local-path screening SHALL begin after skill activation. Litespec MUST NOT claim to secure content loaded before activation.

#### Scenario: Harness auto-loads review instructions

- **WHEN** the harness reads system instructions, applicable `AGENTS.md` files, or the review `SKILL.md` before activating review
- **THEN** those inputs are treated as trusted bootstrap rather than screened local review content

#### Scenario: Bootstrap instructions are not trusted

- **WHEN** the user cannot trust repository instructions auto-loaded by the harness
- **THEN** litespec review stops and directs the user to a harness-level sandbox or pre-load policy

### Requirement: Adversarial Review of Issue and Spec vs Implementation

After trusted bootstrap and skill activation, `litespec-review` SHALL perform a context-aware, adversarial review by first reading only the remote GH issue body, then safely screening every additional local path before reading its contents. Safely approved content includes the queue units, load-bearing specs, relevant decisions, and exact issue-owned review scope. It SHALL probe for interaction bugs, state transitions, wiring gaps, and contract violations, not only syntax or surface compliance.

#### Scenario: Review with GH issue and spec

- **WHEN** `litespec-review` is invoked for a change with a GH issue and a load-bearing spec
- **THEN** it reads the remote issue, screens all selected local paths, then reads approved specs, decisions, and issue-owned changes before probing the implementation

#### Scenario: Review for small fix without issue

- **WHEN** `litespec-review` is invoked on a small fix with no GH issue
- **THEN** it requires a single-parent fix commit, uses the parent as the screening base, enumerates changed paths with NUL-delimited Git output, screens all selected paths before content access, and reviews approved per-path changes rather than inferring scope from a dirty tree

#### Scenario: Root or merge commit is not inferred as a small fix

- **WHEN** the identified small-fix commit has zero or multiple parents
- **THEN** review stops without a verdict because deleted-path preimage ownership is ambiguous

### Requirement: Isolated Issue Branch

The `litespec-plan` skill in clear mode SHALL start from a clean working tree, capture `Base: <sha>`, create and switch to a dedicated `litespec/<change-name>` branch, and record it as `Branch: <branch>` in the GH issue body or local queue mirror. All commits and working-tree changes on that branch SHALL belong exclusively to the issue. Unrelated work MUST use another branch or worktree.

#### Scenario: Clear planning starts clean

- **WHEN** `litespec-plan` in clear mode finds output from `git status --porcelain`
- **THEN** it stops before creating the issue or branch and asks the user to commit, stash, or move the existing work

#### Scenario: Dedicated branch records ownership

- **WHEN** clear planning starts from a clean tree for change name `add-search`
- **THEN** it records the current HEAD as `Base:`, creates `litespec/add-search`, and records `Branch: litespec/add-search`

#### Scenario: Existing branch is not reused

- **WHEN** the intended `litespec/<change-name>` branch already exists
- **THEN** clear planning stops instead of mixing another issue into that branch

### Requirement: Exact Review Scope

After trusted bootstrap and skill activation, `litespec-review` SHALL read only the remote GH issue body before screening additional local content. A local queue fallback SHALL itself be screened before being read for ownership metadata. Review SHALL then verify `Base:` and `Branch:`, current branch identity, and Base ancestry. It SHALL enumerate tracked and untracked path names without contents and add every local contract or reference selected for review. Every selected local path and every parent component SHALL be inspected without following links before content access. Review SHALL reject paths outside the repository and known secret-like names; every parent component MUST be a real directory and every existing leaf MUST be a regular file. A deleted tracked path MUST have regular-file mode at Base and remain absent in the working tree. After a path passes screening, review may read that approved path; newly discovered paths MUST be screened before reading. Review SHALL stop without a verdict when ownership or safe inspection cannot be proved.

#### Scenario: Scope includes tracked and untracked work

- **WHEN** the ownership checks pass and the issue branch contains tracked changes and `??` paths
- **THEN** review screens all selected implementation and contract paths plus their parent components, then examines approved tracked diffs, untracked regular files, specs, and decisions

#### Scenario: Branch mismatch stops review

- **WHEN** the current branch differs from the issue's `Branch:` line
- **THEN** review stops without a verdict and does not infer another scope

#### Scenario: Missing ownership metadata stops review

- **WHEN** `Base:` or `Branch:` is absent or Base is not an ancestor of HEAD
- **THEN** review stops without a verdict

#### Scenario: Unsafe path stops before content access

- **WHEN** any selected local path is secret-like or outside the repository, any component is a symlink, a parent is not a directory, an existing leaf is not a regular file, or a deleted path was not a regular file at Base
- **THEN** review stops without a verdict, reports only the path and reason, and does not read contents or follow a link target

#### Scenario: Local queue is screened before bootstrap

- **WHEN** review uses `specs/queues/<name>.md` instead of a remote GH issue
- **THEN** it screens the queue path and every parent component before reading its Base, Branch, or units

### Requirement: Findings and Verdict

The `litespec-review` skill SHALL report each finding with a **Severity** (`CRITICAL`, `WARNING`, or `SUGGESTION`), a **Location** (`file:line` or unit), **Evidence** (excerpt or observation), and a **Fix direction** (one unambiguous instruction). It SHALL conclude with `PASS` or `CHANGES REQUESTED` after ownership checks pass.

A finding is blocking when it is CRITICAL or WARNING and at least one of: it breaks a unit's `Done means:` or `Verify:`, the issue-owned change contradicts a durable spec or decision, or its location lies inside review scope. SUGGESTIONs and findings outside review scope and every unit contract SHALL be routed without affecting the verdict.

#### Scenario: Pass verdict

- **WHEN** every unit checkbox is checked and no blocking finding remains
- **THEN** review returns `PASS`, even if routed findings accompany it

#### Scenario: Changes requested verdict

- **WHEN** at least one blocking finding remains
- **THEN** review returns `CHANGES REQUESTED`, even if every unit checkbox is checked

#### Scenario: Out-of-scope finding does not block

- **WHEN** a CRITICAL or WARNING lies outside review scope and every unit contract
- **THEN** it is routed and does not affect the verdict

### Requirement: Exhaustive Finding Routing

For each finding, `litespec-review` SHALL apply the first matching rule:
1. A SUGGESTION SHALL be routed to the non-blocking small fix lane at the user's discretion.
2. A CRITICAL or WARNING that breaks a unit's `Done means:` or `Verify:` SHALL be a blocking rebuild: the unit is unchecked and rebuilt with `litespec-build`.
3. A CRITICAL or WARNING inside review scope but outside every unit SHALL be blocking. A trivial finding SHALL route to a direct fix on the issue branch; a non-trivial, correctly shaped finding SHALL be appended as a new unchecked unit on the parent queue and built on the same branch; a shape problem SHALL route to `litespec-plan`.
4. A CRITICAL or WARNING outside review scope and every unit SHALL be non-blocking. A trivial finding SHALL route to the small fix lane; a non-trivial finding SHALL be drafted for a later `litespec-plan` invocation that creates its own queue and isolated branch; a shape problem SHALL route to `litespec-plan`.

A finding that needs a durable ruling SHALL be reported as `"needs decision: <question>"` before applying the matching route. The need for a decision SHALL NOT alter whether the finding blocks.

#### Scenario: Warning breaks a unit

- **WHEN** a WARNING shows that a unit's contract is not satisfied
- **THEN** the unit is unchecked and routed to `litespec-build` as a blocking rebuild

#### Scenario: In-scope finding outside units blocks

- **WHEN** a CRITICAL or WARNING lies inside review scope but breaks no existing unit
- **THEN** it routes to a direct issue-branch fix, a new parent unit, or plan according to its size and shape, and the parent stays open

#### Scenario: Out-of-scope finding routes without blocking

- **WHEN** a CRITICAL or WARNING lies outside review scope and every unit
- **THEN** it routes to the appropriate lane and may coexist with `PASS`

#### Scenario: Finding needs a decision

- **WHEN** a finding requires a durable architectural ruling
- **THEN** review reports `"needs decision"` and preserves the finding's blocking status while the decision is made

### Requirement: Pure Review Role

The `litespec-review` skill MUST NOT write code, modify files, check or uncheck existing checkboxes, or implement fixes. It SHALL report and route findings. Appending a new unchecked unit to the parent GH issue body or local queue is a permitted routing mutation; review MUST NOT otherwise modify the queue.

#### Scenario: Review does not implement

- **WHEN** `litespec-review` runs
- **THEN** it makes no implementation changes

#### Scenario: Parent unit append is routing

- **WHEN** review finds non-trivial, correctly shaped work inside review scope but outside existing units
- **THEN** it appends an unchecked unit to the parent queue without implementing it or changing existing units

### Requirement: Adversarial Scenario Reference

When the change contains stateful code paths, `litespec-review` SHALL load `references/review/adversarial-review.md`, construct worst-case scenarios from the specs, and trace them through implementation. Unconfirmed candidates tagged `Uncertain` SHALL NOT block; confirmed CRITICAL or WARNING findings follow the normal scope and routing rules.

#### Scenario: Adversarial review for stateful code

- **WHEN** the change contains state transitions, multi-entity operations, or concurrent access
- **THEN** review enumerates adversarial scenarios and checks them against the implementation

#### Scenario: Unconfirmed candidate does not block

- **WHEN** an adversarial scenario remains `Uncertain`
- **THEN** it is reported for human triage without affecting the verdict

### Requirement: No Unit for Trivial Findings

`litespec-review` SHALL NOT invent units for trivial findings. A new unit SHALL be drafted only when a finding needs a unit's worth of work and does not break an existing unit contract.

#### Scenario: Trivial in-scope finding

- **WHEN** review finds a trivial issue inside review scope but outside every unit
- **THEN** it routes to a blocking direct fix without creating a unit

#### Scenario: Non-trivial work outside units

- **WHEN** an in-scope finding requires real implementation work that no existing unit covers
- **THEN** review may append an unchecked unit to the parent queue without implementing it

### Requirement: Issue Closure Condition

A GH issue SHALL close only when every unit checkbox is checked and `litespec-review` returns `PASS`. Routed non-blocking findings SHALL NOT prevent closure or reopen the issue.

#### Scenario: Checked units and pass permit closure

- **WHEN** every unit checkbox is checked and review returns `PASS`
- **THEN** the issue may close

#### Scenario: Checked units with blocking finding stay open

- **WHEN** every unit checkbox is checked but review returns `CHANGES REQUESTED`
- **THEN** the issue remains open

### Requirement: No Persistent Finding Tracker

`litespec-review` findings SHALL be ephemeral. A finding SHALL route to an existing unit checkbox, be fixed directly, or become a new queue issue. The review skill SHALL NOT maintain a finding tracker, task list, or persistent finding artifact.

#### Scenario: Re-review reads current state

- **WHEN** `litespec-review` is re-run after fixes
- **THEN** it evaluates the current issue-owned branch, specs, and code and recomputes the verdict without a previous finding log
