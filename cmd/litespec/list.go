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
	if err := checkUnknownFlags(args, map[string]bool{"--specs": true, "--changes": true, "--decisions": true, "--sort": true, "--json": true, "--status": true}); err != nil {
		return err
	}

	var specsOnly, decisionsOnly, asJSON bool
	var sortBy, statusFilter string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--specs":
			specsOnly = true
		case "--changes":
		case "--decisions":
			decisionsOnly = true
		case jsonFlag:
			asJSON = true
		case "--sort":
			if i+1 >= len(args) {
				return fmt.Errorf("--sort requires a value (recent, name, or deps)")
			}
			sortBy = args[i+1]
			i++
		case "--status":
			if i+1 >= len(args) {
				return fmt.Errorf("--status requires a value (proposed, accepted, or superseded)")
			}
			statusFilter = args[i+1]
			i++
		}
	}

	if sortBy == "" {
		sortBy = "recent"
	}
	if decisionsOnly {
		if sortBy == "" || sortBy == "recent" {
			sortBy = "number"
		}
	}
	validSorts := map[string]bool{"recent": true, "name": true, "deps": true, "number": true}
	if !validSorts[sortBy] {
		return fmt.Errorf("--sort must be 'recent', 'name', 'deps', or 'number', got %q", sortBy)
	}
	if statusFilter != "" && !decisionsOnly {
		return fmt.Errorf("--status can only be used with --decisions")
	}
	if statusFilter != "" && statusFilter != "proposed" && statusFilter != "accepted" && statusFilter != "superseded" {
		return fmt.Errorf("--status must be 'proposed', 'accepted', or 'superseded', got %q", statusFilter)
	}
	if decisionsOnly && specsOnly {
		return fmt.Errorf("--decisions and --specs are mutually exclusive")
	}
	// --changes and --decisions are also mutually exclusive
	for _, arg := range args {
		if arg == "--changes" && decisionsOnly {
			return fmt.Errorf("--decisions and --changes are mutually exclusive")
		}
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
			Changes   []internal.ChangeListItemJSON   `json:"changes,omitempty"`
			Specs     []internal.SpecListItemJSON     `json:"specs,omitempty"`
			Decisions []internal.DecisionListItemJSON `json:"decisions,omitempty"`
			Warnings  []string                       `json:"warnings,omitempty"`
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
		} else if decisionsOnly {
			decisions, listErr := internal.ListDecisions(root)
			if listErr != nil {
				return listErr
			}
			decisions = filterDecisionsByStatus(decisions, statusFilter)
			decisions = sortDecisions(decisions, sortBy)
			for _, d := range decisions {
				item := internal.DecisionListItemJSON{
					Number:       d.Number,
					Slug:         d.Slug,
					Title:        d.Title,
					Status:       string(d.Status),
					Supersedes:   d.Supersedes,
					SupersededBy: d.SupersededBy,
				}
				if !d.LastModified.IsZero() {
					item.LastModified = d.LastModified.Format(time.RFC3339)
				}
				out.Decisions = append(out.Decisions, item)
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

	if decisionsOnly {
		decisions, listErr := internal.ListDecisions(root)
		if listErr != nil {
			return listErr
		}
		decisions = filterDecisionsByStatus(decisions, statusFilter)
		decisions = sortDecisions(decisions, sortBy)
		fmt.Println("Decisions:")
		if len(decisions) == 0 {
			fmt.Println("  (none)")
			return nil
		}
		fmt.Println()
		for _, d := range decisions {
			fmt.Printf("  %04d  %-30s  %-10s  %s\n", d.Number, d.Slug, d.Status, d.Title)
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

func filterDecisionsByStatus(decisions []*internal.Decision, statusFilter string) []*internal.Decision {
	if statusFilter == "" {
		return decisions
	}
	var filtered []*internal.Decision
	for _, d := range decisions {
		if string(d.Status) == statusFilter {
			filtered = append(filtered, d)
		}
	}
	return filtered
}

func sortDecisions(decisions []*internal.Decision, sortBy string) []*internal.Decision {
	switch sortBy {
	case "name":
		sort.Slice(decisions, func(i, j int) bool {
			return decisions[i].Slug < decisions[j].Slug
		})
	case "recent":
		sort.Slice(decisions, func(i, j int) bool {
			return decisions[i].LastModified.After(decisions[j].LastModified)
		})
	default: // "number"
		sort.Slice(decisions, func(i, j int) bool {
			return decisions[i].Number < decisions[j].Number
		})
	}
	return decisions
}
