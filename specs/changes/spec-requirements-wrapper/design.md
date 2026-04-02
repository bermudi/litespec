## Architecture

This is a parser/serializer-layer change. The `Spec` struct gains a `Purpose` field. `ParseMainSpec` enforces the `## Requirements` wrapper and optionally captures `## Purpose`. `SerializeSpec` emits both sections. No changes to delta parsing, merge logic, or CLI command flow.

```
Spec struct (before)          Spec struct (after)
┌────────────────────┐       ┌────────────────────┐
│ Capability string   │       │ Capability string   │
│ Requirements []Req  │       │ Purpose string      │ ← NEW
│                     │       │ Requirements []Req  │
└────────────────────┘       └────────────────────┘
```

Data flow unchanged:

```
ParseMainSpec → Spec → MergeDelta → Spec → SerializeSpec
                     ↑                    ↑
                  reads Purpose        emits Purpose
                  reads Requirements   emits ## Requirements
```

`ParseDeltaSpec` is unchanged — deltas use their own operation headers (`## ADDED Requirements`, etc.).

## Decisions

**Strict parse, no backward compat.** Old flat-format specs (no `## Requirements` wrapper) become parse errors. This is intentional — the 4 existing canon specs get migrated in the last phase, and the parser change lands first. Tests that use old-format fixtures must be updated alongside the parser.

**Purpose is captured but not required.** The parser accepts `## Purpose` before `## Requirements` and stores it in `Spec.Purpose`. No H2 other than `## Purpose` is permitted before `## Requirements`. If present, `SerializeSpec` emits it. Validation does not check Purpose content.

**No changes to delta format.** Delta specs keep their `## ADDED/MODIFIED/REMOVED/RENAMED Requirements` headers. These are operation markers, not structural wrappers in the same sense.

## File Changes

- `internal/types.go` — Add `Purpose string` field to `Spec` struct
- `internal/delta.go` — Update `ParseMainSpec` to require `## Requirements`, capture optional `## Purpose`; update `SerializeSpec` to emit both
- `internal/validate.go` — No validation changes (existing spec validation already calls `ParseMainSpec` which will enforce the new format)
- `internal/skill/artifact.go` — Update `artifactSpecs` template to show `## Requirements` wrapper in delta format guidance
- `internal/delta_test.go` — Update all `ParseMainSpec` test fixtures to include `## Requirements` wrapper; update `SerializeSpec` round-trip tests
- `internal/delta_merge_test.go` — No changes (merge operates on `Spec` structs, not raw markdown)
- `internal/validate_test.go` — Update test fixtures that create spec files on disk
- `internal/archive_test.go` — Update test fixtures if they create main specs
- `DESIGN.md` — Update spec format description and `Spec` struct documentation
- `specs/canon/validate/spec.md` — Add `## Requirements` wrapper
- `specs/canon/archive/spec.md` — Add `## Requirements` wrapper
- `specs/canon/status/spec.md` — Add `## Requirements` wrapper
- `specs/canon/instructions/spec.md` — Add `## Requirements` wrapper
