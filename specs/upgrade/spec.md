# upgrade

## Requirements

### Requirement: explicit upgrade command

litespec SHALL provide an `upgrade` command that checks for the latest version and installs it via `go install` when the binary was installed via `go install`.

#### Scenario: upgrade available

- **WHEN** the user runs `litespec upgrade` and the binary is in `$GOBIN` or `$GOPATH/bin` and a newer version exists on GitHub Releases
- **THEN** litespec SHALL run `go install <module>@latest`, stream the output to the user, and print the new version

#### Scenario: already up to date

- **WHEN** the user runs `litespec upgrade` and the installed version matches the latest GitHub release
- **THEN** litespec SHALL print "Already up to date" and exit without running `go install`

#### Scenario: local version newer than remote

- **WHEN** the user runs `litespec upgrade` and the local version is greater than the latest GitHub release
- **THEN** litespec SHALL treat the installation as up to date and exit without running `go install`

#### Scenario: not installed via go install

- **WHEN** the user runs `litespec upgrade` and the binary is not in `$GOBIN` or `$GOPATH/bin`
- **THEN** litespec SHALL print an error explaining that auto-upgrade only supports `go install` installations and exit with a non-zero status

#### Scenario: go install failure

- **WHEN** the user runs `litespec upgrade` and `go install` exits with a non-zero code
- **THEN** litespec SHALL exit with the same non-zero code after streaming the `go install` output

#### Scenario: network error fetching latest version

- **WHEN** the user runs `litespec upgrade` and the HTTP request to GitHub Releases API fails
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

litespec SHALL fetch the latest release tag from `api.github.com/repos/bermudi/litespec/releases/latest`, parse both the remote tag and the local `version` const as semver, and compare them to determine if an upgrade is available.

#### Scenario: newer version on remote

- **WHEN** the remote semver is greater than the local version const
- **THEN** litespec SHALL proceed with `go install`

#### Scenario: equal versions

- **WHEN** the remote semver equals the local version const
- **THEN** litespec SHALL report that the installation is already up to date

#### Scenario: local version greater than remote

- **WHEN** the local semver is greater than the remote semver
- **THEN** litespec SHALL report that the installation is already up to date

### Requirement: post-upgrade hint

After a successful explicit upgrade, litespec SHALL print a hint telling the user to run `litespec update` in their projects to refresh generated artifacts.

#### Scenario: upgrade succeeds

- **WHEN** `go install` completes with exit code 0
- **THEN** litespec SHALL print the new version and a message suggesting `litespec update` to refresh project artifacts

### Requirement: background self-update gate

litespec SHALL perform a silent background `go install` at most once every 7 days when the binary is a `go install` installation. The gate SHALL NOT produce any output or block the main command.

#### Scenario: check interval elapsed

- **WHEN** the timestamp file in `os.UserCacheDir()/litespec/last-update-check` has an mtime older than 7 days and the binary is a `go install` installation
- **THEN** litespec SHALL start `go install <module>@latest` as a background process via `cmd.Start()`, suppress all output, update the timestamp file, and continue normal command execution without blocking

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
