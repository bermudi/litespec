Create delta spec files in the specs/ directory inside the change.

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
- Only include sections that have requirements — omit empty sections