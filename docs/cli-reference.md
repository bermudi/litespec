# CLI Reference

Every command in `litespec --help`, with usage, flags, and examples.

```
Usage: litespec <command> [options]

Workflow (two lanes):
  Small fix: read product/spec/decisions -> edit code -> update spec if contract
  New feature: plan[fuzzy] -> plan[clear] (GH issue) -> grill-me -> build -> review -> close

Commands:
  init [--tools <ids>]              Initialize project structure
  validate [--all|--specs|--decisions|--issue <N>|--queue <path>] [--type T]   Validate specs, decisions, and queues
  view                              Dashboard overview
  update [--tools <ids>]            Regenerate skills and adapters
  digest --issue <N> | --queue <p>  Print expected unit contract digests for a queue
  receipt --issue <N> | --queue <p> --heading "<h>"  Assemble evidence receipt comment files
  issue check --issue <N> --heading "<h>"  Tick exactly one unit checkbox (managed)
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
   --issue <N>  Fetch and validate one GH queue issue
   --queue <path>  Validate one local queue file
   --type       Disambiguate name type: spec|decision (validate)
```

`--json` and `--minimal` are supported where each command's help says so. Success from `validate` always reads `structure ok; implementation semantics not verified` — structure only, never a claim the code is correct. Exit `0` on success, `1` on error, validation failure, or bad arguments. No removed commands (`new`, `list`, `status`, `instructions`, `import`, `preview`, `archive`, `patch`) exist.

## `init`

```bash
litespec init [--tools <ids>] [--json] [--minimal]
```

Scaffold a project: `specs/product.md`, `specs/glossary.md` (if missing), `specs/decisions/`, and the three generated skills in `.agents/skills/`.

| Flag | Description |
|------|-------------|
| `--tools <ids>` | Comma-separated tool IDs (today only `claude`: symlinks skills into `.claude/skills/`) |
| `--json` | Output as JSON |
| `--minimal` | Minimal output |

```bash
litespec init
litespec init --tools claude
```

## `validate`

```bash
litespec validate [<name>|--all|--specs|--decisions|--issue <N>|--queue <path>] [--type T] [--strict] [--json] [--minimal]
```

Lint structure of specs, decisions, and queues. No arguments means `--all`.

Spec checks: requirement bodies contain `SHALL`/`MUST`; every requirement has a `#### Scenario:` with `WHEN`/`THEN`; specs parse. Decision checks: `NNNN-<slug>.md` naming; Status (`proposed`/`accepted`/`superseded`), Context, Decision, Consequences; supersede pointers resolve (superseded decisions point forward); no duplicate numbers/slugs or cycles. Queue checks: exactly one `Base:` (full SHA) and one `Branch:` (`litespec/<name>`) before the first `##`; every unit has identified `Done means:` clauses fully mapped through `Scenarios:`, one executable `Verify:` (fenced block linted with `bash -n`; inline backtick command accepted; vacuous commands rejected), a checkbox, optional unique-nonempty `Read first:`/`Constraints:`/`Depends:` (depends must resolve), `Boundary:` from the closed `filesystem`/`process`/`network` vocabulary with a complete five-risk `Risk cases:` block when present; checked units carry a complete red-green receipt with matching `unit digest:` (see [Workflow](workflow.md)); rebuild requests, re-plan markers, and amendment chains resolve.

| Flag | Description |
|------|-------------|
| `<name>` | One spec or decision by name |
| `--all` | Everything: specs, decisions, queues (the default) |
| `--specs` | Specs only |
| `--decisions` | Decisions only |
| `--issue <N>` | Fetch and validate one GH queue issue |
| `--queue <path>` | Validate one local queue file |
| `--type <T>` | Disambiguate: `spec` or `decision` |
| `--strict` | Warnings become errors |
| `--json` | Output as JSON |
| `--minimal` | Minimal output |

```bash
litespec validate
litespec validate my-spec
litespec validate shared --type spec
litespec validate --all --strict
litespec validate --decisions
litespec validate --issue 42
litespec validate --queue specs/queues/add-auth.md
```

## `view`

```bash
litespec view [--json] [--minimal]
```

Dashboard: product, spec/requirement counts, active/total decisions (spine starred), and open `litespec`-labeled GH issues. Runs `gh issue list --label litespec --json number,title,state,url --state open --limit 10000`; when `gh` or the work tree is missing it silently shows local content only.

| Flag | Description |
|------|-------------|
| `--json` | Full JSON (`summary`, `specs`, `decisions`, `product`, `ghIssues`); with `--minimal`, `summary` only |
| `--minimal` | One tab-separated line: `<N> specs  <M> reqs  <K> decisions  <L> issues` |

```bash
litespec view
litespec view --json
```

## `update`

```bash
litespec update [--tools <ids>] [--json] [--minimal]
```

Regenerate `.agents/skills/<name>/SKILL.md` plus `references/` from the built-in templates. Removes stale `litespec-*` directories, refuses to write through symlinks, leaves project-specific skills (like `the-drill`) alone. Without `--tools`, auto-detects active adapters from existing symlinks.

| Flag | Description |
|------|-------------|
| `--tools <ids>` | Comma-separated tool IDs (today only `claude`) |
| `--json` | Output as JSON |
| `--minimal` | Minimal output |

```bash
litespec update
litespec update --tools claude
```

## `digest`

```bash
litespec digest --issue <N> | --queue <path> [--heading <text>]
```

