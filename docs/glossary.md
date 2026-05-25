# Glossary

The **ubiquitous language** is a concept from Domain-Driven Design (Eric Evans, 2003) — a shared vocabulary that keeps everyone (and every AI session) speaking the same language about the project.

## Why it matters

Without a shared vocabulary, AI agents and humans drift. The same concept gets called different things in different conversations, specs, and code. A ubiquitous language prevents that drift by defining the project's core terms in one place.

## How litespec uses it

The glossary lives in [`specs/glossary.md`](https://github.com/bermudi/litespec/tree/main/specs/glossary.md) — a single, version-controlled markdown file that serves as the project's shared vocabulary.

### Which skills read it

| Skill | Behavior |
|-------|----------|
| **think** | Reads the glossary at session start. Nudges when a concept seems foundational but isn't defined. Suggests creating one if it doesn't exist. |
| **plan** | After writing specs, checks whether new terms were introduced that aren't in the glossary. Offers to update. Offers to seed if no glossary exists. |
| **build** | May consult the glossary as optional terminology context after completing a phase. No enforcement. |
| **review** | May consult the glossary as supplementary terminology context during cross-change review. No enforcement. |

### Graceful degradation

Not every project starts with a glossary. All skills degrade gracefully — if `specs/glossary.md` doesn't exist, no error, no block. The conversation skills (think, plan) may suggest creating one when stable terms emerge.

## How to maintain it

The glossary is **curated**, not auto-generated. The AI proposes terms; the user approves. Only stable, shared, or ambiguous terms belong — not every noun from a proposal.

To add or update terms:

1. Ask your AI agent to update the glossary (the think and plan skills manage it as part of their workflow)
2. Or edit `specs/glossary.md` directly — entries use `- **Term**: definition` format

Each entry is one line: bold the term, follow with a colon and a space, then the definition. Keep definitions concise. Where disambiguation matters, include what a term explicitly does *not* mean.

## Source of truth

The living glossary is [`specs/glossary.md`](https://github.com/bermudi/litespec/tree/main/specs/glossary.md) in the repository. This page explains the concept — it does not duplicate or inline the terms.
