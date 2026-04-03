## Phase 1: Deterministic Delta Ordering

- [ ] Sort delta file entries by filename in `PrepareArchiveWrites` after `os.ReadDir`
- [ ] Add test that multiple delta files merge in lexicographic order regardless of `os.ReadDir` order

## Phase 2: Atomic Write Flow

- [ ] Implement `writeAtomic(path string, content string, backup *string) error` — writes to `path.tmp`, verifies parse, renames to final
- [ ] Implement rollback that restores backup content if atomic write fails partway
- [ ] Refactor `WritePendingSpecs` to use atomic writes with backups
- [ ] Update `cmdArchive` in `main.go` to use the new flow

## Phase 3: Post-Archive Verification

- [ ] After `ArchiveChange` rename, verify archived directory exists in archive path
- [ ] Verify all affected canon specs parse successfully after merge
- [ ] Report verification failures with actionable error messages

## Phase 4: Testing

- [ ] Test atomic write success path (happy path still works)
- [ ] Test rollback when parse fails after merge
- [ ] Test rollback when write fails (simulate permission error)
- [ ] Test deterministic ordering with files named out of order
