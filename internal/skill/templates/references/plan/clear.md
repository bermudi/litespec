Use when the idea is sharp and you need to nail the GH issue (+ spec if load-bearing). This is the clear mode of `plan`.

## What to write — GH issue is proposal + design + queue

1. **Proposal (why/what).** Top of issue body: what we're doing, why, what we're not doing.
2. **Design (how).** Directory, lanes, key decisions — concise, not an essay.
3. **Queue — one `##` per unit.** Each unit:
   ```
   ## <one demo-able outcome>
   Done means: <observable, human-checkable>
   Verify: `<command that fails without the outcome>`
   Status: pending
   ```
   `Verify:` must fail for a plausible state where the outcome is missing. A `go test` that doesn't check output is not a Verify.

4. **Spec if load-bearing.** If the feature is a promise that breaks things when wrong (CLI shape, API, file format), edit `specs/<feature>/spec.md` directly in the same change — not a delta. Keep to 3-5 SHALL requirements, each with a WHEN/THEN scenario.

## Rules

- One unit = one thing you can demo. If you can't demo it, split it.
- One Verify per unit, and that Verify is the gate — `build` must satisfy it before claiming done.
- If building shows the spec is wrong, update the spec in the same PR. Don't force wrong code to match a stale spec.
