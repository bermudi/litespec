package skill

func init() {
	Register("grill", grillTemplate)
}

const grillTemplate = `Interview the user relentlessly about a plan or design until reaching shared understanding.

If this grill call follows exploration in the current session, build on what was already discussed. Do not re-explore from scratch.

Think hard about the implications of each question before asking and use your expertise to guide.

Resolve each branch of the decision tree, one question at a time.

Provide your recommended answer for each question.

**Read code, do not speculate.** If a question can be answered by reading the codebase, read the code instead of asking. If you're unsure how something works, grep for it before posing it as a question to the user.

When a locked architectural ruling emerges that is broader than the current change, suggest creating a decision via ` + "`litespec decide <slug>`" + ` rather than burying it in design.md.

**Backlog scope challenge:** If ` + "`specs/backlog.md`" + ` exists, read it and challenge scope overlaps between the current plan and parked items.

**Language before architecture.** If ` + "`specs/glossary.md`" + ` exists, read it at session start. Before diving into implementation questions, surface and resolve terminology gaps — undefined terms the user relies on, fuzzy definitions, glossary entries that may not match the current codebase. Pin down shared language first; precise terms make architectural questions sharper and shorter. When a new term crystallizes from the discussion, nudge: "This looks like a term for the glossary — want me to add it?" When a glossary term seems misaligned with code, flag it.

When the plan is fully resolved, offer to proceed to propose.`
