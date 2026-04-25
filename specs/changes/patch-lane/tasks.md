# Tasks

## Phase 1: Demote planning artifacts and add content validation

- [x] In `internal/validate.go`, remove `proposal.md`, `design.md`, `tasks.md` from the `requiredFiles` slice (lines 21-38) so their absence no longer fails validation
- [x] Add `validateProposal(content string) []ValidationIssue` enforcing `## Motivation`/`## Why` and `## Scope`/`## What Changes` headings with non-blank bodies; wire into `ValidateChange` only when `proposal.md` exists
- [x] Add `validateDesign(content string) []ValidationIssue` enforcing at least one `## ` heading and ≥3 non-blank lines outside fenced code blocks; wire in only when `design.md` exists
- [x] Extend `validateTasksChecklist` (or add sibling) to require ≥1 checkbox per `## Phase` block; emit error identifying the empty phase heading
- [x] Add tests in `internal/validate_test.go` covering: missing trio passes, empty proposal/design/tasks fails with expected messages, full proposal still passes, legacy `## Why`/`## What Changes` aliases accepted, code-fence-only design content fails
- [x] Run `go build ./...`, `go vet ./...`, `go test ./...` — all green

## Phase 2: Patch-mode detection via metadata and artifact state handling

- [ ] Add `Mode string` field (yaml tag `"mode,omitempty"`) to `ChangeMeta` in `internal/types.go`
- [ ] Add `IsPatchMode(root, name string) bool` in `internal/change.go` — reads `ChangeMeta` via `ReadChangeMeta`, returns true when `Mode == "patch"`; returns false on read error (graceful degradation)
- [ ] Modify `LoadArtifactStates` in `internal/artifact.go` to return only `{specs: ArtifactDone}` for patch-mode changes (omit `proposal`, `design`, `tasks` keys entirely)
- [ ] Verify `LoadChangeContext` in `internal/change.go` still works correctly since it wraps `LoadArtifactStates` — no change needed unless tests reveal otherwise
- [ ] Add tests in `internal/change_test.go` for `IsPatchMode`: true for change with `mode: patch` in metadata, false for change without mode field, false for change with no `.litespec.yaml`, false for change with `mode: ""` (empty string)
- [ ] Add tests in `internal/artifact_test.go` verifying `LoadArtifactStates` returns single-key map for patch mode and full four-key map for full proposal
- [ ] Run `go build ./...`, `go vet ./...`, `go test ./...` — all green

## Phase 3: `litespec patch` command

- [ ] Create `cmd/litespec/patch.go` implementing `cmdPatch(args []string) error` — parses `<name> <capability>` positional args, validates them with the same rules as `litespec new`, refuses if `specs/changes/<name>/` exists, creates `specs/changes/<name>/specs/<capability>/spec.md` with stub content (`# <capability>\n\n## ADDED Requirements\n`), writes `.litespec.yaml` with `mode: patch`
- [ ] Wire `patch` into the command dispatcher in `cmd/litespec/main.go` and add to `printUsage()`
- [ ] Add `printPatchHelp()` following the pattern of other commands
- [ ] Create `cmd/litespec/patch_test.go` covering: happy path (creates expected files, no proposal/design/tasks), missing args error, existing change error, invalid name error
- [ ] End-to-end smoke: from a temp project, run `litespec patch foo bar`, then `litespec validate foo` — both succeed
- [ ] Run `go build ./...`, `go vet ./...`, `go test ./...` — all green

## Phase 4: Status command patch-mode rendering

- [ ] In `cmd/litespec/status.go`, branch on `IsPatchMode` when rendering a single change: emit only the `specs` artifact line followed by `(patch mode)` indicator, suppress proposal/design/tasks lines
- [ ] Apply the same branching in the bulk listing path of `status.go`
- [ ] Add `Mode string` field (json tag `mode,omitempty`) to the JSON status struct in `internal/json.go` (or wherever `BuildChangeStatusJSON` lives); set to `"patch"` for patch-mode changes; ensure non-applicable artifacts are either omitted from the array or marked `status: "n/a"`
- [ ] Update `cmd/litespec/status_test.go` (or `main_test.go`) to verify text output for patch mode (no trio lines, indicator present) and JSON output (`mode: "patch"` present, artifacts handled correctly)
- [ ] Run `go build ./...`, `go vet ./...`, `go test ./...` — all green

## Phase 5: View dashboard patch-mode category

- [ ] In `cmd/litespec/view.go`, partition active changes into patch-mode and full-proposal buckets when collecting changes
- [ ] Render new "Patch Changes" section after Active Changes, before Completed Changes; use `◆` bullet; show change name, born date, touched-relative; no progress bar
- [ ] Update summary section to include patch-changes count alongside draft/active/completed counts
- [ ] Omit the section entirely when no patch changes exist
- [ ] Update existing dashboard tests in `cmd/litespec/main_test.go` / `view_test.go` to verify patch section appears, summary count updates, and absent section when no patch changes
- [ ] Run `go build ./...`, `go vet ./...`, `go test ./...` — all green

## Phase 6: Patch skill, workflow update, and regenerated artifacts

- [ ] Append a `SkillInfo{ID: "litespec-patch", ...}` entry to the `Skills` slice in `internal/paths.go` with description matching the registry guidelines
- [ ] Create `internal/skill/patch.go` registering the template via `init()`; body explains when to use patch vs propose, the `patch → implement → archive` flow, and that no proposal/design/tasks artifacts are produced
- [ ] Update `internal/skill/workflow.go` template to document the patch lane as a sibling of the propose flow with a short "when to choose" guidance
- [ ] Update `internal/skill/skill_test.go` expected skill ID list to include `patch`
- [ ] Run `litespec update` to regenerate `.agents/skills/litespec-patch/SKILL.md` and refresh `.agents/skills/litespec-workflow/SKILL.md`
- [ ] Update `AGENTS.md` Workflow section to document the patch lane and the rule "the delta is the contract; planning artifacts are optional scaffolding"
- [ ] Update `DESIGN.md` to document `IsPatchMode` and where it is used
- [ ] Run `go build ./...`, `go vet ./...`, `go test ./...` — all green
- [ ] End-to-end smoke: create a patch change, run `litespec validate`, `litespec status`, `litespec view`, then `litespec archive`; verify canon updated and archive directory contains no specs/ subtree
