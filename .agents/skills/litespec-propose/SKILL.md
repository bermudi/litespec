---
name: litespec-propose
description: Materialize a complete change proposal with all planning artifacts (proposal, specs, design, tasks). Use when the user wants to create a new change, start a feature, or says "propose".
---

Enter propose mode. Your job is to materialize a complete change proposal from conversation context and codebase understanding onto disk.

---

## Setup

Ask the user what they want to build. Derive a kebab-case change name from the description.

Then check if it already exists:
```bash
litespec status --change <name> --json
```

**If the change exists**, pick up where it left off — check which artifacts are already done and continue from the next missing one. Do not re-create completed artifacts.

**If the change does not exist**, create it:
```bash
litespec new <name>
```

---

## The Loop

Work through artifacts in dependency order. Repeat until all artifacts are created:

1. **Check status:**
```bash
litespec status --change <name> --json
```
   Response: `{changeName, schemaName, isComplete, artifacts: [{id, outputPath, status, missingDeps}]}`

2. **Get instructions for the next "ready" artifact:**
```bash
litespec instructions <artifact-id> --change <name> --json
```
   Response: `{changeName, artifactId, changeDir, outputPath, description, instruction, template, dependencies: [{id, done, path}], unlocks}`

3. **Read dependency files** listed in `dependencies` — these are inputs that inform the current artifact.

4. **Create the artifact file** at `outputPath`, using the template structure as a guide.

5. **Verify the file exists** after writing it. If it did not land, write it again.

6. **Loop** back to step 1 until `isComplete` is true.

---

## Context and Rules Are Constraints, Not Content

Each artifact comes with `instruction`, `template`, and `dependencies`. Use these to guide your thinking and structure — **do not paste them into artifact files**.

- `instruction` and `template` tell you *what to produce and how to shape it*. They are your brief.
- `dependencies` provide *source material to reference and build on*, not text to copy.
- The artifact file should contain original, purposeful content — not a regurgitation of the instructions.

---

## Behavioral Guardrails

- **Verify every file after writing.** Confirm the artifact landed at `outputPath`. If it did not, write it again before moving on.
- **Decide, do not block.** If the user is vague or a detail is unclear, make a reasonable decision and note what you chose in the artifact. The user can correct it during apply or verify. Momentum matters more than perfection at this stage.
- **Resume, do not restart.** If the change already exists, check status and continue from the first incomplete artifact. Never overwrite completed work.
- **Show a summary when done.** After all artifacts are created, print a brief summary of what was created and the file paths. Then suggest next steps:
  - `apply` to start implementing
  - `verify` to review the proposal against specs

---

## What You Are Doing

Turning conversation and codebase understanding into a structured, actionable change proposal. The four artifacts — proposal, specs, design, tasks — form a contract. Get them on disk, get them right enough, move on.
