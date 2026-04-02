## Phase 1: Skill Template

- [ ] Update `internal/skill/verify.go` to add three-way branch: 0 tasks checked → artifact review; some → implementation review; all → pre-archive review
- [ ] Update verify skill description in `internal/paths.go` to mention all three modes

## Phase 2: Documentation

- [ ] Update `DESIGN.md` verify row in Skills table to describe context-aware three-mode behavior
- [ ] Update `AGENTS.md` verify description in Workflow section to reflect all three review modes

## Phase 3: Regenerate and Verify

- [ ] Regenerate `.agents/skills/litespec-verify/SKILL.md` from updated template
- [ ] Run `go build ./...` and `go test ./...` to confirm nothing is broken
