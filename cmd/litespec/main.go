package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
  list [--specs|--changes]                                    List specs or changes
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
	for _, arg := range args {
		switch arg {
		case "--specs":
			specsOnly = true
		case "--changes":
			changesOnly = true
		case jsonFlag:
			asJSON = true
		}
	}

	root, err := internal.FindProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if asJSON {
		type listOutput struct {
			Specs   []string `json:"specs"`
			Changes []string `json:"changes"`
		}

		out := listOutput{}
		if !changesOnly {
			names, err := internal.ListSpecs(root)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
			out.Specs = names
		}
		if !specsOnly {
			names, err := internal.ListChanges(root)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
			out.Changes = names
		}

		data, _ := internal.MarshalJSON(out)
		fmt.Println(string(data))
		return
	}

	if !changesOnly {
		specNames, err := internal.ListSpecs(root)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Specs:")
		if len(specNames) == 0 {
			fmt.Println("  (none)")
		}
		for _, name := range specNames {
			fmt.Printf("  %s\n", name)
		}
	}

	if !specsOnly {
		changeNames, err := internal.ListChanges(root)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Changes:")
		if len(changeNames) == 0 {
			fmt.Println("  (none)")
		}
		for _, name := range changeNames {
			fmt.Printf("  %s\n", name)
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
			ctx, err := internal.LoadChangeContext(root, n)
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
		ctx, err := internal.LoadChangeContext(root, n)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error loading change %q: %v\n", n, err)
			continue
		}
		fmt.Printf("%s\n", n)
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
		changes, _ := internal.ListChanges(root)
		specs, _ := internal.ListSpecs(root)
		isChange := contains(changes, positional)
		isSpec := contains(specs, positional)

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
			for _, name := range changes {
				changeResult, changeErr := internal.ValidateChange(root, name)
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
