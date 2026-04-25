# Design

## Architecture

The patch lane is a parallel entry point into the existing change machinery. The merge engine and validation rules for deltas are unchanged — they already operate on deltas and treat planning artifacts as separate concerns. The archive flow gains one new behavior: stripping the `specs/` subtree from the archived directory after merging deltas to canon (previously, the subtree was retained).

The change introduces one new classification (`IsPatchMode`) and threads it through three consumers: status output, view dashboard categorization, and (passively) the validate flow once the trio becomes optional. The `litespec patch` command itself is a thin scaffolding command that mirrors the structure of `litespec new` but writes only a delta stub.

Component relationships:

```
litespec patch <name> <cap>          (NEW command)
   │
   └── creates specs/changes/<name>/specs/<cap>/spec.md (delta stub)

ValidateChange (existing)            (BEHAVIOR CHANGE)
   ├── trio: required → optional
   └── new content checks: proposal, design, empty-phase

LoadArtifactStates (existing)        (BEHAVIOR CHANGE)
   └── if IsPatchMode → return only {specs: DONE}

cmd/litespec/status.go (existing)    (BEHAVIOR CHANGE)
   └── if IsPatchMode → render only specs line + "(patch mode)"

cmd/litespec/view.go (existing)      (BEHAVIOR CHANGE)
   └── new category: Patch Changes (◆ bullet)

internal/skill/patch.go              (NEW skill template)
internal/paths.go                    (UPDATE: register litespec-patch)
internal/skill/workflow.go           (UPDATE: document patch lane)
```

Data flow stays the same: deltas live in `specs/changes/<name>/specs/`, archive merges them into `specs/canon/`. The only new branching is in display layers and the validation gate.

## Decisions

### Patch mode is declared in metadata, not inferred

**Chosen:** patch mode is stored as `mode: patch` in `.litespec.yaml`. Set by `litespec patch` at creation time. `IsPatchMode` reads the metadata file. When the field is absent (changes created by `litespec new`), the change is full-proposal mode.

**Over alternatives:**
- Inferring from filesystem (absent `proposal.md` + present delta) fails to distinguish "mid-creation full change" from "intentional patch." A full-proposal change that has specs written but no proposal yet looks identical to a patch. Metadata is unambiguous.
- A marker file like `.patch` would be even uglier and is less queryable than YAML.

**Constraints:** if a user manually adds `proposal.md` to a patch-mode change, it remains a patch — the mode is a property of the change, not the current file state. If they want to upgrade to a full change, they remove the `mode` field from `.litespec.yaml`.

### `IsPatchMode` is a single function, not a method on a struct

**Chosen:** `IsPatchMode(root, name string) bool` lives in the internal package as a free function.

**Over alternatives:** attaching it to a `Change` struct would require introducing or extending one. Free function is the smallest correct surface and matches existing helpers like `ChangeExists`.

**Constraints:** callers pass `root` and `name`. This is fine; that's how all the sibling functions work.

### No size policing on patch

**Chosen:** patch accepts any size delta. No heuristic warnings about "this looks too big for patch."

**Over alternatives:** validate could nudge users when a patch-mode change touches > N requirements, but this is gold-plating and creates false positives. Trust the user; same as how `propose` accepts any size.

**Constraints:** users can technically use patch for large changes. Skill text and documentation guide them away from this, but the CLI does not enforce.

### Content validation is structural, not stylistic

**Chosen:** proposal needs `## Motivation` + `## Scope` headings with non-blank bodies. Design needs at least one `## ` heading and ≥3 non-blank lines outside fences. Tasks needs ≥1 checkbox per phase. No prose grading.

**Over alternatives:** richer checks (e.g., "scope must list affected capabilities") slip into linter territory and create false negatives. Structural checks catch the actual problem (empty stubs) without being prescriptive.

**Constraints:** legacy proposals using `## Why` and `## What Changes` are accepted as aliases for `## Motivation` and `## Scope` to avoid breaking existing changes.

### Patch-mode changes are visually distinct in `view`

**Chosen:** patch changes get their own "Patch Changes" section with a `◆` bullet, separate from draft/active/completed. Summary section gets a "Patch Changes" count.

**Over alternatives:** lumping patch changes into "active" works but loses information. A `[patch]` suffix on lines is noisier than a section. A dedicated section reads cleanly and matches how draft/active/completed are already segmented.

**Constraints:** more horizontal complexity in the dashboard; mitigated by omitting the section entirely when no patch changes exist.

### Skill template stays in Go, generated via existing pipeline

**Chosen:** the patch skill is a `internal/skill/patch.go` template registered via `init()`, with a `SkillInfo` entry in `internal/paths.go`. Generated by `litespec update`.

