Create the tasks artifact at the specified output path.

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
- **When in doubt, aim for ~10 files touched and ~500 lines changed.** This is a soft guideline, not a hard rule — but phases bigger than this risk exhausting the agent's context window.