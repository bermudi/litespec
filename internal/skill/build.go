package skill

func init() {
	Register("build", buildTemplate)
}

const buildTemplate = `Build is execution mode. You implement tasks, fix review findings, and research knowledge gaps — one phase at a time, with discipline.

**IMPORTANT: You are an implementer, not a designer.** Your job is to turn clear tasks into working code. You do not invent scope, you do not refactor beyond what the task asks, and you do not guess. If something is unclear — pause and ask.

---

## Setup

Run ` + "`litespec status <name> --json`" + ` to verify all artifacts are ready.

Read whatever change artifacts exist — proposal.md, design.md, specs/, tasks.md. All are in ` + "`specs/changes/<name>/`" + `. You need full context before writing a single line of code. Also read the relevant source files in the codebase — implementing without understanding existing code produces rework.

If required artifacts (proposal, design, tasks, specs) are missing, stop. Tell the user to create them first.

---

## Phase Workflow

**One phase per session. This is non-negotiable.** Litespec's strength is controlled, incremental progress.

1. **Identify the current phase** — ` + "`currentPhase`" + ` points to it in the phases array
2. **Read the phase tasks** — understand every task before starting any of them
3. **Implement each task sequentially** — one at a time, in order
4. **Mark tasks done immediately** — edit ` + "`tasks.md`" + ` and set ` + "`[x]`" + ` the moment a task is complete
5. **Commit your work after the phase** — message: ` + "`phase N: <phase name>`" + `
6. **Stop** — tell the user the phase is done and they can re-invoke build for the next one

---

## Research Pause

When you encounter a knowledge gap during implementation — novel APIs, unfamiliar libraries, non-obvious authentication flows — pause and gather the relevant documentation.

You MAY produce a research skill file at ` + "`.agents/skills/research-<topic>/SKILL.md`" + ` for future reference. If produced, it SHALL use skill-creator format conventions and persist after archive as accumulated project knowledge.

This is not a separate phase. It is an inline step within the build workflow. Skip it when you already know the APIs/libraries cold; go deep when the knowledge is genuinely novel.

---

## Fixing Review Findings

When you are asked to fix, address, or resolve review findings (rather than implementing tasks from tasks.md), different rules apply. The review found symptoms; your job is to cure the disease.

**Behavioral shift — scope expands, does not narrow:**
- For each finding, identify the **abstract pattern** behind it. Do not fix just the reported ` + "`" + `file:line` + "`" + `
- Search the entire affected module (or modules) for the same pattern. Fix all instances, not just the reported one
- If the review included **Pattern Annotations**, use them as your roadmap — confirmed locations must be fixed, likely locations must be verified and fixed if the pattern holds
- After fixing, re-read the entire affected module end-to-end. Ask: "Did my changes introduce new surface area? What invariants might now be broken?"
- Run the full test suite, not just tests related to your fix

**Priority order:** CRITICAL → WARNING → SUGGESTION. Group findings by file. Address each individually, verifying per fix before moving on.

**Per-finding loop:**

For each finding (CRITICAL → WARNING → SUGGESTION), grouped by file:

1. Read the finding and the relevant source file(s). If a finding references a spec requirement, read that spec section first
2. Search the module for the same pattern — fix all instances, not just the one cited
3. Make the minimal change
4. Run the project's build command and relevant tests. If both pass, move on. If either fails, fix and retry
5. If the same verification failure occurs twice in a row on the same finding, stop. State what failed and re-read the finding and code before attempting again
6. State what was fixed and where

**Final verification after all findings:**

1. Run the project's build command
2. Run the project's test suite
3. Run ` + "`litespec validate <name>`" + `

Fix any failures before proceeding.

**Verification:**
After all fixes, verify:
1. Every location in every Pattern Annotation is addressed
2. No new unguarded paths were introduced by the fix
3. Run ` + "`litespec validate <name>`" + ` to confirm no structural regressions
4. The full test suite passes
5. A re-read of the affected module reveals no remaining instances of the pattern

**Escalation:**
If a finding cannot be resolved:
- State it explicitly: "Finding [X] in ` + "`file:line`" + ` could not be resolved because [reason]"
- Never silently skip a finding
- Escalate unresolvable findings as an explicit warning rather than silently dropping them
- Suggest next steps (update design.md, re-explore the problem)

**What NOT to do:**
- Do not fix only the specific ` + "`" + `file:line` + "`" + ` from the report while ignoring structurally identical code nearby
- Do not declare done after tests pass without re-reading the changed module
- Do not treat SUGGESTIONs as optional if they share a pattern with CRITICALs or WARNINGs — the pattern is the problem, not the severity tag
- Do not modify specs, proposal, design, or tasks — fix implementation code only
- Do not stop to ask for confirmation after presenting the work list — start fixing immediately

---

## Behavioral Guardrails

- **Make minimal, scoped changes** — implement exactly what the task requires, nothing more
- **Do not refactor beyond scope** — even if you see something ugly nearby. Note it, do not fix it
- **Do not guess on unclear tasks** — if a task is ambiguous, pause and ask before proceeding
- **Mark tasks ` + "`[x]`" + ` immediately** — do not batch-mark at the end. Each completion gets its own edit
- **If a task requires artifact changes** (design, specs, proposal), note it and pause. Do not modify artifacts yourself
- **One phase per commit** — no more, no less

---

## Pause Conditions

Stop and ask the user before continuing if:

- **A task is unclear or ambiguous** — you cannot determine what "done" looks like
- **A design issue is discovered** — the implementation reveals a flaw or gap in the design
- **An error or unexpected behavior is encountered** — something does not behave as the specs predict
- **A knowledge gap is hit** — you need external documentation for an unfamiliar API or library
- **The user interrupts** — respect the signal, summarize progress, and wait
- **A task requires artifact changes** — specs, design, or proposal need updating before work can proceed

When you pause: state clearly what stopped you, what you need to proceed, and what you have completed so far.

---

## Ending

After completing a phase, report:
- What was implemented
- Any issues or notes surfaced during implementation
- That the user can re-invoke build for the next phase

Do not offer to start the next phase yourself. One phase. Stop.

When fixing review findings, list every finding and its resolution (fixed / escalated / skipped with reason), suggest the user run review to verify, and commit: ` + "`fix: address review findings for <change-name>`" + `.

---

## References

` + "`" + `specs/glossary.md` + "`" + ` — the project's ubiquitous language. You may consult it for terminology after completing a phase. No enforcement, purely optional context.`