**Over alternatives:** writing `.agents/skills/litespec-patch/SKILL.md` directly would violate the documented generation pipeline in `AGENTS.md`.

**Constraints:** none — this is the documented path.

## File Changes

### New files

- **`cmd/litespec/patch.go`** — implements `cmdPatch(args []string) error`. Parses two positional args (name, capability), validates them, refuses if `specs/changes/<name>/` exists, creates the directory tree, writes the stub `spec.md`, and writes `.litespec.yaml` with `mode: patch`. Exit-free per `project-structure` requirements. Implements **Patch Command Scaffold** requirement.

- **`cmd/litespec/patch_test.go`** — happy path + error cases (missing args, existing change, invalid names). Implements the test-coverage requirement from `project-structure`.

- **`internal/skill/patch.go`** — registers a `litespec-patch` skill template via `init()`. Template body explains when to use patch, when not to, and the workflow `patch → implement → archive`. Implements **Patch Lane Skill** requirement.

- **`specs/changes/patch-lane/specs/<capability>/spec.md`** — already created in the specs phase (this directory).

### Modified files

- **`internal/validate.go`** — three changes:
  1. Demote `proposal.md`, `design.md`, `tasks.md` from `requiredFiles` slice (lines 21-38). Move them to a "validate if present" pattern.
  2. Add `validateProposal(content string) []ValidationIssue` for the **Proposal Content Validation** requirement.
  3. Add `validateDesign(content string) []ValidationIssue` for the **Design Content Validation** requirement.
  4. Extend `validateTasksChecklist` (or add a sibling) to enforce per-phase checkbox presence per the **Empty-Phase Detection in Tasks** requirement.

- **`internal/validate_test.go`** — tests for: patch-mode change passes; missing trio passes; full proposal still passes; bad proposal/design/tasks content fails with new error messages.

- **`internal/artifact.go`** — `LoadArtifactStates` checks `IsPatchMode` and returns only `{specs: DONE}` for patch-mode changes. Implements **Patch-Mode Artifact States** requirement.

- **`internal/change.go`** — add `IsPatchMode(root, name string) bool` helper that reads `.litespec.yaml` and returns true when `mode` field is `"patch"`. `LoadChangeContext` already calls `LoadArtifactStates`, so it inherits the patch-mode behavior. Implements **Patch-Mode Change Detection** requirement.

- **`cmd/litespec/status.go`** — when rendering a change, check `IsPatchMode` and emit only the `specs` line plus a `(patch mode)` indicator. JSON output adds `"mode": "patch"` and either omits non-applicable artifacts or marks them `status: "n/a"`. Implements **Patch-Mode Status Display** requirement.

- **`internal/json.go`** (or wherever `BuildChangeStatusJSON` lives) — extend the JSON struct with an optional `Mode string` field. Set to `"patch"` when applicable.

- **`cmd/litespec/view.go`** — add patch-mode change collection alongside draft/active/completed. Render new "Patch Changes" section with `◆` bullet. Update summary line to include patch count. Implements **Patch-Mode Changes In Dashboard** requirement.

- **`internal/paths.go`** — append `SkillInfo{ID: "litespec-patch", Name: "litespec-patch", Description: "..."}` to the `Skills` slice. Implements skill registration per `skill-generation` rules.

- **`internal/skill/workflow.go`** — update template body to document the patch lane as `patch → archive`, with a sentence on when to choose it. Implements **Workflow Skill References Patch Lane** requirement.

- **`internal/skill/skill_test.go`** — extend the expected skill list to include `patch`. Implements the modified `skill-generation` requirement.

- **`AGENTS.md`** — add a paragraph in "Workflow" section documenting the patch lane and the rule "the delta is the contract; planning artifacts are optional scaffolding."

- **`DESIGN.md`** — same documentation update, with a brief explanation of `IsPatchMode` and where it is used.

### No-change files (touched during regeneration)

- **`.agents/skills/litespec-patch/SKILL.md`** — generated by `litespec update` after `internal/skill/patch.go` and `internal/paths.go` are in place. Not hand-edited.
- **`.agents/skills/litespec-workflow/SKILL.md`** — regenerated after template update.
- **`.claude/skills/litespec-patch`** — symlink generated by `litespec update --tools claude` if applicable.

### Verification

- `go build ./...` — type check.
- `go test ./...` — runs new tests in `validate_test.go`, `patch_test.go`, `skill_test.go`.
- `go vet ./...` — sanity.
- Manual smoke: create a patch change, validate, archive, verify canon updated and dashboard categorizes correctly.
