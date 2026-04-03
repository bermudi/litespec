## Architecture

This change adds error reporting to the skill generation pipeline and introduces a validation function that can be called independently. No new packages.

## Decisions

- **Error, don't skip**: Replace `if template == "" { continue }` with an error return. This is a breaking change for any workflow that tolerates missing templates, but correctness trumps convenience here — a missing template is always a bug.
- **Separate validation function**: `ValidateSkillTemplates` provides a way to check template health without generating files. Can be wired into `validate` command later.
- **Test in the skill package**: Tests live alongside the template registration code in `internal/skill/`, not in the parent `internal/` package.

## File Changes

- `internal/skill.go`: Replace silent `continue` with error return in `GenerateSkills` when template is empty.
- `internal/adapter.go`: Replace silent `continue` with error return in `GenerateAdapterCommands` when template is empty.
- `internal/skill/skill.go`: Add `ValidateSkillTemplates(skillIDs []string) []string` function.
- `internal/skill/skill_test.go`: New file — test template registration, frontmatter generation, and validation function.
