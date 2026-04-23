# Ubiquitous Language — Tasks

## Phase 1: Glossary Artifact and Skill

- [ ] Create `specs/glossary.md` with initial project terms seeded from AGENTS.md and DESIGN.md (canon, change, delta, archive, phase, skill, artifact, spec, scenario, proposal, design, tasks)
- [ ] Add `glossary` entry to `Skills` slice in `internal/paths.go`
- [ ] Create `internal/skill/glossary.go` with glossary skill template registered via `init()`
- [ ] Update expected skill IDs in `internal/skill/skill_test.go` to include "glossary"
- [ ] Run `go test ./internal/skill/` and `go build ./cmd/litespec/` to verify

## Phase 2: Skill Integration

- [ ] Update explore template in `internal/skill/explore.go`: add glossary read at session start in "Litespec Awareness" section, add nudge behavior for undefined terms
- [ ] Update grill template in `internal/skill/grill.go`: add glossary read at session start, add nudge for new terms during grilling
- [ ] Update propose template in `internal/skill/propose.go`: add glossary check after specs are written, offer to update with new terms
- [ ] Update apply template in `internal/skill/apply.go`: add passive glossary reference in a "References" section at the end
- [ ] Run `go build ./cmd/litespec/` and `./litespec update` to regenerate skills, verify generated SKILL.md files contain glossary directives

## Phase 3: Docs and Housekeeping

- [ ] Create `docs/glossary.md` explaining the ubiquitous language concept, how litespec uses it, how to maintain it, and linking to `specs/glossary.md` as source of truth
- [ ] Add `Glossary: glossary.md` to `mkdocs.yml` nav
- [ ] Update DESIGN.md: remove `## Glossary` from canonical spec format, remove glossary delta operations, add `specs/glossary.md` as a project artifact, update the Glossary section at the bottom
- [ ] Update AGENTS.md: add glossary to core concepts, mention which skills read it
- [ ] Descope `cross-change-contracts`: update proposal.md to remove glossary structural layer, keep review skill enhancement
- [ ] Descope `cross-change-contracts`: update design.md to remove glossary architecture/decisions/file changes, adapt review to read project glossary
- [ ] Descope `cross-change-contracts`: update tasks.md to remove phases 1-3 (glossary types/parsing/merge/validation), keep phase 4 adapted
- [ ] Descope `cross-change-contracts`: remove `specs/spec-format/spec.md` (all requirements are glossary-specific), remove `specs/validate/spec.md` (all requirements are glossary-specific), update `specs/review/spec.md` to reference project glossary as supplementary context
- [ ] Run `uv run mkdocs build` to verify docs build cleanly
