# CLI Reference

Complete command-line interface reference for litespec.

## Global Flags

| Flag | Description |
|------|-------------|
| `--version`, `-v` | Print version information |
| `--help`, `-h` | Print help message |
| `--json` | Output structured JSON (supported by: `status`, `validate`, `list`, `instructions`) |

## Commands

### `init`

Initialize a new litespec project in the current directory.

```
litespec init [--tools <ids>]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--tools <ids>` | Comma-separated tool IDs (e.g., `claude`) |

**Behavior:**
- Creates `specs/canon/` — canonical spec directory
- Creates `specs/changes/` — change proposals directory
- Creates `specs/changes/archive/` — archived changes directory
- Creates `.agents/skills/` — generated skill files
- Generates skills from canonical specs
- Optionally creates tool-specific symlinks

**Examples:**
```bash
# Basic initialization
litespec init

# Initialize with Claude Code symlinks
litespec init --tools claude
```

**Tips:**
- Run this in your project root
- Use `--tools claude` to set up Claude Code integration automatically
- Skills are generated from `specs/canon/` specs during initialization

---

### `new`

Create a new change directory under `specs/changes/`.

```
litespec new <name>
```

**Arguments:**
- `<name>` — Change name (e.g., `add-auth`, `fix-login-bug`)

**Behavior:**
- Creates `specs/changes/<name>/` directory
- Creates `specs/changes/<name>/specs/` directory for delta specs
- Creates `.litespec.yaml` with metadata (schema, timestamp)
- Fails if change already exists

**Examples:**
```bash
# Create a new change
litespec new add-user-auth

# Create a change for a bug fix
litespec new fix-rate-limit
```

**Tips:**
- Use kebab-case for change names
- Change name appears in archive path as `<date>-<name>`

---

### `list`

List active changes or canonical specs with metadata.

```
litespec list [--specs|--changes] [--sort recent|name] [--json]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--specs` | List specs instead of changes |
| `--changes` | List changes (default) |
| `--sort <field>` | Sort by `recent` (default) or `name` |
| `--json` | Output as JSON |

**Default output (changes):**
- Shows task progress (`✓ Complete`, `X/Y tasks`, `No tasks`)
- Shows relative last-modified time (`just now`, `Xm ago`, `Xh ago`, `Xd ago`, or date)
- Columns aligned to widest name

**Specs output:**
- Lists all canonical specs alphabetically
- Shows requirement count per spec

**Examples:**
```bash
# List changes (default)
litespec list

# List specs
litespec list --specs

# Sort changes by name
litespec list --sort name

# JSON output for scripts
litespec list --json
litespec list --specs --json
```

**JSON Output Format:**

Changes:
```json
{
  "changes": [
    {
      "name": "add-auth",
      "completedTasks": 5,
      "totalTasks": 8,
      "lastModified": "2026-04-02T10:30:00Z",
      "status": "in-progress"
    }
  ]
}
```

Specs:
```json
{
  "specs": [
    {
      "name": "validate",
      "requirementCount": 12
    }
  ]
}
```

**Tips:**
- `--sort` only applies to changes — specs are always alphabetical
- Relative time shows locale date for items older than 30 days
- Use `--json` for integration with other tools

---

### `status`

Show artifact states for a change or all changes.

```
litespec status [<name>] [--json]
```

**Arguments:**
- `<name>` — Optional change name (omit to show all changes)

**Flags:**

| Flag | Description |
|------|-------------|
| `--json` | Output as JSON |

**Artifact States:**
- `BLOCKED` — Dependencies not satisfied
- `READY` — All dependencies exist, artifact does not
- `DONE` — Artifact file exists

**Default output (all changes):**
```
add-auth
  proposal:    DONE
  specs:       READY
  design:      READY
  tasks:       BLOCKED
```

**Named change output:**
```
Change: add-auth
Created: 2026-04-02 10:30:00

  proposal    DONE       Why and what — the motivation, scope, and approach
  specs       READY      Delta specifications
  design      READY      How — the technical approach, architecture decisions
  tasks       BLOCKED    What to do — the phased implementation checklist
```

**Examples:**
```bash
# Show all changes
litespec status

# Show specific change
litespec status add-auth

# JSON output
litespec status --json
litespec status add-auth --json
```

**JSON Output Format:**

All changes:
```json
[
  {
    "changeName": "add-auth",
    "schemaName": "spec-driven",
    "isComplete": false,
    "artifacts": [
      {
        "id": "proposal",
        "outputPath": "proposal.md",
        "status": "done"
      }
    ]
  }
]
```

**Tips:**
- Shows creation timestamp for named changes
- Exits with code 1 if change not found
- Use to track which artifacts are ready for creation

