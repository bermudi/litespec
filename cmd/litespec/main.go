package main

import (
	"fmt"
	"math"
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
	case "completion":
		cmdCompletion(os.Args[2:])
	case "__complete":
		cmdComplete()
	case "view":
		cmdView(os.Args[2:])
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

func cmdInit(args []string) {
	if hasHelpFlag(args) {
		printInitHelp()
		return
	}
	checkUnknownFlags(args, map[string]bool{"--tools": true})

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
	if hasHelpFlag(args) {
		printUpdateHelp()
		return
	}
	checkUnknownFlags(args, map[string]bool{"--tools": true})

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

func cmdView(args []string) {
	if hasHelpFlag(args) {
		printViewHelp()
		return
	}
	checkUnknownFlags(args, map[string]bool{})

	root, err := internal.FindProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	specs, err := internal.ListSpecs(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	changes, err := internal.ListChanges(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	var draft, active, completed []internal.ChangeInfo
	for _, c := range changes {
		if c.TotalTasks == 0 {
			draft = append(draft, c)
		} else if c.CompletedTasks == c.TotalTasks {
			completed = append(completed, c)
		} else {
			active = append(active, c)
		}
	}

	sort.Slice(active, func(i, j int) bool {
		pctI := float64(active[i].CompletedTasks) / float64(active[i].TotalTasks)
		pctJ := float64(active[j].CompletedTasks) / float64(active[j].TotalTasks)
		if pctI != pctJ {
			return pctI < pctJ
		}
		return active[i].Name < active[j].Name
	})

	totalReqs := 0
	for _, s := range specs {
		totalReqs += s.RequirementCount
	}

	totalCompletedTasks := 0
	totalTasks := 0
	for _, c := range active {
		totalCompletedTasks += c.CompletedTasks
		totalTasks += c.TotalTasks
	}

	fmt.Println()
	fmt.Println("Litespec Dashboard")
	fmt.Println()
	sep := strings.Repeat("═", 60)
	fmt.Println(sep)

	fmt.Println("Summary:")
	fmt.Printf("  ● Specifications: %d specs, %d requirements\n", len(specs), totalReqs)
	if len(draft) > 0 {
		fmt.Printf("  ● Draft Changes: %d\n", len(draft))
	}
	fmt.Printf("  ● Active Changes: %d in progress\n", len(active))
	fmt.Printf("  ● Completed Changes: %d\n", len(completed))
	if totalTasks > 0 {
		pct := int(math.Round(float64(totalCompletedTasks) / float64(totalTasks) * 100))
		fmt.Printf("  ● Task Progress: %d/%d (%d%% complete)\n", totalCompletedTasks, totalTasks, pct)
	}

	if len(active) > 0 {
		fmt.Println()
		fmt.Println("Active Changes")
		fmt.Println(strings.Repeat("─", 60))
		for _, c := range active {
			bar := createProgressBar(c.CompletedTasks, c.TotalTasks, 20)
			pct := int(math.Round(float64(c.CompletedTasks) / float64(c.TotalTasks) * 100))
			fmt.Printf("  ◉ %-30s %s %d%%\n", c.Name, bar, pct)
		}
	}

	if len(draft) > 0 {
		fmt.Println()
		fmt.Println("Draft Changes")
		fmt.Println(strings.Repeat("─", 60))
		for _, c := range draft {
			fmt.Printf("  ○ %s\n", c.Name)
		}
	}

	if len(completed) > 0 {
		fmt.Println()
		fmt.Println("Completed Changes")
		fmt.Println(strings.Repeat("─", 60))
		for _, c := range completed {
			fmt.Printf("  ✓ %s\n", c.Name)
		}
	}

	if len(specs) > 0 {
		fmt.Println()
		fmt.Println("Specifications")
		fmt.Println(strings.Repeat("─", 60))
		sort.Slice(specs, func(i, j int) bool {
			return specs[i].RequirementCount > specs[j].RequirementCount
		})
		for _, s := range specs {
			label := "requirement"
			if s.RequirementCount != 1 {
				label = "requirements"
			}
			fmt.Printf("  ▪ %-30s %d %s\n", s.Name, s.RequirementCount, label)
		}
	}

	depMap, err := internal.LoadDepMap(root)
	if err != nil {
		fmt.Println()
		fmt.Println(sep)
		return
	}

	hasDeps := false
	for _, deps := range depMap {
		if len(deps) > 0 {
			hasDeps = true
			break
		}
	}

	if hasDeps {
		fmt.Println()
		fmt.Println("Dependency Graph")
		fmt.Println(strings.Repeat("─", 60))
		renderDependencyGraph(depMap, changes)
	}

	fmt.Println()
	fmt.Println(sep)
	fmt.Printf("\nUse litespec list --changes or litespec list --specs for detailed views\n")
}

func createProgressBar(completed, total, width int) string {
	if total == 0 {
		return strings.Repeat("─", width)
	}
	pct := float64(completed) / float64(total)
	filled := int(math.Round(pct * float64(width)))
	empty := width - filled
	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", empty) + "]"
}

func renderDependencyGraph(depMap map[string][]string, changes []internal.ChangeInfo) {
	activeNames := make(map[string]bool)
	for _, c := range changes {
		activeNames[c.Name] = true
	}

	reverseMap := make(map[string][]string)
	for name, deps := range depMap {
		for _, dep := range deps {
			reverseMap[dep] = append(reverseMap[dep], name)
		}
	}

	related := make(map[string]bool)
	for name := range activeNames {
		if len(depMap[name]) > 0 || len(reverseMap[name]) > 0 {
			related[name] = true
		}
		for _, dep := range depMap[name] {
			if activeNames[dep] {
				related[dep] = true
			}
		}
	}

	var unrelated []string
	for name := range activeNames {
		if !related[name] {
			unrelated = append(unrelated, name)
		}
	}

	if len(related) == 0 {
		if len(unrelated) > 0 {
			sort.Strings(unrelated)
			fmt.Println("\nUnrelated:")
			for _, name := range unrelated {
				fmt.Printf("  - %s\n", name)
			}
		}
		return
	}

	var roots []string
	for name := range related {
		if len(depMap[name]) == 0 {
			roots = append(roots, name)
		}
	}

	sort.Strings(roots)

	seen := make(map[string]bool)

	var printNode func(name string, prefix string, isLast bool)
	printNode = func(name string, prefix string, isLast bool) {
		if seen[name] {
			return
		}
		seen[name] = true

		connector := "├── "
		if isLast {
			connector = "└── "
		}
		fmt.Printf("%s%s%s\n", prefix, connector, name)

		children := reverseMap[name]
		sort.Strings(children)
		for i, child := range children {
			newPrefix := prefix
			if isLast {
				newPrefix += "    "
			} else {
				newPrefix += "│   "
			}
			printNode(child, newPrefix, i == len(children)-1)
		}
	}

	for i, root := range roots {
		printNode(root, "", i == len(roots)-1)
	}

	if len(unrelated) > 0 {
		sort.Strings(unrelated)
		fmt.Println("\nUnrelated:")
		for _, name := range unrelated {
			fmt.Printf("  - %s\n", name)
		}
	}
}

func cmdList(args []string) {
	if hasHelpFlag(args) {
		printListHelp()
		return
	}
	checkUnknownFlags(args, map[string]bool{"--specs": true, "--changes": true, "--sort": true, "--json": true})

	var specsOnly, asJSON bool
	var sortBy string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--specs":
			specsOnly = true
		case "--changes":
		case jsonFlag:
			asJSON = true
		case "--sort":
			if i+1 >= len(args) {
				fmt.Fprintf(os.Stderr, "error: --sort requires a value (recent, name, or deps)\n")
				os.Exit(1)
			}
			sortBy = args[i+1]
			i++
		}
	}

	if sortBy == "" {
		sortBy = "recent"
	}
	if sortBy != "recent" && sortBy != "name" && sortBy != "deps" {
		fmt.Fprintf(os.Stderr, "error: --sort must be 'recent', 'name', or 'deps', got %q\n", sortBy)
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
		if specsOnly {
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
		} else {
			changes, listErr := internal.ListChanges(root)
			if listErr != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", listErr)
				os.Exit(1)
			}
			sortChanges(changes, sortBy, root)
			for _, c := range changes {
				item := internal.ChangeListItemJSON{
					Name:           c.Name,
					CompletedTasks: c.CompletedTasks,
					TotalTasks:     c.TotalTasks,
					LastModified:   c.LastModified.Format(time.RFC3339),
					Status:         internal.ChangeListStatus(c.CompletedTasks, c.TotalTasks),
				}
				meta, metaErr := internal.ReadChangeMeta(root, c.Name)
				if metaErr == nil && len(meta.DependsOn) > 0 {
					item.DependsOn = meta.DependsOn
				}
				out.Changes = append(out.Changes, item)
			}
		}

		data, _ := internal.MarshalJSON(out)
		fmt.Println(string(data))
		return
	}

	if specsOnly {
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
		} else {
			fmt.Println()
			maxName := maxNameWidthSpecs(specs)
			nameHeaderWidth := max(maxName, 4)
			fmt.Printf("  %-*s  %s\n", nameHeaderWidth, "Name", "Requirements")
			for _, s := range specs {
				fmt.Printf("  %-*s  %d\n", nameHeaderWidth, s.Name, s.RequirementCount)
			}
		}
		return
	}

	changes, listErr := internal.ListChanges(root)
	if listErr != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", listErr)
		os.Exit(1)
	}
	sortChanges(changes, sortBy, root)
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

