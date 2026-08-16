Use when the idea is sharp and you need to nail the GH issue (+ spec if load-bearing). This is the clear mode of `plan`.

## What to write

1. **GH issue body — the queue.** One `##` per unit. Each unit:
   ```
   ## <one demo-able outcome>
   Done means: <observable, human-checkable>
   Verify: `<command that fails without the outcome>`
   - [ ] pending
   ```
   `Verify:` must fail for a plausible state where the outcome is missing. A `go test` that doesn't check output is not a Verify.

2. **Spec if load-bearing.** If the feature is a promise that breaks things when wrong (CLI shape, API, file format), edit `specs/<feature>/spec.md` directly in the same change — not a delta. Keep to 3-5 SHALL requirements, each with a WHEN/THEN scenario.

3. **Overflow only if needed.** If the shape won't fit in the issue body, add `specs/changes/<name>/proposal.md` (why/what) and `design.md` (how). Otherwise no files in `specs/changes/`.

## Rules

- One unit = one thing you can demo. If you can't demo it, split it.
- One Verify per unit, and that Verify is the gate — `build` must satisfy it before claiming done.
- If building shows the spec is wrong, update the spec in the same PR. Don't force wrong code to match a stale spec.
