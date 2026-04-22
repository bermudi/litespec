# Design — Architectural Decision Records

## Overview

Introduce `decisions/` as a fourth artifact class alongside canon, changes, and archive. Decisions are numbered, narrative, persistent files that live independently of the change workflow. Implementation is intentionally small: a parser, a validator, one new CLI command (`decide`), and integrations into three existing commands (`list`, `view`, `validate`).

The feature deliberately mirrors existing litespec conventions — filesystem-first, convention-over-configuration, no YAML frontmatter — so there is nothing new to learn beyond the artifact itself.

## Architecture

### Filesystem layout

```
specs/
├── canon/
├── changes/
├── decisions/                              ← new, optional
│   ├── 0001-single-shared-workspace.md
│   ├── 0002-beta-tools-session-bound.md
│   └── 0003-recursion-cap-three.md
```

Each file is a flat markdown document. No subdirectories, no categories, no tags in v1. Ordering is by number. Grouping, when needed, is done at read time by `status` or by mtime.

### Parser (`internal/decision.go`)

A new file `internal/decision.go` defines:

```go
type DecisionStatus string

const (
    StatusProposed   DecisionStatus = "proposed"
    StatusAccepted   DecisionStatus = "accepted"
    StatusSuperseded DecisionStatus = "superseded"
)

type Decision struct {
    Number         int
    Slug           string
    Title          string
    Status         DecisionStatus
    Context        string
    Decision       string
    Consequences   string
    Supersedes     []string
    SupersededBy   []string
    FilePath       string
    LastModified   time.Time
}

func ParseDecision(path string) (*Decision, error)
func ListDecisions(root string) ([]*Decision, error)
func FindDecisionBySlug(root, slug string) (*Decision, error)
```

Parsing uses the existing markdown scanning patterns from `change.go` (H1 extraction, H2 section splitting). No new dependencies.

The filename is the source of truth for number and slug — the parser extracts `NNNN` and `<slug>` via regex `^(\d{4})-([a-z0-9-]+)\.md$`. Files not matching this pattern are ignored by `ListDecisions` (not errored), matching how litespec treats unexpected entries elsewhere.

### Validator (`internal/validate.go`)

Add `ValidateDecision(root, slug string) *ValidationResult` and extend `ValidateAll` to include decisions. Checks:

- **Structure** — all required H2 sections present; status is a valid enum; title (H1) is non-empty.
- **Uniqueness** — no duplicate numbers (duplicate slugs are impossible because filenames must differ on disk; the command layer enforces slug uniqueness at creation).
- **Supersede integrity** — `## Supersedes: <slug>` target exists and has status `superseded`; `## Superseded-By: <slug>` target exists; decisions with status `superseded` MUST have a `## Superseded-By` pointer.
- **Cycles** — `supersedes`/`supersededBy` graph is checked for cycles using the same DFS pattern as `deps.go`.

Reuse the `ValidationResult` shape so the JSON output integrates for free with `--json`.

### CLI command: `decide`

New file `cmd/litespec/decide.go` following the per-command pattern in `project-structure`. Signature:

```go
func cmdDecide(args []string) error
```

Behavior:
1. Validate slug regex: `^[a-z0-9][a-z0-9-]*$`, no leading/trailing hyphen.
2. `ListDecisions(root)` to compute next number (highest + 1, or 1).
3. Check for slug collision across all existing decisions.
4. Write scaffold to `specs/decisions/NNNN-<slug>.md`.

Scaffold template (embedded via Go string constant, not a separate template file — too small to justify one):

```markdown
# <Title Placeholder>

## Status

proposed

## Context

<!-- What forces are at play? What constraints apply? -->

## Decision

<!-- What we decided and why. Use SHALL/MUST where intent is normative. -->

## Consequences

<!-- What becomes easier? What becomes harder? What must change elsewhere? -->
```

No `--status` flag at creation — every new decision starts `proposed`. Moving to `accepted` is a manual edit (we are not a workflow engine for decisions; git history tells that story).

### CLI integrations

**`list --decisions`** — new flag handled in `cmd/litespec/list.go`. Mutually exclusive with `--changes` and `--specs`. The `--sort` flag gains `number` as a valid value (default for decisions). Output shape mirrors the existing column-aligned table from `list`'s enriched listing.

**`view`** — update `cmd/litespec/view.go` to render a Decisions section. Placement: between Specifications and the optional Dependency Graph. Two-group layout (active + superseded summary) keeps the section bounded regardless of how many superseded decisions accumulate.

**`validate --decisions`** — new bulk flag. Positional name resolution gains a third lookup (`FindDecisionBySlug`) after change and spec lookups. `--type decision` is added to the existing disambiguation flag.

