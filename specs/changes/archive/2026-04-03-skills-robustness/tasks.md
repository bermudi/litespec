## Phase 1: Error Reporting

- [x] Replace `if template == "" { continue }` with error return in `GenerateSkills` (`internal/skill.go:30-32`)
- [x] Replace `if template == "" { continue }` with error return in `GenerateAdapterCommands` (`internal/adapter.go:25-27`)
- [x] Verify CLI commands (`init`, `update`) surface these errors properly

## Phase 2: Template Validation

- [x] Add `ValidateSkillTemplates(skillIDs []string) []string` to `internal/skill/skill.go`
- [x] Wire `ValidateSkillTemplates` into `ValidateAll` (optional, as warning)
- [x] Update `go run ./cmd/litespec validate --all` to check template health

## Phase 3: Testing

- [x] Create `internal/skill/skill_test.go` with tests for `Get()` returning non-empty for all known IDs
- [x] Test frontmatter YAML generation (name and description fields present)
- [x] Test `ValidateSkillTemplates` with missing and complete registrations
- [x] Run `go test ./...` to verify no regressions
