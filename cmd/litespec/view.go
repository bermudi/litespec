package main

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/bermudi/litespec/internal"
)

func cmdView(args []string) error {
	if hasHelpFlag(args) {
		printViewHelp()
		return nil
	}
	if err := checkUnknownFlags(args, map[string]bool{}); err != nil {
		return err
	}

	root, err := internal.FindProjectRoot()
	if err != nil {
		return err
	}

	specs, err := internal.ListSpecs(root)
	if err != nil {
		return err
	}

	changes, err := internal.ListChanges(root)
	if err != nil {
		return err
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
	fmt.Printf("  ● Draft Changes: %d\n", len(draft))
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
		return nil
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
	return nil
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
