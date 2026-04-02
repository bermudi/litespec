## Phase 1: Completion Backend

- [x] Create `internal/completion.go` with `Completion` struct (`Candidate`, `Description`), `Complete(root string, words []string) []Completion`, and `CompletionScript(shell string) (string, error)`
- [x] Implement the word-position parser in `Complete()` — identify command, subcommand, flag position, and flag-argument position from the word list
- [x] Implement static candidate resolvers: `completeCommands()`, `completeArtifactIDs()`, `completeShells()`, per-command flag lists
- [x] Implement dynamic candidate resolvers: `completeChangeNames(root)`, `completeSpecNames(root)`, `completeToolIDs()` — all with silent error fallback
- [x] Wire per-command dispatch: map each command (init, new, list, status, validate, instructions, archive, update, completion) to its completable positions

## Phase 2: Shell Scripts

- [x] Create `internal/completion/scripts/litespec.bash` — `_litespec()` function using `complete -F`, parsing `COMP_WORDS`/`COMP_CWORD`, calling `__complete`, feeding first field to `COMPREPLY`
- [x] Create `internal/completion/scripts/litespec.zsh` — `#compdef litespec`, `_arguments` for flags, dispatch for subcommands, `_describe` for descriptions
- [x] Create `internal/completion/scripts/litespec.fish` — `complete -c litespec` with `-n` conditions, `-d` descriptions, dynamic candidate calls
- [x] Embed all three scripts via `//go:embed` in `internal/completion.go` and wire `CompletionScript()` to return the matching template

## Phase 3: CLI Wiring and Documentation

- [x] Add `case "completion"` and `case "__complete"` to the `main.go` switch, plus `cmdCompletion()` and `cmdComplete()` functions
- [x] Update `printUsage()` in `main.go` to include the `completion` command
- [x] Add completion tests: `internal/completion_test.go` covering static candidates, dynamic candidate resolution (with temp dirs), word-position parsing edge cases, and invalid shell rejection
- [x] Update `DESIGN.md` CLI commands table with `completion` and `__complete`
- [x] Build and run `go test ./...`, `go vet ./...` to verify everything passes
