# Glossary

The **ubiquitous language** is a concept from Domain-Driven Design — a shared, precise vocabulary that keeps humans and AI agents aligned when they talk, spec, and code. In litespec, that vocabulary lives in `specs/glossary.md` and is read by the planning and review skills.

## Why it matters

Without a shared vocabulary, the same concept gets a different name in every conversation. Drift accumulates: product.md says one thing, the spec says another, the code uses a third, and the next AI session invents a fourth. A ubiquitous language prevents that drift by making the core terms explicit and durable.

It is not a dictionary of every domain noun. It is a curated list of terms whose meaning affects how the system is specified, built, or reviewed.

## How litespec uses it

The glossary lives in the repo as `specs/glossary.md`. It is optional, but litespec works better when it exists.

| Skill | How it uses the glossary |
|-------|--------------------------|
| `litespec-plan` | Reads it in fuzzy and clear modes. When a new, stable, or ambiguous term appears, it nudges you to add it. It can also seed a new glossary. |
| `litespec-build` | May consult it as optional terminology context while implementing a unit. No enforcement. |
| `litespec-review` | May consult it while checking spec vs. implementation. No enforcement. |

If `specs/glossary.md` is missing, the skills degrade gracefully: no error, no block. The plan skill may suggest creating one once stable terms emerge.

## How to maintain it

The glossary is **curated**, not auto-generated. A skill proposes; a human approves. Only add terms that are:

- **Shared** — used in specs, code, decisions, and everyday conversation
- **Stable** — likely to mean the same thing in six months
- **Ambiguous** — easy to confuse with something else

To add or update terms, either:

1. Ask your agent to update it while running `litespec-plan`.
2. Edit `specs/glossary.md` directly.

Each entry is one line: a bold term, a colon and a space, then a concise definition. Where it helps, note what the term explicitly does **not** mean.

```markdown
- **Widget**: a customer-visible unit of work in the dashboard. Not a database row. Not a UI component.
```

## Source of truth

The terms themselves live in `specs/glossary.md`. This page explains the concept and the workflow — it does not duplicate or inline the terms.

