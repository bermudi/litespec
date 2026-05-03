# Design: adopt-list-backlog

## Architecture

The backlog listing feature follows the established pattern for `list` sub-flags (`--specs`, `--decisions`):

1. **Parser layer** (`internal/backlog.go`) — pure data extraction, no I/O beyond reading the file
2. **CLI layer** (`cmd/litespec/list.go`) — flag parsing, mutual exclusivity, output formatting
3. **JSON types** (`internal/json.go`) — structured output shapes
4. **Command spec** (`internal/commandspec.go`) — shell completion registration

## Component Relationships

```
cmdList()
  ├── ParseBacklogItems()  → []BacklogItem
  │     ├── normalizeBacklogSection()  → (key, ok)
  │     └── extractBacklogTitle()      → string
  └── BacklogItemJSON       → JSON output
```

`normalizeBacklogSection` is the shared authority for section name → key mapping. Both `ParseBacklog` (counts) and `ParseBacklogItems` (items) delegate to it, ensuring consistency.

## Data Flow

1. `cmdList` parses `--backlog` flag, validates mutual exclusivity
2. Calls `ParseBacklogItems(BacklogPath(root))`
3. For text output: iterates items, tracks current section to emit section headers, prints titles with `▪` prefix
4. For JSON output: maps each `BacklogItem` to `BacklogItemJSON`, marshals via `MarshalJSON`

## Section Key Mapping

| Markdown Header       | Internal Key     | Display Label     |
|-----------------------|------------------|-------------------|
| `## Deferred`         | `deferred`       | Deferred          |
| `## Open Questions`   | `open-questions` | Open Questions    |
| `## Future Versions`  | `future`         | Future            |
| `## Future`           | `future`         | Future            |
| `## Other`            | `other`          | Other             |
| anything else         | *(rejected)*     | —                 |

## Title Extraction

Bullet lines (`- ` or `* `) are stripped of their marker, then trimmed of whitespace. Then:
- If the trimmed line starts with `**`, extract text between opening and closing `**`
- Otherwise, return the full trimmed line as-is

## Patterns

- **Graceful absence**: Missing/empty backlog returns nil — no error, no output noise
- **Convention over configuration**: Section names are hardcoded, not configurable
- **Text vs JSON symmetry**: Same data, two presentations, gated by `--json`
