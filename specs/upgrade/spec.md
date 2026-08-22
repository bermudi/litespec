# upgrade

## Requirements

### Requirement: explicit upgrade command

litespec SHALL provide an `upgrade` command that checks for the latest version and installs it via `go install` when the binary was installed via `go install`. `upgrade` SHALL track the latest version per channel: a stable local version SHALL only consider stable tags (tags without a `-` prerelease suffix, e.g. `v0.20.2`), while a prerelease local version (e.g. `v2.0.0-beta.2`) SHALL consider both stable and prerelease tags and upgrade to the latest of either.

#### Scenario: stable upgrade available

- **WHEN** the user runs `litespec upgrade` from a stable install and a newer stable version exists on GitHub
- **THEN** litespec SHALL run `go install <module>@<latest stable tag>`, stream the output to the user, and print the new version

#### Scenario: beta upgrade to newer beta

- **WHEN** the user runs `litespec upgrade` from `v2.0.0-beta.2` and `v2.0.0-beta.4` is the latest tag
- **THEN** litespec SHALL upgrade to `v2.0.0-beta.4`

#### Scenario: beta upgrade to stable

- **WHEN** the user runs `litespec upgrade` from `v2.0.0-beta.2` and `v2.0.0` is the latest stable
- **THEN** litespec SHALL upgrade to `v2.0.0`, because a stable release is newer than its prerelease

#### Scenario: already up to date

- **WHEN** the user runs `litespec upgrade` and the installed version matches the latest version for its channel
- **THEN** litespec SHALL print "Already up to date" and exit without running `go install`

#### Scenario: local version newer than remote

- **WHEN** the local version is greater than or equal to the latest version for its channel (e.g. local `v2.0.0-beta.2` is newer by major than latest stable `v0.20.2` on the stable channel)
- **THEN** litespec SHALL treat the installation as up to date and exit without running `go install`

#### Scenario: stable does not follow betas

- **WHEN** a stable local version such as `v0.20.2` is installed and `v2.0.0-beta.4` is the latest prerelease
- **THEN** `upgrade` SHALL ignore the beta and report already up to date

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

litespec SHALL fetch the list of tags from `api.github.com/repos/bermudi/litespec/tags`, parse tags as semver including prerelease identifiers, and compare per semver precedence (a stable version is newer than its prerelease; a higher major wins across lines) to determine if an upgrade is available. For a stable local version, `upgrade` SHALL consider only stable tags; for a prerelease local version, it SHALL consider both stable and prerelease tags.

#### Scenario: newer stable version on remote (stable channel)

- **WHEN** the local version is stable and the latest stable semver is greater than it
- **THEN** litespec SHALL proceed with `go install`

#### Scenario: equal versions

- **WHEN** the latest semver for the channel equals the local version const
- **THEN** litespec SHALL report that the installation is already up to date

#### Scenario: local version greater than remote

- **WHEN** the local semver is greater than the latest semver for its channel
- **THEN** litespec SHALL report that the installation is already up to date

#### Scenario: stable channel ignores betas

- **WHEN** the local version is stable `v0.20.2` and the tag list contains `v2.0.0-beta.4` alongside `v0.20.2`
- **THEN** `upgrade` SHALL consider only `v0.20.2` as the latest and ignore the beta

#### Scenario: beta channel sees betas

- **WHEN** the local version is `v2.0.0-beta.2` and the tag list contains `v2.0.0-beta.4` alongside `v0.20.2`
- **THEN** `upgrade` SHALL consider `v2.0.0-beta.4` as the latest and upgrade to it

#### Scenario: beta versus stable with same base

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

- **WHEN** the timestamp file in `os.UserCacheDir()/litespec/last-update-check` has an mtime older than 7 days and the binary is a `go install` installation and the latest version for its channel is newer than the local version
- **THEN** litespec SHALL start `go install <module>@<latest tag for its channel>` as a background process via `cmd.Start()`, suppress all output, update the timestamp file, and continue normal command execution without blocking

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
