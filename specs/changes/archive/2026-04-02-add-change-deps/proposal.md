## Motivation

Parallel changes often have implicit ordering requirements — change B cannot be archived before change A because it builds on capabilities A introduces. Today litespec has no way to express this. Users must remember ordering manually or maintain a separate `IMPLEMENTATION_ORDER.md` file that the tool cannot validate.

Additionally, multiple active changes can target the same canonical spec with overlapping delta operations, and litespec provides no warning about potential conflicts.

## Scope

- Add an optional `dependsOn` field to `.litespec.yaml` metadata, accepting a list of change names
- Dependency resolution against active and archived changes (active takes priority on name collision)
- Validation: missing references (error), dependency cycles (error), overlap detection across active changes (warning, suppressed when a `dependsOn` edge already exists between overlapping changes)
- Archive soft-block: warn and block if a dependency is still active, override with existing `--allow-incomplete` flag
- `litespec list --sort deps` topological ordering in the tabular view
- `litespec view` command (mirroring OpenSpec's dashboard) with a dependency graph section shown when any active change has `dependsOn`
- Auto-derived overlap warnings in `validate --all` based on delta target analysis

## Non-Goals

- `provides` / `requires` capability markers — not needed; `dependsOn` is explicit
- `touches` advisory field — overlap is auto-derived from delta targets
- `parent` / child split relationships — defer until a concrete need arises
- `--depends-on` CLI flag on `litespec new` — the propose skill writes metadata
- Blocking archive on active dependencies — soft block only (warning + `--allow-incomplete` escape hatch)
- Dependency-aware parallel archival or automatic sequencing