Print each queue unit's identity and expected contract digest — one tab-separated `occurrence  heading  digest` line per unit. Build pastes this digest into the receipt verbatim; `validate` recomputes it from the unit's contract fields, so any post-evidence edit to heading, `Done means:`, `Scenarios:`, risks, or `Verify:` fails validation until plan witnesses an amendment.

| Flag | Description |
|------|-------------|
| `--issue <N>` | GH issue number (requires `gh`) |
| `--queue <path>` | Local queue file |
| `--heading <text>` | Only lines whose heading equals this exactly (every duplicate occurrence listed; no match exits non-zero) |

```bash
litespec digest --issue 42
litespec digest --queue specs/queues/add-auth.md --heading "Enforce per-IP cap"
```

## `receipt`

```bash
litespec receipt --issue <N> | --queue <path> --heading "<text>" [--occurrence <K>]
  --pre-sha <sha> --pre-status <n> --pre-out <file>
  --post-sha <sha> --post-status <n> --post-out <file>
  [--rebuild] [--recovered-from <id>] [--post] [--out <dir>]
```

Assemble one validator-clean evidence receipt for the resolved unit (exact heading + positive same-heading occurrence, same identity as `digest`) and emit numbered `receipt-0001.md`, `receipt-0002.md`, … comment files. Issue mode also prints the exact `gh issue comment` commands in posting order; queue mode emits files only. Emit-only by default — nothing posts, nothing ticks boxes. The receipt self-parses through the evidence grammar before anything is written; ambiguous/unknown/out-of-range headings, unreadable outputs, equal SHAs, zero pre status, non-zero post status, and non-ancestor pre all refuse before any write. Over-long receipts split across comments (never truncated); a mid-sequence write failure removes already-written files so no partial set remains. Emitted paths and printed commands are absolute, so posting works from any directory.

| Flag | Description |
|------|-------------|
| `--issue <N>` | GH issue number (requires `gh`) |
| `--queue <path>` | Local queue file |
| `--heading <text>` | Exact unit heading (required) |
| `--occurrence <K>` | Which duplicate heading (default: must be unambiguous) |
| `--pre-sha <sha>` | Full pre commit SHA |
| `--pre-status <n>` | Pre exit status (must be non-zero) |
| `--pre-out <file>` | File holding raw pre output |
| `--post-sha <sha>` | Full post commit SHA |
| `--post-status <n>` | Post exit status (default 0) |
| `--post-out <file>` | File holding raw post output |
| `--rebuild` | Include the rebuild routing identity (rebuilding a reviewed unit) |
| `--recovered-from <id>` | Append-only recovery provenance: earlier complete receipt ID |
| `--post` | Opt-in: run the printed `gh issue comment` commands in order; first `gh` failure stops the chain with posted/unposted named, no retry |
| `--out <dir>` | Emission directory (default: current directory; created when missing) |

```bash
go test ./internal/ratelimit -run TestCounterWindow > /tmp/pre.out; echo "exit=$?"
# ... implement, commit ...
go test ./internal/ratelimit -run TestCounterWindow > /tmp/post.out
litespec receipt --issue 42 --heading "Sliding window counter" \
  --pre-sha <pre> --pre-status 1 --pre-out /tmp/pre.out \
  --post-sha <post> --post-out /tmp/post.out
```

Build normally runs this for you; reach for it directly when assembling a receipt by hand.

## `issue check`

```bash
litespec issue check --issue <N> --heading "<text>" [--occurrence <K>]
```

Tick exactly one unit checkbox. Resolves by the same heading + occurrence identity as `digest`, then writes back only when the result differs from the fetched body by that single `- [ ]` → `- [x]` flip with `Base:`/`Branch:` byte-unchanged. Anything else — ambiguous or unknown heading, out-of-range occurrence, already-checked target, any other body delta, `gh` failure — refuses with no write. No offline lane, no broader body management: hand-editing issue bodies is retired.

```bash
litespec issue check --issue 42 --heading "Sliding window counter"
```

## `upgrade`

```bash
litespec upgrade [--json] [--minimal]
```

Check the GitHub tags for the channel's latest (stable ignores prereleases; prereleases follow both) and `go install` it when newer. `go install` placements only — elsewhere it errors. After a successful upgrade it reminds you to run `litespec update` in your projects. A silent background check runs at most weekly without output or blocking.

```bash
litespec upgrade
```

## `completion`

```bash
litespec completion <shell>
```

Print a completion script derived from the `CommandSpecs` registry. `bash`, `zsh`, or `fish`; no flags, exactly one argument.

```bash
# Persist
litespec completion bash > ~/.local/share/bash-completion/completions/litespec
litespec completion zsh > ~/.zfunc/_litespec        # ~/.zfunc in fpath, then compinit
litespec completion fish > ~/.config/fish/completions/litespec.fish
# Or try for one session
eval "$(litespec completion bash)"
```

## Tool adapters

`--tools` on `init`/`update` creates adapter symlinks pointing at canonical `.agents/skills/`.

| Tool ID | Name | Skills directory |
|---------|------|------------------|
| `claude` | Claude Code | `.claude/skills/` |

Pass `--tools claude` once; later `update` runs auto-detect and refresh. Adding a CLI command or flag means updating the `CommandSpecs` registry in `internal/commandspec.go` — completions derive from it.

## See also

- [Workflow](workflow.md) — unit shape, evidence, routing, closure
- [Concepts](concepts.md) — durable vs disposable
- [Getting Started](getting-started.md) — install and first commands
