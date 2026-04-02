## Phase 1: Skill Template

- [x] Update `internal/skill/verify.go` to add three-way branch: 0 tasks checked → artifact review; some → implementation review; all → pre-archive review
- [x] Update verify skill description in `internal/paths.go` to mention all three modes

## Phase 2: Documentation

- [x] Update `DESIGN.md` verify row in Skills table to describe context-aware three-mode behavior
- [x] Update `AGENTS.md` verify description in Workflow section to reflect all three review modes

## Phase 3: Regenerate and Verify

- [x] Regenerate `.agents/skills/litespec-verify/SKILL.md` from updated template
- [x] Run `go build ./...` and `go test ./...` to confirm nothing is broken
