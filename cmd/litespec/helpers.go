package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/bermudi/litespec/internal"
)

func hasHelpFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return true
		}
	}
	return false
}

func checkUnknownFlags(args []string, validFlags map[string]bool) error {
	for _, arg := range args {
		if strings.HasPrefix(arg, "--") && !validFlags[arg] {
			return fmt.Errorf("unknown flag %s", arg)
		}
	}
	return nil
}

func printInitHelp() {
	fmt.Print(`Usage: litespec init [--tools <ids>]

Initialize a new litespec project in the current directory.

Creates:
  specs/canon/      Canonical spec directory
  specs/changes/    Change proposals directory
  .agents/skills/   Generated skill files

Flags:
  --tools <ids>     Comma-separated tool IDs (e.g., claude)

Examples:
  litespec init
  litespec init --tools claude
`)
}

func printUpdateHelp() {
	fmt.Print(`Usage: litespec update [--tools <ids>]

Regenerate skills and adapter commands from current specs.

Flags:
  --tools <ids>     Comma-separated tool IDs (e.g., claude)

Examples:
  litespec update
  litespec update --tools claude
`)
}

func printNewHelp() {
	fmt.Print(`Usage: litespec new <name>

Create a new change directory under specs/changes/.

Arguments:
  <name>            Change name (e.g., add-auth)

Examples:
  litespec new add-auth
`)
}

func printListHelp() {
	fmt.Print(`Usage: litespec list [--specs|--changes] [--sort recent|name|deps] [--json]

List active changes in the project (default) or specs with --specs.

Flags:
  --specs           List specs instead of changes
  --changes         List changes (default)
  --sort <field>    Sort changes by 'recent' (default), 'name', or 'deps'
  --json            Output as JSON

Examples:
  litespec list
  litespec list --changes --sort name
  litespec list --sort deps
  litespec list --specs --json
`)
}

func printStatusHelp() {
	fmt.Print(`Usage: litespec status [<name>] [--json]

Show artifact states for a change or all changes.

Arguments:
  <name>            Change name (omit to show all changes)

Flags:
  --json            Output as JSON

Examples:
  litespec status
  litespec status my-change
  litespec status --json
`)
}

func printValidateHelp() {
	fmt.Print(`Usage: litespec validate [<name>] [--all|--changes|--specs] [--type T] [--strict] [--json]

Validate changes and specs.

Arguments:
  <name>            Validate a specific change or spec by name

Flags:
  --all             Validate all changes and specs
  --changes         Validate all changes only
  --specs           Validate all specs only
  --type <T>        Disambiguate name: change|spec
  --strict          Treat warnings as errors
  --json            Output as JSON

Examples:
  litespec validate
  litespec validate my-change
  litespec validate --all --strict
  litespec validate shared --type spec
`)
}

func printInstructionsHelp() {
	fmt.Print(`Usage: litespec instructions <artifact> [--json]

Get artifact-specific instructions for writing proposals, specs, designs, or tasks.

Arguments:
  <artifact>        One of: proposal, specs, design, tasks

Flags:
  --json            Output as JSON

Examples:
  litespec instructions proposal
  litespec instructions design --json
`)
}

func printArchiveHelp() {
	fmt.Print(`Usage: litespec archive <name> [--allow-incomplete]

Apply deltas and archive a completed change.

Arguments:
  <name>            Change name to archive

Flags:
  --allow-incomplete    Archive even with incomplete tasks

Examples:
  litespec archive my-change
  litespec archive my-change --allow-incomplete
`)
}

func printViewHelp() {
	fmt.Print(`Usage: litespec view

Display a dashboard overview of specs, changes, and their dependency relationships.

Examples:
  litespec view
`)
}

func printUpgradeHelp() {
	fmt.Print(`Usage: litespec upgrade

Check for the latest version and upgrade via go install.

Only available for binaries installed via 'go install'.

Examples:
  litespec upgrade
  litespec upgrade --help
`)
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func contains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}

func sortChanges(changes []internal.ChangeInfo, sortBy string, root string) {
	switch sortBy {
	case "name":
		sort.Slice(changes, func(i, j int) bool {
			return changes[i].Name < changes[j].Name
		})
	case "deps":
		depMap, err := internal.LoadDepMap(root)
		if err != nil {
			fmt.Fprintf(os.Stderr, "WARN  could not load dependency map, falling back to alphabetical sort\n")
			sort.Slice(changes, func(i, j int) bool {
				return changes[i].Name < changes[j].Name
			})
			return
		}
		cycles := internal.DetectCycles(depMap)
		if len(cycles) > 0 {
			for _, cycle := range cycles {
				fmt.Fprintf(os.Stderr, "WARN  dependency cycle: %s\n", strings.Join(cycle, " -> "))
			}
			sort.Slice(changes, func(i, j int) bool {
				return changes[i].Name < changes[j].Name
			})
			return
		}
		sorted := internal.TopologicalSort(changes, depMap)
		copy(changes, sorted)
	default:
		sort.Slice(changes, func(i, j int) bool {
			return changes[i].LastModified.After(changes[j].LastModified)
		})
	}
}

func changeStatusText(c internal.ChangeInfo) string {
	if c.TotalTasks == 0 {
		return "No tasks"
	}
	if c.CompletedTasks == c.TotalTasks {
		return "✓ Complete"
	}
	return fmt.Sprintf("%d/%d tasks", c.CompletedTasks, c.TotalTasks)
}

func maxNameWidthChanges(changes []internal.ChangeInfo) int {
	m := 0
	for _, c := range changes {
		if len(c.Name) > m {
			m = len(c.Name)
		}
	}
	return m
}

func maxNameWidthSpecs(specs []internal.SpecInfo) int {
	m := 0
	for _, s := range specs {
		if len(s.Name) > m {
			m = len(s.Name)
		}
	}
	return m
}

func validateChangeName(name string) error {
	if name == "" {
		return fmt.Errorf("change name cannot be empty")
	}
	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return fmt.Errorf("change name cannot contain path separators")
	}
	if strings.Contains(name, "..") {
		return fmt.Errorf("change name cannot contain path traversal (..)")
	}
	if name != strings.TrimSpace(name) {
		return fmt.Errorf("change name cannot have leading or trailing whitespace")
	}
	if len(name) > 100 {
		return fmt.Errorf("change name cannot exceed 100 characters (got %d)", len(name))
	}
	for _, reserved := range []string{"canon", "changes", "archive"} {
		if name == reserved {
			return fmt.Errorf("change name %q is reserved", name)
		}
	}
	return nil
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

func saveToolIDs(root string, toolIDs []string) error {
	cfg, err := internal.ReadProjectConfig(root)
	if err != nil {
		return err
	}
	cfg.Tools = toolIDs
	return internal.WriteProjectConfig(root, cfg)
}
