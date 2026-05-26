Create the design artifact at the specified output path.

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
- Before writing a claim about what existing code does ("X moves from A to B", "Y is deleted", "Z calls W"), re-read the actual source file and verify the claim is true against current code. Do not trust memory from the exploration phase — the file may differ from what you remember.