---

### `validate`

Validate changes and specs for structure, delta syntax, and dangling deltas.

```
litespec validate [<name>] [--all|--changes|--specs] [--type change|spec] [--strict] [--json]
```

**Arguments:**
- `<name>` — Optional change or spec name

**Flags:**

| Flag | Description |
|------|-------------|
| `--all` | Validate all changes and specs |
| `--changes` | Validate all changes only |
| `--specs` | Validate all specs only |
| `--type <T>` | Disambiguate name: `change` or `spec` |
| `--strict` | Treat warnings as errors |
| `--json` | Output as JSON |

**Validation Checks:**
- Artifact structure and existence
- Delta spec syntax (ADDED/MODIFIED/REMOVED/RENAMED markers)
- Dangling deltas (references to non-existent requirements)
- Spec format requirements (SHALL/MUST in body, scenarios for ADDED/MODIFIED)

**Default behavior (no arguments):**
- Validates all changes and all specs (equivalent to `--all`)

**Examples:**
```bash
# Validate all (default)
litespec validate

# Validate specific change
litespec validate my-change

# Validate specific spec
litespec validate auth

# Disambiguate ambiguous name
litespec validate shared --type change

# Validate all changes only
litespec validate --changes

# Validate with strict mode
litespec validate --all --strict

# JSON output
litespec validate --json
```

**JSON Output Format:**
```json
{
  "valid": false,
  "errors": [
    {
      "severity": "error",
      "message": "Requirement 'non-existent' not found in spec",
      "file": "specs/changes/my-change/specs/auth/spec.md"
    }
  ],
  "warnings": [
    {
      "severity": "warning",
      "message": "No scenarios defined",
      "file": "specs/changes/my-change/specs/auth/spec.md"
    }
  ],
  "summary": {
    "total": 2,
    "invalid": 1
  }
}
```

**Exit Codes:**
- `0` — Validation passed
- `1` — Validation failed or error occurred

**Tips:**
- Dangling delta detection catches broken refs during validation, not just at archive time
- `--strict` is useful in CI/CD pipelines
- Use `--type` when a name exists as both a change and a spec

---

### `instructions`

Get artifact-specific instructions for writing proposals, specs, designs, or tasks.

```
litespec instructions <artifact> [--json]
```

**Arguments:**
- `<artifact>` — One of: `proposal`, `specs`, `design`, `tasks`

**Flags:**

| Flag | Description |
|------|-------------|
| `--json` | Output as JSON |

**Behavior:**
- Returns static artifact creation guidance
- No change context required
- Used by AI skills to understand how to create artifacts

**Examples:**
```bash
# Get proposal instructions
litespec instructions proposal

# Get design instructions as JSON
litespec instructions design --json

# Get tasks instructions
litespec instructions tasks
```

**JSON Output Format:**
```json
{
  "artifactId": "design",
  "description": "How — technical approach, architecture decisions, data flow, file changes",
  "instruction": "... detailed instructions ...",
  "template": "... propose skill template ...",
  "outputPath": "design.md"
}
```

**Exit Codes:**
- `0` — Success
- `1` — Unknown artifact or error

**Tips:**
- Each artifact has unique guidance (proposal focuses on motivation/scope, specs on delta format, design on architecture, tasks on phases)
- Instructions are generated from canonical specs
- Use this to understand expected artifact structure

---

### `archive`

Apply deltas and archive a completed change.

```
litespec archive <name> [--allow-incomplete]
```

**Arguments:**
- `<name>` — Change name to archive

**Flags:**

| Flag | Description |
|------|-------------|
| `--allow-incomplete` | Archive even with incomplete tasks |

