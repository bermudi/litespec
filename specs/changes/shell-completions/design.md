## Architecture

Two new commands added to the existing `os.Args` switch in `main.go`:

```
litespec completion <shell>     → cmdCompletion() → internal.CompletionScript(shell)
litespec __complete <words...>  → cmdComplete()  → internal.Complete(root, words)
```

The `__complete` backend is a pure function that takes the project root and the current command-line words, parses the cursor position against the known command/flag grammar, and returns `[]Completion` (candidate + description pairs).

Shell script templates live in `internal/completion/scripts/` as `.bash`, `.zsh`, `.fish` files, embedded via `go:embed`. The `CompletionScript(shell)` function returns the matching template as a string.

```
cmd/litespec/main.go          adds "completion" and "__complete" cases to switch
internal/completion.go         Complete() + CompletionScript() logic
internal/completion/scripts/   go:embed'd shell scripts
  litespec.bash
  litespec.zsh
  litespec.fish
```

## Decisions

### No CLI framework
Cobra/urfave would add completions for free but would require rewriting the entire CLI routing. The hand-rolled approach keeps the existing `main.go` intact, adds two switch cases, and keeps the dependency graph flat. The trade-off is maintaining three shell scripts and a completion resolver, but the CLI surface is small enough (~8 commands, ~15 flags) that this is manageable.

### Cobra-style `__complete` protocol
One backend mechanism for all three shells. Each shell script translates its native completion hook into `litespec __complete <words>` and parses the `candidate\tdescription` output. This avoids duplicating completion logic across three script languages.

### `candidate\tdescription` without directives
Cobra's `:N` directive system (controlling trailing space, file fallback, nospace) solves problems this CLI doesn't have. There's no position where file completion is needed — every completable slot resolves to a fixed or runtime-queried set. The shell scripts handle "nothing to complete" by not wiring up those positions.

### `go:embed` for shell scripts
Single binary, no runtime file deps. The scripts are ~50-80 lines each, rarely change. `go:embed` is stdlib.

### Silent fallback on errors
Completion is UX sugar. If `FindProjectRoot()` fails or `ListChanges()` errors, `__complete` prints nothing and exits 0. The user gets no completions rather than an error message interrupting their flow.

## File Changes

### `internal/completion.go` (new)
- `Completion` struct: `Candidate string`, `Description string`
- `Complete(root string, words []string) []Completion` — parses word position, dispatches to per-command resolvers
- `CompletionScript(shell string) (string, error)` — returns embedded script for the requested shell
- Internal helpers: `completeCommands()`, `completeChangeNames(root)`, `completeSpecNames(root)`, `completeArtifactIDs()`, `completeToolIDs()`, `completeShells()`
- Per-command parsers that understand flag positions (e.g., after `--change`, after `--tools`)

### `internal/completion/scripts/litespec.bash` (new)
- `_litespec()` function using `complete -F`
- Parses `COMP_WORDS`/`COMP_CWORD`, calls `litespec __complete "${COMP_WORDS[@]}"`, splits on tab, feeds first field to `COMPREPLY`
- ~60 lines

### `internal/completion/scripts/litespec.zsh` (new)
- `#compdef litespec`
- `_litespec()` function using `_arguments` for flags, dispatch for subcommands
- Calls `litespec __complete` for dynamic candidates, parses tab-separated output
- Uses `_describe` to show descriptions inline
- ~70 lines

### `internal/completion/scripts/litespec.fish` (new)
- `complete -c litespec` with `-n` conditions per command scope
- `-d` descriptions for commands and flags
- Calls `litespec __complete` for dynamic candidates
- ~80 lines

### `cmd/litespec/main.go` (modified)
- Add `case "completion"` → `cmdCompletion(os.Args[2:])`
- Add `case "__complete"` → `cmdComplete(os.Args[2:])`
- New `cmdCompletion(args []string)` — validates shell arg, calls `internal.CompletionScript(shell)`, prints to stdout
- New `cmdComplete(args []string)` — resolves root (silently ignores error), calls `internal.Complete(root, os.Args[1:])`, prints `candidate\tdescription` lines

### `DESIGN.md` (modified)
- Add `completion` and `__complete` to the CLI commands table
