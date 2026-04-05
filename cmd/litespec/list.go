package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/bermudi/litespec/internal"
)

func cmdList(args []string) error {
	if hasHelpFlag(args) {
		printListHelp()
		return nil
	}
	if err := checkUnknownFlags(args, map[string]bool{"--specs": true, "--changes": true, "--sort": true, "--json": true}); err != nil {
		return err
	}

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
				return fmt.Errorf("--sort requires a value (recent, name, or deps)")
			}
			sortBy = args[i+1]
			i++
		}
	}

	if sortBy == "" {
		sortBy = "recent"
	}
	if sortBy != "recent" && sortBy != "name" && sortBy != "deps" {
		return fmt.Errorf("--sort must be 'recent', 'name', or 'deps', got %q", sortBy)
	}

	root, err := internal.FindProjectRoot()
	if err != nil {
		return err
	}

	if _, err := os.Stat(filepath.Join(root, internal.ProjectDirName)); err != nil {
		return fmt.Errorf("not a litespec project. Run 'litespec init' first")
	}

	if asJSON {
		type listOutput struct {
			Changes  []internal.ChangeListItemJSON `json:"changes,omitempty"`
			Specs    []internal.SpecListItemJSON   `json:"specs,omitempty"`
			Warnings []string                      `json:"warnings,omitempty"`
		}

		out := listOutput{}
		if specsOnly {
			specs, listErr := internal.ListSpecs(root)
			if listErr != nil {
				return listErr
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
				return listErr
			}
			sortChanges(changes, sortBy, root)
			for _, c := range changes {
				item := internal.ChangeListItemJSON{
					Name:           c.Name,
					CompletedTasks: c.CompletedTasks,
					TotalTasks:     c.TotalTasks,
					Status:         internal.ChangeListStatus(c.CompletedTasks, c.TotalTasks),
					DependsOn:      c.DependsOn,
				}
				if !c.LastModified.IsZero() {
					item.LastModified = c.LastModified.Format(time.RFC3339)
				}
				if !c.Created.IsZero() {
					item.Born = c.Created.Format(time.RFC3339)
				}
				out.Changes = append(out.Changes, item)
			}
		}

		data, err := internal.MarshalJSON(out)
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	if specsOnly {
		specs, listErr := internal.ListSpecs(root)
		if listErr != nil {
			return listErr
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
		return nil
	}

	changes, listErr := internal.ListChanges(root)
	if listErr != nil {
		return listErr
	}
	sortChanges(changes, sortBy, root)
	fmt.Println("Changes:")
	if len(changes) == 0 {
		fmt.Println("  (none)")
	}
	maxName := maxNameWidthChanges(changes)
	for _, c := range changes {
		status := changeStatusText(c)
		born := ""
		if !c.Created.IsZero() {
			born = c.Created.Format("2006-01-02")
		}
		relTime := ""
		if !c.LastModified.IsZero() {
			relTime = internal.FormatRelativeTime(c.LastModified)
		}
		fmt.Printf("  %-*s  %-16s %-12s %s\n", maxName, c.Name, status, born, relTime)
	}
	return nil
}
