package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/bermudi/litespec/internal"
)

const jsonFlag = "--json"

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "--version", "-v":
		fmt.Printf("litespec v%s\n", version)
	case "--help", "-h":
		printUsage()
	case "init":
		cmdInit(os.Args[2:])
	case "list":
		cmdList(os.Args[2:])
	case "status":
		cmdStatus(os.Args[2:])
	case "validate":
		cmdValidate(os.Args[2:])
	case "instructions":
		cmdInstructions(os.Args[2:])
	case "archive":
		cmdArchive(os.Args[2:])
	case "new":
		cmdNew(os.Args[2:])
	case "update":
		cmdUpdate(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Print(`Usage: litespec <command> [options]

Commands:
  init [--tools <ids>]                                        Initialize project structure
  new <name>                                                  Create a new change
  list [--specs|--changes] [--sort recent|name]                   List specs or changes
  status [<name>]                                             Show artifact states
  validate [<name>] [--all|--changes|--specs] [--type T]      Validate changes and specs
  instructions <artifact>                                     Get artifact instructions
  archive <name>                                              Apply deltas and archive change
  update [--tools <ids>]                                      Regenerate skills and adapters

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
  --sort       Sort changes by recent or name (list, default: recent)
`)
}

func cmdInit(args []string) {
	var tools string
	for i := 0; i < len(args); i++ {
		if args[i] == "--tools" && i+1 < len(args) {
			tools = args[i+1]
			i++
		}
	}

	root, err := internal.FindProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if err := internal.InitProject(root); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Created specs/ directory structure")

	if err := internal.GenerateSkills(root); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Generated .agents/skills/")

	if tools != "" {
		toolIDs := splitCSV(tools)
		if err := internal.GenerateAdapterCommands(root, toolIDs); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Generated adapter commands for: %s\n", tools)
	}

	fmt.Println("Project initialized.")
}

func cmdUpdate(args []string) {
	var tools string
	for i := 0; i < len(args); i++ {
		if args[i] == "--tools" && i+1 < len(args) {
			tools = args[i+1]
			i++
		}
	}

	root, err := internal.FindProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if _, err := os.Stat(filepath.Join(root, internal.ProjectDirName)); err != nil {
		fmt.Fprintf(os.Stderr, "error: not a litespec project. Run 'litespec init' first.\n")
		os.Exit(1)
	}

	if err := internal.GenerateSkills(root); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Updated .agents/skills/")

	if tools != "" {
		toolIDs := splitCSV(tools)
		if err := internal.GenerateAdapterCommands(root, toolIDs); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Updated adapter symlinks for: %s\n", tools)
	}
}

