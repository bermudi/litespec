package skill

func init() {
	Register("workflow", workflowTemplate)
}

const workflowTemplate = `Explain the litespec workflow and help the user understand where they are in it.

**The workflow is unidirectional:**

` + "```" + `
explore → grill → propose → [research →] apply → review → archive
                                          │
                                      adopt (separate path)
` + "```" + `

**explore** — Think freely. No artifacts. Read code, ask questions, map architecture. Never implement.
**grill** — Stress-test a plan. Relentless Q&A on tradeoffs, risks, edge cases. No artifacts.
**propose** — Materialize the change. Creates proposal.md, specs/, design.md, tasks.md.
**research** — Gather external knowledge (APIs, libraries, schemas). Produces research skills. Optional.
**apply** — Implement one phase at a time. One session per phase, one commit per phase.
**review** — Context-aware review: artifacts (pre-impl), code vs specs (during), both (pre-archive).
**archive** — Apply deltas to canonical specs and move the change to archive. The commit to implemented.
**adopt** — Reverse-engineer specs from existing code. Separate path, does not use propose/apply.

---

## Progressive Discovery

Do not dump the full workflow on the user. Detect their current state and explain what matters next.

### Detect state

Run these commands silently:

` + "```bash" + `
litespec list --json
litespec status --json
` + "```" + `

### If no project exists

The user needs ` + "`litespec init`" + `. Explain that init creates the specs directory and generates skills for their AI tools.

### If project exists but no changes

The user is at the start. Explain explore → grill → propose as the path to creating their first change. Mention adopt as an alternative if they have existing code to document.

### If changes exist

Find the most relevant change (user-mentioned, or most recently touched) and explain its current phase:

**No tasks yet (draft)** — The change was created but not proposed. Next: write artifacts or run litespec-propose.

**Tasks exist, not all done (active)** — The change is being implemented. Show progress and identify the current phase (first unchecked tasks block). Next: litespec-apply for that phase.

**All tasks done (ready to archive)** — Implementation is complete. Next: litespec-review then litespec-archive.

### If archived changes exist

Point to the canonical specs as the source of truth. Changes in specs/canon/ describe the implemented system.

---

## When the user asks "what do I do next?"

1. Check project state
2. Identify the most relevant active change
3. Tell them the exact next command or skill to invoke
4. Briefly explain why that phase comes next

Example: "You have one active change (auth-refactor) at 2/4 phases. The next unchecked tasks are in Phase 3. Run litespec-apply to implement them."

---

## Common questions

**"Can I skip explore/grill?"** — Yes. If you already know what you want, go straight to propose.

**"Something is wrong after propose, can I edit?"** — No backward flow. Start over from explore/grill. This prevents drift between plan and implementation.

**"What is research for?"** — External knowledge gaps. Skip it if you know the APIs/libraries cold.

**"When do I review?"** — Three times: after propose (artifacts), during apply (code vs specs), before archive (both). The review skill adapts automatically.

**"What is adopt?"** — A separate path. Give it code, it reverse-engineers specs. No propose, no apply, no archive.`
