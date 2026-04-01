# litespec

A lean, AI-native spec-driven development CLI.

`go 1.26+` | [Design Doc](DESIGN.md) | Inspired by [OpenSpec](https://github.com/Fission-AI/OpenSpec)

---

## What is litespec

litespec is a CLI tool that gives AI coding agents structured, spec-driven workflows. It reimagines OpenSpec with stronger opinions: fewer concepts, leaner skills, unidirectional flow, and proper dangling-delta validation.

The CLI is a read-only context provider. The AI writes artifacts directly. litespec just tells it what to do, where things are, and whether they're valid.

## Key Ideas

- **Convention over configuration** — zero config files. All defaults, all the time.
- **Unidirectional workflow** — `explore → grill → propose → apply → verify → archive`. No going backward.
- **Lean skills** — minimal tokens, zero boilerplate. Enriched enough to guide, short enough to not waste context.
- **`.agents/skills/` is canonical** — one home for all skills. Symlink adapter for Claude Code (`--tools claude`).
- **Git-native** — specs live in your repo. Branch per change, per-phase commits (future).
- **CLI is read-only** — structured data out, never writes from the CLI side.
- **Dangling delta detection** — catches broken deltas during `validate`, not just at archive time.

## Workflow

```
explore → grill → propose → apply → verify → archive
                     ↑                          │
                  continue                  adopt (separate path)
```

| Step | What happens |
|------|-------------|
| `explore` | Ephemeral thinking. No artifacts. Conversational. |
| `grill` | Relentless Q&A. Resolves every branch of the design tree before moving on. |
| `propose` | Materializes everything: change dir, proposal, specs, design, tasks. This is the commit point. |
| `continue` | Creates the next missing artifact one at a time. For partial proposals. |
| `apply` | Implements tasks per phase. One phase per invocation. |
| `verify` | Pure AI review of code vs specs. |
| `adopt` | Reverse-engineers specs from existing code. Separate path. |
| `archive` | Applies delta operations, moves change to archive. |

If something's wrong after `propose`, start over from `explore`/`grill`. No backward flow.

## Installation

```bash
go install github.com/bermudi/litespec/cmd/litespec@latest
```

Or build from source:

```bash
git clone https://github.com/bermudi/litespec.git
cd litespec
go build -o litespec ./cmd/litespec
```

Then move the binary somewhere on your PATH (e.g. `~/.local/bin`).

## Quick Start

```bash
# Initialize a project
litespec init

# Optional: symlink skills into .claude/skills/ for Claude Code
litespec init --tools claude

# Create a new change
litespec new add-user-auth

# See what's going on
litespec status --change add-user-auth

# Check everything is valid
litespec validate

# When done, merge and archive
litespec archive add-user-auth
```

Then use the skills in `.agents/skills/` with your AI agent. The skills tell the AI what to do — litespec tells the AI what exists.

## Commands

| Command | Purpose |
|---------|---------|
| `init [--tools <ids>]` | Scaffold `specs/` dir + generate skills (+ optional tool symlinks) |
| `new <name>` | Create a new change directory |
| `list [--specs\|--changes]` | List specs or changes |
| `status [--change <name>]` | Show artifact states (BLOCKED / READY / DONE) |
| `validate [--change <name>] [--all] [--strict]` | Validate structure, delta syntax, dangling deltas |
| `instructions <artifact>` | Return enriched instructions for AI to create an artifact |
| `archive <name>` | Apply deltas + move change to archive |
| `update [--tools <ids>]` | Regenerate skills (+ optional tool symlinks) without touching specs |

All commands support `--json` for structured output.

## Project Structure

```
project/
├── specs/
│   ├── specs/                        # source of truth (current capabilities)
│   │   └── <capability>/
│   │       └── spec.md
│   └── changes/                      # active changes
│       ├── <name>/
│       │   ├── .litespec.yaml        # metadata (schema + timestamp)
│       │   ├── proposal.md           # why & what
│       │   ├── design.md             # how (technical decisions)
│       │   ├── tasks.md              # phased implementation checklist
│       │   └── specs/                # delta specs
│       │       └── <capability>/
│       │           └── spec.md
│       └── archive/                  # completed changes (YYYY-MM-DD-<name>/)
└── .agents/skills/                   # generated skills (canonical)
    ├── litespec-explore/
    ├── litespec-grill/
    ├── litespec-propose/
    ├── litespec-continue/
    ├── litespec-apply/
    ├── litespec-verify/
    ├── litespec-adopt/
    └── litespec-archive/
```

## Delta Specs

Changes describe modifications to specs using delta markers:

| Marker | Behavior |
|--------|----------|
| `## ADDED Requirements` | Append to end of main spec |
| `## MODIFIED Requirements` | Replace matching requirement by header |
| `## REMOVED Requirements` | Delete from main spec |
| `## RENAMED Requirements` | Change section header, preserve content |

Applied in strict order at archive time: **RENAMED → REMOVED → MODIFIED → ADDED**.

litespec catches dangling deltas (references to non-existent requirements) during `validate` — not just at archive time. This is the kind of thing that saves you from a bad merge at the worst possible moment.

## Status

This is an active experiment. Decisions made yesterday may be revised today if we find something better.
