# review

## Requirements

### Requirement: Adversarial Review of Issue and Spec vs Implementation

The `litespec-review` skill SHALL perform a context-aware, adversarial review by reading the GH issue body, its units with `Done means:` and `Verify:`, any load-bearing `specs/<feature>/spec.md`, relevant `specs/decisions/`, and the exact issue-owned review scope. It SHALL probe for interaction bugs, state transitions, wiring gaps, and contract violations, not only syntax or surface compliance.

#### Scenario: Review with GH issue and spec

- **WHEN** `litespec-review` is invoked for a change with a GH issue and a load-bearing spec
- **THEN** it reads the issue, spec, decisions, and issue-owned changes before probing the implementation

#### Scenario: Review for small fix without issue

- **WHEN** `litespec-review` is invoked on a small fix with no GH issue
- **THEN** it requires the user to identify the fix commit, screens that commit's paths before content access, and reviews approved per-path changes rather than inferring scope from a dirty tree

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

Before reviewing a queue issue, `litespec-review` SHALL verify that `Base:` and `Branch:` exist, that the current branch matches `Branch:`, and that `Base:` is an ancestor of `HEAD`. Review SHALL enumerate tracked and untracked path names without reading contents, using NUL-delimited Git output. It SHALL reject paths outside the repository, known secret-like names, symlinks, and non-regular files before content inspection. File-type inspection MUST NOT follow symlinks. After every path passes screening, review scope SHALL contain each tracked diff and each untracked regular file. Review SHALL stop without a verdict when ownership or safe inspection cannot be proved.

#### Scenario: Scope includes tracked and untracked work

- **WHEN** the ownership checks pass and the issue branch contains tracked changes and `??` paths
- **THEN** review screens all path names and file types first, then examines each safe tracked diff and untracked regular file as issue-owned work

#### Scenario: Branch mismatch stops review

- **WHEN** the current branch differs from the issue's `Branch:` line
- **THEN** review stops without a verdict and does not infer another scope

#### Scenario: Missing ownership metadata stops review

- **WHEN** `Base:` or `Branch:` is absent or Base is not an ancestor of HEAD
- **THEN** review stops without a verdict

#### Scenario: Unsafe path stops before content access

- **WHEN** review scope contains a secret-like path, symlink, path outside the repository, or non-regular file
- **THEN** review stops without a verdict, reports only the path and reason, and does not read contents or follow a link target

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
