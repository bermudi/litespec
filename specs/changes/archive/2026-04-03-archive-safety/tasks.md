## Phase 1: Deterministic Delta Ordering

- [x] Sort delta file entries by filename in `PrepareArchiveWrites` after `os.ReadDir`
- [x] Add test that multiple delta files merge in lexicographic order regardless of `os.ReadDir` order

## Phase 2: Atomic Write Flow

- [x] Implement `writeAtomic(path string, content string, backup *string) error` — writes to `path.tmp`, verifies parse, renames to final
- [x] Implement rollback that restores backup content if atomic write fails partway
- [x] Refactor `WritePendingSpecs` to use atomic writes with backups
- [x] Update `cmdArchive` in `main.go` to use the new flow

## Phase 3: Post-Archive Verification

- [x] After `ArchiveChange` rename, verify archived directory exists in archive path
- [x] Verify all affected canon specs parse successfully after merge
- [x] Report verification failures with actionable error messages

## Phase 4: Testing

- [x] Test atomic write success path (happy path still works)
- [x] Test rollback when parse fails after merge
- [x] Test rollback when write fails (simulate permission error)
- [x] Test deterministic ordering with files named out of order
