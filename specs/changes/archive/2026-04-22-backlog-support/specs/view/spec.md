# view

## ADDED Requirements

### Requirement: Backlog Summary in Dashboard

When `specs/backlog.md` exists, the dashboard summary section SHALL display a backlog line showing item counts per recognized category. The recognized H2 section names are `## Deferred`, `## Open Questions`, and `## Future Versions`. Items are counted as top-level lines starting with `- ` (no leading whitespace) under each H2 section. Items under unrecognized H2 sections SHALL be counted as "other." The backlog line SHALL only include categories that have items (e.g., `● Backlog: 3 deferred, 2 open questions` when no future items exist). When `specs/backlog.md` does not exist, the backlog line SHALL be omitted entirely.

#### Scenario: Backlog with all categories

- **WHEN** `specs/backlog.md` exists with 2 items under `## Deferred`, 3 under `## Open Questions`, and 1 under `## Future Versions`
- **THEN** the summary shows `● Backlog: 2 deferred, 3 open questions, 1 future`

#### Scenario: Backlog with unknown sections

- **WHEN** `specs/backlog.md` exists with 1 item under `## Deferred` and 2 items under `## Nice-to-Have`
- **THEN** the summary shows `● Backlog: 1 deferred — 2 other`

#### Scenario: No backlog file

- **WHEN** `specs/backlog.md` does not exist
- **THEN** the backlog line is omitted from the summary

#### Scenario: Nested bullets not counted

- **WHEN** `specs/backlog.md` has a top-level `- ` item followed by indented `  - ` sub-items under `## Deferred`
- **THEN** only the top-level item is counted

#### Scenario: Empty backlog file

- **WHEN** `specs/backlog.md` exists but contains no items (empty or only headings)
- **THEN** the backlog line is omitted from the summary
