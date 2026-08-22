# upgrade

## Requirements

### Requirement: explicit upgrade command

litespec SHALL provide an `upgrade` command that checks for the latest stable version and installs it via `go install` when the binary was installed via `go install`. `upgrade` SHALL only consider stable tags (tags without a `-` prerelease suffix, e.g. `v0.20.2`); prerelease tags such as `v2.0.0-beta.2` SHALL be ignored.

#### Scenario: upgrade available

- **WHEN** the user runs `litespec upgrade` and the binary is in `$GOBIN` or `$GOPATH/bin` and a newer stable version exists on GitHub
- **THEN** litespec SHALL run `go install <module>@<latest stable tag>`, stream the output to the user, and print the new version

#### Scenario: already up to date

- **WHEN** the user runs `litespec upgrade` and the installed version matches the latest stable tag
- **THEN** litespec SHALL print "Already up to date" and exit without running `go install`

#### Scenario: local version newer than remote

- **WHEN** the user runs `litespec upgrade` and the local version is greater than or equal to the latest stable tag, or the local prerelease (e.g. `v2.0.0-beta.2`) is newer by major than the latest stable (`v0.20.2`)
- **THEN** litespec SHALL treat the installation as up to date and exit without running `go install`

#### Scenario: not installed via go install

- **WHEN** the user runs `litespec upgrade` and the binary is not in `$GOBIN` or `$GOPATH/bin`
- **THEN** litespec SHALL print an error explaining that auto-upgrade only supports `go install` installations and exit with a non-zero status

#### Scenario: go install failure

- **WHEN** the user runs `litespec upgrade` and `go install` exits with a non-zero code
- **THEN** litespec SHALL exit with the same non-zero code after streaming the `go install` output

#### Scenario: network error fetching latest version

- **WHEN** the user runs `litespec upgrade` and the HTTP request to the GitHub Tags API fails
- **THEN** litespec SHALL print an error describing the failure and exit with a non-zero status

### Requirement: go install detection

litespec SHALL detect whether the running binary was installed via `go install` by checking if the executable path is within `$GOBIN` or `$GOPATH/bin`.

#### Scenario: binary in GOBIN

- **WHEN** the running binary's path starts with the value of the `GOBIN` environment variable
- **THEN** litespec SHALL treat it as a `go install` installation

#### Scenario: binary in GOPATH/bin

- **WHEN** `GOBIN` is unset and the running binary's path starts with `$GOPATH/bin` (defaulting to `~/go/bin` if `GOPATH` is unset)
- **THEN** litespec SHALL treat it as a `go install` installation

#### Scenario: binary elsewhere

- **WHEN** the running binary's path is not within `$GOBIN` or `$GOPATH/bin`
- **THEN** litespec SHALL treat it as a non-go-install installation

### Requirement: module path discovery

litespec SHALL derive its module path from `runtime/debug.ReadBuildInfo()` to construct the `go install` command, ensuring the path survives module renames without code changes.

#### Scenario: module path resolved

- **WHEN** litespec reads build info and finds a non-empty module path
- **THEN** litespec SHALL use that path as the `go install` target

#### Scenario: module path unavailable

- **WHEN** `ReadBuildInfo()` returns an empty module path
- **THEN** litespec SHALL print an error that the module path could not be determined and exit with a non-zero status

### Requirement: version comparison

litespec SHALL fetch the list of tags from `api.github.com/repos/bermudi/litespec/tags`, exclude prerelease tags (those containing `-` after the `v` prefix, e.g. `v2.0.0-beta.2`), parse the remaining stable tags and the local `version` const as semver including prerelease identifiers, and compare per semver precedence (a stable version is newer than its prerelease; a higher major wins across lines) to determine if an upgrade is available.

#### Scenario: newer stable version on remote

- **WHEN** the latest stable semver is greater than the local version const
- **THEN** litespec SHALL proceed with `go install`

#### Scenario: equal versions

- **WHEN** the latest stable semver equals the local version const
- **THEN** litespec SHALL report that the installation is already up to date

#### Scenario: local version greater than remote

- **WHEN** the local semver is greater than the latest stable semver
- **THEN** litespec SHALL report that the installation is already up to date

#### Scenario: prerelease tags are ignored

- **WHEN** the tag list contains `v2.0.0-beta.4` alongside `v0.20.2`
- **THEN** `upgrade` SHALL consider only `v0.20.2` as the latest stable and ignore the beta

#### Scenario: prerelease local versus stable remote

- **WHEN** the local version is `v2.0.0-beta.2` and the latest stable is `v0.20.2` (lower major)
- **THEN** the local version is considered newer, so `upgrade` reports already up to date

#### Scenario: prerelease local versus its stable

- **WHEN** the local version is `v2.0.0-beta.2` and the latest stable is `v2.0.0`
- **THEN** the stable is newer, so `upgrade` proceeds with `go install`

### Requirement: post-upgrade hint

After a successful explicit upgrade, litespec SHALL print a hint telling the user to run `litespec update` in their projects to refresh generated artifacts.

#### Scenario: upgrade succeeds

- **WHEN** `go install` completes with exit code 0
- **THEN** litespec SHALL print the new version and a message suggesting `litespec update` to refresh project artifacts

### Requirement: background self-update gate

litespec SHALL perform a silent background `go install` at most once every 7 days when the binary is a `go install` installation. The gate SHALL NOT produce any output or block the main command.

#### Scenario: check interval elapsed

- **WHEN** the timestamp file in `os.UserCacheDir()/litespec/last-update-check` has an mtime older than 7 days and the binary is a `go install` installation and the latest stable tag is newer than the local version
- **THEN** litespec SHALL start `go install <module>@<latest stable tag>` as a background process via `cmd.Start()`, suppress all output, update the timestamp file, and continue normal command execution without blocking

#### Scenario: check interval not elapsed

- **WHEN** the timestamp file exists and its mtime is within 7 days
- **THEN** litespec SHALL skip the background self-update entirely

#### Scenario: not a go install installation

- **WHEN** the binary is not in `$GOBIN` or `$GOPATH/bin`
- **THEN** litespec SHALL skip the background self-update entirely regardless of timestamp

### Requirement: cache directory

litespec SHALL store the update check timestamp in the platform-standard cache directory resolved by `os.UserCacheDir()`, under a `litespec/` subdirectory.

#### Scenario: cache directory does not exist

- **WHEN** `os.UserCacheDir()/litespec/` does not exist
- **THEN** litespec SHALL create it before writing the timestamp file

#### Scenario: timestamp file does not exist

- **WHEN** the timestamp file does not exist
- **THEN** litespec SHALL treat it as if the interval has elapsed and trigger the background self-update
