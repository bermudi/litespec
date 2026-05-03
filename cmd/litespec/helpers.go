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

func printPatchHelp() {
	fmt.Print(`Usage: litespec patch <name> <capability>

Create a patch-mode change with only a delta spec. No proposal, design, or tasks.

Patch mode is for small, single-capability changes where the delta is the contract.
For larger changes or anything needing design discussion, use 'litespec new' instead.

Arguments:
  <name>            Change name (e.g., add-verbose-flag)
  <capability>     Capability to patch (e.g., cli)

Examples:
  litespec patch add-verbose-flag cli
  litespec patch fix-output-format status
`)
}

func printNewHelp() {
	fmt.Print(`Usage: litespec new <name> [--json]

Create a new change directory under specs/changes/ and show the artifact shape.

Arguments:
  <name>            Change name (e.g., add-auth)

Flags:
  --json            Output artifact states as JSON

Examples:
  litespec new add-auth
  litespec new add-auth --json
`)
}

func printListHelp() {
	fmt.Print(`Usage: litespec list [--specs|--changes|--decisions|--backlog] [--sort <mode>] [--status <state>] [--json]

List active changes in the project (default), specs with --specs, decisions with --decisions, or backlog items with --backlog.

Flags:
  --specs           List specs instead of changes
  --changes         List changes (default)
  --decisions       List architectural decision records
  --backlog         List backlog items by section
  --sort <field>    Sort by 'recent' (default), 'name', 'deps', or 'number' (decisions)
  --status <state>  Filter decisions by status: proposed, accepted, superseded (requires --decisions)
  --json            Output as JSON

Examples:
  litespec list
  litespec list --changes --sort name
  litespec list --sort deps
  litespec list --specs --json
  litespec list --decisions
  litespec list --decisions --status accepted --sort recent
  litespec list --backlog
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
	fmt.Print(`Usage: litespec validate [<name>] [--all|--changes|--specs|--decisions] [--type T] [--strict] [--json]

Validate changes, specs, and decisions.

Arguments:
  <name>            Validate a specific change, spec, or decision by name

Flags:
  --all             Validate all changes, specs, and decisions
  --changes         Validate all changes only
  --specs           Validate all specs only
  --decisions       Validate all decisions only
  --type <T>        Disambiguate name: change|spec|decision
  --strict          Treat warnings as errors
  --json            Output as JSON

Examples:
  litespec validate
  litespec validate my-change
  litespec validate --all --strict
  litespec validate shared --type spec
  litespec validate --decisions
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

Apply deltas to canonical specs and archive a change (marks it as implemented).

Arguments:
  <name>            Change name to archive

Flags:
  --allow-incomplete    Archive even with incomplete tasks or unarchived dependencies

Examples:
  litespec archive my-change
  litespec archive my-change --allow-incomplete
`)
}

func printViewHelp() {
	fmt.Print(`Usage: litespec view [--json]

Display a dashboard overview of specs, changes, and their dependency relationships.

Flags:
  --json  Output as JSON

Examples:
  litespec view
  litespec view --json
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

func pluralize(word string, count int) string {
	if count == 1 {
		return word
	}
	if strings.HasSuffix(word, "y") {
		return word[:len(word)-1] + "ies"
	}
	return word + "s"
}
