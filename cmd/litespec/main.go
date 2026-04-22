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

const jsonFlag = "--json"

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
	case "list":
		return cmdList(os.Args[2:])
	case "status":
		return cmdStatus(os.Args[2:])
	case "validate":
		return cmdValidate(os.Args[2:])
	case "instructions":
		return cmdInstructions(os.Args[2:])
	case "archive":
		return cmdArchive(os.Args[2:])
	case "new":
		return cmdNew(os.Args[2:])
	case "update":
		return cmdUpdate(os.Args[2:])
	case "upgrade":
		return cmdUpgrade(os.Args[2:])
	case "completion":
		return cmdCompletion(os.Args[2:])
	case "__complete":
		cmdComplete()
		return nil
	case "preview":
		return cmdPreview(os.Args[2:])
	case "view":
		return cmdView(os.Args[2:])
	case "decide":
		return cmdDecide(os.Args[2:])
	case "import":
		return cmdImport(os.Args[2:])
	default:
		printUsage()
		return fmt.Errorf("unknown command: %s", os.Args[1])
	}
}

func printUsage() {
	fmt.Print(`Usage: litespec <command> [options]

Commands:
  init [--tools <ids>]                                        Initialize project structure
  new <name>                                                  Create a new change
  list [--specs|--changes|--decisions] [--sort recent|name|deps|number] [--status <state>]   List specs, changes, or decisions
  status [<name>]                                             Show artifact states
  validate [<name>] [--all|--changes|--specs|--decisions] [--type T]      Validate changes, specs, and decisions
  instructions <artifact>                                     Get artifact instructions
  archive <name>                                              Apply deltas and archive change
  preview <name> [--json]                                     Preview what archive would do to canon specs
  view                                                        Dashboard overview with dependency graph
  decide <slug>                                              Create a new architectural decision record
  import [--source <dir>] [--dry-run] [--force]               Import OpenSpec project to litespec
  update [--tools <ids>]                                      Regenerate skills and adapters
  upgrade                                                     Check for and install the latest version
  completion <shell>                                          Generate shell completion script (bash, zsh, fish)

Tools:
  claude    Symlink skills into .claude/skills/ for Claude Code

Flags:
   --version    Print version
   --help       Print this help message
   --json       Output structured JSON (status, validate, list, instructions)
   --strict     Treat warnings as errors (validate)
   --all        Validate all changes, specs, and decisions
   --changes    Validate all changes only
   --specs      Validate all specs only
   --decisions  Validate all decisions only
   --type       Disambiguate name type: change|spec|decision (validate)
    --sort       Sort changes by recent, name, or deps (list, default: recent)
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
