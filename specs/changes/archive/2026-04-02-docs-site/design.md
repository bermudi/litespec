## Architecture

This change is purely additive — no existing litespec code is modified. It adds a documentation layer on top of the repo:

```
repo root/
├── mkdocs.yml              ← MkDocs config (new)
├── pyproject.toml           ← Python deps for mkdocs (new)
├── docs/                    ← Doc content (new)
│   ├── index.md
│   ├── concepts.md
│   ├── getting-started.md
│   ├── tutorial.md
│   ├── workflow.md
│   ├── cli-reference.md
│   ├── delta-specs.md
│   └── project-structure.md
├── .github/
│   └── workflows/
│       └── docs.yml         ← Deploy workflow (new)
└── README.md                ← Trimmed (modified)
```

## Decisions

### MkDocs Material over alternatives
Chosen because it's the standard for developer docs, looks professional with zero config, and has great search/navigation built in. mdbook (Rust) was considered but Material's polish is hard to beat. Docusaurus is overkill for a Go CLI.

### Manual CLI reference (for now)
The `cli-reference.md` page is maintained manually rather than auto-generated from Go code. This keeps the change small. Auto-generation from cobra/doc can be added as a future change.

### Docs as source of truth
The `docs/` pages are canonical. README.md is a lightweight pointer. This avoids drift between two sets of documentation.

### `pyproject.toml` with uv
Consistent with the project's tooling conventions. No global installs, no pip.

### Tutorial with worked example
The `tutorial.md` page shows a complete worked example from init to archive with real AI output at each stage. This is critical for onboarding — users need to see what it feels like to use litespec, not just read command tables. OpenSpec does this well and it's worth emulating.

### Concepts page for "why"
The `concepts.md` page explains what specs are, why spec-driven development works, and when it applies. This is the convincing document for skeptical readers. OpenSpec's `concepts.md` is their best page for this reason.

## File Changes

| File | Action | Why |
|------|--------|-----|
| `mkdocs.yml` | Create | MkDocs configuration — site name, theme, nav structure |
| `pyproject.toml` | Create | Declares mkdocs + mkdocs-material as deps |
| `docs/index.md` | Create | Landing page — pitch, what + why |
| `docs/concepts.md` | Create | Philosophy: what specs are, why they work, progressive rigor |
| `docs/getting-started.md` | Create | Installation, init |
| `docs/tutorial.md` | Create | Worked "first change" walkthrough from init to archive with real output |
| `docs/workflow.md` | Create | The explore→grill→... flow with named patterns (Quick Feature, Exploratory, Adopt) |
| `docs/cli-reference.md` | Create | Every command, every flag |
| `docs/delta-specs.md` | Create | Delta format explained with before/after merge example |
| `docs/project-structure.md` | Create | What goes where and why |
| `.github/workflows/docs.yml` | Create | GitHub Actions: build + deploy to Pages on push to main |
| `README.md` | Modify | Trim to brief summary + link to docs site |
