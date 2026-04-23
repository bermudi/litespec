# Delta Preview Command

## Motivation

From the meta-friction log:

> No delta preview — I'd have paid for `litespec preview decisions` showing what canon specs would look like after merge. For a change touching four capabilities with ADDED-only deltas, it's fine — but the moment I had MODIFIED or RENAMED in a real change, I'd want to eyeball the merge before archive.

The `archive` command applies deltas atomically but provides no visibility into the final merged specs. Users must trust the merge logic or manually reconstruct the outcome. A preview command closes this gap by showing what would happen without making any changes.

## Scope

### New command: `litespec preview <change-name>`

- Accepts a positional change name.
- Reuses the existing `PrepareArchiveWrites` function to compute the merged result.
- Displays a structural summary per capability: capability name, new/modified indicator, and a list of operations with requirement names.
- Prints a footer with aggregate counts: capabilities affected, requirements added, modified, removed, renamed.
- Supports `--json` for structured output suitable for scripting and CI pipelines.

### Edge cases covered

- New capabilities (no existing canon spec) are flagged as `NEW SPEC`.
- Empty changes produce a clear "No changes to preview" message.
- Failed merges surface the error without writing anything.
- Non-existent changes report a clear error.

## Non-Goals

- **No `--diff` flag in this change.** Unified diff output is reserved for a future enhancement.
- **No `--full` merged spec output.** The structural summary is intentionally concise.
- **No side effects.** Preview never writes to canon, archive, or temporary files.
- **No preview for archived changes.** Archived changes have no deltas; the command errors.
- **No validation blocking.** Preview works even on changes that fail validation, so users can see what the merge would look like while they are still iterating.
