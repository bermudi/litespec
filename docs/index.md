# litespec

A lean, AI-native spec-driven development CLI.

litespec gives AI coding agents structured workflows that keep your codebase aligned with your specifications. It's a reimagining of OpenSpec with stronger opinions: fewer concepts, leaner skills, unidirectional flow, and proper dangling-delta validation.

The CLI is a read-only context provider. The AI writes artifacts directly. litespec tells it what to do, where things are, and whether they're valid.

---

## What makes litespec different

**Convention over configuration** — zero config files. All defaults, all the time. No stub `config.yaml` to fill in — it works out of the box.

**Unidirectional workflow** — `explore → grill → propose → apply → verify → archive`. No going backward. If something's wrong after propose, start over. This prevents partial states and confusion.

**Lean skills** — minimal tokens, zero boilerplate. Each skill is focused instructions, not pages of boilerplate that waste your AI context.

**Git-native** — specs live in your repo. Branch per change, per-phase commits. Your PRs carry the spec history.

**Read-only CLI** — the AI never writes through the CLI. It writes artifact files directly. The CLI exists to give the AI structured data (status, instructions, validation).

**Dangling delta detection** — catches broken deltas during `validate`, not just at archive time. This saves you from bad merges at the worst possible moment.

---

## Why use litespec

- **Structured workflows for AI agents** — give your AI a clear path from idea to implementation
- **Specs as source of truth** — your capabilities are documented, tested, and versioned
- **Isolated changes** — each proposed modification lives in its own directory with proposal, specs, design, and tasks
- **Delta-based merging** — modify specs with ADDED/MODIFIED/REMOVED/RENAMED markers that merge cleanly at archive time
- **Artifact-specific instructions** — `litespec instructions <artifact>` gives the AI targeted guidance for each artifact type
- **Works with your tools** — Claude Code, Cursor, and more via skill symlinks

---

## The workflow

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

---

## Quick start

```bash
# Initialize a project
litespec init

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

---

## Get started

[Installation & Setup](getting-started.md) → [Tutorial: Your First Change](tutorial.md) → [Concepts](concepts.md) → [CLI Reference](cli-reference.md)
