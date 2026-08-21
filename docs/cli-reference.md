# CLI Reference

Complete command-line reference for litespec v2.

```
Usage: litespec <command> [options]

Workflow (two lanes):
  Small fix: read product/spec/decisions -> edit code -> update spec if contract
  New feature: plan[fuzzy] -> plan[clear] (GH issue) -> grill-me -> build -> review -> close

Commands:
  init [--tools <ids>]              Initialize project structure
  validate [--all|--specs|--decisions|--issue N|--queue PATH] [--type T]   Validate specs, decisions, and queues
  view                              Dashboard overview
  update [--tools <ids>]            Regenerate skills and adapters
  upgrade                           Check for and install the latest version
  completion <shell>                Generate shell completion script (bash, zsh, fish)

Tools:
  claude    Symlink skills into .claude/skills/ for Claude Code

Flags:
   --version    Print version
   --help       Print this help message
   --json       Output structured JSON (validate, view)
   --strict     Treat warnings as errors (validate)
   --all        Validate all specs, decisions, and queues
   --specs      Validate all specs only
   --decisions  Validate all decisions only
   --issue N    Fetch and validate one GH queue issue
   --queue PATH Validate one local queue file
   --type       Disambiguate name type: spec|decision (validate)
```

The flags above are the top-level defaults. Each command also supports its own flags and aliases, shown below.

## Global Flags

| Flag | Description |
|------|-------------|
| `--version` | Print version information |
| `--help` | Print help message |
| `--json` | Output structured JSON where supported |
| `--strict` | Treat warnings as errors (`validate` only) |
| `--all` | Validate all specs, decisions, and queues (`validate` only) |
| `--specs` | Validate all specs only (`validate` only) |
| `--decisions` | Validate all decisions only (`validate` only) |
| `--issue N` | Fetch and validate one GH queue issue (`validate` only) |
| `--queue <path>` | Validate one local queue file (`validate` only) |
| `--type <T>` | Disambiguate name type: `spec` or `decision` (`validate` only) |

## `init`

Usage:

```bash
litespec init [--tools <ids>] [--json] [--minimal]
```

Description:

Initialize a new litespec project in the current directory.

Creates:

- `specs/product.md`
- `specs/glossary.md`
- `specs/decisions/`
- `.agents/skills/` (`litespec-plan`, `litespec-build`, `litespec-review`)

Flags:

| Flag | Description |
|------|-------------|
| `--tools <ids>` | Comma-separated tool IDs (e.g., `claude`) |
| `--json` | Output as JSON |
| `--minimal` | Minimal output |

Examples:

```bash
litespec init
litespec init --tools claude
litespec init --json
```

## `validate`

Usage:

```bash
litespec validate [<name>|--all|--specs|--decisions|--issue N|--queue <path>] [--type T] [--strict] [--json] [--minimal]
```

Description:

Validate the structure of specs, decisions, and queue issues/files.

Spec checks:

- Requirement body text contains `SHALL` or `MUST`.
- Each requirement has at least one `#### Scenario:` with `WHEN` and `THEN`.
- Specs are valid Markdown and parseable.

Decision checks:

- Required sections: Context, Decision, Consequences.
- Valid status: `proposed`, `accepted`, or `superseded`.
- No duplicate numbers or slugs.
- Supersede pointers resolve and point to `superseded` decisions.
- No supersede cycles.

Queue checks:

- Exactly one `Base:` and `Branch:` ownership line before the first `##` heading.
- `Base:` is a full commit ID and `Branch:` matches `litespec/<change-name>`.
- Every unit has `Done means:`, an executable `Verify:`, and a checkbox.
- `Depends:` references resolve to units in the same queue.

Flags:

| Flag | Description |
|------|-------------|
| `<name>` | Validate a specific spec or decision by name |
| `--all` | Validate all specs, decisions, and queues |
| `--specs` | Validate all specs only |
| `--decisions` | Validate all decisions only |
| `--issue N` | Fetch and validate one GH queue issue |
| `--queue <path>` | Validate one local queue file |
| `--type <T>` | Disambiguate name: `spec` or `decision` |
| `--strict` | Treat warnings as errors |
| `--json` | Output as JSON |
| `--minimal` | Minimal output |

Default behavior with no arguments is equivalent to `--all`.

Examples:

```bash
litespec validate
litespec validate my-spec
litespec validate --all --strict
litespec validate shared --type spec
litespec validate --decisions
litespec validate --issue 42
litespec validate --queue specs/queues/add-auth.md
```

## `view`

Usage:

```bash
litespec view [--json] [--minimal]
```

Description:

Display a dashboard overview of product, specs, decisions, and open `litespec` GH issues.

If `gh` is installed and the project is a Git repository with a GitHub remote, `view` calls `gh issue list --label litespec` and shows open queue issues. Otherwise it shows only local specs and decisions.

Flags:

| Flag | Description |
|------|-------------|
| `--json` | Output as JSON |
| `--minimal` | Minimal output |

Examples:

```bash
litespec view
litespec view --json
```

## `update`

Usage:

```bash
litespec update [--tools <ids>] [--json] [--minimal]
```

Description:

Regenerate skills and adapter symlinks from the built-in templates.

- Writes `.agents/skills/<skill>/SKILL.md` for each generated skill.
- Writes reference files from `internal/skill/templates/references/`.
- Cleans stale skill directories and symlinks.
- Auto-detects active adapters (e.g., existing `.claude/skills/` symlinks).
- Does not modify `specs/` content.

Flags:

| Flag | Description |
|------|-------------|
| `--tools <ids>` | Comma-separated tool IDs (e.g., `claude`) |
| `--json` | Output as JSON |
| `--minimal` | Minimal output |

Examples:

```bash
litespec update
litespec update --tools claude
```

## `upgrade`

Usage:

```bash
litespec upgrade [--json] [--minimal]
```

Description:

Check for the latest release and upgrade via `go install github.com/bermudi/litespec/cmd/litespec@latest`.

Only works for binaries installed with `go install`. Exits with no change if already up to date.

Flags:

| Flag | Description |
|------|-------------|
| `--json` | Output as JSON |
| `--minimal` | Minimal output |

Examples:

```bash
litespec upgrade
litespec upgrade --json
```

## `completion`

Usage:

```bash
litespec completion <shell>
```

Description:

Generate a shell completion script.

Supported shells: `bash`, `zsh`, `fish`.

Examples:

```bash
# Bash
litespec completion bash > ~/.local/share/bash-completion/completions/litespec
eval "$(litespec completion bash)"

# Zsh
litespec completion zsh > ~/.zsh/completion/_litespec
fpath=(~/.zsh/completion $fpath)
autoload -U compinit && compinit

# Fish
litespec completion fish > ~/.config/fish/completions/litespec.fish
```

`completion` has no flags and accepts exactly one shell argument.

## Tool Adapters

The `--tools` flag for `init` and `update` creates tool-specific adapter symlinks that point to the canonical `.agents/skills/` directory.

| Tool ID | Name | Skills Directory |
|---------|------|------------------|
| `claude` | Claude Code | `.claude/skills/` |

Run `litespec init --tools claude` once. Subsequent `litespec update` calls auto-detect and refresh the symlinks.

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | Error, validation failed, or invalid arguments |

## See Also

- [Workflow](workflow.md)
- [Concepts](concepts.md)
- [Getting Started](getting-started.md)
