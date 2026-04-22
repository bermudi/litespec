# Exploration: `litespec preview` Command

## Problem Statement

From the meta-friction log:

> No delta preview — I'd have paid for `litespec preview decisions` showing what canon specs would look like after merge. For a change touching four capabilities with ADDED-only deltas, it's fine — but the moment I had MODIFIED or RENAMED in a real change, I'd want to eyeball the merge before archive.

The `archive` command applies deltas atomically but provides no visibility into what the final merged specs will look like. Users must trust the merge logic or manually reconstruct the outcome mentally.

## Current Archive Flow

```
cmdArchive()
  └── PrepareArchiveWrites(root, name)  ← loads + merges
        ├── Read delta specs from specs/changes/<name>/<cap>/
        ├── Load main spec from specs/canon/<cap>/spec.md
        ├── MergeDelta(mainSpec, deltas)  ← handles ADDED/MODIFIED/REMOVED/RENAMED
        └── Returns []PendingWrite with merged content
  └── WritePendingSpecsAtomic(writes)
  └── ArchiveChange(root, name)
```

The `MergeDelta` function (`internal/delta.go:234-385`) processes operations in order:
1. **RENAMED** - renames requirements, checks for target collisions
2. **REMOVED** - removes requirements by name
3. **MODIFIED** - replaces content and scenarios
4. **ADDED** - appends new requirements

## Desired Behavior

```bash
$ litespec preview decisions

=== Preview: decisions → canon specs ===

▸ decisions (NEW SPEC)
  + ADDED: Decision Records
  + ADDED: Decision File Format
  + ADDED: Decision Validation
  + ADDED: Decision Listing
  + ADDED: Decision Viewing

▸ list (MODIFIED)
  + ADDED: List Decisions Flag

▸ validate (MODIFIED)
  ~ MODIFIED: JSON Output for Validate  ← no net change to this req
  + ADDED: Decision Validation
  + ADDED: Type Disambiguation Includes Decision

▸ view (MODIFIED)
  + ADDED: Dashboard Decision Count
  + ADDED: Decision Section in View

═══════════════════════════════════════════════════════════
4 capabilities affected • 10 requirements added • 0 modified • 0 removed • 0 renamed
```

For changes with MODIFIED/REMOVED/RENAMED, show side-by-side or unified diff:

```bash
$ litespec preview some-complex-change

▸ auth (MODIFIED)

  ~ MODIFIED: Login Requirement
    BEFORE:
      The system SHALL authenticate users via password.
    
    AFTER:
      The system SHALL authenticate users via password or SSO.
    
    SCENARIOS: 2 added, 1 removed

  - REMOVED: Legacy OAuth  ← entire requirement deleted
  
  → RENAMED: Two-Factor → MFA  ← requirement renamed
```

## Implementation Approaches

### Option A: Structural Summary (Minimal)

Reuse `PrepareArchiveWrites` but output structured text instead of writing.

**Pros:**
- Minimal code (~100 lines)
- Uses existing merge logic
- Fast to implement

**Cons:**
- Doesn't show full merged spec content
- Harder to verify intent for complex MODIFIED

### Option B: Full Content Preview (Complete)

Show the full merged spec as it would appear in canon, annotated with delta markers.

**Pros:**
- Complete visibility into final state
- Can diff against current canon

**Cons:**
- More verbose output
- Requires formatting decisions

### Option C: Diff Mode (Git-style)

Show unified diff of each capability's spec.md before/after.

**Pros:**
- Familiar format for developers
- Shows exactly what changes

**Cons:**
- Requires computing "before" state
- May be noisy for ADDED-only changes

## Recommended Approach: Hybrid (A + C)

Default to structural summary (Option A), add `--diff` flag for full unified diff (Option C).

### Data Flow

```go
func cmdPreview(args []string) error {
    // 1. Parse flags (--diff, --json)
    
    // 2. Get PendingWrites via existing PrepareArchiveWrites
    writes, err := internal.PrepareArchiveWrites(root, name)
    
    // 3. For each write, compute delta summary
    for _, w := range writes {
        before := loadCurrentSpec(w.Path)  // may not exist for NEW
        after := parseMergedSpec(w.Content)
        
        summary := computeDeltaSummary(before, after)
        printSummary(summary)
    }
}
```

### New Internal Function: `ComputeDeltaSummary`

```go
type DeltaSummary struct {
    Capability    string
    IsNew         bool
    Added         []string        // requirement names
    Modified      []ModSummary    // old→new content
    Removed       []string        // requirement names  
    Renamed       []RenameSummary // old→new names
}

type ModSummary struct {
    Name        string
    OldContent  string
    NewContent  string
    OldScenarios []string
    NewScenarios []string
}
```

This function compares the current canon spec (if exists) against the merged result and categorizes changes.

### CLI Design

```bash
litespec preview <change-name> [flags]

Flags:
  --diff       Show unified diff instead of summary
  --json       Output structured JSON
  --full       Show complete merged spec content
```

## Edge Cases

1. **New capability** (no existing canon spec)
   - Show all requirements as "ADDED (NEW SPEC)"

2. **Multiple delta files for same capability**
   - Already handled by `PrepareArchiveWrites` which sorts and merges all files

3. **Failed merge** (conflicting operations)
   - `MergeDelta` returns error — preview should show it clearly

4. **Empty change** (no delta specs)
   - Show "No changes to preview"

5. **Validation errors**
   - Preview should still work on invalid changes (shows what WOULD happen if fixed)
   - Or add `--strict` to fail fast on validation errors

## Files to Modify

- `cmd/litespec/main.go` — add "preview" to command switch
- `cmd/litespec/preview.go` — new file with preview command
- `internal/preview.go` — (optional) delta summary computation

## Open Questions

1. **Should preview show validation warnings?**
   - Yes, but non-blocking. The point is to see the merge structure.

2. **Should preview work on archived changes?**
   - No, archived changes have no deltas to preview.

3. **Output format for scripting?**
   - `--json` flag returns structured data for CI/CD pipelines.

## Related Friction Points

This addresses the "Validation passed is opaque" issue indirectly — preview shows WHAT would be merged, but doesn't detail WHAT was checked. A separate enhancement might add `validate --verbose`.

---

## Decision

**Proceed with Option A (Structural Summary) as MVP**, with `--diff` flag reserved for future enhancement.

The structural summary provides the 80% value with 20% effort. Users get visibility into:
- Which capabilities are affected
- How many requirements of each operation type
- Net change counts

This is sufficient for the "decisions" change (ADDED-only) and provides foundation for richer diff output later.
