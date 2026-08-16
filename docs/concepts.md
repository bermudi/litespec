# Concepts

## What a Spec IS (and ISN'T)

A spec is a contract — `specs/<feature>/spec.md` with load-bearing requirements written as SHALL/MUST and scenarios as WHEN/THEN. No `canon/` — edit the file directly. Only load-bearing features get a spec; small fixes edit it in place. GH issue is the queue — it holds proposal + design + queue (units with Done means + Verify), not `specs/changes/` or `backlog.md`.

## Why Spec-Driven Works

You think about **what** before **how**. `litespec-plan` in fuzzy mode asks 2-3 questions before any file; clear mode nails the GH issue + spec draft. Each unit is one demo-able outcome with a Verify that would fail without it — `build` must satisfy Verify before ticking the box. `review` is adversarial and checks spec vs implementation.

## What Makes a Good Requirement

SHALL/MUST is mandatory — body text must contain it. Each requirement has named scenarios (`#### Scenario: <name>`) with WHEN/THEN. One responsibility per requirement. Codebase-design reference enforces thin vertical slices, not horizontal layering.

## When to Use Litespec (and When Not To)

Use for load-bearing capabilities (CLI, API, file formats) that survive 6 months. Don't for one-offs, trivial refactors, or prototypes. Two lanes keep rigor proportional: small fix is zero ceremony (read product/spec/decisions -> edit -> done), new feature uses plan fuzzy/clear -> grill-me -> build -> review -> close issue.

## Durable vs Disposable

Durable (curated, small): `specs/product.md` (mental models + 2-3 flows), `specs/<feature>/spec.md`, `specs/decisions/NNNN-slug.md` (spine: true for load-bearing), `specs/glossary.md`. Disposable: GH issues (closed). If being stale would mislead, keep it; else delete after merge.

## The Bottom Line

Specs are communication — between present and future you, between you and teammates. GH issue is the queue that keeps what doesn't rot, drops the ceremony.
