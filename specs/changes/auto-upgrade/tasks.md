## Phase 1: shared detection and utility functions
- [x] Implement `isGoInstall() bool` in `cmd/litespec/upgrade.go` — checks executable path against `$GOBIN` and `$GOPATH/bin` (go install detection)
- [x] Implement `getModulePath() (string, error)` in `cmd/litespec/upgrade.go` — reads module path from `runtime/debug.ReadBuildInfo()` (module path discovery)
- [x] Implement `parseSemver(tag string) (major, minor, patch int, err error)` in `cmd/litespec/upgrade.go` — parses semver strings, strips leading `v` (version comparison)
- [x] Implement `fetchLatestVersion() (string, error)` in `cmd/litespec/upgrade.go` — GET `api.github.com/repos/bermudi/litespec/releases/latest`, extracts tag name from JSON response (version comparison)

## Phase 2: explicit upgrade command
- [x] Implement `cmdUpgrade(args []string) error` in `cmd/litespec/upgrade.go` — includes `--help` flag handling, calls `isGoInstall()`, `getModulePath()`, `fetchLatestVersion()`, compares semver, runs `go install` with streamed output, prints post-upgrade hint, streams errors on `go install` failure, prints error on network failure (explicit upgrade command, version comparison, post-upgrade hint)
- [x] Implement `printUpgradeHelp()` in `cmd/litespec/helpers.go` — follows existing `print*Help()` pattern
- [x] Add `"upgrade"` case to command switch in `cmd/litespec/main.go` `run()` function, routing to `cmdUpgrade()`
- [x] Add `upgrade` entry to `printUsage()` in `cmd/litespec/main.go`

## Phase 3: background update gate
- [x] Implement `maybeBackgroundUpgrade()` in `cmd/litespec/main.go` — checks `isGoInstall()`, reads timestamp from `os.UserCacheDir()/litespec/last-update-check`, fires `cmd.Start("go", "install", module+"@latest")` with suppressed output if > 7 days, stamps timestamp (background self-update gate, cache directory)
- [x] Call `maybeBackgroundUpgrade()` in `cmd/litespec/main.go` `run()`, after `len(os.Args) < 2` check and before command switch

## Phase 4: tests
- [x] Test `isGoInstall()` — binary in `GOBIN`, binary in `GOPATH/bin`, binary elsewhere, `GOPATH` defaulting to `~/go/bin` (go install detection)
- [x] Test `parseSemver()` — valid versions, `v` prefix, invalid strings (version comparison)
- [x] Test `getModulePath()` — with build info present, with empty path (module path discovery)
- [x] Test `fetchLatestVersion()` — using `httptest.NewServer` to mock GitHub API, success case, malformed JSON, non-200 response (version comparison)
- [x] Test `cmdUpgrade()` — up-to-date, upgrade available, local newer than remote, not go install, module path unavailable, go install failure, network error (explicit upgrade command)
- [x] Test `maybeBackgroundUpgrade()` — timestamp within 7 days (skip), timestamp expired (fire), no timestamp file (fire), not go install (skip) (background self-update gate, cache directory)
- [x] Test `TestCLIUpgradeHelp` — verify help output contains expected strings
