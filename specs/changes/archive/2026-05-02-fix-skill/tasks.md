# Tasks: Fix Skill

## Phase 1: Add fix skill registration and template

- [x] Add `SkillInfo` entry for `fix` to `Skills` slice in `internal/paths.go` (ID "fix", name "litespec-fix")
- [x] Create `internal/skill/fix.go` with `init()` registering the fix template
- [x] Run `go test ./internal/skill/` to verify template registration
- [x] Run `litespec update` to regenerate skills and verify `.agents/skills/litespec-fix/SKILL.md` is created

## Phase 2: Update review handoff and workflow

- [x] Update review template ending in `internal/skill/review.go` to reference fix skill instead of apply
- [x] Update workflow diagram and Skills table in `DESIGN.md`
- [x] Run `litespec update` to regenerate all skills
- [x] Run `litespec validate fix-skill` to confirm no structural regressions
- [x] Run `go test ./...` and `go vet ./...` to verify build health
