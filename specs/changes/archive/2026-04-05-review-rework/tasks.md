## Phase 1: Skill Code Rework

- [x] Rename `internal/skill/verify.go` to `internal/skill/review.go`, update `Register("verify", ...)` to `Register("review", ...)`, rename const `verifyTemplate` to `reviewTemplate`, replace skill-name `verify`/`Verify` references (`verify mode` → `review mode`, `Verify mode` → `Review mode`, `the current verify behavior` → `the current review behavior`). Do NOT change natural-language "verify" on line 100 ("Only flag what you can verify from reading the code").
- [x] Delete `internal/skill/continue.go`
- [x] Update `internal/paths.go`: remove `continue` entry, rename `verify` entry to `review` (ID `"review"`, Name `"litespec-review"`, description referencing `review` instead of `verify`), reorder to `explore, grill, propose, review, apply, adopt, archive`
- [x] Update `internal/skill/skill_test.go` expected skill list from `["explore", "grill", "propose", "continue", "apply", "verify", "adopt", "archive"]` to `["explore", "grill", "propose", "review", "apply", "adopt", "archive"]`
- [x] Update `internal/skill/propose.go` line 100 (`verify` → `review`) and line 104 (`` `verify` `` → `` `review` `` in suggested next steps)
- [x] Update `internal/skill/adopt.go` line 87: change `verify` → `review` in suggested next steps
- [x] Run `go build ./...`, `go test ./...`, `go vet ./...` to verify everything compiles and passes

## Phase 2: Canon Spec Rename

- [x] Rename `specs/canon/verify/` directory to `specs/canon/review/`
- [x] Update `specs/canon/review/spec.md`: change heading from `# verify` to `# review` and all prose references from `verify` to `review`

## Phase 3: Generated Skills

- [x] Delete `.agents/skills/litespec-continue/` directory
- [x] Delete `.agents/skills/litespec-verify/` directory
- [x] Run `go run ./cmd/litespec update` to regenerate `.agents/skills/litespec-review/SKILL.md` from updated template

## Phase 4: Documentation

- [x] Update `AGENTS.md`: workflow line, verify description in Workflow section, remove continue reference, update skill count
- [x] Update `DESIGN.md`: directory structure (remove continue, rename verify), workflow diagram, skills table (remove continue row, rename verify row)
- [x] Update `README.md`: workflow line
- [x] Update `docs/index.md`: workflow diagram, step table (remove continue row, rename verify row)
- [x] Update `docs/workflow.md`: remove continue section, rename verify section to review, update all workflow diagrams, update named patterns, update Decision Flow diagram
- [x] Update `docs/project-structure.md`: skills directory listing (remove continue, rename verify)
- [x] Update `docs/cli-reference.md`: skills directory listing (remove continue, rename verify)
- [x] Update `docs/tutorial.md`: section heading `## Verification` → `## Review` (line 317), "run verify" → "run review" (line 319), `litespec verify` → `litespec review` (line 322), `**Verify**` → `**Review**` in summary list (line 403)
- [x] Run `go build ./...`, `go test ./...` to verify final state
