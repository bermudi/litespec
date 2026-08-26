package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"
)

var version = "dev"

func init() {
	if v := resolveVersion(); v != "" {
		version = v
	}
}

func resolveVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "" || info.Main.Version == "(devel)" {
		return ""
	}
	return strings.TrimPrefix(info.Main.Version, "v")
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		printUsage()
		return nil
	}

	maybeBackgroundUpgrade()

	switch os.Args[1] {
	case "--version", "-v":
		fmt.Printf("litespec v%s\n", version)
		return nil
	case "--help", "-h":
		printUsage()
		return nil
	case "init":
		return cmdInit(os.Args[2:])
	case "validate":
		return cmdValidate(os.Args[2:])
	case "view":
		return cmdView(os.Args[2:])
	case "update":
		return cmdUpdate(os.Args[2:])
	case "digest":
		return cmdDigest(os.Args[2:])
	case "upgrade":
		return cmdUpgrade(os.Args[2:])
	case "completion":
		return cmdCompletion(os.Args[2:])
	case "__complete":
		cmdComplete()
		return nil
	default:
		printUsage()
		return fmt.Errorf("unknown command: %s", os.Args[1])
	}
}

func printUsage() {
	fmt.Print(`Usage: litespec <command> [options]

Workflow (two lanes):
  Small fix: read product/spec/decisions -> edit code -> update spec if contract
  New feature: plan[fuzzy] -> plan[clear] (GH issue) -> grill-me -> build -> review -> close

Commands:
  init [--tools <ids>]              Initialize project structure
  validate [--all|--specs|--decisions|--issue <N>|--queue <path>] [--type T]   Validate specs, decisions, and queues
  view                              Dashboard overview
  update [--tools <ids>]            Regenerate skills and adapters
  digest --issue <N> | --queue <p>  Print expected unit contract digests for a queue
  upgrade                           Check for and install the latest version
  completion <shell>                Generate shell completion script (bash, zsh, fish)

Tools:
  claude    Symlink skills into .claude/skills/ for Claude Code

Flags:
   --version    Print version
   --help       Print this help message
   --json       Output structured JSON (validate, view)
   --strict     Treat warnings as errors (validate)
   --all        Validate all specs, decisions, and queues
   --specs      Validate all specs only
   --decisions  Validate all decisions only
   --issue <N>  Fetch and validate one GH queue issue
   --queue <path>  Validate one local queue file
   --type       Disambiguate name type: spec|decision (validate)
`)
}

const backgroundUpgradeInterval = 7 * 24 * time.Hour

func maybeBackgroundUpgrade() {
	if !isGoInstall() {
		return
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return
	}
	stampDir := filepath.Join(cacheDir, "litespec")
	stampFile := filepath.Join(stampDir, "last-update-check")

	info, err := os.Stat(stampFile)
	if err == nil && time.Since(info.ModTime()) < backgroundUpgradeInterval {
		return
	}

	modulePath, err := modulePathFn()
	if err != nil {
		return
	}

	if err := os.MkdirAll(stampDir, 0o755); err != nil {
		return
	}
	_ = os.WriteFile(stampFile, nil, 0o644)

	local := version
	if local == "dev" || local == "" {
		local = "0.0.0"
	}
	latest, err := fetchLatestVersionFor(local)
	if err != nil {
		return
	}
	cmp, err := compareSemver(local, latest)
	if err != nil || cmp >= 0 {
		return
	}
	startBackgroundInstall(modulePath, latest)
}
