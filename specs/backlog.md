# Backlog

## Deferred

- **Universal `--json` on every command** — the wiki (Zanie Blue, Ronacher) argues tool output must be structured, context-reducing, and fast. litespec's `--json` flag is inconsistent: present on `status`, `validate`, `list`, `instructions` but missing or partial on others. `view` is pure text output. Agents consuming litespec output get verbose text when they could get structured data. Scope:
  - Universal `--json` on every command
  - `view --json` for dashboard data
  - `validate --json` already exists, verify completeness
  - `instructions --json` for programmatic artifact templates
  - `--minimal` flag that outputs only actionable signal (error count, phases remaining, current task)

- **Tracer bullet phases in tasks.md** — the wiki says implementation should start with a thin vertical slice hitting every layer. litespec's phased tasks are sequential (foundation → core → integration) — exactly the "horizontal" approach tracer bullets replace. Tasks.md phases encourage layered implementation (schema first, then logic, then wiring) rather than vertical slices. Fix: add tracer bullet guidance to the propose skill template. When generating tasks.md, the first phase should be a vertical slice that exercises the full stack. Example: "Phase 1: Tracer — wire a single endpoint from route → handler → DB → response, with a test." This is a prompt change, not a code change.
- **Apply skill should auto-verify before commit** — apply marks tasks as `[x]` after writing code but never runs tests. The wiki's compounding booboos pattern: small errors accumulate when unchecked. An agent can complete a phase, mark all tasks done, and commit broken code. Fix: add to the apply skill — after marking tasks done but before committing, run the project's verify command. If it fails, fix before committing. This is a skill change, not a CLI change — the apply skill already reads AGENTS.md for build commands; it just needs to run them.
- **Smarter context per phase in skills** — the wiki's Memento Strategy: clear context, then reload only what matters. litespec's apply skill reads "whatever change artifacts exist" — proposal, design, specs, tasks, plus source files. An agent in phase 3 doesn't need the full proposal and design. It needs the current tasks, the relevant specs, and the files it's modifying. Fix: add context partitioning to skills. Phase 1: read all artifacts. Subsequent phases: read only specs + tasks + the files modified in the previous phase. This is a skill template change, not a CLI change. Or add a `litespec context <name> --phase N` command that outputs only the relevant files/paths for that phase.

## Open Questions

## Future Versions
- **Cross-change spec links** — the wiki's content architecture (threads → concepts → authors → projects, all cross-linked with `[[wikilinks]]`) is more navigable than litespec's flat `specs/canon/<capability>/spec.md` structure. Canonical specs are isolated files. A requirement that depends on another spec has no explicit link. Design decisions reference specs by name but there's no machine-readable link. Fix: add optional `dependsOn` to spec requirements (references other specs), similar to how change dependencies work. `validate` could check that referenced specs exist. Enables the wiki's cross-reference pattern at the spec level.
