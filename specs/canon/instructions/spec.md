# instructions

### Requirement: Instructions Without Change Context

The `litespec instructions <artifact>` command MUST NOT require a `--change` flag. It SHALL return static artifact creation guidance for the given artifact ID. The `--json` flag is supported for structured output. Valid artifact IDs are `proposal`, `specs`, `design`, and `tasks`. This is a breaking change from the previous `--json` output format which included `changeName`, `changeDir`, `schemaName`, `dependencies`, and `unlocks` fields — these are removed.

#### Scenario: Instructions for an artifact

- **WHEN** `litespec instructions proposal` is run
- **THEN** artifact creation guidance for `proposal` is printed

#### Scenario: Instructions with JSON

- **WHEN** `litespec instructions design --json` is run
- **THEN** structured JSON with `artifactId`, `description`, `instruction`, `template`, and `outputPath` fields is printed

#### Scenario: Unknown artifact

- **WHEN** `litespec instructions unknown-artifact` is run
- **THEN** an error is printed listing valid artifact IDs
