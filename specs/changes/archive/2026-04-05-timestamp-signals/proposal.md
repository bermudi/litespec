## Motivation

Users currently see two disconnected time signals across different commands:

- `status <name>` shows "Created" from the `.litespec.yaml` `created` field
- `list` shows relative time ("3d ago") derived from filesystem mtimes, sorted by most recent
- `view` shows no timestamps at all

The "born" signal (when a change was created) and "last touched" signal (when it was last modified) are both valuable — born tells you how old an effort is, last touched tells you whether it's stale. Neither signal is consistently surfaced across all output formats.

## Scope

- **`list` text output**: Add "born" column showing the `created` timestamp. Rename the existing relative-time column to clarify it's "last touched" (filesystem-derived mtime).
- **`list --json` output**: Add `born` field (RFC3339) to `ChangeListItemJSON`. The existing `lastModified` field already represents last-touched.
- **`view` dashboard**: Show born and last-touched per change entry.
- **`ChangeInfo` struct**: Add `Created time.Time` field, populated from `.litespec.yaml` during `ListChanges`.
- **`WriteChangeMeta` / `UpdateChangeDeps`**: New helpers added as part of the bug fix to ensure `created` is preserved on metadata re-writes.

## Non-Goals

- Changing how `created` is stored — it stays in `.litespec.yaml` as a UTC timestamp.
- Changing how "last touched" is derived — it stays filesystem-mtime-based.
- Addressing mtime accuracy after `git clone` — accepted as a known limitation.
- Adding time-based filtering or sorting by "born" — out of scope for this change.
