# litespec

A lean, AI-native spec-driven development CLI.

`go 1.26+` | [Design Doc](DESIGN.md) | [Docs](https://bermudi.github.io/litespec/) | Inspired by [OpenSpec](https://github.com/Fission-AI/OpenSpec)

---

litespec gives AI coding agents structured workflows that keep your codebase aligned with your specifications. It's a reimagining of OpenSpec with stronger opinions: fewer concepts, leaner skills, unidirectional flow, and proper dangling-delta validation.

## Quick Start

```bash
# Install
go install github.com/bermudi/litespec/cmd/litespec@latest

# Initialize
litespec init

# Create a change
litespec new add-feature

# When done
litespec archive add-feature
```

## Documentation

**[Full Documentation → https://bermudi.github.io/litespec/](https://bermudi.github.io/litespec/)**

- [Getting Started](https://bermudi.github.io/litespec/getting-started/) — Installation and setup
- [Tutorial](https://bermudi.github.io/litespec/tutorial/) — Complete walkthrough from init to archive
- [CLI Reference](https://bermudi.github.io/litespec/cli-reference/) — All commands and flags
- [Workflow](https://bermudi.github.io/litespec/workflow/) — The spec-driven development workflow
- [Concepts](https://bermudi.github.io/litespec/concepts/) — Philosophy and why it works

## What Makes litespec Different

- **Convention over configuration** — zero config files. All defaults.
- **Unidirectional workflow** — `explore → grill → propose → [research →] apply → review → archive`. No backward flow.
- **Lean skills** — minimal tokens, zero boilerplate.
- **Git-native** — specs live in your repo. Branch per change, per-phase commits.
- **Read-only CLI** — the AI never writes through the CLI. It writes artifact files directly.
- **Dangling delta detection** — catches broken deltas during `validate`, not just at archive time.

## Status

This is an active experiment. Decisions made yesterday may be revised today if we find something better.
