---
name: the-drill
description: 'Run the full end-of-session release ritual — commit, archive, version bump, release, verify upgrade. Use when the user says "you know the drill", "do the drill", or wants to ship and verify. This skill is the final step after a change is implemented: it commits any pending work, archives the change (marking it as implemented), bumps the version, creates a GitHub release, and verifies the upgrade works locally.'
---

You are running the drill. This is the end-of-session ship ritual for litespec. Execute each step in order. Do not skip steps. Do not skip the verification at the end.

First, determine which flow you're in:

- **Change flow**: There is an open change in `specs/changes/` that was completed this session. Steps: commit → archive → version bump → release → verify upgrade.
- **Hotfix flow**: There is no open change — just a bug fix, refactor, or other work done outside the spec workflow. Steps: commit → version bump → release → verify upgrade.

Check session context (recent commands, conversation history) and `specs/changes/` to determine the flow. If a change exists and is complete, use the change flow. If `specs/changes/` is empty or the change was already archived, use the hotfix flow.

---

## Step 1: Commit pending changes

Run `git status` and `git diff`. If there are modified or untracked files that belong to the current work, commit them with a descriptive message. If the working tree is clean, move on.

Do not commit files unrelated to the current work (e.g., editor configs, temp files).

## Step 2: Archive (change flow only)

Skip this step entirely in the hotfix flow.

This archives the specific change that was implemented in this session — not any other change. Archiving merges deltas into canonical specs and moves the change to the archive, marking it as implemented.

1. Run `litespec validate <name>` — if errors, fix them
2. Run `litespec archive <name>` to merge delta specs and move to archive
3. Commit the archive: `git add specs/ && git commit -m "archive: <name>"`

If the change was already archived in a previous step (check `git log`), skip this step.

## Step 3: Version bump and release

1. Get the current version: `git tag --sort=-v:refname | head -1`
2. Ask the user what the next version should be (patch, minor, or major). Suggest the appropriate bump based on what changed — patches for bugfixes, minor for features, major for breaking changes.
3. Create the git tag: `git tag <version>`
4. Push main first: `git push origin main`
5. Push the tag: `git push origin <version>`
6. Draft release notes from `git log <prev-version>..<new-version> --oneline`. Write them in the project's existing style:

```
## What's Changed

- **Feature area**: concise description of what changed
- Another change
```

Use bold for the feature area when it maps to a distinct capability. Plain text for smaller changes.

7. Create the release:
```
gh release create <version> --title "<version>" -n "<release notes>"
```

## Step 4: Verify upgrade

This is the most important step. Do not skip it.

The Go module proxy caches versions and may not have the new tag immediately. The `upgrade` command installs the explicit tag with `GOPROXY=https://proxy.golang.org,direct` fallback, but you still need to verify the binary actually changed.

1. Run `litespec upgrade`
2. Run `litespec --version`
3. Confirm the output matches the version you just released

If `--version` still shows the old version, the proxy hasn't propagated yet. Wait a moment and try:
```
GOPROXY=https://proxy.golang.org,direct go install github.com/bermudi/litespec/cmd/litespec@<version>
```
Then verify again. Do not report success until `litespec --version` shows the new version.

---

## Completion

Report the result:
- What was committed
- What was archived (if change flow)
- Version released
- Confirmation that upgrade verification passed
