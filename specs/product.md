# Product

## Mental Models

- Agent-native workflows over file-heavy ceremony
- GH issue is the queue — proposal + design + queue live in the issue body
- Lean skills with progressive disclosure (fuzzy/clear/grilling/codebase-design/domain-modeling)

## Flows

1. Small fix — zero ceremony: You say "fix typo" -> agent reads product + relevant spec + decisions/glossary -> edits code -> updates the one specs/<feature>/spec.md if contract change -> done. No new, no issue required.

2. New feature — plan[fuzzy] (read code, ask 2-3 questions, no files) -> plan[clear] (write GH issue: proposal + design + units with Verify; also draft spec if load-bearing) -> you: "looks good" or "grill-me" -> build: one unit at a time -> review -> close GH issue

## What we are

- A lean spec-driven CLI for AI coding agents
- Convention over configuration, zero config files

## What we aren't

- A delta-merge system (edit specs directly)
- A backlog tracker (GH issues are the backlog)
