## Motivation

litespec has no mechanism to keep itself current. Users who installed via `go install` must manually remember to re-run the install command when a new version ships. In practice, people run stale CLIs for weeks.

Most modern CLIs (brew, rtx, uv) solve this with a timestamp-gated background check that adds near-zero startup overhead. litespec should do the same — constrained to `go install` users only, avoiding the complexity and risk of custom binary self-replacement.

## Scope

- Add a `litespec upgrade` command that explicitly checks for and installs the latest version via `go install`
- Add a background update gate that silently runs `go install` if more than 7 days have elapsed since the last check
- Restrict all self-update behavior to binaries installed via `go install` (detected by checking if the binary lives in `$GOBIN` or `$GOPATH/bin`)
- Use `os.UserCacheDir()` (`~/.cache/litespec/` on Linux, `~/Library/Caches/litespec/` on macOS) for the timestamp file
- Derive the module path from `runtime/debug.ReadBuildInfo()` so it survives module renames
- `litespec upgrade` streams `go install` output and prints a hint to run `litespec update` in projects after upgrading
- Background gate suppresses all output and fires `go install` via `cmd.Start()` (fire and forget)

## Non-Goals

- Self-replacing binary downloads (downloading a binary and writing over the running executable)
- Support for package managers (homebrew, nix, etc.) — they manage their own updates
- A `--check` or dry-run flag for `litespec upgrade`
- Configurable update interval (hardcoded at 7 days)
- Any config file or persistent user preferences
- Automatic artifact regeneration after binary updates (just a printed hint)
- Merging `upgrade` into the existing `update` command (they serve different scopes)