func cmdList(args []string) {
	var specsOnly, changesOnly, asJSON bool
	var sortBy string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--specs":
			specsOnly = true
		case "--changes":
			changesOnly = true
		case jsonFlag:
			asJSON = true
		case "--sort":
			if i+1 < len(args) {
				sortBy = args[i+1]
				i++
			}
		}
	}

	if sortBy == "" {
		sortBy = "recent"
	}
	if sortBy != "recent" && sortBy != "name" {
		fmt.Fprintf(os.Stderr, "error: --sort must be 'recent' or 'name', got %q\n", sortBy)
		os.Exit(1)
	}

	root, err := internal.FindProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if asJSON {
		type listOutput struct {
			Changes []internal.ChangeListItemJSON `json:"changes,omitempty"`
			Specs   []internal.SpecListItemJSON   `json:"specs,omitempty"`
		}

		out := listOutput{}
		if !changesOnly {
			specs, listErr := internal.ListSpecs(root)
			if listErr != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", listErr)
				os.Exit(1)
			}
			sort.Slice(specs, func(i, j int) bool {
				return specs[i].Name < specs[j].Name
			})
			for _, s := range specs {
				out.Specs = append(out.Specs, internal.SpecListItemJSON{
					Name:             s.Name,
					RequirementCount: s.RequirementCount,
				})
			}
		}
		if !specsOnly {
			changes, listErr := internal.ListChanges(root)
			if listErr != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", listErr)
				os.Exit(1)
			}
			sortChanges(changes, sortBy)
			for _, c := range changes {
				out.Changes = append(out.Changes, internal.ChangeListItemJSON{
					Name:           c.Name,
					CompletedTasks: c.CompletedTasks,
					TotalTasks:     c.TotalTasks,
					LastModified:   c.LastModified.Format(time.RFC3339),
					Status:         internal.ChangeListStatus(c.CompletedTasks, c.TotalTasks),
				})
			}
		}

		data, _ := internal.MarshalJSON(out)
		fmt.Println(string(data))
		return
	}

	if !changesOnly {
		specs, listErr := internal.ListSpecs(root)
		if listErr != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", listErr)
			os.Exit(1)
		}
		sort.Slice(specs, func(i, j int) bool {
			return specs[i].Name < specs[j].Name
		})
		fmt.Println("Specs:")
		if len(specs) == 0 {
			fmt.Println("  (none)")
		}
		maxName := maxNameWidthSpecs(specs)
		for _, s := range specs {
			fmt.Printf("  %-*s  requirements %d\n", maxName, s.Name, s.RequirementCount)
		}
	}

	if !specsOnly {
		changes, listErr := internal.ListChanges(root)
		if listErr != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", listErr)
			os.Exit(1)
		}
		sortChanges(changes, sortBy)
		fmt.Println("Changes:")
		if len(changes) == 0 {
			fmt.Println("  (none)")
		}
		maxName := maxNameWidthChanges(changes)
		for _, c := range changes {
			status := changeStatusText(c)
			relTime := internal.FormatRelativeTime(c.LastModified)
			fmt.Printf("  %-*s  %-16s %s\n", maxName, c.Name, status, relTime)
		}
	}
}

