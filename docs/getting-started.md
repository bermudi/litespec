# Getting Started

This guide walks you through installing litespec v2 and running the commands you'll use every day. v2 is leaner: the GH issue is the queue, feature specs live at `specs/<feature>/spec.md`, and there are three generated skills.

## Prerequisites

litespec is a Go CLI. You need:

- **Go 1.26 or later** — [Install Go](https://go.dev/dl/)

Check your version:

```bash
go version
```

## Installation

### Install via `go install` (recommended)

```bash
go install github.com/bermudi/litespec/cmd/litespec@latest
```

The binary lands in `~/go/bin/litespec`. Add that directory to your PATH:

```bash
export PATH="$HOME/go/bin:$PATH"
```

Put that in your shell profile (`~/.bashrc`, `~/.zshrc`, etc.) to persist it.

### Build from source

If you want an unreleased commit or to hack on litespec:

```bash
git clone https://github.com/bermudi/litespec.git
cd litespec
go build -o litespec ./cmd/litespec
mv litespec ~/.local/bin/
```

### Verify the binary

Check the help output to confirm the v2 command set:

```bash
$ litespec --help
Usage: litespec <command> [options]

Workflow (two lanes):
  Small fix: read product/spec/decisions -> edit code -> update spec if contract
  New feature: plan[fuzzy] -> plan[clear] (GH issue) -> grill-me -> build -> review -> close

Commands:
  init [--tools <ids>]              Initialize project structure
  validate [--all|--specs|--decisions] [--type T]   Validate specs and decisions
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
   --all        Validate all specs and decisions
   --specs      Validate all specs only
   --decisions  Validate all decisions only
   --type       Disambiguate name type: spec|decision (validate)
```

## Initialize a project

Run this in the project root:

```bash
$ litespec init --tools claude
Created specs/ directory structure
Generated .agents/skills/
Generated adapter commands for: claude
Project initialized.

GH issue is the queue — proposal + design + queue live in the GH issue body. Run `litespec view` to see product + specs + GH issues.
```

This scaffolds:

```
your-project/
├── specs/
│   ├── product.md       # mental models + flows
│   ├── glossary.md      # ubiquitous language
│   └── decisions/       # durable rulings
├── .agents/
│   └── skills/
│       ├── litespec-plan/
│       ├── litespec-build/
│       └── litespec-review/
└── .claude/
    └── skills/          # symlinks to .agents/skills/ (Claude Code)
```

`.agents/skills/` is the canonical skill directory. The `.claude/skills/` path is only generated when you pass `--tools claude` or run `litespec update --tools claude`.

## Regenerate skills

After you change specs or decisions, refresh the generated skills:

```bash
$ litespec update
Updated .agents/skills/
Updated adapter symlinks for: claude
```

`litespec update` auto-detects active adapters, so you usually don't need `--tools` again.

To check for a newer binary, run `litespec upgrade` (only works for `go install` installations).

## Enable shell completions

### Bash

```bash
litespec completion bash > ~/.local/share/bash-completion/completions/litespec
```

Or source it directly:

```bash
source <(litespec completion bash)
```

### Zsh

```bash
litespec completion zsh > ~/.zsh/completion/_litespec
```

Add to your `.zshrc`:

```bash
fpath=(~/.zsh/completion $fpath)
autoload -U compinit && compinit
```

### Fish

```bash
litespec completion fish > ~/.config/fish/completions/litespec.fish
```

## View the dashboard

`litespec view` shows product, specs, and open GH issues:

```text

Litespec Dashboard

════════════════════════════════════════════════════════════
Product:
  specs/product.md — # Product
  product: mental models + flows

Summary:
  ● Specifications: 0 specs, 0 requirements

Specifications
────────────────────────────────────────────────────────────
  (no feature specs yet — add specs/<feature>/spec.md)

════════════════════════════════════════════════════════════

```

If `gh` is installed and authenticated, open issues appear under `GH Issues (open)`.

## Start a feature

In v2, `litespec-plan` in `clear` mode creates the GH issue with the `litespec` label. The issue body is proposal + design + queue (64k limit). No folder is created — the GH issue is the queue.

If `gh` is unavailable, `plan[clear]` writes the same body to `specs/queues/<name>.md`, where `<name>` is the change name chosen during planning.

The issue body is the plan. The `litespec-build` skill works through it one unit at a time.

## Validate your specs

`litespec validate` lints `specs/<feature>/spec.md` and decisions:

- each requirement body contains `SHALL` or `MUST`
- each load-bearing requirement has at least one `#### Scenario:` with `WHEN` and `THEN`
- decisions follow the `NNNN-<slug>.md` format

Before you add specs:

```bash
$ litespec validate
ok: 0 capabilities, 0 requirements, 0 scenarios
```

## Next steps

- [Tutorial: Your First Feature](tutorial.md) — a complete feature cycle
- [Workflow](workflow.md) — small fix vs new feature
- [Concepts](concepts.md) — what makes a good spec
- [CLI Reference](cli-reference.md) — command details

## Troubleshooting

### `litespec: command not found`

Add `~/go/bin` to your PATH and reload your shell.

### `litespec view` doesn't show GH issues

Check that `gh` is installed, authenticated, and that the repo has open issues labeled `litespec`. `view` runs `gh issue list --label litespec --state open`.

### `validate` reports errors

Every requirement body must contain `SHALL` or `MUST`. Every load-bearing requirement needs at least one `#### Scenario:` with both `WHEN` and `THEN`.
