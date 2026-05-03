# backlog-item-parser

## ADDED Requirements

### Requirement: section normalization

The parser SHALL recognize exactly four `## ` sections in backlog.md: `Deferred`, `Open Questions`, `Future Versions` (or `Future`), and `Other`. Section headers are matched case-insensitively and with trimmed whitespace. Unrecognized `## ` headers are excluded from item extraction and surfaced as warnings by validation.

#### Scenario: Standard sections recognized

- **WHEN** backlog.md contains `## Deferred`, `## Open Questions`, `## Future Versions`, and `## Other` sections
- **THEN** all items under these sections are counted and extracted with the correct section key

#### Scenario: Future shorthand

- **WHEN** backlog.md contains `## Future` instead of `## Future Versions`
- **THEN** items under that section are recognized as "future" section items

#### Scenario: Unrecognized section ignored for items

- **WHEN** backlog.md contains `## Nice-to-Have` with items
- **THEN** `ParseBacklog` records the section name as unrecognized (but does not count those items under any category), while `ParseBacklogItems` skips them entirely

### Requirement: item extraction with title parsing

`ParseBacklogItems` SHALL extract each top-level list item (`- ` or `* ` with no leading whitespace) as a `BacklogItem` containing the section key and the item title. Titles are extracted from `**bold**` markdown if present; otherwise the full line text (after the bullet marker) is used as the title.

#### Scenario: Bold title extraction

- **WHEN** a backlog item line is `- **Universal JSON** — description text`
- **THEN** the extracted title is `Universal JSON`

#### Scenario: Plain text fallback

- **WHEN** a backlog item line is `- Plain text item`
- **THEN** the extracted title is `Plain text item`

#### Scenario: Unclosed bold marker

- **WHEN** a backlog item line is `- **Title with no closing`
- **THEN** the extracted title is `Title with no closing` (the leading `**` is stripped, remainder returned)

#### Scenario: Nested bullets ignored

- **WHEN** a backlog item is followed by indented sub-bullets (leading whitespace)
- **THEN** only the top-level item is extracted; nested bullets are skipped

#### Scenario: Asterisk bullets supported

- **WHEN** backlog items use `* ` instead of `- ` as the bullet marker
- **THEN** items are extracted identically to dash bullets

### Requirement: graceful absence handling

Both `ParseBacklog` and `ParseBacklogItems` SHALL return nil (not an error) when the backlog file does not exist. An empty backlog file (sections present but no items) also returns nil.

#### Scenario: Missing file

- **WHEN** `specs/backlog.md` does not exist
- **THEN** `ParseBacklog` returns `nil, nil` and `ParseBacklogItems` returns `nil, nil`

#### Scenario: Empty recognized sections

- **WHEN** backlog.md contains recognized section headers but no list items
- **THEN** `ParseBacklog` returns `nil, nil`

#### Scenario: Unrecognized section with no items

- **WHEN** backlog.md contains only unrecognized section headers (e.g., `## Nice-to-Have`) with no items
- **THEN** `ParseBacklog` returns a non-nil summary with the header name in `Unrecognized`

### Requirement: CRLF line ending support

The parser SHALL handle Windows-style `\r\n` line endings correctly, stripping `\r` from each line before processing.

#### Scenario: CRLF backlog file

- **WHEN** backlog.md uses `\r\n` line endings
- **THEN** all sections and items are parsed identically to `\n`-only files
