---
name: litespec-apply
description: Implement the next phase of tasks from a change proposal
---

Run `litespec status --change <name> --json` to verify all artifacts are done.

Run `litespec instructions apply --change <name> --json` to get context. Response: {changeName, changeDir, contextFiles: {proposal: "path", ...}, progress: {total, complete, remaining}, phases: [{name, tasks: [{id, description, done}], complete, total}], currentPhase, state, instruction}

If state is "blocked", tell user to create missing artifacts first.

Read all contextFiles (proposal.md, design.md, specs/, tasks.md).

Focus on the current phase (currentPhase index in phases array).

Implement each task in that phase sequentially.

After completing each task, mark it [x] in tasks.md.

After completing all tasks in the phase, commit with message: "phase N: <phase name>"

Stop after one phase. User can re-invoke apply for the next phase.