func cmdStatus(args []string) {
	var name string
	var asJSON bool
	for _, arg := range args {
		switch arg {
		case jsonFlag:
			asJSON = true
		default:
			if !strings.HasPrefix(arg, "-") && name == "" {
				name = arg
			}
		}
	}

	root, err := internal.FindProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if name != "" {
		if !internal.ChangeExists(root, name) {
			fmt.Fprintf(os.Stderr, "error: change %q not found\n", name)
			os.Exit(1)
		}

		ctx, err := internal.LoadChangeContext(root, name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		if asJSON {
			status := internal.BuildChangeStatusJSON(ctx)
			data, _ := internal.MarshalJSON(status)
			fmt.Println(string(data))
			return
		}

		fmt.Printf("Change: %s\n", name)
		if !ctx.Created.IsZero() {
			fmt.Printf("Created: %s\n", ctx.Created.Format("2006-01-02 15:04:05"))
		}
		fmt.Println()
		for _, art := range internal.Artifacts {
			fmt.Printf("  %-12s %-10s %s\n", art.ID, ctx.Artifacts[art.ID], art.Description)
		}
		return
	}

	changes, err := internal.ListChanges(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if asJSON {
		var statuses []internal.ChangeStatusJSON
		for _, n := range changes {
			ctx, err := internal.LoadChangeContext(root, n.Name)
			if err != nil {
				continue
			}
			statuses = append(statuses, internal.BuildChangeStatusJSON(ctx))
		}
		data, _ := internal.MarshalJSON(statuses)
		fmt.Println(string(data))
		return
	}

	if len(changes) == 0 {
		fmt.Println("No active changes.")
		return
	}
	for _, n := range changes {
		ctx, err := internal.LoadChangeContext(root, n.Name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error loading change %q: %v\n", n.Name, err)
			continue
		}
		fmt.Printf("%s\n", n.Name)
		for _, art := range internal.Artifacts {
			fmt.Printf("  %-12s %s\n", art.ID+":", ctx.Artifacts[art.ID])
		}
	}
}

func cmdValidate(args []string) {
	var positional string
	var flagAll, flagChanges, flagSpecs, strict, asJSON bool
	var typeFilter string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--all":
			flagAll = true
		case "--changes":
			flagChanges = true
		case "--specs":
			flagSpecs = true
		case "--strict":
			strict = true
		case jsonFlag:
			asJSON = true
		case "--type":
			if i+1 < len(args) {
				typeFilter = args[i+1]
				i++
			}
		default:
			if !strings.HasPrefix(args[i], "-") && positional == "" {
				positional = args[i]
			}
		}
	}

	hasBulk := flagAll || flagChanges || flagSpecs

	if positional != "" && hasBulk {
		fmt.Fprintf(os.Stderr, "error: positional name and bulk flags (--all, --changes, --specs) are mutually exclusive\n")
		os.Exit(1)
	}

	if typeFilter != "" && positional == "" {
		fmt.Fprintf(os.Stderr, "error: --type requires a positional name\n")
		os.Exit(1)
	}

	if typeFilter != "" && hasBulk {
		fmt.Fprintf(os.Stderr, "error: --type cannot be used with bulk flags\n")
		os.Exit(1)
	}

	if typeFilter != "" && typeFilter != "change" && typeFilter != "spec" {
		fmt.Fprintf(os.Stderr, "error: --type must be 'change' or 'spec', got %q\n", typeFilter)
		os.Exit(1)
	}

	root, err := internal.FindProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	var result *internal.ValidationResult

	if positional != "" {
		changeList, _ := internal.ListChanges(root)
		specList, _ := internal.ListSpecs(root)
		changeNames := make([]string, len(changeList))
		for i, c := range changeList {
			changeNames[i] = c.Name
		}
		specNames := make([]string, len(specList))
		for i, s := range specList {
			specNames[i] = s.Name
		}
		isChange := contains(changeNames, positional)
		isSpec := contains(specNames, positional)

		if typeFilter == "change" {
			isSpec = false
		} else if typeFilter == "spec" {
			isChange = false
		}

		if isChange && isSpec {
			fmt.Fprintf(os.Stderr, "error: %q is ambiguous — exists as both a change and a spec. Use --type change or --type spec\n", positional)
			os.Exit(1)
		}

		if !isChange && !isSpec {
			fmt.Fprintf(os.Stderr, "error: %q not found as a change or spec\n", positional)
			os.Exit(1)
		}

		if isChange {
			result, err = internal.ValidateChange(root, positional)
		} else {
			result, err = internal.ValidateSpec(root, positional)
		}
	} else {
		validateSpecs := flagSpecs || flagAll || (!flagChanges && !flagSpecs)
		validateChanges := flagChanges || flagAll || (!flagChanges && !flagSpecs)

		result = &internal.ValidationResult{Valid: true}

		if validateSpecs {
			specResult, specErr := internal.ValidateSpecs(root)
			if specErr != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", specErr)
				os.Exit(1)
			}
			result.Errors = append(result.Errors, specResult.Errors...)
			result.Warnings = append(result.Warnings, specResult.Warnings...)
		}

		if validateChanges {
			changes, listErr := internal.ListChanges(root)
			if listErr != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", listErr)
				os.Exit(1)
			}
			for _, ci := range changes {
				changeResult, changeErr := internal.ValidateChange(root, ci.Name)
				if changeErr != nil {
					fmt.Fprintf(os.Stderr, "error: %v\n", changeErr)
					os.Exit(1)
				}
				result.Errors = append(result.Errors, changeResult.Errors...)
				result.Warnings = append(result.Warnings, changeResult.Warnings...)
			}
		}

		result.Valid = len(result.Errors) == 0
		if strict && len(result.Warnings) > 0 {
			result.Valid = false
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if asJSON {
		out := internal.BuildValidationResultJSON(result)
		data, _ := internal.MarshalJSON(out)
		fmt.Println(string(data))
		if !result.Valid || (strict && len(result.Warnings) > 0) {
			os.Exit(1)
		}
		return
	}

	failed := !result.Valid
	for _, issue := range result.Errors {
		fmt.Printf("ERROR  %s: %s\n", issue.File, issue.Message)
	}
	for _, issue := range result.Warnings {
		fmt.Printf("WARN   %s: %s\n", issue.File, issue.Message)
	}

	if strict && len(result.Warnings) > 0 {
		failed = true
	}

	if !failed {
		fmt.Println("Validation passed.")
	} else {
		os.Exit(1)
	}
}

