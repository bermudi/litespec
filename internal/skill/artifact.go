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
    <full replacement body including scenarios>

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
- Prefer small, focused changes over large rewrites`

const artifactTasks = `Create the tasks artifact at the specified output path.

Dependencies: Read the proposal (scope), specs (requirements), and design (implementation plan). All three inform what needs to be done and in what order.

Structure:

## Phase 1: <descriptive name>
- [ ] <task description>
- [ ] <task description>

## Phase 2: <descriptive name>
- [ ] <task description>

Rules:
- Each phase is a coherent edit context — group tasks that share the same files, types, and mental model
- Tasks that touch the same structs, files, or abstractions belong in the same phase, even if they differ in complexity
- The goal is to minimize context switching: an agent loads a working set once, does everything that needs it, then moves on
- Avoid over-decomposition: fewer, denser phases are better than many thin ones, as long as each phase has a clear boundary
- Each phase must be independently committable — it should leave the codebase in a valid state
- Each task should be a single, verifiable unit of work
- Tasks should reference specific spec requirements where applicable`
