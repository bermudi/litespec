# cli-input

## ADDED Requirements

### Requirement: Universal JSON Output Flag

Every CLI command that produces substantive output SHALL accept a `--json` flag that emits structured JSON to stdout. The flag SHALL be supported on: `init`, `new`, `patch`, `list`, `status`, `validate`, `instructions`, `archive`, `preview`, `view`, `decide`, `update`, `upgrade`, and `import`. The `completion` and `__complete` commands are excluded — they produce shell-specific output that is not consumed as data.

In `--json` mode, commands SHALL emit only valid JSON to stdout. Human-facing prose, tips, and decorative formatting SHALL be suppressed. Errors and warnings encountered during execution SHALL be reported in the JSON structure, not on stderr (unless the command fails before producing any output, in which case stderr is acceptable).

#### Scenario: init with --json

- **WHEN** `litespec init --json` is run in a new directory
- **THEN** JSON is printed to stdout with keys for created directories, generated skills, and adapter symlinks, and no prose is printed

#### Scenario: init --json on already-initialized project

- **WHEN** `litespec init --json` is run in a directory that already has a `specs/` directory
- **THEN** JSON is printed with `initialized: false` and a `message` field explaining the project already exists

#### Scenario: archive with --json

- **WHEN** `litespec archive my-change --json` is run on a valid change
- **THEN** JSON is printed to stdout with keys for the archived change name, updated capabilities, and confirmation status

#### Scenario: archive --json with validation errors

- **WHEN** `litespec archive my-change --json` is run on a change with dangling deltas
- **THEN** an error is printed to stderr and exit code is 1 (no JSON output, since archive fails before producing results)

#### Scenario: decide with --json

- **WHEN** `litespec decide my-decision --json` is run
- **THEN** JSON is printed to stdout with keys for the decision number, slug, and file path

#### Scenario: decide --json with no arguments

- **WHEN** `litespec decide --json` is run without a slug
- **THEN** an error is printed to stderr and exit code is 1 (no JSON output)

#### Scenario: update with --json

- **WHEN** `litespec update --json` is run
- **THEN** JSON is printed to stdout with keys for updated skills and adapter symlinks

#### Scenario: update --json when nothing changed

- **WHEN** `litespec update --json` is run and no skills or adapters needed updating
- **THEN** JSON is printed with `skillsUpdated: true` and empty adapters list

#### Scenario: upgrade with --json

- **WHEN** `litespec upgrade --json` is run and an upgrade is available
- **THEN** JSON is printed to stdout with keys for the previous version, new version, and upgrade status

#### Scenario: upgrade --json already up to date

- **WHEN** `litespec upgrade --json` is run and the installed version matches the latest release
- **THEN** JSON is printed with `upgraded: false`, `currentVersion`, and a `message` field

#### Scenario: upgrade --json not a go-install binary

- **WHEN** `litespec upgrade --json` is run and the binary is not in GOBIN/GOPATH
- **THEN** an error is printed to stderr and exit code is 1 (no JSON output)

#### Scenario: import with --json

- **WHEN** `litespec import --json` is run on an OpenSpec project
- **THEN** JSON is printed to stdout with keys for imported canon specs, changes, archives, and any warnings

#### Scenario: import --json with no OpenSpec project

- **WHEN** `litespec import --json` is run on a directory without OpenSpec structure
- **THEN** an error is printed to stderr and exit code is 1 (no JSON output)

#### Scenario: --json flag registered in completion

- **WHEN** shell completion is invoked for any command supporting `--json`
- **THEN** `--json` appears as a completion candidate

### Requirement: Minimal Output Flag

Every CLI command that supports `--json` SHALL also accept a `--minimal` flag. When `--minimal` is set, the command SHALL emit only the core actionable data — no prose, no tips, no decorative formatting, no contextual suggestions. The output format depends on whether `--json` is also present:

- `--minimal --json`: emit JSON with only the action-relevant fields (no hints, no suggestions, no supplementary text)
- `--minimal` alone: emit the tersest human-readable text output — only the primary result

`--minimal` is designed for LLM context windows where token budget matters and supplementary text is noise.

#### Scenario: validate with --minimal --json

- **WHEN** `litespec validate --all --minimal --json` is run with no issues
- **THEN** JSON is printed with only `{"valid": true}` and no summary or counts

#### Scenario: validate with --minimal --json and errors

- **WHEN** `litespec validate --all --minimal --json` is run with errors
- **THEN** JSON is printed with `{"valid": false, "errors": [...]}` and no summary counts

#### Scenario: status with --minimal

- **WHEN** `litespec status my-change --minimal` is run
- **THEN** only the artifact states are printed (one per line, e.g., `proposal: done  specs: done`) with no header, no tips, no formatting

#### Scenario: list with --minimal

- **WHEN** `litespec list --minimal` is run
- **THEN** only change names are printed, one per line, with no status column, no born dates, no header

#### Scenario: view with --minimal

- **WHEN** `litespec view --minimal` is run
- **THEN** only the summary counts are printed (one per line) with no progress bars, no decorative separators, no tips

#### Scenario: --minimal without --json produces text

- **WHEN** `litespec validate --minimal` is run
- **THEN** terse text output is printed (e.g., `OK` or error lines) with no structured JSON

### Requirement: Minimal JSON Field Selection

When both `--minimal` and `--json` are set, commands SHALL emit a strict subset of the full `--json` output. The minimal subset SHALL contain only fields that represent the primary result of the command — validation status, error lists, created resources, or count summaries. Hints, suggestions, descriptions, and supplementary metadata SHALL be omitted.

#### Scenario: Full JSON includes summary, minimal does not

- **WHEN** `litespec validate --all --json` is run
- **THEN** the output includes `summary` with counts
- **WHEN** `litespec validate --all --minimal --json` is run
- **THEN** the output omits `summary` and includes only `valid` and `errors`

#### Scenario: Full JSON includes hints, minimal does not

- **WHEN** `litespec upgrade --json` is run and an upgrade succeeds
- **THEN** the output includes a `hint` field suggesting `litespec update`
- **WHEN** `litespec upgrade --minimal --json` is run and an upgrade succeeds
- **THEN** the output omits the `hint` field
