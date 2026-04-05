## Architecture

This change extends two existing output surfaces (`list` and `view`) to surface two timestamp signals that already exist in the system:

- **Born** — `ChangeMeta.Created` from `.litespec.yaml`, set once at `CreateChange` time
- **Last touched** — filesystem mtime, already computed by `GetLastModified()` and stored in `ChangeInfo.LastModified`

The data flow is: `ListChanges()` already calls `ReadChangeMeta()` per change for dependency info. We extend it to also populate `ChangeInfo.Created` from the same read. Then `list` and `view` formatters consume both fields.

```
.litespec.yaml          filesystem
  created: t1             mtime: t2
       │                     │
       ▼                     ▼
  ReadChangeMeta()     GetLastModified()
       │                     │
       └─────► ChangeInfo ◄──┘
               Created  LastModified
                   │         │
              ┌────┴─────────┴────┐
              ▼                   ▼
          cmdList()          cmdView()
```

## Decisions

### ChangeInfo gets Created, not ChangeMeta passthrough

`ChangeInfo` currently carries `LastModified` but not `Created`. Adding `Created` to `ChangeInfo` is cleaner than having `cmdList` and `cmdView` each independently call `ReadChangeMeta` — they already receive `ChangeInfo` from `ListChanges`.

### ListChanges reads meta once, populates both fields

Currently `ListChanges` reads `.litespec.yaml` only when `--sort deps` triggers it via `LoadDepMap`. But `cmdList` already calls `ReadChangeMeta` per change in the JSON path for `DependsOn`. Moving the `ReadChangeMeta` call into `ListChanges` itself means the data is always available, eliminating per-command reads.

### View timestamp format: parenthetical after name

Format: `◉ add-auth  (born 2026-04-01, touched 3d ago)`. This keeps the existing visual hierarchy (bullet, name, progress bar) while appending temporal context inline. Avoids adding a separate column layout to the dashboard.

### List text format: four-column layout

Format: `name  status  born  last-touched`. Born as `YYYY-MM-DD`, last-touched as relative time. Matches the existing three-column layout with one addition.

## File Changes

### `internal/change.go`

- **`ChangeInfo` struct** — add `Created time.Time` field
- **`ListChanges()`** — call `ReadChangeMeta` for each change to populate `Created` on `ChangeInfo`

### `internal/json.go`

- **`ChangeListItemJSON`** — add `Born string` field (`json:"born"`)

### `cmd/litespec/list.go`

- **Text output** — add born column (YYYY-MM-DD) between status and last-touched columns
- **JSON output** — populate `Born` field from `c.Created.Format(time.RFC3339)`

### `cmd/litespec/view.go`

- **Active changes section** — append `(born YYYY-MM-DD, touched Xm ago)` after progress bar
- **Draft changes section** — append `(born YYYY-MM-DD, touched Xm ago)` after name
- **Completed changes section** — append `(born YYYY-MM-DD, touched Xm ago)` after name
- **Dependency graph nodes** — append `(born YYYY-MM-DD, touched Xm ago)` after name
