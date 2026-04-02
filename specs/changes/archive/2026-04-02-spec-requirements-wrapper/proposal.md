## Motivation

Canonical specs currently place `### Requirement:` blocks directly under the `# <capability>` heading with no structural wrapper. This works for parsing but creates ambiguity: any prose between the H1 and first requirement is silently ignored, and there's no clear boundary between descriptive content and formal requirements.

## Scope

- Add a required `## Requirements` H2 section header to the canonical spec format
- `ParseMainSpec` MUST require `## Requirements` before accepting `### Requirement:` blocks
- `SerializeSpec` MUST emit `## Requirements` wrapper
- Optional `## Purpose` H2 is permitted between the H1 and `## Requirements` — captured by the parser but not validated
- Any H2 other than `## Purpose` before `## Requirements` is a parse error
- `ParseDeltaSpec` is unchanged — delta specs use their own operation headers (`## ADDED/MODIFIED/REMOVED/RENAMED Requirements`)
- Existing canon specs (validate, archive, status, instructions) MUST be migrated to include `## Requirements`
- Tests, DESIGN.md, and skill instructions updated to reflect the new format

## Non-Goals

- Adding or requiring `## Purpose` content (optional, not validated)
- Changing delta spec format
- Changing proposal, design, or tasks format
- Import from OpenSpec (separate change)
