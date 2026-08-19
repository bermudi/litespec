# view

## Requirements

### Requirement: Project Root Discovery

The `litespec view` command SHALL require a project root by locating a `specs/` directory in the current directory or any ancestor. If no `specs/` directory is found, the command SHALL exit with an error.

#### Scenario: Outside a project

- **WHEN** the user runs `litespec view` in a directory with no `specs/` directory in any ancestor
- **THEN** the command prints an error and exits non-zero

### Requirement: Output Flags

`litespec view` SHALL accept `--json` and `--minimal` flags. `--json` SHALL output structured JSON, `--minimal` SHALL output a compact tab-separated summary, and the two together SHALL output the minimal JSON summary.

#### Scenario: Default text output

- **WHEN** the user runs `litespec view` with no flags
- **THEN** the command prints the full text dashboard

#### Scenario: JSON output

- **WHEN** the user runs `litespec view --json`
- **THEN** the command prints the full JSON representation and no text dashboard

### Requirement: Minimal Text Output

`litespec view --minimal` SHALL output a single tab-separated summary line formatted as `<N> specs<TAB><M> reqs<TAB><K> decisions<TAB><L> issues`, where `K` is the count of active (non-superseded) decisions.

#### Scenario: Minimal tab-separated summary

- **WHEN** the user runs `litespec view --minimal` in a project with 3 specs, 7 requirements, 2 active decisions, and 1 open GH issue
- **THEN** the output is a single line containing `3 specs`, `7 reqs`, `2 decisions`, and `1 issues` separated by tab characters

### Requirement: Dashboard Header and Footer

The text dashboard SHALL print the title `Litespec Dashboard` followed by a 60-character outer border of `═` characters at the top, and SHALL close with the same 60-character `═` border.

#### Scenario: Borders surround the dashboard

- **WHEN** the user runs `litespec view`
- **THEN** the output contains `Litespec Dashboard`, then a line of 60 `═` characters, and ends with a matching line of 60 `═` characters

### Requirement: Product Section

The text dashboard SHALL display a `Product:` section. When `specs/product.md` exists, it SHALL trim surrounding whitespace, truncate the file content to 400 characters (appending `…` if truncated), take the first line, and print `specs/product.md — <firstLine>` followed by `product: mental models + flows`. When `specs/product.md` is missing, it SHALL print `specs/product.md — missing (run litespec init to scaffold)` and `product: not yet initialized`.

#### Scenario: Product file present

- **WHEN** `specs/product.md` exists with first line `Widget Manager`
- **THEN** the Product section shows `specs/product.md — Widget Manager` and `product: mental models + flows`

#### Scenario: Product file missing

- **WHEN** `specs/product.md` does not exist
- **THEN** the Product section shows `specs/product.md — missing (run litespec init to scaffold)` and `product: not yet initialized`

### Requirement: Summary Section

The text dashboard SHALL display a `Summary:` section containing `● Specifications: <N> specs, <M> requirements`. When decisions are loaded successfully and at least one exists, it SHALL also print `● Decisions: <active>/<total>`, where `active` is the count of non-superseded decisions. When GH issues are loaded successfully and at least one exists, it SHALL also print `● GH Issues: <N> open`.

#### Scenario: Summary with all counts

- **WHEN** the project has specs, active decisions, and open GH issues
- **THEN** the Summary section shows the specifications, decisions, and GH issues counts

#### Scenario: Summary with only specifications

- **WHEN** the project has specs but no decisions and no open GH issues
- **THEN** the Summary section shows only `● Specifications: <N> specs, <M> requirements`

### Requirement: Specifications Section

The text dashboard SHALL display a `Specifications` section underlined with 60 `─` characters. For each feature spec it SHALL list the spec name padded to 30 characters, its requirement count, the singular or plural label `requirement`/`requirements`, and the path `(specs/<name>/spec.md)`, sorted by requirement count descending and prefixed with a `▪` bullet. When no specs exist, it SHALL print the placeholder `(no feature specs yet — add specs/<feature>/spec.md)`.

#### Scenario: Specs sorted by requirement count

- **WHEN** `specs/auth/spec.md` has 1 requirement and `specs/view/spec.md` has 5 requirements
- **THEN** the Specifications section lists `view` before `auth` with their counts and paths

#### Scenario: No feature specs

