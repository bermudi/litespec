## Motivation

Skills are the primary interface between litespec and AI agents. If a skill silently fails to generate, agents lose access to workflow commands without any indication. The current system skips skills with missing templates and has zero test coverage for the skill generation pipeline itself. This is a fragile foundation for the tool's most important integration point.

## Scope

- Report errors when skill templates are missing instead of silently skipping
- Validate that all registered skills have corresponding templates at generation time
- Add unit tests for the skill generation pipeline
- Validate template frontmatter structure

## Non-Goals

- Changing skill template content or quality (separate effort)
- Adding new skills
- Modifying the adapter symlink mechanism
