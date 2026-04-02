## Phase 1: Types and Parser

- [ ] Add `Purpose string` field to `Spec` struct in `internal/types.go`
- [ ] Update `ParseMainSpec` in `internal/delta.go` to require `## Requirements` wrapper and capture optional `## Purpose`
- [ ] Update `SerializeSpec` in `internal/delta.go` to emit `## Requirements` wrapper and optional `## Purpose`

## Phase 2: Tests

- [ ] Update all `ParseMainSpec` test fixtures in `internal/delta_test.go` to include `## Requirements` wrapper
- [ ] Update `TestSerializeRoundTrip` to verify `## Requirements` in output
- [ ] Update `TestSerializeSpecWithNoScenarios` to use new format
- [ ] Add test: spec without `## Requirements` wrapper produces parse error
- [ ] Add test: spec with `## Purpose` captures purpose text
- [ ] Add test: spec with unsupported H2 before `## Requirements` produces parse error
- [ ] Add test: round-trip preserves Purpose field
- [ ] Run `go test ./...` and fix any remaining test failures

## Phase 3: Documentation and Skill Instructions

- [ ] Update spec format description in `DESIGN.md`
- [ ] Update `artifactSpecs` in `internal/skill/artifact.go` to reference `## Requirements` wrapper

## Phase 4: Migrate Existing Canon Specs

- [ ] Add `## Requirements` wrapper to `specs/canon/validate/spec.md`
- [ ] Add `## Requirements` wrapper to `specs/canon/archive/spec.md`
- [ ] Add `## Requirements` wrapper to `specs/canon/status/spec.md`
- [ ] Add `## Requirements` wrapper to `specs/canon/instructions/spec.md`
- [ ] Run `go build ./...` and `go test ./...` to verify everything passes