- **WHEN** the `specs/` directory contains no feature spec directories
- **THEN** the Specifications section shows the placeholder `(no feature specs yet — add specs/<feature>/spec.md)`

### Requirement: Decisions Section

The text dashboard SHALL display a `Decisions` section underlined with 60 `─` characters when decisions are loaded successfully and at least one exists. It SHALL list active (non-superseded) decisions sorted by number ascending, printing `*` before the number when the decision has `spine: true`, then the four-digit zero-padded number, the slug padded to 30 characters, and the status. When superseded decisions exist, it SHALL print a single `superseded: <N>` line after the active list. If no decisions exist or loading fails, the section SHALL be omitted entirely.

#### Scenario: Active and superseded decisions

- **WHEN** the project contains an accepted spine decision `0001-auth.md`, a proposed decision `0002-cache.md`, and a superseded decision `0003-old.md`
- **THEN** the Decisions section contains a line starting with `  *0001` for the spine decision, a line starting with `   0002` for the proposed decision, and `superseded: 1`

#### Scenario: No decisions

- **WHEN** the project has no `specs/decisions/` directory
- **THEN** the dashboard contains no Decisions section

### Requirement: GH Issues Section

The text dashboard SHALL display a `GH Issues (open)` section underlined with 60 `─` characters when `gh` is available and at least one open issue is returned, printing each issue as `#<number> <title> <url>`. When `gh` is not on `PATH` and no issues are found, it SHALL display a `GH Issues` section with the notice `(gh not available — showing local specs only)`. When `gh` is present but returns no open issues, the section SHALL be omitted.

#### Scenario: Open issues listed

- **WHEN** `gh` is on `PATH`, the project is in a git work tree, and `gh issue list` returns one open issue titled `Fix auth`
- **THEN** the GH Issues section lists the issue with `#<number>`, `Fix auth`, and its URL

#### Scenario: gh not available

- **WHEN** `gh` is not installed on `PATH`
- **THEN** the dashboard shows the `GH Issues` notice `gh not available — showing local specs only`

### Requirement: GitHub Issue Fetch

`litespec view` SHALL fetch open GitHub issues by running `gh issue list --json number,title,state,url --state open --limit 50` from the project root. It SHALL require `gh` to be on `PATH` and the project to be in a git work tree; if the `.git` directory is absent, it SHALL fall back to `git rev-parse --is-inside-work-tree`. Any failure SHALL be silent and result in an empty issue list.

#### Scenario: gh invocation

- **WHEN** the project has a `.git` directory and `gh` is installed
- **THEN** the command invokes `gh issue list` with the exact JSON fields, `--state open`, and `--limit 50`

#### Scenario: Missing git work tree

- **WHEN** the project has no `.git` directory and `git rev-parse` does not output `true`
- **THEN** the GH Issues section is empty and no error is shown

### Requirement: JSON Output

With `--json`, `litespec view` SHALL output a JSON object containing `summary` (`specs`, `requirements`, optional `decisions` with `active` and `total`, and optional `ghIssues` count), `specs` (`name` and `requirementCount`), `decisions` (active decisions sorted by number ascending with `number`, `slug`, `status`, and `spine`), `product` (`path`, `exists`, and `preview` when the product file exists), and `ghIssues` (`number`, `title`, `state`, `url`). When `--minimal` is also supplied, only the `summary` object SHALL be returned.

#### Scenario: Full JSON

- **WHEN** the user runs `litespec view --json` in a project with specs, decisions, a product, and GH issues
- **THEN** the output contains `summary`, `specs`, `decisions`, `product`, and `ghIssues` fields

#### Scenario: Minimal JSON

- **WHEN** the user runs `litespec view --json --minimal`
- **THEN** the output contains only `summary` with counts

### Requirement: Data Loading

`litespec view` SHALL load feature specs via `internal.ListSpecs` and sum their requirement counts for the summary. It SHALL load decisions via `internal.ListDecisions` and tolerate errors by omitting the Decisions section. It SHALL load `specs/product.md` and tolerate a missing file by showing the missing message.

#### Scenario: Requirement count from specs

- **WHEN** the project contains `specs/auth/spec.md` with 2 requirements and `specs/view/spec.md` with 5 requirements
- **THEN** the summary shows `7 requirements`

#### Scenario: Tolerate missing decisions

- **WHEN** `specs/decisions/` cannot be read
- **THEN** the dashboard omits the Decisions section and continues
