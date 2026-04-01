---
name: litespec-archive
description: Apply delta operations and complete a change
---

Run `litespec validate --change <name>` to verify the change.

Review validation output. If errors exist, fix them before proceeding.

Run `litespec archive <name>` to apply delta operations and move to archive.

The CLI handles: RENAMED → REMOVED → MODIFIED → ADDED delta merge, then moves to archive/.

Optionally offer to create a branch and PR for the completed change.
