# cli-completion

## ADDED Requirements

### Requirement: Completion Command

The CLI SHALL provide a public `completion <shell>` command that prints a shell-native completion script to stdout. The `shell` argument MUST be one of `bash`, `zsh`, or `fish`. Any other value SHALL produce an error message to stderr and exit with code 1.

#### Scenario: Valid shell argument
- **WHEN** the user runs `litespec completion bash`
- **THEN** the command prints a bash-native completion script to stdout and exits with code 0

#### Scenario: Invalid shell argument
- **WHEN** the user runs `litespec completion powershell`
- **THEN** the command prints an error to stderr listing valid shells and exits with code 1

#### Scenario: Missing shell argument
- **WHEN** the user runs `litespec completion` with no argument
- **THEN** the command prints a usage message to stderr and exits with code 1

### Requirement: Hidden Complete Command

The CLI SHALL provide a hidden `__complete` command that receives the command-line words as positional arguments and prints completion candidates to stdout, one per line, in the format `candidate\tdescription`. This command SHALL NOT appear in help output or command completion suggestions. Errors during completion resolution SHALL produce no output (silent fallback).

#### Scenario: Complete command names
- **WHEN** the user invokes `litespec __complete litespec` (completing the first word after `litespec`)
- **THEN** the command prints all public command names (init, new, list, status, validate, instructions, archive, update, completion) with their descriptions, one per line, excluding `__complete`

#### Scenario: Complete change names
- **WHEN** the user invokes `litespec __complete litespec status ` (cursor after the positional argument slot for `status`)
- **THEN** the command prints the names of active changes from `specs/changes/`, one per line

#### Scenario: Error during filesystem access
- **WHEN** `__complete` is invoked outside a litespec project and attempts to list changes
- **THEN** the command prints nothing and exits with code 0

### Requirement: Dynamic Completion Resolution

The `__complete` command SHALL resolve candidates dynamically from runtime state for the following positions:
- Change names from `ListChanges()` (filesystem)
- Spec names from `ListSpecs()` (filesystem)
- Tool IDs from the `Adapters` var
- Artifact IDs from the `Artifacts` var (proposal, specs, design, tasks)

Static completions (command names, flag names, shell names) SHALL be hardcoded in the completion resolver.

#### Scenario: Complete tool IDs
- **WHEN** the user invokes `litespec __complete litespec init --tools ` (cursor after `--tools `)
- **THEN** the command prints each ID from the `Adapters` var with its description

#### Scenario: Complete artifact IDs
- **WHEN** the user invokes `litespec __complete litespec instructions ` (cursor after `instructions `)
- **THEN** the command prints each ID from the `Artifacts` var with its description

#### Scenario: Complete flags for a command
- **WHEN** the user invokes `litespec __complete litespec validate --`
- **THEN** the command prints all flags valid for `validate` (--all, --changes, --specs, --strict, --json, --type) with their descriptions

### Requirement: Bash Completion Script

The completion script for bash SHALL register a `_litespec()` completion function using `complete -F`. The function SHALL parse `COMP_WORDS` and `COMP_CWORD`, invoke `litespec __complete` with the current word list, and feed candidates to `COMPREPLY`. Descriptions from the `__complete` output SHALL be ignored (bash has no native description support).

#### Scenario: Bash tab-completes commands
- **WHEN** the user sources the bash script and types `litespec v<tab>`
- **THEN** bash offers `validate` as a completion

#### Scenario: Bash tab-completes change names
- **WHEN** the user types `litespec status <tab>` in a project with changes `foo` and `bar`
- **THEN** bash offers `foo` and `bar` as completions

### Requirement: Zsh Completion Script

The completion script for zsh SHALL use `#compdef litespec` and implement a completion function using `_arguments` for flag definitions and a dispatch function that calls `litespec __complete` for dynamic candidates.

#### Scenario: Zsh tab-completes commands with descriptions
- **WHEN** the user sources the zsh script and types `litespec <tab>`
- **THEN** zsh offers all public commands with their descriptions shown inline

#### Scenario: Zsh tab-completes change names
- **WHEN** the user types `litespec archive <tab>` in a project with changes
- **THEN** zsh offers change names as completions

### Requirement: Fish Completion Script

The completion script for fish SHALL use `complete -c litespec` with `-n` conditions for command-scoped completions and `-d` for descriptions. It SHALL call `litespec __complete` for dynamic candidates.

#### Scenario: Fish tab-completes commands with descriptions
- **WHEN** the user sources the fish script and types `litespec <tab>`
- **THEN** fish offers all public commands with descriptions

#### Scenario: Fish tab-completes flags
- **WHEN** the user types `litespec list -<tab>`
- **THEN** fish offers `--specs`, `--changes`, and `--json` with descriptions

### Requirement: Completion Output Format

The `__complete` command SHALL print each candidate as a single line in the format `candidate\tdescription` where `\t` is a literal tab character. Shell-specific scripts SHALL parse this format and use the description field as appropriate for their shell.

#### Scenario: Output format with tab separator
- **WHEN** `__complete` returns candidates for commands
- **THEN** each line contains the command name, a tab character, and a one-line description (e.g., `init\tInitialize project structure`)
