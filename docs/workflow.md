# Workflow

The litespec workflow is unidirectional — you move forward through phases, never backward. This design prevents the "drifting proposal" problem where plans and implementation get out of sync.

```
explore → grill → propose → review → apply → review → archive
                                          │
                                      adopt (separate path)
```

Each step has a clear purpose and produces specific artifacts. Choose the right pattern for your situation.

## Workflow Steps

### explore: Thinking Mode

**What happens:** Conversational exploration of ideas, problems, or directions. No artifacts are created.

**What the AI does:** Reads code, asks questions, draws diagrams, investigates architecture. It's a thinking partner, not an implementer.

**Artifacts created:** None (ephemeral context kept in the AI's window)

**When to use:** When you have a vague idea, need to investigate the codebase, or want to think through a problem before committing to a change.

**Example:** "Thinking about adding shell completions. How would that work with our CLI structure?"

---

### grill: Stress-Testing Mode

**What happens:** Relentless Q&A about a plan or design. Every branch of the decision tree is resolved before proceeding.

**What the AI does:** Interviews you about tradeoffs, risks, edge cases, and assumptions. It explores the codebase to answer questions when possible, asks when not.

**Artifacts created:** None (ephemeral)

**When to use:** When a design decision, architecture choice, or plan would benefit from structured interrogation. Not every question needs grilling — but major changes should be stress-tested.

**Example:** "Grill me on this caching design. I want to find the holes before we implement it."

---

### propose: Materialization Mode

**What happens:** Creates a complete change proposal with all planning artifacts. This is the commit point.

**What the AI does:**
1. Creates a change directory: `specs/changes/<name>/`
2. Generates all artifacts in dependency order:
   - `proposal.md` — motivation, scope, non-goals
   - `specs/` — delta specs describing what changes
   - `design.md` — technical decisions and architecture
   - `tasks.md` — phased implementation checklist

**Artifacts created:** Complete proposal (proposal, specs, design, tasks)

**When to use:** When you're ready to create a change and have clarity on what you want to build.

**Example:** "Propose a docs-site feature using MkDocs."

---

### apply: Implementation Mode

**What happens:** Implements tasks one phase at a time. Each phase is one agent session, one commit.

**What the AI does:**
1. Reads all artifacts (proposal, specs, design, tasks)
2. Identifies the current phase (first phase with unchecked tasks)
3. Implements each task sequentially
4. Marks tasks complete as they finish
5. Commits after the phase: `phase N: <phase name>`

**Artifacts created:** Code changes, commits

**When to use:** When you're ready to write code and all planning artifacts are complete.

**Example:** "Apply" the docs-site change. Let's start with Phase 1.

---

### review: Review Mode

**What happens:** Context-aware AI review that adapts to the change lifecycle.

**What the AI does:** Detects task completion state and chooses an appropriate review:
- **Artifact review** (0 tasks checked): Evaluates proposal, specs, design, tasks for quality, consistency, and readiness
- **Implementation review** (some tasks checked): Compares implemented code against specs
- **Pre-archive review** (all tasks checked): Comprehensive review of both artifacts and code

**Artifacts created:** Review report with CRITICAL, WARNING, SUGGESTION findings

**When to use:** Before starting implementation (artifact review), during implementation (implementation review), or before archiving (pre-archive review).

**Example:** "Review" the docs-site change. We're in Phase 2 and I want to check if the code matches the specs.

---

### archive: Finalization Mode

**What happens:** Validates task completion, merges delta specs into the canonical source of truth, and moves the change to archive.

**How to run it:** Use the CLI commands directly — no dedicated skill is needed:
1. `litespec validate <name>` — verify artifacts exist, delta syntax is valid, no dangling deltas
2. `litespec archive <name>` — applies deltas, strips the change's `specs/` subtree, and moves to `specs/changes/archive/YYYY-MM-DD-<name>/`

You can also pass `--allow-incomplete` to bypass the task-completion check.

**Artifacts created:** Updated canonical specs (`specs/canon/`), archived change

**When to use:** When all implementation is done and you're ready to finalize the change.

**Example:** `litespec archive docs-site` — all phases are complete.

---

### adopt: Reverse-Engineering Mode (Separate Path)

**What happens:** Reverse-engineers specs from existing code. A separate workflow for documenting capabilities that already exist.

**What the AI does:**
1. Reads the provided file or directory thoroughly
2. Builds a mental model of the code's purpose, dependencies, and behavior
3. Creates a change proposal with specs documenting discovered capabilities
4. Generates proposal, specs, design, and tasks

**Artifacts created:** Complete proposal describing existing code

**When to use:** When you have code that needs spec'ing but no spec exists yet. This is the on-ramp for existing codebases.

**Example:** "Adopt" the auth package. It's well-tested but has no spec.

---

## Named Workflow Patterns

### Quick Feature Pattern

**Flow:** propose → apply → archive

**When to use:** Simple, well-understood features where you know what you want to build without exploration.

**Example:** Adding a `--json` flag to an existing command.

**Why skip explore/grill:** The change is straightforward — adding a flag doesn't require architectural thinking or stress-testing.

---

### Exploratory Pattern

**Flow:** explore → grill → propose → review → apply → review → archive

**When to use:** Complex features, architectural changes, or anything with significant uncertainty.

**Example:** "Thinking about a real-time collaboration feature. How would this work with our existing data model?"

**Why use explore:** You need to investigate the codebase, understand integration points, and think through tradeoffs.

**Why use grill:** The design decisions matter — getting the architecture wrong would be expensive to fix.

---

### Adopt Pattern (Separate Path)

**Flow:** adopt → archive

**When to use:** Documenting existing code that has no spec yet. This is the reverse of the normal workflow — you're extracting specs from code rather than writing code from specs.

**Example:** "We have a config parser that works but no tests. Adopt it to understand what it does, then we can review it."

**Why adopt first:** The code already exists. You need to understand it before you can review or improve it.

---

## Decision Flow

Which pattern should you use?

```
Is this existing code without a spec?
│
└─ Yes → Use adopt → archive

No → How much uncertainty do you have?
│
├─ Zero → Use Quick Feature (propose → review → apply → review → archive)
│
└─ Some or a lot → Use Exploratory (explore → grill → propose → review → apply → review → archive)
```

**Guidelines:**
- If the feature is a clear, scoped addition (e.g., add a flag), skip explore/grill.
- If the change affects architecture, data models, or integration points, use explore.
- If the design has tradeoffs or risks that need stress-testing, use grill.
- If you're not sure, start with explore. It's always safe to explore.

---

## Why No Backward Flow?

The workflow is unidirectional for a reason. If you discover a problem during apply:

```
Wrong approach during apply?
│
├─ Do NOT edit proposal/specs/design to match the code
│
└─ START OVER from explore/grill → create a new proposal
```

**Why?** Because if you update artifacts to match what you implemented, you've lost the contract. The artifacts should describe what you *intended*, not what you accidentally built. If the implementation reveals design flaws, start over with better planning.

The archive step enforces this: it validates that all tasks are complete before merging specs. You can't archive a half-baked change. This discipline prevents gradual drift where code and specs diverge over time.

---

## Real Example: The docs-site Change

The `docs-site` change in this repo followed the Exploratory pattern:

1. **explore:** Investigated MkDocs alternatives, sketched directory structure, debated manual vs auto-generated docs
2. **grill:** Stress-tested the choice of MkDocs Material, questioned the scope (why not add search now?), verified that docs-as-source-of-truth made sense
3. **propose:** Created a complete proposal with specs, design, and tasks
4. **apply:** Implemented in three phases — infrastructure (pyproject.toml, mkdocs.yml), content (8 doc pages), deployment (GitHub Actions)
5. **review:** Ran artifact review, implementation review, and pre-archive review
6. **archive:** Merged delta specs into `specs/canon/docs-site/spec.md` and moved to archive

This change had significant uncertainty (which docs engine? what scope? deployment strategy?), so it benefited from the full exploratory workflow. A simpler change like `shell-completions` could have used Quick Feature.

---

## Next Steps

- [Tutorial](tutorial.md) — worked walkthrough of a complete change from init to archive
- [Concepts](concepts.md) — philosophy behind spec-driven development
- [CLI Reference](cli-reference.md) — command details