func cmdStatus(args []string) {
	if hasHelpFlag(args) {
		printStatusHelp()
		return
	}
	checkUnknownFlags(args, map[string]bool{"--json": true})

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
	if hasHelpFlag(args) {
		printValidateHelp()
		return
	}
	checkUnknownFlags(args, map[string]bool{"--all": true, "--changes": true, "--specs": true, "--strict": true, "--json": true, "--type": true})

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
			if i+1 >= len(args) {
				fmt.Fprintf(os.Stderr, "error: --type requires a value (change or spec)\n")
				os.Exit(1)
			}
			typeFilter = args[i+1]
			i++
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

		if validateSpecs && validateChanges {
			result, err = internal.ValidateAll(root, strict)
		} else if validateSpecs {
			result, err = internal.ValidateSpecs(root)
		} else {
			changes, listErr := internal.ListChanges(root)
			if listErr != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", listErr)
				os.Exit(1)
			}
			result = &internal.ValidationResult{Valid: true}
			for _, ci := range changes {
				changeResult, changeErr := internal.ValidateChange(root, ci.Name)
				if changeErr != nil {
					fmt.Fprintf(os.Stderr, "error: %v\n", changeErr)
					os.Exit(1)
				}
				result.Errors = append(result.Errors, changeResult.Errors...)
				result.Warnings = append(result.Warnings, changeResult.Warnings...)
			}

			depMap, depErr := internal.LoadDepMap(root)
			if depErr == nil {
				cycles := internal.DetectCycles(depMap)
				for _, cycle := range cycles {
					path := strings.Join(cycle, " -> ")
					result.Errors = append(result.Errors, internal.ValidationIssue{
						Severity: internal.SeverityError,
						Message:  fmt.Sprintf("dependency cycle detected: %s", path),
					})
				}

				overlaps := internal.DetectOverlaps(root, changes, depMap)
				result.Warnings = append(result.Warnings, overlaps...)
			}

			result.Valid = len(result.Errors) == 0
			if strict && len(result.Warnings) > 0 {
				result.Valid = false
			}
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
	if hasHelpFlag(args) {
		printInstructionsHelp()
		return
	}
	checkUnknownFlags(args, map[string]bool{"--json": true})

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
	if hasHelpFlag(args) {
		printArchiveHelp()
		return
	}
	checkUnknownFlags(args, map[string]bool{"--allow-incomplete": true})

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
	if len(filtered) > 1 {
		fmt.Fprintf(os.Stderr, "error: unexpected arguments. Usage: litespec archive <name> [--allow-incomplete]\n")
		os.Exit(1)
	}

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

	dependents, depErr := internal.GetDependents(root, name)
	if depErr != nil {
		fmt.Fprintf(os.Stderr, "error checking dependents: %v\n", depErr)
		os.Exit(1)
	}
	if len(dependents) > 0 {
		if allowIncomplete {
			fmt.Fprintf(os.Stderr, "WARN  active changes depend on %q: %s\n", name, strings.Join(dependents, ", "))
		} else {
			fmt.Fprintf(os.Stderr, "ERROR  active changes depend on %q: %s. Archive them first or use --allow-incomplete.\n", name, strings.Join(dependents, ", "))
			os.Exit(1)
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
	if hasHelpFlag(args) {
		printNewHelp()
		return
	}
	checkUnknownFlags(args, map[string]bool{})
	if len(args) > 1 {
		fmt.Fprintf(os.Stderr, "error: unexpected arguments. Usage: litespec new <name>\n")
		os.Exit(1)
	}

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

func hasHelpFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return true
		}
	}
	return false
}

func cmdCompletion(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "usage: litespec completion <shell>\n")
		fmt.Fprintf(os.Stderr, "Supported shells: bash, zsh, fish\n")
		os.Exit(1)
	}

	shell := args[0]
	script, err := internal.CompletionScript(shell)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(script)
}

func cmdComplete() {
	root, _ := internal.FindProjectRoot()

	words := os.Args[2:]
	completions := internal.Complete(root, words)
	for _, c := range completions {
		fmt.Printf("%s\t%s\n", c.Candidate, c.Description)
	}
}

func checkUnknownFlags(args []string, validFlags map[string]bool) {
	for _, arg := range args {
		if strings.HasPrefix(arg, "--") && !validFlags[arg] {
			fmt.Fprintf(os.Stderr, "error: unknown flag %s\n", arg)
			os.Exit(1)
		}
	}
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
