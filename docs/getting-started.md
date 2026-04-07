# Getting Started

This guide walks you through installing litespec and setting up your first project.

## Prerequisites

litespec is a Go CLI tool. You need:

- **Go 1.26 or later** — [Install Go](https://go.dev/dl/)

Check your Go version:

```bash
go version
```

## Installation

### Install via `go install` (Recommended)

This is the quickest way to get litespec:

```bash
go install github.com/bermudi/litespec/cmd/litespec@latest
```

The binary will be installed to `~/go/bin/litespec`. Make sure this directory is on your PATH:

```bash
export PATH="$HOME/go/bin:$PATH"
```

Add this line to your shell profile (`~/.bashrc`, `~/.zshrc`, etc.) to persist it across sessions.

### Build from Source

If you prefer to build from source or want to modify litespec:

```bash
git clone https://github.com/bermudi/litespec.git
cd litespec
go build -o litespec ./cmd/litespec
```

Then move the binary somewhere on your PATH:

```bash
mv litespec ~/.local/bin/
```

### Verify Installation

Confirm litespec is installed and check its version:

```bash
litespec --version
```

You should see `litespec v0.1.0` (or a newer version).

## Initialize a Project

Navigate to your project directory and run:

```bash
litespec init
```

This creates the litespec project structure:

```
your-project/
├── specs/
│   ├── canon/          # Canonical specs (current capabilities)
│   ├── changes/        # Active change proposals
│   └── changes/archive/ # Completed changes
└── .agents/
    └── skills/         # Generated skill files for AI agents
```

### Set Up Tool Symlinks

If you're using **Claude Code**, generate symlinks so Claude can find the skills:

```bash
litespec init --tools claude
```

This creates symlinks in `.claude/skills/` pointing to the generated skills in `.agents/skills/`. Subsequent `litespec update` commands will auto-detect and refresh these symlinks without needing `--tools` again.

## Enable Shell Completions

litespec provides shell completions for bash, zsh, and fish.

### Bash

```bash
litespec completion bash > ~/.local/share/bash-completion/completions/litespec
```

Or source it directly in your `.bashrc`:

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

## Verify Your Setup

Run a quick check to confirm everything is working:

```bash
# List available commands
litespec --help

# Check project structure
ls -la specs/
ls -la .agents/skills/

# Validate everything is set up correctly
litespec validate
```

## Next Steps

Now that litespec is installed and initialized:

- Read the [Tutorial](tutorial.md) for a complete walkthrough from init to archive
- Learn about the [Workflow](workflow.md) for spec-driven development
- Explore the [CLI Reference](cli-reference.md) for all commands and flags

## Troubleshooting

### `litespec: command not found`

Make sure `~/go/bin` is on your PATH. Add this to your shell profile:

```bash
export PATH="$HOME/go/bin:$PATH"
```

### Symlink errors with `--tools claude`

On Windows, symlinks may require developer mode or administrator privileges. You can also manually copy the skills directory if symlinks aren't working.

### Completions not working

Restart your shell after installing completions. For zsh, run `compinit` manually:

```bash
autoload -U compinit && compinit
```

### `validate` reports errors after init

If you see validation errors immediately after init, ensure you're in a git repository. litespec expects to work within a version-controlled project.
