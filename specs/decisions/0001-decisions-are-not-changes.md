# Decisions Are Not Changes

## Status

accepted

## Context

Litespec needed a home for locked architectural reasoning that spans multiple changes. Three existing artifact types fell short: canon specs describe *what*, not *why*; design.md is scoped to a single change and gets archived; research skills hold external knowledge, not internal rulings.

## Decision

Decisions SHALL be a separate artifact type from changes. They SHALL have no dependsOn, no tasks, no deltas, and no archive lifecycle. They SHALL live in `specs/decisions/` as numbered markdown files and persist independently of the change workflow.

## Consequences

Clean separation between planning (changes) and standing rulings (decisions). Changes can cite decisions in prose without structural coupling. Decisions can be superseded by newer decisions without involving the change workflow at all. The trade-off is one more artifact type to learn, but the concept is small and opt-in.
