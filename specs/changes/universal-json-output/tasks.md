## Phase 1: Flag infrastructure and JSON types for init, decide, update, upgrade

- [ ] Add `--json` and `--minimal` flags to all relevant commands in `internal/commandspec.go`: `init`, `archive`, `decide`, `update`, `upgrade`, `import` get `--json`; every `--json` command gets `--minimal`
- [ ] Add `minimalFlag = "--minimal"` constant and `parseOutputFlags(args) (asJSON, asMinimal bool)` helper to `cmd/litespec/helpers.go`
- [ ] Update help text functions (`printInitHelp`, `printArchiveHelp`, `printDecideHelp`, `printUpdateHelp`, `printUpgradeHelp`, `printImportHelp`) and `printUsage` to document `--json` and `--minimal`
- [ ] Add `--json`/`--minimal` output to `cmdInit` — local `initResultJSON` struct, idempotent handling (`initialized: false` on re-run)
- [ ] Add `--json`/`--minimal` output to `cmdDecide` — local `decideResultJSON` struct
- [ ] Add `--json`/`--minimal` output to `cmdUpdate` — local `updateResultJSON` struct
- [ ] Add `--json`/`--minimal` output to `cmdUpgrade` — local `upgradeResultJSON` struct, handles already-up-to-date and not-go-install cases
- [ ] Add CLI tests for `--json` on each command (init, decide, update, upgrade) including error/no-op cases
- [ ] Run `go build`, `go vet`, `go test ./...`

## Phase 2: JSON types for archive and import

- [ ] Add `--json`/`--minimal` output to `cmdArchive` — local `archiveResultJSON` struct, validation errors go to stderr (no JSON)
- [ ] Add `--json`/`--minimal` output to `cmdImport` — local `importResultJSON` struct, includes warnings array
- [ ] Add CLI tests for `archive --json` and `import --json` including error cases
- [ ] Run `go build`, `go vet`, `go test ./...`

## Phase 3: Wire `--minimal` into existing commands

Deferrable — this phase is additive and orthogonal to Phases 1-2. Can be a separate patch if desired.

- [ ] Add `--minimal` parsing to `cmdStatus`, `cmdList`, `cmdValidate`, `cmdInstructions`, `cmdView`, `cmdPreview`, `cmdNew`, `cmdPatch`
- [ ] Implement minimal output paths for each: terse text for `--minimal`, field-filtered JSON for `--minimal --json`
- [ ] Add CLI tests for `validate --minimal --json` and `list --minimal`
- [ ] Run `go build`, `go vet`, `go test ./...`
