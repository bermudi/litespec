package main

import (
	"fmt"
	"os"
	"os/exec"
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
	case "new":
		return cmdNew(os.Args[2:])
	case "validate":
		return cmdValidate(os.Args[2:])
	case "view":
		return cmdView(os.Args[2:])
	case "update":
		return cmdUpdate(os.Args[2:])
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
  new <name> [--issue N]            Link to GH issue (v2: no folder)
  validate [--all|--specs|--decisions] [--type T]   Validate specs and decisions
  view                              Dashboard overview
  update [--tools <ids>]            Regenerate skills and adapters
  upgrade                           Check for and install the latest version
  completion <shell>                Generate shell completion script (bash, zsh, fish)

Tools:
  claude    Symlink skills into .claude/skills/ for Claude Code

Flags:
   --version    Print version
   --help       Print this help message
   --json       Output structured JSON (validate, view)
   --strict     Treat warnings as errors (validate)
   --all        Validate all specs and decisions
   --specs      Validate all specs only
   --decisions  Validate all decisions only
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

	modulePath, err := getModulePath()
	if err != nil {
		return
	}

	if err := os.MkdirAll(stampDir, 0o755); err != nil {
		return
	}
	_ = os.WriteFile(stampFile, nil, 0o644)

	cmd := exec.Command("go", "install", modulePath+"@latest")
	cmd.Stdout = nil
	cmd.Stderr = nil
	_ = cmd.Start()
}
