# Backlog Support

## Motivation

Litespec has no concept of "maybe later." Every idea must go through the full explore → grill → propose pipeline to exist on disk. This means deferred roadmap items, open questions, and future version plans either live in the user's head or get forced through a ceremony they don't warrant.

Decisions (`specs/decisions/`) solved the "resolved rulings" gap. This change solves the inverse — ideas and questions that need to persist across sessions without the full change proposal ceremony.

## Scope

- A single `specs/backlog.md` file as a lightweight parking lot for deferred work, open questions, and future version plans
- `litespec view` parses the file and adds a summary line to the dashboard (e.g., `● Backlog: 3 deferred, 2 open questions, 4 future`)
- Three recognized H2 sections (`## Deferred`, `## Open Questions`, `## Future Versions`) counted and reported individually; any other H2s rolled into an "other" count
- Parsing counts top-level `- ` bullets only (no nesting, no checkbox awareness)
- Lightweight skill template prompts in explore, propose, review, and grill skills so the AI reads and references the backlog naturally

## Non-Goals

- No `litespec list --backlog` flag — the file is readable directly
- No validation of backlog.md format — it's a scratchpad, zero ceremony
- No structured cross-references from changes to backlog items — prose is sufficient
- No CLI command to manage backlog items — the AI writes the file directly (consistent with "CLI is read-only context provider")
- No wiring of `specs/explorations/` — stays as-is, out of scope
- No checkbox-aware parsing or graduation tracking
- No scaffolding by `litespec init` — the file is created on demand by the AI when the first item is parked
