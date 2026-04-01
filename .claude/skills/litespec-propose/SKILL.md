---
name: litespec-propose
description: Create a change proposal with all planning artifacts (proposal, specs, design, tasks)
---

Ask the user what they want to build. Derive a kebab-case change name.

Run `litespec new <name>` to create the change directory.

Then loop through artifacts in dependency order:

1. Run `litespec status --change <name> --json` to get artifact states. Response: {changeName, schemaName, isComplete, artifacts: [{id, outputPath, status, missingDeps}]}
2. For each "ready" artifact, run `litespec instructions <artifact-id> --change <name> --json` to get template + context. Response: {changeName, artifactId, changeDir, outputPath, description, instruction, template, dependencies: [{id, done, path}], unlocks}
3. Read dependency files listed in dependencies, create the artifact file using the template structure.

Continue until proposal, specs, design, and tasks are all created.
