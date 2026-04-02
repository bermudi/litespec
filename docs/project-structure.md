# Project Structure

litespec uses a structured directory layout that makes specs the single source of truth while keeping work in progress isolated.

## Overview

```
project/
├── specs/                       # All spec-driven content
│   ├── canon/                   # Accepted capabilities (source of truth)
│   ├── changes/                  # Active work in progress
│   │   ├── <change-name>/        # Active change
│   │   └── archive/             # Completed changes
│   └── changes/archive/
│       └── YYYY-MM-DD-<name>/   # Archived change
└── .agents/skills/               # Generated AI skills (canonical)
```

## The `specs/` Directory

The `specs/` directory is the heart of litespec — all specification-driven development lives here.

### `specs/canon/` — Source of Truth

Canonical specs represent accepted capabilities. These are the definitive requirements that your codebase should satisfy.

```
specs/canon/
├── validate/
│   └── spec.md                  # Validation capabilities
├── archive/
│   └── spec.md                  # Archive behavior
├── status/
│   └── spec.md                  # Status reporting
└── <capability>/
    └── spec.md
```

Each capability has exactly one `spec.md` file. This is the canonical spec format.

#### Canonical Spec Format

```markdown
# <capability>

## Purpose                        (optional)

## Requirements                   (required)

### Requirement: <name>
<body text — must contain SHALL or MUST>

#### Scenario: <name>
- **WHEN** <condition>
- **THEN** <expected outcome>
```

Key rules:
- `## Requirements` is required — all `### Requirement:` blocks live here
- `## Purpose` is optional prose context before requirements
- ADDED and MODIFIED requirements must have at least one scenario
- Requirement body text must contain `SHALL` or `MUST`
- Scenarios use opaque WHEN/THEN text format

Example from `specs/canon/validate/spec.md`:

```markdown
# validate

## Requirements

### Requirement: JSON Output for Validate
The `litespec validate` command MUST support a `--json` flag that returns structured JSON output...

#### Scenario: Validate single change with JSON flag
- **WHEN** `litespec validate <change-name> --json` is run
- **THEN** output is valid JSON with `valid`, `errors`, and `warnings` fields
```

### `specs/changes/` — Work in Progress

Active changes live here, isolated from the canonical source of truth.

```
specs/changes/
├── <change-name>/               # Active change
│   ├── .litespec.yaml          # Metadata
│   ├── proposal.md             # Why & what
│   ├── design.md               # How (technical decisions)
│   ├── tasks.md                # Phased implementation checklist
│   └── specs/                  # Delta specs
│       └── <capability>/
│           └── spec.md
└── archive/                    # Completed changes (moved here)
    └── YYYY-MM-DD-<name>/
```

## Active Change Directory

Each active change contains all planning artifacts and delta specs needed to implement it.

### `.litespec.yaml` — Change Metadata

```yaml
schema: spec-driven
created: 2026-04-02T21:06:51Z
```

Minimal metadata. No phase tracking — that's derived from `tasks.md`.

### `proposal.md` — Why & What

Answers fundamental questions:

```markdown
# <change-name>

## Motivation
Why this change exists. What problem it solves.

## Scope
What's included. Specific deliverables and boundaries.

## Non-Goals
What's explicitly out of scope.
```

### `design.md` — How

Technical decisions and architecture:

```markdown
## Architecture
High-level structure. How pieces fit together.

## Decisions
Key trade-offs and choices with rationale.

## File Changes
Table of files being modified or created.
```

### `tasks.md` — Phased Checklist

Implementation organized into focused phases:

```markdown
## Phase 1: Foundation
- [ ] Set up database schema
- [ ] Create migration

## Phase 2: Core Logic
- [ ] Implement auth service
- [ ] Add middleware
```

Phase tracking is automatic — the first phase with unchecked tasks is the current phase.

### `specs/` — Delta Specs

Delta specs propose modifications to canonical specs using ADDED/MODIFIED/REMOVED/RENAMED operations:

```markdown
# <capability>

## ADDED Requirements
### Requirement: New Feature
The system SHALL do something new...

## MODIFIED Requirements
### Requirement: Existing Feature
The system SHALL now do this instead of that...

## REMOVED Requirements
### Requirement: Old Feature
(No body — just marks it for removal)

## RENAMED Requirements
### Requirement: Old Name
(Header changes, content preserved)
```

Delta operations merge in strict order at archive time: RENAMED → REMOVED → MODIFIED → ADDED.

## Archived Change Directory

Once a change is complete, `litespec archive` applies the deltas to `canon/` and moves the remaining planning artifacts to the archive.

```
specs/changes/archive/
└── 2026-04-02-shell-completions/
    ├── .litespec.yaml
    ├── proposal.md
    ├── design.md
    └── tasks.md                # No specs/ subtree!
```

**Critical:** Archived changes contain only planning artifacts. The `specs/` subtree is stripped because its contents have been merged into `specs/canon/` — the single source of truth.

## The `.agents/skills/` Directory

AI skills are generated into `.agents/skills/` — this is the canonical location.

```
.agents/skills/
├── litespec-explore/           # Thinking mode
├── litespec-grill/             # Relentless Q&A
├── litespec-propose/           # Create change + artifacts
├── litespec-continue/          # Add missing artifact
├── litespec-apply/             # Implement one phase
├── litespec-verify/            # Context-aware review
├── litespec-adopt/             # Reverse-engineer from code
└── litespec-archive/           # Merge deltas + archive
```

Each skill is a single `SKILL.md` file with YAML frontmatter:

```markdown
---
name: litespec-propose
description: Materialize a complete change proposal...
---

Enter propose mode. Your job is to...
```

Skills are lean — minimal token usage, no boilerplate. The `--tools claude` option creates symlinks in `.claude/skills/` for Claude Code integration.

## Why This Structure

### Git-Native

litespec works with git, not against it:
- Each change maps cleanly to a feature branch (`change/<name>`)
- Phases align naturally with per-phase commits
- Archive time is the merge point into main

### Spec-Driven

Specs are the source of truth:
- `canon/` contains accepted capabilities
- Delta specs in `specs/changes/<name>/specs/` propose modifications
- Archive merges deltas atomically

### Clear Separation

Active work is isolated:
- Active changes in `specs/changes/<name>/` don't affect the source of truth
- Archived changes preserve intent without creating parallel spec trees
- Skills provide workflow guidance without cluttering the codebase

### Progressive Rigor

The structure supports different workflows:
- **Quick Feature:** explore → grill → propose → apply → archive
- **Exploratory:** explore → grill (no artifacts if it doesn't pan out)
- **Adopt:** reverse-engineer specs from existing code

## File System Summary

```
project/
├── specs/
│   ├── canon/                    # Accepted capabilities
│   │   ├── <capability>/
│   │   │   └── spec.md          # Canonical spec
│   │   └── ...
│   └── changes/
│       ├── <name>/              # Active change
│       │   ├── .litespec.yaml
│       │   ├── proposal.md
│       │   ├── design.md
│       │   ├── tasks.md
│       │   └── specs/
│       │       └── <capability>/
│       │           └── spec.md  # Delta spec
│       └── archive/
│           └── YYYY-MM-DD-<name>/
│               ├── .litespec.yaml
│               ├── proposal.md
│               ├── design.md
│               └── tasks.md      # Planning artifacts only
└── .agents/skills/
    └── litespec-<skill>/
        └── SKILL.md              # Canonical skill file
```

This structure makes it easy to:
- See what's accepted vs. what's in progress
- Understand the evolution of a capability through archived changes
- Guide AI agents through the spec-driven development process
- Maintain a single source of truth while keeping work isolated
