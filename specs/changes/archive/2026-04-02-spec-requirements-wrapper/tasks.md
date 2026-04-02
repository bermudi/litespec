## Phase 1: Types and Parser

- [x] Add `Purpose string` field to `Spec` struct in `internal/types.go`
- [x] Update `ParseMainSpec` in `internal/delta.go` to require `## Requirements` wrapper and capture optional `## Purpose`
- [x] Update `SerializeSpec` in `internal/delta.go` to emit `## Requirements` wrapper and optional `## Purpose`

## Phase 2: Tests

- [x] Update all `ParseMainSpec` test fixtures in `internal/delta_test.go` to include `## Requirements` wrapper
- [x] Update `TestSerializeRoundTrip` to verify `## Requirements` in output
- [x] Update `TestSerializeSpecWithNoScenarios` to use new format
- [x] Add test: spec without `## Requirements` wrapper produces parse error
- [x] Add test: spec with `## Purpose` captures purpose text
- [x] Add test: spec with unsupported H2 before `## Requirements` produces parse error
- [x] Add test: round-trip preserves Purpose field
- [x] Run `go test ./...` and fix any remaining test failures

## Phase 3: Documentation and Skill Instructions

- [x] Update spec format description in `DESIGN.md`
- [x] Update `artifactSpecs` in `internal/skill/artifact.go` to reference `## Requirements` wrapper

## Phase 4: Migrate Existing Canon Specs

- [x] Add `## Requirements` wrapper to `specs/canon/validate/spec.md`
- [x] Add `## Requirements` wrapper to `specs/canon/archive/spec.md`
- [x] Add `## Requirements` wrapper to `specs/canon/status/spec.md`
- [x] Add `## Requirements` wrapper to `specs/canon/instructions/spec.md`
- [x] Run `go build ./...` and `go test ./...` to verify everything passes
