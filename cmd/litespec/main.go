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
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Print(`Usage: litespec <command> [options]

Commands:
  init [--tools <ids>]                            Initialize project structure
  new <name>                                      Create a new change
  list [--specs|--changes]                        List specs or changes
  status [--change <name>]                        Show artifact states
  validate [--change <name>] [--all] [--strict]   Validate changes and specs
  instructions <artifact> [--change <name>]       Get artifact instructions
  archive <name>                                  Apply deltas and archive change

Flags:
  --version    Print version
  --help       Print this help message
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
	var changeName string
	var asJSON bool
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--change":
			if i+1 < len(args) {
				changeName = args[i+1]
				i++
			}
		case jsonFlag:
			asJSON = true
		}
	}

	root, err := internal.FindProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if asJSON && changeName != "" {
		ctx, err := internal.LoadChangeContext(root, changeName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		status := internal.BuildChangeStatusJSON(ctx)
		data, _ := internal.MarshalJSON(status)
		fmt.Println(string(data))
		return
	}

	if asJSON && changeName == "" {
		changes, err := internal.ListChanges(root)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		var statuses []internal.ChangeStatusJSON
		for _, name := range changes {
			ctx, err := internal.LoadChangeContext(root, name)
			if err != nil {
				continue
			}
			statuses = append(statuses, internal.BuildChangeStatusJSON(ctx))
		}
		data, _ := internal.MarshalJSON(statuses)
		fmt.Println(string(data))
		return
	}

	if changeName == "" {
		changes, err := internal.ListChanges(root)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if len(changes) == 0 {
			fmt.Println("No active changes.")
			return
		}
		for _, name := range changes {
			ctx, err := internal.LoadChangeContext(root, name)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error loading change %q: %v\n", name, err)
				continue
			}
			fmt.Printf("%s\n", name)
			for _, art := range internal.Artifacts {
				fmt.Printf("  %-12s %s\n", art.ID+":", ctx.Artifacts[art.ID])
			}
		}
		return
	}

	ctx, err := internal.LoadChangeContext(root, changeName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Change: %s\n", changeName)
	if !ctx.Created.IsZero() {
		fmt.Printf("Created: %s\n", ctx.Created.Format("2006-01-02 15:04:05"))
	}
	fmt.Println()
	for _, art := range internal.Artifacts {
		fmt.Printf("  %-12s %-10s %s\n", art.ID, ctx.Artifacts[art.ID], art.Description)
	}
}

func cmdValidate(args []string) {
	var changeName string
	var all, strict, asJSON bool
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--change":
			if i+1 < len(args) {
				changeName = args[i+1]
				i++
			}
		case "--all":
			all = true
		case "--strict":
			strict = true
		case jsonFlag:
			asJSON = true
		}
	}

	root, err := internal.FindProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	var result *internal.ValidationResult

	if all || changeName == "" {
		result, err = internal.ValidateAll(root, strict)
	} else {
		result, err = internal.ValidateChange(root, changeName)
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
		fmt.Fprintf(os.Stderr, "usage: litespec instructions <artifact> [--change <name>] [--json]\n")
		os.Exit(1)
	}

	artifactID := args[0]
	var changeName string
	var asJSON bool
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--change":
			if i+1 < len(args) {
				changeName = args[i+1]
				i++
			}
		case jsonFlag:
			asJSON = true
		}
	}

	root, err := internal.FindProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if changeName == "" {
		fmt.Fprintf(os.Stderr, "error: --change <name> is required\n")
		os.Exit(1)
	}

	if artifactID == "apply" {
		instr, err := internal.BuildApplyInstructionsJSON(root, changeName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if asJSON {
			data, _ := internal.MarshalJSON(instr)
			fmt.Println(string(data))
			return
		}
		fmt.Printf("Change: %s\n", changeName)
		fmt.Printf("State: %s\n", instr.State)
		fmt.Printf("Progress: %d/%d (%d remaining)\n", instr.Progress.Complete, instr.Progress.Total, instr.Progress.Remaining)
		if instr.CurrentPhase < len(instr.Phases) {
			fmt.Printf("Current Phase: %s\n", instr.Phases[instr.CurrentPhase].Name)
		}
		fmt.Println(instr.Instruction)
		return
	}

	artifactInfo := internal.GetArtifact(artifactID)
	if artifactInfo == nil {
		fmt.Fprintf(os.Stderr, "unknown artifact: %s (valid: proposal, specs, design, tasks, apply)\n", artifactID)
		os.Exit(1)
	}

	if asJSON {
		instr, err := internal.BuildArtifactInstructionsJSON(root, changeName, artifactID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		data, _ := internal.MarshalJSON(instr)
		fmt.Println(string(data))
		return
	}

	template := internal.GetSkillTemplate("propose")
	fmt.Printf("Change: %s\n", changeName)
	fmt.Println(template)
}

func cmdArchive(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "usage: litespec archive <change-name>\n")
		os.Exit(1)
	}

	name := args[0]
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

	changeSpecsDir := internal.ChangeSpecsPath(root, name)
	entries, err := os.ReadDir(changeSpecsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading change specs: %v\n", err)
		os.Exit(1)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		capability := entry.Name()
		capDir := filepath.Join(changeSpecsDir, capability)
		files, readErr := os.ReadDir(capDir)
		if readErr != nil {
			continue
		}

		var deltas []*internal.DeltaSpec
		for _, f := range files {
			if filepath.Ext(f.Name()) != ".md" {
				continue
			}
			data, readErr := os.ReadFile(filepath.Join(capDir, f.Name()))
			if readErr != nil {
				fmt.Fprintf(os.Stderr, "error reading delta spec %s: %v\n", f.Name(), readErr)
				os.Exit(1)
			}
			delta, parseErr := internal.ParseDeltaSpec(string(data))
			if parseErr != nil {
				fmt.Fprintf(os.Stderr, "error parsing delta spec %s: %v\n", f.Name(), parseErr)
				os.Exit(1)
			}
			deltas = append(deltas, delta)
		}

		if len(deltas) == 0 {
			continue
		}

		mainSpecDir := filepath.Join(internal.SpecsPath(root), capability)
		mainSpecPath := filepath.Join(mainSpecDir, "spec.md")
		mainData, readErr := os.ReadFile(mainSpecPath)

		var mainSpec *internal.Spec
		if readErr != nil {
			mainSpec = &internal.Spec{Capability: deltas[0].Capability}
		} else {
			mainSpec, err = internal.ParseMainSpec(string(mainData))
			if err != nil {
				fmt.Fprintf(os.Stderr, "error parsing main spec for %s: %v\n", capability, err)
				os.Exit(1)
			}
		}

		merged, err := internal.MergeDelta(mainSpec, deltas)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error merging delta for %s: %v\n", capability, err)
			os.Exit(1)
		}

		if err := os.MkdirAll(mainSpecDir, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "error creating spec directory: %v\n", err)
			os.Exit(1)
		}
		if err := os.WriteFile(mainSpecPath, []byte(internal.SerializeSpec(merged)), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "error writing spec: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Updated spec: %s\n", capability)
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
