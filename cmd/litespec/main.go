package main

import (
	"fmt"
	"os"
)

const jsonFlag = "--json"

const version = "0.1.0"

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
	case "completion":
		return cmdCompletion(os.Args[2:])
	case "__complete":
		cmdComplete()
		return nil
	case "view":
		return cmdView(os.Args[2:])
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
  list [--specs|--changes] [--sort recent|name|deps]                   List specs or changes
  status [<name>]                                             Show artifact states
  validate [<name>] [--all|--changes|--specs] [--type T]      Validate changes and specs
  instructions <artifact>                                     Get artifact instructions
  archive <name>                                              Apply deltas and archive change
  view                                                        Dashboard overview with dependency graph
  update [--tools <ids>]                                      Regenerate skills and adapters
  completion <shell>                                          Generate shell completion script (bash, zsh, fish)

Tools:
  claude    Symlink skills into .claude/skills/ for Claude Code

Flags:
   --version    Print version
   --help       Print this help message
   --json       Output structured JSON (status, validate, list, instructions)
   --strict     Treat warnings as errors (validate)
   --all        Validate all changes and specs
   --changes    Validate all changes only
   --specs      Validate all specs only
   --type       Disambiguate name type: change|spec (validate)
   --sort       Sort changes by recent, name, or deps (list, default: recent)
`)
}