### Skill updates

`internal/skill/grill.go`, `propose.go`, and `review.go` gain short additions to their templates:

- **grill** — when a ruling emerges that is broader than the current change, suggest `litespec decide`.
- **propose** — during design.md authoring, prompt the AI to check whether any language sounds like a standing rule and cite an existing decision or create one.
- **review** — during artifact review, flag imperative language in `design.md` that looks like a cross-cutting rule ("all subagents must…", "we will never…") and recommend promotion.

These are prompt additions, not enforcement. All authoring happens through `litespec decide` and direct file editing, not through skill-mediated actions.

## Key Decisions

### Numbered filenames, not hashed or dated

**Decision:** `NNNN-<slug>.md` with a global monotonic counter.

**Alternatives considered:**
- **Date prefix (`2026-04-22-slug.md`)** — easy to sort chronologically but makes cross-reference awkward (decision slugs appear in prose without the date). Mirrors archive layout but archive is internal machinery; decisions are cited by humans.
- **Hash-based (`d7a3f1-slug.md`)** — stable under concurrent creation but unreadable. Nobody wants to read "per d7a3f1".
- **Slug-only (`slug.md`)** — ambiguous when two decisions touch the same concept at different times. Numbers give a natural supersede progression.

Numbered filenames match ADR conventions in the wider industry (MADR, Nygard) without importing their ceremony.

### No frontmatter

**Decision:** Status, supersedes, and superseded-by are parsed from H2 sections, not YAML frontmatter.

**Rationale:** Litespec has no frontmatter anywhere else. Changes use `.litespec.yaml` as a sidecar, but decisions are standalone and don't warrant a sidecar per file. Keeping everything in the markdown body means the decision renders correctly in any markdown viewer with no pre-processing.

### Decisions are not changes

**Decision:** Decisions have no `dependsOn`, no tasks, no deltas, no archive lifecycle.

**Rationale:** The whole point of the feature is a separate concept for locked reasoning. If decisions needed a workflow, they'd just be changes. The supersede pointer is the only lifecycle primitive — create a new decision, mark the old one superseded, point forward.

### Validator only checks declared pointers

**Decision:** Validation checks that `## Supersedes` and `## Superseded-By` resolve, but does NOT require decisions to be cited by any change or spec. Decisions can exist in isolation.

**Rationale:** Citation is prose. Forcing structural links would push us toward an IDL, which `cross-change-contracts` explicitly decided against. Dangling decisions are not an error — they may be consulted by humans outside the codebase.

### Skill nudges, not enforcement

**Decision:** Skills suggest promoting rulings to decisions but never block.

**Rationale:** A rigid promotion rule would fight reality. Some rulings are genuinely change-scoped even when they sound broad. Let the author decide; the skill just raises the question.

## File Changes

**New files:**
- `internal/decision.go` — parser, data types, directory scanning
- `internal/decision_test.go` — parser and list tests
- `cmd/litespec/decide.go` — CLI command

**Modified files:**
- `internal/validate.go` — `ValidateDecision`, extend `ValidateAll`, `--type decision`
- `internal/validate_test.go` — decision validation tests
- `internal/paths.go` — add `SkillInfo` entries if any skill templates reference decisions; otherwise unchanged
- `internal/json.go` — decision JSON serialization
- `cmd/litespec/list.go` — `--decisions` flag, status filter, number sort
- `cmd/litespec/view.go` — Decisions section rendering
- `cmd/litespec/main.go` — register `decide` command
- `cmd/litespec/completion.go` — completions for `decide`, `--decisions`
- `internal/completion.go` — dynamic slug completions for `validate <name>` and `--type decision`
- `internal/skill/grill.go`, `propose.go`, `review.go` — short prompt additions
- `DESIGN.md`, `AGENTS.md` — document the new artifact type

**No changes to:** canon spec format, delta merge logic, archive flow, existing change metadata.

## Risks and Mitigations

- **Risk:** Decisions become a dumping ground for everything the author doesn't want to specify properly.
  **Mitigation:** The four required sections (Context, Decision, Consequences, Status) create just enough friction to make decisions a deliberate act. The review skill's advisory flag helps surface misuse.

- **Risk:** Numbering collisions under concurrent edits (two people run `decide` simultaneously on different machines).
  **Mitigation:** Same failure mode as `git` for any file-creating tool. Resolution is manual — rename one file. Not worth solving at the tool level.

- **Risk:** The feature is small, and if nobody uses it, we've added CLI surface for nothing.
  **Mitigation:** It is opt-in and has zero runtime cost when unused (`specs/decisions/` doesn't need to exist). If it turns out nobody uses it after a month of real use, we remove it — cheap to delete.
