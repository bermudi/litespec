## Motivation

litespec is a CLI tool with 8 commands and a growing flag surface. Users currently have to remember every command, flag, and valid value (change names, artifact IDs, tool IDs) from memory or `--help`. Shell completions are table stakes for any CLI that expects regular use — they reduce cognitive load, prevent typos, and make the tool discoverable.

## Scope

Add dynamic shell completion support for bash, zsh, and fish:

- A public `litespec completion <shell>` command that prints a shell-native completion script to stdout
- A hidden `litespec __complete <words...>` backend that resolves dynamic candidates (change names, spec names, artifact IDs, tool IDs) from the runtime state
- Native completion scripts per shell — fish gets descriptions, zsh gets proper `_arguments` handling, bash gets standard `complete -F`
- Completion output format: one `candidate\tdescription` pair per line
- Errors during completion produce no output (silent fallback)

Completable positions:
- `$1`: all public commands (init, new, list, status, validate, instructions, archive, update, completion)
- `init --tools`, `update --tools`: tool IDs from the `Adapters` var at runtime
- `status --change`, `validate --change`, `instructions --change`: change names from filesystem
- `instructions <artifact>`: artifact IDs (proposal, specs, design, tasks, apply)
- `archive <name>`: change names from filesystem
- `completion <shell>`: static list (bash, zsh, fish)
- Static flags per command (`--json`, `--specs`, `--changes`, `--all`, `--strict`, `--allow-incomplete`)

## Non-Goals

- No `--install` flag or auto-detection of completion directories. Users eval or pipe the output themselves.
- No completion for flag values that are free-text (e.g., `new <name>`).
- No Cobra-style `:N` directive system — the shell scripts handle fallback logic natively.
- No migration to a CLI framework (Cobra, urfave/cli). This stays hand-rolled.
