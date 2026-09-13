# Getting Started

Install litespec, scaffold a project, and learn the commands you'll use every day.

## Prerequisites

- **Go 1.26.1 or later** — [Install Go](https://go.dev/dl/)
- **`gh` (recommended)** — authenticated GitHub CLI, so `view` and the queue commands can reach your issues. Without it, litespec falls back to local `specs/queues/<name>.md` files.

```bash
go version
gh auth status
```

## Installation

```bash
go install github.com/bermudi/litespec/v2/cmd/litespec@latest
export PATH="$HOME/go/bin:$PATH"
```

Persist the `PATH` line in your shell profile (`~/.bashrc`, `~/.zshrc`). To hack on litespec itself:

```bash
git clone https://github.com/bermudi/litespec.git
cd litespec
go build -o litespec ./cmd/litespec
```

Confirm the install:

```bash
litespec --help
```

## Initialize a project

Run this in the project root:

```bash
litespec init
```

This scaffolds:

```
your-project/
├── specs/
│   ├── product.md       # mental models + flows
│   ├── glossary.md      # ubiquitous language
│   └── decisions/       # durable rulings
└── .agents/
    └── skills/
        ├── litespec-plan/
        ├── litespec-build/
        └── litespec-review/
```

`.agents/skills/` is the canonical skill directory — nearly every AI coding agent discovers it natively. For Claude Code, which doesn't read `.agents/`, pass `--tools claude` once:

```bash
litespec init --tools claude
```

That symlinks the three skills into `.claude/skills/`. Later `litespec update` runs auto-detect and refresh those symlinks, so you only pass `--tools` again to add a new adapter.

After init, three commands carry the daily loop:

```bash
litespec view       # product + specs + decisions + open litespec GH issues
litespec validate   # structure check; success means structure ok, semantics not verified
litespec update     # regenerate the three skills from built-in templates
```

`litespec upgrade` checks for a newer binary (`go install` installations only) and reminds you to run `litespec update` afterward so generated skills match the new version.

## The two lanes

**Small fix** needs no issue. Tell the agent the fix; it reads the product, the relevant spec, decisions, and glossary; edits code; updates the spec in place if the contract changed.

**New feature** goes through the issue queue. The `litespec-plan` skill starts from a clean tree, records the current commit as `Base:`, creates `litespec/<change-name>`, and writes the labeled GH issue — ownership lines first, then proposal, design, and queue. If `gh` is unavailable, it writes the same body to `specs/queues/<name>.md`. `litespec-build` works through the queue one unit at a time on the recorded branch; unrelated work uses another branch or worktree.

## Validate your specs

`litespec validate` lints structure only — it never runs your code and never claims the implementation is correct:

- each requirement body contains `SHALL` or `MUST`
- each requirement has at least one `#### Scenario:` with `WHEN` and `THEN`
- decisions follow the `NNNN-<slug>.md` format with Status/Context/Decision/Consequences
- queue issues carry valid `Base:` and `Branch:` ownership lines, identified `Done means:` clauses mapped to named scenarios, and one executable `Verify:` per unit

```bash
$ litespec validate
structure ok; implementation semantics not verified: 1 capability, 2 requirements, 3 scenarios, 0 units
```

Scope the check when you need to:

```bash
litespec validate my-spec              # one spec or decision by name
litespec validate --specs              # specs only
litespec validate --decisions          # decisions only
litespec validate --issue 42           # one GH queue issue
litespec validate --queue specs/queues/add-auth.md
```

A checked unit must carry a complete red-green receipt — verbatim command, `unit digest:`, distinct pre/post SHAs, non-zero pre and zero post statuses, two nonempty fenced outputs, matching scope lines. A prose `Evidence:` label is not a receipt. The [Tutorial](tutorial.md) shows what a real one looks like; [Workflow](workflow.md) explains the full protocol.

## Enable shell completions

```bash
# Bash — current session
source <(litespec completion bash)
# Bash — persist
litespec completion bash > ~/.local/share/bash-completion/completions/litespec

# Zsh — persist (ensure ~/.zfunc is in your fpath, then compinit)
litespec completion zsh > ~/.zfunc/_litespec

# Fish — persist
litespec completion fish > ~/.config/fish/completions/litespec.fish
```

## Which AI tools work?

Any agent that reads `.agents/skills/` works out of the box — that's nearly all of them. Claude Code is the one exception: it only reads `.claude/skills/`, hence the `--tools claude` symlinks above. No other adapters exist; new ones are added only when a concrete tool needs one.

## Next steps

- [Tutorial](tutorial.md) — a complete feature cycle, issue body to closed issue
- [Workflow](workflow.md) — unit shape, evidence protocol, review routing
- [Concepts](concepts.md) — what makes a good spec
- [CLI Reference](cli-reference.md) — every command and flag

## Troubleshooting

### `litespec: command not found`

Add `~/go/bin` to your `PATH` and reload your shell.

### `litespec view` shows no GH issues

Check that `gh` is installed and authenticated, and that the repo has open issues labeled `litespec`. `view` runs `gh issue list --label litespec --state open --limit 10000`, silently showing only local specs when it can't reach GitHub.

### `validate` reports errors

Every requirement body must contain `SHALL` or `MUST`, and every requirement needs at least one `#### Scenario:` with both `WHEN` and `THEN`. Every queue unit needs `Done means:` with bracketed clause IDs, a `Scenarios:` block covering every ID, one executable `Verify:`, and a checkbox.
