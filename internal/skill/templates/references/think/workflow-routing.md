The user wants to know where they are in the litespec workflow and what to do next.

**The workflow is unidirectional:**

```
explore → grill → propose → apply → review → archive
                                    │
                                adopt (separate path)

patch → archive  (lightweight lane for small, single-capability changes)
```

Detect the user's current state:

```bash
litespec list --json
```

**Interpreting litespec list --json:**
- changes[].status: "in-progress" = active, "complete" = ready to archive
- changes[].completedTasks / totalTasks: 0/0 = draft, N/M = active, M/M = ready
- changes[].lastModified: use to find the most recently touched change

**No project exists** — the user needs `litespec init`.

**Project exists but no changes** — read `references/onboarding.md` to distinguish first-time users from experienced users.

**Changes exist** — find the most relevant change and explain its current phase:
- **No tasks yet (draft)**: totalTasks == 0. Next: write tasks.md or use litespec-plan.
- **Tasks exist, not all done (active)**: Next: litespec-build for the current phase.
- **All tasks done (ready to archive)**: Next: litespec-review, then the user runs `litespec archive <name>`.

**Key gotchas:**
- explore and grill are ephemeral — no artifacts. To save thinking, move to propose.
- propose is the commit point — once artifacts exist, the plan is committed.
- Phase tracking comes from tasks.md checkboxes — the first unchecked block is the current phase.
- Archive is a human decision — the agent never runs `litespec archive`.

When the user asks "what do I do next?", use this response template:

> **Current state:** [X active changes, Y ready to archive]
> **Most relevant:** [change-name] at [N/M tasks]
> **Next step:** [specific skill]
> **Why:** [brief reason]

Common questions — read `references/faq.md` when the user asks workflow questions.
