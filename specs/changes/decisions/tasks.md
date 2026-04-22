# Tasks — Architectural Decision Records

## Phase 1: Parser and Data Model

- [ ] Add `Decision`, `DecisionStatus` types and status constants in `internal/decision.go`
- [ ] Implement `ParseDecision(path)` — extracts title, status, sections, supersede pointers
- [ ] Implement `ListDecisions(root)` — scans `specs/decisions/`, skips non-matching filenames
- [ ] Implement `FindDecisionBySlug(root, slug)` — accepts either `NNNN-slug` or `slug`
- [ ] Unit tests: valid file, missing sections, bad status, missing title, supersede parsing, filename regex edge cases
- [ ] Unit tests: `ListDecisions` on empty dir, missing dir, dir with non-conforming files

## Phase 2: Validation

- [ ] Implement `ValidateDecision(root, slug)` — structure, status enum, supersede pointer resolution
- [ ] Extend `ValidateAll` to include decisions when present
- [ ] Add supersede cycle detection (reuse DFS from `deps.go`)
- [ ] Add duplicate number detection across the directory
- [ ] Unit tests: each validation error path, supersede cycles, duplicate numbers, forward-pointer requirement for superseded status
- [ ] Extend `ValidationResult` JSON output to include decision errors
- [ ] Unit tests: JSON shape for decision errors

## Phase 3: CLI — decide command

- [ ] Implement `cmd/litespec/decide.go` — slug validation, next-number calculation, collision check, scaffold write
- [ ] Register `decide` in `cmd/litespec/main.go` dispatcher
- [ ] Happy-path test: first decision, subsequent decisions, scaffold content
- [ ] Error-path tests: invalid slug, duplicate slug, write failure
- [ ] Add shell completion for `decide` subcommand in `internal/completion/`

## Phase 4: CLI — list and validate integration

- [ ] Add `--decisions` flag to `validate`, `--type decision` disambiguation
- [ ] Extend positional name resolution in `validate` to check decision slugs
- [ ] Add `--decisions` flag to `list`, with `--status` filter and `--sort number|recent|name`
- [ ] Mutual-exclusion check: `--decisions` vs `--changes` / `--specs`
- [ ] Implement decision row rendering and JSON output for `list`
- [ ] Dynamic slug completion for `validate <name>` and `--type decision`
- [ ] Unit tests: each new flag combination, JSON shape for `list --decisions --json`
- [ ] Ambiguity handling test: name matches both change and decision slug

## Phase 5: CLI — view integration

- [ ] Implement Decisions section in `cmd/litespec/view.go`
- [ ] Active-vs-superseded grouping, number-sorted display, omission when empty
- [ ] Add decision count line to summary section
- [ ] Unit tests: no decisions, only active, only superseded, mixed, summary formatting

## Phase 6: Skills and Documentation

- [ ] Update `internal/skill/grill.go` — prompt author to suggest `decide` for broad rulings
- [ ] Update `internal/skill/propose.go` — during design.md authoring, prompt to check for standing rules
- [ ] Update `internal/skill/review.go` — flag imperative cross-cutting language in design.md
- [ ] Regenerate skills with `litespec update`
- [ ] Update `DESIGN.md` — document `specs/decisions/` in directory structure and add a Decisions section
- [ ] Update `AGENTS.md` — add decisions to Core Concepts
- [ ] Update `README.md` — mention `litespec decide` in command table if one exists

## Phase 7: End-to-End Verification

- [ ] Manual E2E: `litespec decide foo` → edit → `litespec validate --decisions` → `litespec list --decisions` → `litespec view`
- [ ] Manual E2E: create decision, supersede it, verify pointers validate and list/view reflect status
- [ ] Run full test suite: `go build ./...`, `go vet ./...`, `go test ./...`
- [ ] Self-dogfood: create the first real decision in this repo capturing the litespec decision "decisions are not changes"