func cmdInstructions(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "usage: litespec instructions <artifact> [--json]\n")
		os.Exit(1)
	}

	artifactID := args[0]
	var asJSON bool
	for _, arg := range args[1:] {
		if arg == jsonFlag {
			asJSON = true
		}
	}

	artifactInfo := internal.GetArtifact(artifactID)
	if artifactInfo == nil {
		fmt.Fprintf(os.Stderr, "unknown artifact: %s (valid: proposal, specs, design, tasks)\n", artifactID)
		os.Exit(1)
	}

	if asJSON {
		instr, err := internal.BuildArtifactInstructionsStandaloneJSON(artifactID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		data, _ := internal.MarshalJSON(instr)
		fmt.Println(string(data))
		return
	}

	instruction := internal.GetSkillTemplate(internal.ArtifactInstructionID(artifactID))
	fmt.Println(instruction)
}

func cmdArchive(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "usage: litespec archive <change-name> [--allow-incomplete]\n")
		os.Exit(1)
	}

	allowIncomplete := false
	filtered := args[:0]
	for _, a := range args {
		if a == "--allow-incomplete" {
			allowIncomplete = true
			continue
		}
		filtered = append(filtered, a)
	}
	name := filtered[0]

	root, err := internal.FindProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	result, err := internal.ValidateChange(root, name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if !result.Valid {
		for _, issue := range result.Errors {
			fmt.Fprintf(os.Stderr, "ERROR  %s: %s\n", issue.File, issue.Message)
		}
		fmt.Fprintf(os.Stderr, "Validation failed. Fix errors before archiving.\n")
		os.Exit(1)
	}
	for _, issue := range result.Warnings {
		fmt.Printf("WARN   %s: %s\n", issue.File, issue.Message)
	}

	if !allowIncomplete {
		tasksPath := filepath.Join(internal.ChangePath(root, name), "tasks.md")
		tasksData, tasksErr := os.ReadFile(tasksPath)
		if tasksErr == nil {
			completed, total := internal.TaskCompletion(string(tasksData))
			if completed < total {
				fmt.Fprintf(os.Stderr, "ERROR  %d/%d tasks completed. Finish tasks or use --allow-incomplete.\n", completed, total)
				os.Exit(1)
			}
		}
	}

	writes, err := internal.PrepareArchiveWrites(root, name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if err := internal.WritePendingSpecs(writes); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	for _, w := range writes {
		fmt.Printf("Updated spec: %s\n", w.Capability)
	}

	if err := internal.ArchiveChange(root, name); err != nil {
		fmt.Fprintf(os.Stderr, "error archiving change: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Change %q archived successfully.\n", name)
}

func cmdNew(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "usage: litespec new <change-name>\n")
		os.Exit(1)
	}

	name := args[0]
	root, err := internal.FindProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if err := internal.CreateChange(root, name); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(internal.ChangePath(root, name))
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

func sortChanges(changes []internal.ChangeInfo, sortBy string) {
	switch sortBy {
	case "name":
		sort.Slice(changes, func(i, j int) bool {
			return changes[i].Name < changes[j].Name
		})
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
	max := 0
	for _, c := range changes {
		if len(c.Name) > max {
			max = len(c.Name)
		}
	}
	return max
}

func maxNameWidthSpecs(specs []internal.SpecInfo) int {
	max := 0
	for _, s := range specs {
		if len(s.Name) > max {
			max = len(s.Name)
		}
	}
	return max
}
