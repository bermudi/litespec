package main

import (
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/bermudi/litespec/internal"
)

func newFlagSet(name string, usage func()) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	// Silence fs.Usage — Go's flag.Parse calls it for both --help and
	// unknown flags. We only want help for --help, not on typos.
	fs.Usage = func() {}
	helpRegistry[fs] = usage
	return fs
}

// helpRegistry pairs FlagSets with their help printers.
var helpRegistry = map[*flag.FlagSet]func(){}

// parseFlagSet wraps fs.Parse to handle help and unknown-flag errors.
// Returns (false, nil) if help was shown (caller should return nil).
// Returns (false, err) for parse errors (caller should return err).
// Returns (true, nil) for successful parse (caller should proceed).
func parseFlagSet(fs *flag.FlagSet, args []string) (bool, error) {
	args = reorderForFlagSet(fs, args)
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			if helpFn, ok := helpRegistry[fs]; ok {
				helpFn()
			}
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// reorderForFlagSet moves flags before positional args so flag.Parse
// doesn't stop at the first positional arg. String flags consume
// their value arg; bool flags don't.
func reorderForFlagSet(fs *flag.FlagSet, args []string) []string {
	var flags, positional []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			positional = append(positional, arg)
			continue
		}
		if arg == "--" {
			positional = append(positional, args[i+1:]...)
			break
		}
		flags = append(flags, arg)
		name := strings.TrimLeft(arg, "-")
		if eqIdx := strings.IndexByte(name, '='); eqIdx >= 0 {
			continue
		}
		if f := fs.Lookup(name); f != nil {
			if !isBoolFlag(f) && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
				flags = append(flags, args[i])
			}
		}
	}
	return append(flags, positional...)
}

func isBoolFlag(f *flag.Flag) bool {
	type boolFlag interface{ IsBoolFlag() bool }
	bf, ok := f.Value.(boolFlag)
	return ok && bf.IsBoolFlag()
}

func printInitHelp() {
	fmt.Print(`Usage: litespec init [--tools <ids>] [--json] [--minimal]

Initialize a new litespec project in the current directory.

Creates:
  specs/product.md    Product mental models + flows
  specs/glossary.md   Glossary (if missing)
  specs/decisions/    Decisions directory
  .agents/skills/   Generated skill files

Flags:
  --tools <ids>     Comma-separated tool IDs (e.g., claude)
  --json            Output as JSON
  --minimal         Minimal output

Examples:
  litespec init
  litespec init --tools claude
  litespec init --json
`)
}

func printUpdateHelp() {
	fmt.Print(`Usage: litespec update [--tools <ids>] [--json] [--minimal]

Regenerate skills and adapter commands from current specs.

Flags:
  --tools <ids>     Comma-separated tool IDs (e.g., claude)
  --json            Output as JSON
  --minimal         Minimal output

Examples:
  litespec update
  litespec update --tools claude
  litespec update --json
`)
}

func printValidateHelp() {
	fmt.Print(`Usage: litespec validate [<name>|--all|--specs|--decisions|--issue N|--queue <path>] [--type T] [--strict] [--json] [--minimal]

Validate specs, decisions, and queues.

Arguments:
  <name>            Validate a specific spec or decision by name

Flags:
  --all             Validate all specs, decisions, and queues
  --specs           Validate all specs only
  --decisions       Validate all decisions only
  --issue <N>       Fetch and validate a single GH issue by number
  --queue <path>    Validate a single local queue markdown file
  --type <T>        Disambiguate name: spec|decision
  --strict          Treat warnings as errors
  --json            Output as JSON
  --minimal         Minimal output

Examples:
  litespec validate
  litespec validate my-spec
  litespec validate --all --strict
  litespec validate shared --type spec
  litespec validate --decisions
  litespec validate --queue specs/queues/add-auth.md
`)
}

func printViewHelp() {
	fmt.Print(`Usage: litespec view [--json] [--minimal]

Display a dashboard overview of product, specs, decisions, and open GH issues.

Flags:
  --json            Output as JSON
  --minimal         Minimal output

Examples:
  litespec view
  litespec view --json
`)
}

func printUpgradeHelp() {
	fmt.Print(`Usage: litespec upgrade [--json] [--minimal]

Check for the latest version and upgrade via go install.

Only available for binaries installed via 'go install'.

Flags:
  --json            Output as JSON
  --minimal         Minimal output

Examples:
  litespec upgrade
  litespec upgrade --json
`)
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func validateToolIDs(toolIDs []string) error {
	validIDs := internal.ValidToolIDs()
	for _, id := range toolIDs {
		found := false
		for _, valid := range validIDs {
			if id == valid {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("unknown tool %q (valid: %s)", id, strings.Join(validIDs, ", "))
		}
	}
	return nil
}

func pluralize(word string, count int) string {
	if count == 1 {
		return word
	}
	if strings.HasSuffix(word, "y") {
		return word[:len(word)-1] + "ies"
	}
	return word + "s"
}

func hasNonExemptWarnings(warnings []internal.ValidationIssue) bool {
	for _, w := range warnings {
		if !w.StrictExempt {
			return true
		}
	}
	return false
}
