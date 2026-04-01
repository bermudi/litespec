---
name: litespec-continue
description: Create exactly one missing artifact for an existing change, then stop. Use when the user wants to fill in the next missing piece of a change or says "continue".
---

Continue the next missing artifact for an existing change. Create exactly ONE artifact, then STOP.

---

## Setup

Run `litespec list --json` to see changes. If no name given, prompt user to select.

Run `litespec status --change <name> --json` to get artifact states.
Response: {changeName, schemaName, isComplete, artifacts: [{id, outputPath, status, missingDeps}]}

---

## State Branching

Check the response and branch:

### isComplete is true
All artifacts exist. Do NOT create anything. Tell the user:
- The change is fully planned
- Suggest next steps: apply (start implementing) or archive (finalize)

### At least one artifact is "ready"
Pick the **first** ready artifact. This is your one task.

Run `litespec instructions <artifact-id> --change <name> --json` for that artifact.
Response: {changeName, artifactId, changeDir, outputPath, description, instruction, template, dependencies: [{id, done, path}], unlocks}

### All remaining artifacts are "blocked"
Nothing can be created. Tell the user exactly which artifacts are blocked and which dependencies are missing:
- "Artifact X needs Y and Z to exist first"
- Suggest they check if earlier artifacts were created correctly

---

## Creating the Artifact

1. Read every dependency file listed in `dependencies` — these are your inputs
2. Follow the `instruction` and `template` structure exactly
3. Write the file to `outputPath` within the change directory

### Context and rules

The `instruction` field may include **context** (background info) and **rules** (constraints). These are guardrails for YOU — they shape how you think about and construct the artifact. Do NOT copy them verbatim into the output file. The output file should contain the artifact content itself.

---

## THE RULE

**Create exactly ONE artifact. Then STOP.**

Do not chain into the next artifact. Do not check if more are ready. Do not loop. One artifact. Done.

Report:
- Which artifact was created
- What artifacts it unlocked (check `unlocks` from the instructions response)
- That the user should invoke continue again for the next one