**Archive Process:**
1. **Validate** — Run full validation on the change
2. **Check tasks** — Verify all tasks complete (unless `--allow-incomplete`)
3. **Merge deltas** — Apply RENAMED→REMOVED→MODIFIED→ADDED to `specs/canon/`
4. **Strip specs/** — Remove change's `specs/` directory
5. **Move** — Relocate to `specs/changes/archive/<YYYY-MM-DD>-<name>/`

**Delta Merge Order:**
1. `RENAMED` — Establish correct headers
2. `REMOVED` — Eliminate requirements
3. `MODIFIED` — Update remaining requirements
4. `ADDED` — Append new requirements

**Examples:**
```bash
# Archive completed change
litespec archive add-auth

# Archive incomplete change (bypass task check)
litespec archive add-auth --allow-incomplete
```

**Exit Codes:**
- `0` — Success
- `1` — Validation failed, tasks incomplete, or error

**Tips:**
- Archived directory contains only: `.litespec.yaml`, `proposal.md`, `design.md`, `tasks.md`
- `specs/` subtree is stripped — content merged into canonical specs
- Canon creates new directories if capability doesn't exist
- Date prefix in archive path prevents conflicts

---

### `update`

Regenerate skills and adapter commands from current specs.

```
litespec update [--tools <ids>]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--tools <ids>` | Comma-separated tool IDs (e.g., `claude`) |

**Behavior:**
- Regenerates all skills in `.agents/skills/`
- Updates tool-specific symlinks if `--tools` provided
- Does not modify `specs/` directory
- Fails if not a litespec project

**Examples:**
```bash
# Regenerate skills only
litespec update

# Update skills and recreate Claude symlinks
litespec update --tools claude
```

**Exit Codes:**
- `0` — Success
- `1` — Not a litespec project or error

**Tips:**
- Use after modifying canonical specs to regenerate skills
- Faster than `init` for skill refresh
- Does not create project structure (assumes it exists)

---

### `completion`

Generate shell completion script.

```
litespec completion <shell>
```

**Arguments:**
- `<shell>` — One of: `bash`, `zsh`, `fish`

**Supported Shells:**
- `bash` — Uses `_litespec()` completion function
- `zsh` — Uses `#compdef litespec` with `_arguments`
- `fish` — Uses `complete -c litespec` commands

**Examples:**
```bash
# Generate bash completion
litespec completion bash

# Generate zsh completion
litespec completion zsh

# Generate fish completion
litespec completion fish

# Install bash completion
litespec completion bash | sudo tee /etc/bash_completion.d/litespec

# Install zsh completion (add to ~/.zshrc)
litespec completion zsh > ~/.zsh/completion/_litespec
```

**Exit Codes:**
- `0` — Success
- `1` — Invalid shell or error

**Tips:**
- Source the script to enable completion
- Completions are dynamic — query project state for change/spec names
- Hidden `__complete` command provides completion data

---

## Hidden Commands

### `__complete`

Internal backend for dynamic shell completions.

```
litespec __complete <words...>
```

**Behavior:**
- Receives command-line words as positional arguments
- Prints completion candidates: `candidate\tdescription` (one per line)
- Errors during resolution produce no output (silent fallback)

**Dynamic Completions:**
- Change names from filesystem
- Spec names from filesystem
- Tool IDs from adapter config
- Artifact IDs from artifact registry
- Flags and static values hardcoded

**Examples:**
```bash
# Complete command names
litespec __complete litespec

# Complete change names for status
litespec __complete litespec status

# Complete tool IDs for --tools
litespec __complete litespec init --tools
```

**Tips:**
- Not intended for direct use — used by shell scripts
- Tab character separates candidate from description
- Bash ignores description, zsh/fish use it

---

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | Error, validation failed, or invalid arguments |

All commands follow consistent exit code behavior for scripting and automation.

---

## Tool Integration

### Claude Code

Generate symlinks for Claude Code integration:

```bash
litespec init --tools claude
# or
litespec update --tools claude
```

This creates symlinks in `.claude/skills/` pointing to `.agents/skills/`.

**Supported Tools:**

| Tool ID | Name | Skills Directory |
|---------|------|-----------------|
| `claude` | Claude Code | `.claude/skills/` |

---

## Project Structure

```
project/
├── specs/
│   ├── canon/                    # Source of truth
│   │   └── <capability>/
│   │       └── spec.md
│   └── changes/                  # Active changes
│       ├── <name>/
│       │   ├── .litespec.yaml
│       │   ├── proposal.md
│       │   ├── design.md
│       │   ├── tasks.md
│       │   └── specs/
│       │       └── <capability>/
│       │           └── spec.md
│       └── archive/              # Completed changes
│           └── <date>-<name>/
└── .agents/skills/               # Generated skills
    ├── litespec-explore/
    ├── litespec-grill/
    ├── litespec-propose/
    ├── litespec-continue/
    ├── litespec-apply/
    ├── litespec-verify/
    ├── litespec-adopt/
    └── litespec-archive/
```

---

## Delta Spec Operations

| Marker | Behavior |
|--------|----------|
| `## ADDED Requirements` | Append to end of main spec |
| `## MODIFIED Requirements` | Replace matching requirement by header |
| `## REMOVED Requirements` | Delete from main spec |
| `## RENAMED Requirements` | Change section header, preserve content |

**Merge Order:** RENAMED → REMOVED → MODIFIED → ADDED

---

## See Also

- [Concepts](concepts.md) — Philosophy and core ideas behind litespec
- [Getting Started](getting-started.md) — Installation and first steps
- [Tutorial](tutorial.md) — Worked walkthrough of a complete change
