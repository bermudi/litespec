# Patch Lane

## Motivation

The current change workflow (`propose → apply → archive`) is correctly sized for changes with real spec impact, but it imposes painful ceremony on small changes — adding a CLI flag, tweaking a phrase in skill text, a one-function behavior change. For these, authoring `proposal.md`, `design.md`, and `tasks.md` is transcription, not thinking. The actual contract — the delta against canon — is the only artifact that matters for system correctness.

When the friction is high enough, users skip litespec entirely and edit code directly, which causes the spec rot litespec exists to prevent. The framework loses by being too heavy for the 80% case.

A lighter lane is needed: one that keeps the non-negotiable (canon stays current) and drops the negotiable (planning artifacts that exist to scaffold thinking that already happened).

Two related observations enable this:

1. **The CLI rigorously validates the delta** (parse, scenario presence, dangling-delta detection, name collisions). It only checks **existence** of `proposal.md` / `design.md` / `tasks.md`, not their content. An empty `proposal.md` passes today. So the planning artifacts are already weak signals.
2. **The delta is the contract.** Archive merges deltas into canon regardless of whether the planning artifacts exist or are useful. They are scaffolding for humans and AIs, not inputs to the spec engine.

## Scope

Introduce a **patch lane** for delta-only changes, alongside the existing full proposal lane.

### CLI changes

- New command: `litespec patch <name> <capability>` — scaffolds a delta-only change. Creates `specs/changes/<name>/specs/<capability>/spec.md` with a delta stub. No `proposal.md`, no `design.md`, no `tasks.md`, no `.litespec.yaml`.
- `litespec validate` — demote `proposal.md`, `design.md`, `tasks.md` from required to optional. Validate them only when they exist.
- `litespec validate` — add lightweight content validation when planning artifacts exist:
  - `proposal.md`: must contain `## Motivation` (or `## Why`) and `## Scope` (or `## What Changes`) headings, each with at least one non-blank body line. Errors if missing or empty.
  - `design.md`: must contain at least one `## ` heading and at least 3 non-blank content lines outside fenced code blocks. Catches stub files without prescribing structure.
  - `tasks.md`: existing phase + checkbox checks remain. Add: every phase must contain at least one checkbox. (Currently a phase with zero tasks passes.)

### Artifact state machine changes

- `LoadArtifactStates` and `LoadChangeContext` — recognize patch-mode changes (inferred from absence of `proposal.md` and presence of at least one delta in `specs/`). In patch mode, planning artifacts report `N/A` instead of `BLOCKED` or `READY`. The change's lifecycle is just `specs DONE`.
- `litespec status` — patch-mode changes display only the `specs` artifact line plus a one-line note indicating patch mode.
- `litespec view` — patch-mode changes display in a distinct category (or with a `[patch]` marker), separate from draft/active/ready-to-archive.

### Skills and docs

- New skill `litespec-patch` — describes when to use the patch lane (small, single-capability changes; new flags; minor behavioral tweaks) and when not (multi-capability changes, REMOVED requirements, anything needing design discussion → use `propose`). Skill instructs the AI to write the delta, implement, then archive.
- Update `litespec-workflow` skill — document the patch lane as a sibling of the propose workflow: `patch → archive`.
- Update `AGENTS.md` and `DESIGN.md` — document the lane and the rule: *the delta is the contract; planning artifacts are optional scaffolding*.
- Regenerate skills via the standard `litespec update` flow after the Go template is registered.

### Archive behavior

- No changes required. Archive already validates via `ValidateChange`; once the trio is optional, patch-mode changes pass cleanly. Task-completion gate already short-circuits when `tasks.md` is absent. Archive merges deltas into canon as today.

## Non-Goals

- **No size or complexity policing on patch.** No heuristics that flag "this delta is too big for patch." Trust the user. Same as today's `propose` accepting any size.
- **No new metadata field for patch mode.** Mode is inferred from artifact presence, not declared in `.litespec.yaml`. Convention over configuration.
- **No changes to the `propose` skill** beyond a passing reference to the patch lane. Real changes still get the full treatment.
- **No `research` integration with patch.** If a change needs research, it is not a patch.
- **No changes to `review`.** Review already adapts to whatever artifacts exist; it will read what is present and skip what is not.
- **No retroactive migration of existing changes.** Patch lane applies to new changes only.
- **No content checks beyond "not empty / has expected sections".** Validate is not a linter for proposal quality. The checks added here are existence + minimal structure, no prose grading.
- **No removal of `proposal.md` / `design.md` / `tasks.md`** from the system. They remain available and are still produced by the `propose` flow. The change is purely about making them optional and adding a lighter alternative.
