# Tasks: Fix Skill

## Phase 1: Add fix skill registration and template

- [ ] Add `SkillInfo` entry for `fix` to `Skills` slice in `internal/paths.go` (ID "fix", name "litespec-fix")
- [ ] Create `internal/skill/fix.go` with `init()` registering the fix template
- [ ] Run `go test ./internal/skill/` to verify template registration
- [ ] Run `litespec update` to regenerate skills and verify `.agents/skills/litespec-fix/SKILL.md` is created

## Phase 2: Update review handoff and workflow

- [ ] Update review template ending in `internal/skill/review.go` to reference fix skill instead of apply
- [ ] Update workflow diagram and Skills table in `DESIGN.md`
- [ ] Run `litespec update` to regenerate all skills
- [ ] Run `litespec validate fix-skill` to confirm no structural regressions
- [ ] Run `go test ./...` and `go vet ./...` to verify build health
