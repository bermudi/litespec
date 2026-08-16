# litespec

A lean, AI-native spec-driven development CLI.

litespec gives AI coding agents structured workflows that keep your codebase aligned with specifications. GH issue is the queue — proposal + design + queue live in the GH issue body, not `specs/changes/`. Three lean skills, two lanes, direct spec edits.

---

## What makes litespec different

**GH issue is the queue** — not `specs/changes/` or `QUEUE.md`. 64k limit, no overflow. Offline fallback: `specs/changes/<name>/QUEUE.md`.

**Two lanes** — small fix (zero ceremony, no issue) vs new feature (plan fuzzy -> clear -> grill-me -> build one unit at a time -> review -> close issue).

**3 skills only** — `litespec-plan` (fuzzy/clear + grilling/codebase-design/domain-modeling), `litespec-build` (one unit with Done means + Verify), `litespec-review` (adversarial). Generated via `litespec update` into `.agents/skills/`.

**Direct spec edits** — `specs/<feature>/spec.md` with SHALL/MUST + WHEN/THEN. No `canon/`, no `backlog.md`, no ADDED/MODIFIED.

---

## Why use litespec

- **Structured workflows for AI** — clear path from fuzzy idea to demo-able unit
- **Specs as source of truth** — only load-bearing contracts, curated and small
- **GH-native queue** — issues are the backlog, `view` shows `gh issue list` when available

---

## The workflow

Two lanes, unidirectional, no backward flow. `litespec init` scaffolds `specs/product.md`, `specs/glossary.md`, `specs/decisions/` + skills. `litespec new <name> --issue N` links to GH issue. `litespec validate` lints specs + decisions + GH issue Verify. `litespec view` shows product + specs + GH issues + spine.

---

## Get started

[Installation & Setup](getting-started.md) → [Tutorial: Your First Change](tutorial.md) → [Concepts](concepts.md) → [CLI Reference](cli-reference.md)
