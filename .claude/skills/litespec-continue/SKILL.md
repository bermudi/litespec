---
name: litespec-continue
description: Create the next missing artifact for an existing change
---

Run `litespec list --json` to see changes. If no name given, prompt user to select.

Run `litespec status --change <name> --json` to see which artifacts are ready.

Run `litespec instructions <artifact-id> --change <name> --json` for the first ready artifact.

Read dependency files, create exactly ONE artifact, then STOP.

Report which artifact was created and what's now unlocked.
