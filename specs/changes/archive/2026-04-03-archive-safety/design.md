## Architecture

The archive flow gains a write-verify-commit pattern: prepare writes → write to temp → verify → swap into canon → archive change directory → verify archive. Rollback cleans up temp files and restores original canon specs if any step fails.

## Decisions

- **Temp files adjacent to target**: Write temp files in the same directory as the target canon spec (e.g., `spec.md.tmp`) so they're on the same filesystem, making the rename atomic.
- **Backup before swap**: Read existing canon spec content before overwriting so it can be restored on failure. Store backup content in the `PendingWrite` struct via a `Backup []byte` field.
- **Atomicity scope is per-archive-operation**: All capabilities in a single archive are treated as one unit — if any capability's write fails, all are rolled back.
- **Sort in PrepareArchiveWrites**: Sort delta filenames with `sort.Slice` after `os.ReadDir` to ensure deterministic merge order. This is the minimal fix — no new sorting infrastructure needed.
- **Verify parse after merge**: Call `ParseMainSpec` on the merged content before writing. This catches serialization bugs early.

## File Changes

- `internal/change.go`: 
  - In `PrepareArchiveWrites`: sort delta file entries by name before processing.
  - Add `WritePendingSpecsAtomic` function that writes to temp files, verifies parse, then renames to final location.
  - Add rollback logic that restores original content on failure.
  - Add post-archive verification in `ArchiveChange` or a new wrapper.
- `cmd/litespec/main.go`: Update `cmdArchive` to use the new atomic write flow.
- `internal/change_test.go`: Add tests for deterministic merge order, rollback on failure, and post-archive verification.
