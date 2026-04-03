# list
## MODIFIED Requirements
### Requirement: Sort Flag

The `litespec list` command SHALL support a `--sort` flag accepting `recent` (default), `name`, or `deps`. When `--sort recent`, changes SHALL be ordered by last-modified time descending (most recent first). When `--sort name`, changes SHALL be ordered alphabetically ascending. When `--sort deps`, changes SHALL be ordered by topological sort of their dependency graph: changes with no dependencies first, then changes whose dependencies are already listed, with lexicographic tie-breaking at each level. Changes with no `dependsOn` field are treated as roots. The `--sort` flag SHALL only apply to changes — specs are always sorted alphabetically. The `--sort` flag with `--specs` only SHALL have no effect.

#### Scenario: Default sort is recent

- **WHEN** `litespec list` is run with no `--sort` flag
- **THEN** changes are ordered by most recently modified first

#### Scenario: Sort by name

- **WHEN** `litespec list --sort name` is run
- **THEN** changes are ordered alphabetically by name

#### Scenario: Sort by dependency order

- **WHEN** `litespec list --sort deps` is run and change B depends on change A
- **THEN** change A appears before change B in the output

#### Scenario: Sort deps with unrelated changes

- **WHEN** `litespec list --sort deps` is run and change B depends on A, while C has no dependencies
- **THEN** A and C appear before B, with A and C ordered alphabetically

#### Scenario: Sort deps with no dependencies

- **WHEN** `litespec list --sort deps` is run and no change has `dependsOn`
- **THEN** changes are sorted alphabetically as a fallback

#### Scenario: Sort deps with cycles

- **WHEN** `litespec list --sort deps` is run and a dependency cycle exists among active changes
- **THEN** changes involved in the cycle are sorted alphabetically and a warning is printed to stderr
