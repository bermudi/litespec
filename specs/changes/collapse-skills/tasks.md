## Phase 1: Replace skill registry and templates

- [x] Update `internal/paths.go`: replace 11 `SkillInfo` entries with 4 (think, plan, build, review) with concise descriptions
- [x] Delete 11 template files from `internal/skill/`: explore.go, grill.go, propose.go, research.go, review.go, apply.go, adopt.go, workflow.go, glossary.go, patch.go, fix.go
- [x] Create `internal/skill/think.go`: merge explore + grill + workflow templates, register via `init()`
- [x] Create `internal/skill/plan.go`: merge propose + patch + adopt + glossary templates, register via `init()`
- [x] Create `internal/skill/build.go`: merge apply + fix templates, add research pause condition, register via `init()`
- [x] Create `internal/skill/review.go`: port existing review template, register via `init()`
- [x] Update `internal/skill/skill_test.go`: change `knownIDs` to `["think", "plan", "build", "review", "artifact-proposal", "artifact-specs", "artifact-design", "artifact-tasks"]`
- [x] Run `go build ./...` and `go test ./...` — verify build and tests pass
- [x] Run `go run ./cmd/litespec update` — verify 4 skill directories generated, legacy directories cleaned up

## Phase 2: Update canonical spec

- [ ] Update `specs/canon/skill-generation/spec.md`: apply the delta from this change (MODIFIED requirements updated, ADDED requirements included, REMOVED requirements deleted)
- [ ] Run `litespec validate --specs` — verify canonical spec validates clean

## Phase 3: Update documentation

- [ ] Update `AGENTS.md`: remove research phase from workflow, update Core Concepts to reference 4 skills, update Key Design Decisions, update Skill Generation conventions
- [ ] Update `DESIGN.md`: replace 7-directory skill tree with 4, update skill-related sections
- [ ] Run `go build ./...` — verify no breakage
