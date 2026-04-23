# Design: Backlog Support

## Architecture

The backlog is a single markdown file (`specs/backlog.md`) with no structural coupling to the rest of litespec. The only programmatic touchpoint is `view`, which parses the file to produce a summary line. Everything else — creating, editing, graduating items — is done by the AI writing the file directly.

```
specs/backlog.md  ──reads──►  internal/backlog.go (ParseBacklog)
                                      │
                                      ▼
                              cmd/litespec/view.go (summary line)
```

Skill templates get one-line additions — no Go code, just prompt text.

## Decisions

### Backlog parsing lives in `internal/backlog.go`

A new `ParseBacklog` function reads the file and returns category counts. This follows the pattern of `decision.go` — a domain parser in `internal/` consumed by the CLI. The function takes a file path, not a root, since there's only one file to read (unlike decisions which are a directory of files).

Alternative: inline the parsing in `view.go`. Rejected because the parsing logic is testable in isolation and `view.go` is already 300+ lines.

### Format contract

Three H2 section names are recognized: `## Deferred`, `## Open Questions`, `## Future Versions`. Any other H2 triggers a new category counted as "other." Items are top-level `- ` lines (no leading whitespace). Nested bullets, blank lines, and prose are ignored.

The dashboard label for each recognized section:
- `## Deferred` → "deferred"
- `## Open Questions` → "open questions"  
- `## Future Versions` → "future"
- anything else → "other"

### Dashboard placement

The backlog summary line appears in the summary section after the decisions line (or after task progress if no decisions exist). This mirrors how decisions were added — optional line, omitted when absent.

## File Changes

### `internal/backlog.go` (new)

New file with:
- `BacklogSummary` struct: `Deferred int`, `OpenQuestions int`, `Future int`, `Other int`
- `BacklogPath(root string) string`: returns `filepath.Join(root, ProjectDirName, "backlog.md")`
- `ParseBacklog(path string) (*BacklogSummary, error)`: reads the file, walks lines, counts top-level `- ` items under each H2 section. Returns `nil, nil` when file doesn't exist or when the file exists but has zero total items (not an error — an empty backlog is the same as no backlog).

### `internal/backlog_test.go` (new)

Tests for `ParseBacklog`: all categories populated, unknown sections counted as other, nested bullets ignored, missing file returns nil, empty file returns zero counts.

### `cmd/litespec/view.go` (modify)

After the decisions summary line in `cmdView`, call `ParseBacklog` and emit the backlog summary line if non-nil. Format: `● Backlog: N deferred, N open questions, N future` with categories omitted when zero, and ` — N other` appended when other > 0.

### `internal/paths.go` (modify)

Add `BacklogFileName = "backlog.md"` constant for consistency with other path constants.

### `internal/skill/explore.go` (modify)

Add one directive to the explore template: read `specs/backlog.md` if it exists for context on parked items and open questions.

### `internal/skill/propose.go` (modify)

Add one directive to the propose template: check if the proposal materializes a backlog item and suggest removing it from `specs/backlog.md`.

### `internal/skill/review.go` (modify)

Add one directive to the review template: when a change explicitly defers scope, suggest adding the deferred items to `specs/backlog.md`.

### `internal/skill/grill.go` (modify)

Add one directive to the grill template: read `specs/backlog.md` to challenge scope overlaps between the current plan and parked items.
