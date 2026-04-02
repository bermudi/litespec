# Tasks

## Phase 1: Internal — Validation and Instructions Logic

- [x] Add `ValidateSpec(root, name string) (*ValidationResult, error)` in `internal/validate.go` — singular counterpart to existing `ValidateSpecs`
- [x] Add `BuildArtifactInstructionsStandaloneJSON(artifactID string)` in `internal/json.go` — returns `artifactId`, `description`, `instruction`, `template`, `outputPath` without change context
- [x] Remove `BuildArtifactInstructionsJSON` and `ArtifactInstructionsJSON` from `internal/json.go` — replaced by standalone version

## Phase 2: CLI — Argument Parsing Rewrite and Docs

- [ ] Rewrite `cmdValidate()` — accept positional `<name>`, add `--changes`/`--specs`/`--all` (combinable), add `--type change|spec`, remove `--change`
- [ ] Rewrite `cmdStatus()` — accept positional `<name>`, remove `--change`
- [ ] Rewrite `cmdInstructions()` — remove `--change` requirement, return static artifact guidance
- [ ] Update `printUsage()` to reflect new command surfaces
- [ ] Update `DESIGN.md` CLI commands table

## Phase 3: Tests

- [ ] Add tests for `ValidateSpec()`
- [ ] Add tests for `BuildArtifactInstructionsStandaloneJSON()`
- [ ] Add CLI-level tests for validate argument parsing (positional, bulk flags, type detection, ambiguity, mutual exclusivity)
- [ ] Add CLI-level tests for status positional name
- [ ] Add CLI-level tests for instructions without `--change`
