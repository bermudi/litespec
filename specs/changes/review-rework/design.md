## Architecture

This is a rename-and-remove change. No new capabilities or behavioral logic is introduced. The three existing review modes (artifact, implementation, pre-archive) remain identical — only the skill name changes from `verify` to `review`. The `continue` skill is removed entirely since `propose` already handles incremental artifact creation.

The canonical workflow is reordered to include a pre-apply review gate:

```
BEFORE:  explore → grill → propose → apply → verify → archive
AFTER:   explore → grill → propose → review → apply → review → archive
```

The `adopt` skill remains a separate path, unchanged.

## Decisions

1. **Direct rename, not delta merge for `specs/canon/verify/`** — The capability directory is renamed to `review/` with all prose updated (verify → review). This happens during implementation, not through the delta merge at archive time. The delta specs in this change (ADDED on `review`, REMOVED on `verify`) document the intent for traceability.

2. **Keep three-mode detection in the review skill** — The workflow reorder means the first review is always artifact review by construction, but the skill should still work correctly regardless of when invoked. Someone might skip the pre-apply review and jump straight to mid-apply.

3. **Drop `continue` without replacement** — `propose` already handles the resume case (line 20: "pick up where it left off"). The one-at-a-time behavior is not valuable enough to justify a separate skill, registration, tests, and generated directory.

4. **Leave archived change references untouched** — Historical records under `specs/changes/archive/` remain as-is.

5. **Update trigger words in skill descriptions** — The review skill description uses `says "review"` as the trigger. The old `verify` trigger word is not carried forward. The updated description text in `internal/paths.go` changes from `says "verify"` to `says "review"` while keeping the rest identical.
6. **Only replace skill-name references, not natural-language "verify"** — Several files use "verify" as natural English meaning "to confirm/check", not as the skill name. These must be left unchanged: `archive.go:7` ("to verify the change" — describes what `litespec validate` does), `apply.go:15` ("to verify all artifacts are ready"), and `verify.go:100` in the template ("Only flag what you can verify from reading the code"). Only skill-name references (e.g. "verify mode", "exit verify mode", `Register("verify", ...)`) are substituted.
7. **ADDED review spec intentionally adds ID/name constraint** — The `Updated Skill Description` requirement in the ADDED review spec adds "The skill ID MUST be `review` and the skill name MUST be `litespec-review`." This is not present in the original canon `verify/spec.md`. It's an intentional scope expansion to make the requirement more explicit.

## File Changes

### Skill code (rename + remove)

- `internal/paths.go` — Remove `continue` entry. Rename `verify` entry to `review` (ID, Name, Description updated). Reorder to: explore, grill, propose, review, apply, adopt, archive.
- `internal/skill/verify.go` — Rename file to `review.go`. Update `Register("verify", ...)` to `Register("review", ...)`. Rename const from `verifyTemplate` to `reviewTemplate`. Replace skill-name references: `verify mode` → `review mode`, `Verify mode` → `Review mode`, `the current verify behavior` → `the current review behavior`. Do NOT change natural-language "verify" on line 100 ("Only flag what you can verify from reading the code").
- `internal/skill/continue.go` — Delete file entirely.
- `internal/skill/skill_test.go` — Update expected skill list from `["explore", "grill", "propose", "continue", "apply", "verify", "adopt", "archive"]` to `["explore", "grill", "propose", "review", "apply", "adopt", "archive"]`.

### Skill template cross-references

- `internal/skill/propose.go` — Line 100: change `verify` → `review` in "during apply or verify". Line 104: change `` `verify` `` → `` `review` `` in suggested next steps.
- `internal/skill/adopt.go` — Line 87: change `verify` → `review` in suggested next steps.
- `internal/skill/archive.go` — No changes needed. Line 7 uses "verify" as natural English ("to verify the change" = to confirm the change), not as a skill name.
- `internal/skill/apply.go` — No changes needed. Line 15 uses "verify" as natural English ("to verify all artifacts are ready"), not as a skill name.

### Canon spec (rename)

- `specs/canon/verify/` — Rename directory to `specs/canon/review/`. Update `spec.md` heading from `# verify` to `# review` and all prose references from `verify` to `review`.

### Generated skill (regenerate)

- `.agents/skills/litespec-continue/` — Delete directory.
- `.agents/skills/litespec-verify/` — Delete directory.
- `.agents/skills/litespec-review/` — Regenerated from updated template via `litespec update`.

### Documentation

- `AGENTS.md` — Update workflow line, verify description in Workflow section, and all `verify`/`continue` references.
- `DESIGN.md` — Update directory structure (remove continue, rename verify), workflow diagram, skills table (remove continue row, rename verify row).
- `README.md` — Update workflow line.
- `docs/index.md` — Update workflow diagram and step table (remove continue row, rename verify row).
- `docs/workflow.md` — Remove continue section, rename verify section to review, update all workflow diagrams, update named patterns (remove Incremental Pattern since `continue` no longer exists; merge its use case into the Exploratory pattern), update Decision Flow diagram (replace Incremental branch with Exploratory pattern since `continue` no longer exists), update natural-language `verify` references in adopt section (lines 196, 198).
- `docs/project-structure.md` — Update skills directory listing (remove continue, rename verify).
- `docs/cli-reference.md` — Update skills directory listing (remove continue, rename verify).
- `docs/tutorial.md` — Lines 317 (section heading `## Verification` → `## Review`), 319 ("run verify" → "run review"), 322 (`litespec verify` → `litespec review` in command example), and 403 (`**Verify**` → `**Review**` in summary list).
