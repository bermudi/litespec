# Tasks — Architectural Decision Records

## Phase 1: Parser and Data Model

- [x] Add `Decision`, `DecisionStatus` types and status constants in `internal/decision.go`
- [x] Implement `ParseDecision(path)` — extracts title, status, sections, supersede pointers
- [x] Implement `ListDecisions(root)` — scans `specs/decisions/`, skips non-matching filenames
- [x] Implement `FindDecisionBySlug(root, slug)` — accepts either `NNNN-slug` or `slug`
- [x] Unit tests: valid file, missing sections, bad status, missing title, supersede parsing, filename regex edge cases
- [x] Unit tests: `ListDecisions` on empty dir, missing dir, dir with non-conforming files

## Phase 2: Validation

- [x] Implement `ValidateDecision(root, slug)` — structure, status enum, supersede pointer resolution
- [x] Extend `ValidateAll` to include decisions when present
- [x] Add supersede cycle detection (reuse DFS from `deps.go`)
- [x] Add duplicate number detection across the directory
- [x] Unit tests: each validation error path, supersede cycles, duplicate numbers, forward-pointer requirement for superseded status
- [x] Extend `ValidationResult` JSON output to include decision errors
- [x] Unit tests: JSON shape for decision errors

## Phase 3: CLI — decide command

- [x] Implement `cmd/litespec/decide.go` — slug validation, next-number calculation, collision check, scaffold write
- [x] Register `decide` in `cmd/litespec/main.go` dispatcher
- [x] Happy-path test: first decision, subsequent decisions, scaffold content
- [x] Error-path tests: invalid slug, duplicate slug, write failure
- [x] Add shell completion for `decide` subcommand in `internal/completion/`

## Phase 4: CLI — list and validate integration

- [x] Add `--decisions` flag to `validate`, `--type decision` disambiguation
- [x] Extend positional name resolution in `validate` to check decision slugs
- [x] Add `--decisions` flag to `list`, with `--status` filter and `--sort number|recent|name`
- [x] Mutual-exclusion check: `--decisions` vs `--changes` / `--specs`
- [x] Implement decision row rendering and JSON output for `list`
- [x] Dynamic slug completion for `validate <name>` and `--type decision`
- [x] Unit tests: each new flag combination, JSON shape for `list --decisions --json`
- [x] Ambiguity handling test: name matches both change and decision slug

## Phase 5: CLI — view integration

- [x] Implement Decisions section in `cmd/litespec/view.go`
- [x] Active-vs-superseded grouping, number-sorted display, omission when empty
- [x] Add decision count line to summary section
- [x] Unit tests: no decisions, only active, only superseded, mixed, summary formatting

## Phase 6: Skills and Documentation

- [x] Update `internal/skill/grill.go` — prompt author to suggest `decide` for broad rulings
- [x] Update `internal/skill/propose.go` — during design.md authoring, prompt to check for standing rules
- [x] Update `internal/skill/review.go` — flag imperative cross-cutting language in design.md
- [x] Regenerate skills with `litespec update`
- [x] Update `DESIGN.md` — document `specs/decisions/` in directory structure and add a Decisions section
- [x] Update `AGENTS.md` — add decisions to Core Concepts
- [x] Update `README.md` — no command table exists, skipped

## Phase 7: End-to-End Verification

- [x] Manual E2E: `litespec decide foo` → edit → `litespec validate --decisions` → `litespec list --decisions` → `litespec view`
- [x] Manual E2E: create decision, supersede it, verify pointers validate and list/view reflect status
- [x] Run full test suite: `go build ./...`, `go vet ./...`, `go test ./...`
- [x] Self-dogfood: create the first real decision in this repo capturing the litespec decision "decisions are not changes"
