# Product

## Mental Models

- Agent-native workflows over file-heavy ceremony
- GH issue is the queue — proposal + design + queue live in the issue body
- Lean skills with progressive disclosure (fuzzy/clear/grilling/codebase-design/domain-modeling)

## Flows

1. Small fix — zero ceremony: You say "fix typo" -> agent reads product + relevant spec + decisions/glossary -> edits code -> updates the one specs/<feature>/spec.md if contract change -> done. No new, no issue required.

2. New feature — plan[fuzzy] (read code, grill, no files) -> plan[clear] (write GH issue with boundary/failure-policy units, scenario mappings, and Verify; draft load-bearing spec) -> build one unit -> review with cumulative advisory coverage -> rebuild at most twice per unchanged unit contract, otherwise re-plan its shape -> close only after PASS

## What we are

- A lean spec-driven CLI for AI coding agents
- Convention over configuration, zero config files

## What we aren't

- A delta-merge system (edit specs directly)
- A backlog tracker (GH issues are the backlog)
