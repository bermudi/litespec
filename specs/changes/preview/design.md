# Design — Delta Preview Command

## Overview

Introduce `litespec preview <change-name>` as a read-only counterpart to `archive`. The command computes what the canonical specs would look like after archiving a change, then prints a structural summary without writing anything to disk. Implementation is thin: reuse `PrepareArchiveWrites`, compare merged results against current canon, format the delta, and print.

## Architecture

### Data Flow

```
cmdPreview(args)
  └── Parse positional name + --json flag
  └── PrepareArchiveWrites(root, name)  ← existing function
        ├── Reads delta specs
        ├── Loads current canon specs
        ├── MergeDelta in memory
        └── Returns []PendingWrite
  └── For each PendingWrite:
        ├── Load current spec from disk (if exists)
        ├── Parse merged content
        └── ComputeDeltaSummary(before, after)
  └── Render summaries (text or JSON)
```

### Delta Summary Computation

A new internal package function compares the current canon spec (if any) against the merged result:

```go
type PreviewOperation struct {
    Type        string // "ADDED", "MODIFIED", "REMOVED", "RENAMED"
    Requirement string
    OldName     string // set only for RENAMED
}

type PreviewCapability struct {
    Name       string
    IsNew      bool
    Operations []PreviewOperation
}

type PreviewTotals struct {
    Capabilities int
    Added        int
    Modified     int
    Removed      int
    Renamed      int
}

type PreviewResult struct {
    Capabilities []PreviewCapability
    Totals       PreviewTotals
}

func ComputePreviewResult(writes []PendingWrite, root string) (*PreviewResult, error)
```

The comparison logic:
1. For each `PendingWrite`, determine if the target path exists in canon.
2. Parse both the current spec (if it exists) and the merged content.
3. Walk requirements in merge order: compare names and content to categorize operations.
   - Requirement exists in merged but not in canon → `ADDED`
   - Requirement exists in merged but not in canon, with matching content to a removed requirement → `RENAMED` (detected by content-matching heuristic)
   - Requirement exists in canon but not in merged and not matched as a rename source → `REMOVED`
   - Requirement exists in both but content differs → `MODIFIED`
   - Requirement exists in both with identical content → omitted (no net change)

### Text Formatter

```go
func FormatPreviewText(result *PreviewResult) string
```

Output shape:

```
=== Preview: <change-name> → canon specs ===

▸ auth (MODIFIED)
  ~ MODIFIED: Login Requirement
  + ADDED: Session Timeout
  - REMOVED: Legacy OAuth
  → RENAMED: Two-Factor → MFA

▸ rate-limit (NEW SPEC)
  + ADDED: Rate Limiting
  + ADDED: Burst Handling

═══════════════════════════════════════════════════════════
2 capabilities affected • 3 added • 1 modified • 1 removed • 1 renamed
```

- `▸` prefix for capability headers
- `+` for ADDED, `~` for MODIFIED, `-` for REMOVED, `→` for RENAMED
- Footer uses `═` separator

### JSON Formatter

```go
func FormatPreviewJSON(result *PreviewResult) ([]byte, error)
```

Reuses the existing JSON marshaling pattern (`internal/json.go`) and checks for marshal errors.

### CLI Integration

New file `cmd/litespec/preview.go`:

```go
func cmdPreview(args []string) error
```

Behavior:
1. Validate positional name is present.
2. Resolve change directory; reject archived changes.
3. Call `PrepareArchiveWrites(root, name)`.
4. Call `ComputePreviewResult(writes, root)`.
5. Render via text or JSON formatter.
6. Print to stdout.

Register in `cmd/litespec/main.go` under the `"preview"` dispatch key.

## Key Decisions

### Reuse PrepareArchiveWrites unchanged

**Decision:** Preview calls `PrepareArchiveWrites` directly. No fork of merge logic.

**Rationale:** The whole point of preview is to show exactly what archive would do. If preview used its own merge path, it could drift from archive behavior and become misleading.

### Compute summary by diffing parsed specs, not by trusting delta markers alone

**Decision:** The summary is derived by comparing the parsed current canon spec against the parsed merged spec, not by echoing the delta file markers.

**Rationale:** A MODIFIED delta might produce no net change if the "new" content happens to match the current canon content. A summary that shows `~ MODIFIED: Foo` when Foo is unchanged would be noise. Diffing the actual before/after state ensures the summary reflects real changes.

### No --diff or --full flags in MVP

**Decision:** Only structural summary and `--json` are implemented now.

**Rationale:** The exploration evaluated three options and selected Option A as the 80/20 solution. The structural summary solves the stated friction ("I'd have paid for `litespec preview decisions` showing what canon specs would look like after merge") with minimal code. Diff output can be added later without changing the core design.

### Preview works on invalid changes

**Decision:** Preview does not run validation before computing the merge.

**Rationale:** Users often want to see the merge structure while they are still drafting a change. Blocking preview on validation failures would defeat this use case. If the merge itself fails (e.g., conflicting operations), the error is surfaced clearly.

## File Changes

**New files:**
- `cmd/litespec/preview.go` — CLI command implementation
- `internal/preview.go` — `ComputePreviewResult`, text formatter, JSON formatter
- `internal/preview_test.go` — unit tests for diffing, formatting, edge cases

**Modified files:**
- `cmd/litespec/main.go` — register `preview` command
- `cmd/litespec/completion.go` — add `preview` to shell completions
- `DESIGN.md` — add `preview` to CLI commands table
- `AGENTS.md` — mention preview command in workflow or core concepts if relevant

**No changes to:** delta merge logic, archive flow, spec format, validation system.

## Risks and Mitigations

- **Risk:** `PrepareArchiveWrites` gains a new behavior (e.g., writing temp files) that breaks preview's read-only guarantee.
  **Mitigation:** Preview only reads the returned `PendingWrite` structs. If `PrepareArchiveWrites` ever acquires side effects, preview remains unaffected because it never calls `WritePendingSpecsAtomic`.

- **Risk:** Summary becomes inaccurate if merge logic changes.
  **Mitigation:** Since preview reuses `PrepareArchiveWrites`, any merge change automatically propagates. Tests should include a cross-check that preview and archive produce the same merged content.

- **Risk:** MODIFIED requirements with identical content are silently omitted, confusing users who expected to see them.
  **Mitigation:** This is correct behavior — no net change means no line item. If users want to see every delta marker, they can read the delta spec files directly. The summary answers "what will canon look like?", not "what did I write in deltas?".
