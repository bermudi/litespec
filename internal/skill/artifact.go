package skill

func init() {
	Register("artifact-proposal", artifactProposal)
	Register("artifact-specs", artifactSpecs)
	Register("artifact-design", artifactDesign)
	Register("artifact-tasks", artifactTasks)
}

const artifactProposal = `Create the proposal artifact at the specified output path.

This is the first artifact — it has no dependencies. Build it from the conversation context and codebase exploration.

Structure:

## Motivation
Why this change is needed. What problem does it solve or what opportunity does it address?

## Scope
What is included in this change. Be specific about:
- Which capabilities are affected
- What behavior changes
- What new functionality is introduced

## Non-Goals
What is explicitly NOT included. This prevents scope creep and sets clear boundaries.

Rules:
- Be concrete, not aspirational — describe a specific change, not a wishlist
- Keep it focused — if the scope feels too large, suggest splitting into multiple changes
- The proposal informs everything downstream: specs describe the scope in detail, design explains how to implement it, tasks break it into phases`

const artifactSpecs = `Create delta spec files in the specs/ directory inside the change.

Dependencies: Read the proposal to understand scope and affected capabilities.

For each affected capability, create a spec.md file under specs/<capability>/ with delta requirements using this format:

    # <capability>
    ## ADDED Requirements
    ### Requirement: <name>
    <body text — must contain SHALL or MUST>

    #### Scenario: <short name>
    - **WHEN** <condition>
    - **THEN** <expected outcome>

    ## MODIFIED Requirements
    ### Requirement: <name>
    <write only what should exist after the change — unchanged parts you want to preserve, plus the changed behavior, including scenarios>

    ## REMOVED Requirements
    ### Requirement: <name>

    ## RENAMED Requirements
    ### Requirement: <old> → <new>

Note: Delta specs use operation headers (## ADDED/MODIFIED/REMOVED/RENAMED Requirements). The canonical specs in specs/canon/ use ## Requirements and optionally ## Purpose. Deltas are merged into canonical specs at archive time.

Rules:
- Every ADDED and MODIFIED requirement must include at least one #### Scenario: block
- REMOVED requirements are name-only — no body or scenarios
- RENAMED requirements change the heading only; content and scenarios carry over under the new name
- Body text for ADDED and MODIFIED requirements must contain SHALL or MUST
- Read existing main specs in specs/canon/ to understand what already exists before writing deltas
- Only include sections that have requirements — omit empty sections`

const artifactDesign = `Create the design artifact at the specified output path.

Dependencies: Read the proposal (motivation/scope) and specs (requirements) to inform the technical approach.

Structure:

## Architecture
How the change fits into the existing system. Component relationships, data flow, and state management.

## Decisions
Key technical decisions and their trade-offs. For each:
- What was chosen
- Why it was chosen over alternatives
- What constraints or assumptions it introduces

## File Changes
Concrete list of files that will be created, modified, or deleted. For each:
- Path
- What changes and why
- How it relates to the spec requirements

Rules:
- Be specific about file paths — vague paths like "a new file in internal/" are not actionable
- Reference spec requirements by name so the link is traceable
- If the change touches existing code, describe the impact on callers
- Prefer small, focused changes over large rewrites
- Before writing a claim about what existing code does ("X moves from A to B", "Y is deleted", "Z calls W"), re-read the actual source file and verify the claim is true against current code. Do not trust memory from the exploration phase — the file may differ from what you remember.`

const artifactTasks = `Create the tasks artifact at the specified output path.

Dependencies: Read the proposal (scope), specs (requirements), and design (implementation plan). All three inform what needs to be done and in what order.

Structure:

## Phase 1: <descriptive name>
- [ ] <task description>
- [ ] <task description>

## Phase 2: <descriptive name>
- [ ] <task description>

Rules:
- Each phase is a commit boundary — it must leave the codebase in a valid, buildable, test-passing state
- A valid phase ends with a commit message that describes one thing: 'phase 2: Add delta merge logic'
- If a phase wouldn't survive 'go build && go test' (or your project's equivalent), it's incomplete — add verification tasks
- Each task should be a single, verifiable unit of work
- Tasks should reference specific spec requirements where applicable

Phase sizing — not too fat, not too thin:
- **A phase must change behavior.** If it only does cleanup, docs, or test backfill without introducing or modifying functional code, fold it into the phase that created that code. Tests and docs belong with the code they cover.
- **One sentence without "and."** If the phase name needs "and" to describe what it does, it is probably two phases. "Add delta parser" is one phase. "Add delta parser and wire up validation and update CLI" is three.
- **Stay within 2–3 packages.** If a phase requires reading and modifying code across more packages than that, the agent will lose context. Split it.
- **Group by shared files and mental model.** "Add parser and validator" is one phase if both touch the same types. Split when the commit message would become a list.
- **When in doubt, aim for ~10 files touched and ~500 lines changed.** This is a soft guideline, not a hard rule — but phases bigger than this risk exhausting the agent's context window.`
