# Decisions Are Not Changes

## Status

accepted

## Context

Litespec needs a home for locked architectural reasoning that spans multiple features. Product and feature specs describe *what*, GitHub issues hold temporary proposals and designs, and agent skills hold operating instructions; none is a durable ruling.

## Decision

Decisions SHALL be a separate, optional artifact type. They SHALL live in `specs/decisions/` as numbered markdown files and persist independently of GitHub issues and feature specs.

## Consequences

Clean separation between temporary planning and standing rulings. GitHub issues and feature specs can cite decisions in prose without structural coupling. Decisions can be superseded by newer decisions without involving the issue workflow. The trade-off is one more artifact type to learn, but the concept is small and opt-in.
