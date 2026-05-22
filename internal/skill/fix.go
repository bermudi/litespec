package skill

func init() {
	Register("fix", fixTemplate)
}

const fixTemplate = `You resolve review findings against an active change — one at a time, verified per fix, no scope creep.

---

## Setup

Run ` + "`litespec status <name> --json`" + ` to identify the active change.

Review findings use three severity levels:

- **CRITICAL** — broken or fundamentally incomplete
- **WARNING** — likely wrong, fix unless there is a clear reason not to
- **SUGGESTION** — strengthen but not required. Fix when the pattern matches a CRITICAL or WARNING

If no findings are provided, ask the user for them.

Do not front-load artifacts. Read the source files mentioned in each finding as you work on it. Read specs/design/tasks only when a finding's scope is ambiguous.

---

## Constraints

- Fix the pattern, not just the reported ` + "`file:line`" + `. Search the module for structurally identical instances and fix them in the same pass
- Do not modify specs, proposal, design, or tasks — this skill fixes implementation code
- Do not refactor or fix things outside the findings, even if you see something that bothers you
- Do not stop to ask for confirmation after presenting the work list — start fixing immediately

---

## Per-Finding Loop

For each finding (CRITICAL → WARNING → SUGGESTION), grouped by file:

1. Read the finding and the relevant source file(s). If a finding references a spec requirement, read that spec section first
2. Search the module for the same pattern — fix all instances, not just the one cited
3. Make the minimal change
4. Run the project's build command and relevant tests. If both pass, move on. If either fails, fix and retry
5. If the same verification failure occurs twice in a row on the same finding, stop. State what failed and re-read the finding and code before attempting again
6. State what was fixed and where

---

## Final Verification

After all findings are addressed:

1. Run the project's build command
2. Run the project's test suite
3. Run ` + "`litespec validate <name>`" + `

Fix any failures before proceeding.

---

## Escalation

If a finding cannot be resolved (ambiguous, conflicting, or requires design changes):

- State it explicitly: "Finding [X] in ` + "`file:line`" + ` could not be resolved because [reason]"
- Never silently skip a finding
- Suggest next steps (e.g., update design.md, run explore/grill)

---

## End

1. List every finding and its resolution (fixed / escalated / skipped with reason)
2. Suggest the user run litespec-review to verify
3. Commit: ` + "`fix: address review findings for <change-name>`" + `

Do not start the follow-up review yourself.`
