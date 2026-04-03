## Architecture

Auto-upgrade introduces two new code paths that share a common detection layer:

```
run()
├── if len(os.Args) >= 2
│   └── maybeBackgroundUpgrade()    ← new, after arg check, before command dispatch
│       ├── isGoInstall()
│       └── cmd.Start("go", "install", modulePath+"@latest")
└── switch os.Args[1]:
    └── "upgrade" → cmdUpgrade()  ← new command
        ├── isGoInstall()
        ├── getModulePath()
        ├── fetchLatestVersion()
        ├── compareSemver()
        ├── exec.Command("go", "install", modulePath+"@latest")
        └── print hint
```

Both paths depend on `isGoInstall()` and module path discovery via `runtime/debug.ReadBuildInfo()`. The background gate runs after the argument check but before command dispatch — so it fires for every real invocation but not for bare `litespec` with no arguments. The explicit `upgrade` command is a normal CLI subcommand.

No changes to existing commands. `update` remains artifact-only.

## Decisions

### go install over self-replace
Self-replacing the binary (download, checksum, rename) duplicates what the Go toolchain already handles — module verification, checksum DB, build caching. `go install` is the single supported path. Users who installed via other methods are unaffected.

**Trade-off:** requires the Go toolchain to be present. Acceptable since `go install` is the only supported install method for this feature.

### Timestamp always stamped on background gate
After `cmd.Start()` the timestamp is always updated, even if the background `go install` process eventually fails (network down, etc.). This prevents accumulating zombie check attempts on every invocation during an outage. A 7-day retry interval is generous — missing one cycle is negligible.

### Fire-and-forget over blocking
`cmd.Start()` without `cmd.Wait()` means the background install runs independently. The current process is unaffected. If the parent exits, the child `go install` process continues and completes on its own. No goroutine lifecycle management needed.

### os.UserCacheDir() for timestamp storage
Uses XDG-compliant cache location (`~/.cache/litespec/` on Linux, `~/Library/Caches/litespec/` on macOS). Ephemeral state that can be deleted without consequence. No config file needed.

### runtime/debug.ReadBuildInfo() for module path
Self-discovering the module path from build metadata means no hardcoded strings. Survives module renames, works for forks. The build info is embedded at compile time by the Go toolchain at zero runtime cost.

### Background gate after arg check
The gate fires only when `len(os.Args) >= 2`, meaning a real command is present. Bare `litespec` invocations that just print usage are skipped — no point firing a background install for a process that exits immediately.

## File Changes

### `cmd/litespec/upgrade.go` (new)
- `cmdUpgrade(args []string) error` — handles `litespec upgrade` subcommand, includes `--help` flag handling via `hasHelpFlag()` + `printUpgradeHelp()`
- `isGoInstall() bool` — checks if running binary is in `$GOBIN` or `$GOPATH/bin`
- `getModulePath() (string, error)` — reads module path from `runtime/debug.ReadBuildInfo()`
- `fetchLatestVersion() (string, error)` — GET `api.github.com/repos/bermudi/litespec/releases/latest`, extracts tag name
- `parseSemver(tag string) (major, minor, patch int, error)` — parses semver strings

### `cmd/litespec/helpers.go` (modified)
- Add `printUpgradeHelp()` — help text for the upgrade command (follows existing `print*Help()` pattern)

### `cmd/litespec/main.go` (modified)
- Add `maybeBackgroundUpgrade()` function — checks `isGoInstall()`, reads timestamp from cache dir, fires `go install` via `cmd.Start()` if > 7 days elapsed
- Call `maybeBackgroundUpgrade()` after the `len(os.Args) < 2` check, before the command switch
- Add `"upgrade"` case to command switch in `run()`

### `cmd/litespec/main.go` — `printUsage()` (modified)
- Add `upgrade` to the commands list in help output

### `cmd/litespec/main_test.go` (modified)
- Tests for `isGoInstall()` with various `GOBIN`/`GOPATH` combinations
- Tests for `maybeBackgroundUpgrade()` — timestamp gating, go-install detection skip
- Tests for `cmdUpgrade()` — explicit upgrade flow, version comparison, error paths
- Tests for `getModulePath()` — with and without build info
- Tests for `parseSemver()` — valid and invalid semver strings
- Tests for `fetchLatestVersion()` — with httptest mock server
- `TestCLIUpgradeHelp` — verifies help output contains expected strings
