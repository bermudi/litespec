## Architecture

The skill system is a pipeline: `SkillInfo` entries in `paths.go` define metadata, Go templates in `internal/skill/` provide content, and `litespec update` generates `.agents/skills/<name>/SKILL.md` files. This change replaces 11 pipeline entries with 4. The pipeline itself is unchanged — `GenerateSkills`, `GenerateAdapterCommands`, `ValidateSkillTemplates` all work the same way, just against a smaller set.

The four new skills are internal consolidations. The CLI surface (`litespec new`, `litespec patch`, `litespec validate`, etc.) is unaffected. Agent behavior changes only in how skills are loaded and routed — agents now see 4 skill descriptions instead of 11.

## Decisions

### Merge by stance, not by workflow step

**Chosen:** Group skills by behavioral stance (think/plan/build/review) rather than workflow step (explore/grill/propose/etc.).

**Why:** Stance is what the model needs to classify correctly. A user saying "let me think about this" should load one skill, not guess between explore, grill, and workflow. Stance-based groups have non-overlapping descriptions and non-overlapping behavioral postures.

**Trade-off:** Each skill is larger (~10-15KB vs ~3-7KB). The model must self-route within the skill to the right section. This is acceptable because (a) the skill can detect mode from `litespec status` output, and (b) intra-skill routing is more reliable than inter-skill routing since the full context is already loaded.

### Kill research as a standalone skill

**Chosen:** Fold research into the build skill as a pause condition — "when you hit a knowledge gap, gather docs and optionally produce a research skill file."

**Why:** No user thinks "time for the research phase." It's ceremony around something a competent agent does naturally. The one good idea (persist research as skill files for future reference) is two sentences in the build skill template, not a standalone skill.

**Trade-off:** Less structure around research. But the former research skill was already optional and its 5KB of instructions mostly described a process (triage, gather, write) that any decent model does without hand-holding.

### Kill workflow as a standalone skill

**Chosen:** Fold workflow routing into the think skill. Think detects current phase and suggests next steps.

**Why:** Every skill already runs `litespec status` at startup. Workflow was a meta-skill describing the system rather than doing work. Its 5KB of "detect state, suggest next step" is redundant when think can include the same routing logic as a section.

### Kill glossary as a standalone skill

**Chosen:** Fold glossary management into the plan skill as a section.

**Why:** Glossary is a format rule ("maintain specs/glossary.md"), not a behavioral stance. It's tightly coupled to the planning phase — you create/update glossary entries when writing proposals. Making it a standalone skill forces the model to classify "I want to update the glossary" as a separate action when it's really part of planning.

### Kill patch as a standalone skill

**Chosen:** Fold patch instructions into the plan skill.

**Why:** Patch is just "create a lightweight change without full planning artifacts." It's a mode of planning, not a different stance. The plan skill can detect patch mode from `.litespec.yaml` and skip proposal/design/tasks accordingly.

### Keep review separate from build

**Chosen:** Four skills instead of three. Review is its own skill.

**Why:** Review's adversarial stance conflicts with build's builder stance. A skill file containing both "write production code" and "adversarially attack production code" contaminates the context window before the first user token. The model is primed as a builder and produces softer reviews. Review's existing design already runs adversarial review *before* reading code to avoid anchoring bias — that stance can't coexist with builder instructions in the same priming context.

## File Changes

### `internal/paths.go`
- **Change:** Replace 11 `SkillInfo` entries with 4 (think, plan, build, review). Remove: explore, grill, propose, research, apply, adopt, workflow, glossary, patch, fix.
- **Why:** Single source of truth for skill metadata. Satisfies "Four Skill Registration" requirement.

### `internal/skill/` — Delete 11 files
- **Delete:** `explore.go`, `grill.go`, `propose.go`, `research.go`, `review.go`, `apply.go`, `adopt.go`, `workflow.go`, `glossary.go`, `patch.go`, `fix.go`
- **Why:** Templates for removed skills. Their content is absorbed into the 4 new templates.

### `internal/skill/` — Create 4 files
- **Create:** `think.go` — Merges explore + grill + workflow. Registers template via `init()`. Contains: exploration mode, grilling mode, workflow status detection, glossary/backlog reading at session start. ~120 lines.
- **Create:** `plan.go` — Merges propose + patch + adopt + glossary. Registers template via `init()`. Contains: proposal creation, patch-mode detection, adopt flow, glossary management section, backlog graduation. ~200 lines.
- **Create:** `build.go` — Merges apply + fix + research-pause. Registers template via `init()`. Contains: phased implementation, fix workflow (CRITICAL→WARNING→SUGGESTION), research pause condition, glossary passive reference. ~180 lines.
- **Create:** `review.go` — Review only, unchanged stance. Registers template via `init()`. Contains: lifecycle-aware review (artifact/implementation/pre-archive), adversarial-first approach, compliance review. ~320 lines (porting from existing review.go at 314 lines).

### `internal/skill/skill_test.go`
- **Change:** Update `knownIDs` from `["explore", "grill", "propose", "review", "apply", "adopt", "glossary", "patch", "fix", "research", "workflow", ...]` to `["think", "plan", "build", "review", ...]` (keeping artifact IDs).
- **Why:** Tests verify all registered templates are non-empty. Satisfies "Skill Generation Tests" requirement.

### `internal/skill/skill.go`
- **No change.** The `Register`, `Get`, `ValidateSkillTemplates` infrastructure is generic and works with any set of skill IDs.

### `internal/skill/artifact.go`
- **No change.** Artifact templates (proposal, specs, design, tasks) are independent of domain skills.

### `.agents/skills/` — Generated
- **After `litespec update`:** 4 directories generated (litespec-think, litespec-plan, litespec-build, litespec-review). 11 legacy directories removed. Adapter symlinks updated.

### `AGENTS.md`
- **Change:** Update workflow diagram (remove research, collapse skill names). Update Core Concepts (remove research phase, remove individual skill names). Update Key Design Decisions (remove research skills section). Update Skill Generation conventions to reference 4 skills.

### `DESIGN.md`
- **Change:** Update directory tree from 7 skill directories to 4. Update skill-related sections to reflect new names.
