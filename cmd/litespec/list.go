package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/bermudi/litespec/internal"
)

type listData struct {
	kind string // "specs", "decisions", "backlog", "changes"

	specs     []internal.SpecInfo
	decisions []*internal.Decision
	backlog   []internal.BacklogItem
	changes   []internal.ChangeInfo
}

func cmdList(args []string) error {
	if hasHelpFlag(args) {
		printListHelp()
		return nil
	}
	if err := checkUnknownFlags(args, map[string]bool{"--specs": true, "--changes": true, "--decisions": true, "--backlog": true, "--sort": true, "--json": true, "--status": true, "--minimal": true}); err != nil {
		return err
	}

	var specsOnly, decisionsOnly, backlogOnly, asJSON, asMinimal bool
	var sortBy, statusFilter string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--specs":
			specsOnly = true
		case "--changes":
		case "--decisions":
			decisionsOnly = true
		case "--backlog":
			backlogOnly = true
		case jsonFlag:
			asJSON = true
		case minimalFlag:
			asMinimal = true
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
	if backlogOnly && (specsOnly || decisionsOnly) {
		return fmt.Errorf("--backlog is mutually exclusive with --specs and --decisions")
	}
	for _, arg := range args {
		if arg == "--changes" && decisionsOnly {
			return fmt.Errorf("--decisions and --changes are mutually exclusive")
		}
		if arg == "--changes" && backlogOnly {
			return fmt.Errorf("--backlog and --changes are mutually exclusive")
		}
	}

	root, err := internal.FindProjectRoot()
	if err != nil {
		return err
	}

	if _, err := os.Stat(filepath.Join(root, internal.ProjectDirName)); err != nil {
		return fmt.Errorf("not a litespec project. Run 'litespec init' first")
	}

	data, err := fetchListData(root, specsOnly, decisionsOnly, backlogOnly, sortBy, statusFilter)
	if err != nil {
		return err
	}

	switch {
	case asJSON:
		return renderListJSON(data, asMinimal)
	case asMinimal:
		return renderListMinimalText(data)
	default:
		return renderListText(data)
	}
}

func fetchListData(root string, specsOnly, decisionsOnly, backlogOnly bool, sortBy, statusFilter string) (*listData, error) {
	d := &listData{}

	switch {
	case specsOnly:
		d.kind = "specs"
		specs, err := internal.ListSpecs(root)
		if err != nil {
			return nil, err
		}
		sort.Slice(specs, func(i, j int) bool {
			return specs[i].Name < specs[j].Name
		})
		d.specs = specs

	case decisionsOnly:
		d.kind = "decisions"
		decisions, err := internal.ListDecisions(root)
		if err != nil {
			return nil, err
		}
		decisions = filterDecisionsByStatus(decisions, statusFilter)
		decisions = sortDecisions(decisions, sortBy)
		d.decisions = decisions

	case backlogOnly:
		d.kind = "backlog"
		items, err := internal.ParseBacklogItems(internal.BacklogPath(root))
		if err != nil {
			return nil, err
		}
		d.backlog = items

	default:
		d.kind = "changes"
		changes, err := internal.ListChanges(root)
		if err != nil {
			return nil, err
		}
		sortChanges(changes, sortBy, root)
		d.changes = changes
	}

	return d, nil
}

func renderListJSON(d *listData, asMinimal bool) error {
	if asMinimal {
		return renderListMinimalJSON(d)
	}

	type listOutput struct {
		Changes   []internal.ChangeListItemJSON   `json:"changes,omitempty"`
		Specs     []internal.SpecListItemJSON     `json:"specs,omitempty"`
		Decisions []internal.DecisionListItemJSON `json:"decisions,omitempty"`
		Backlog   []internal.BacklogItemJSON      `json:"backlog,omitempty"`
	}

	out := listOutput{}
	switch d.kind {
	case "specs":
		for _, s := range d.specs {
			out.Specs = append(out.Specs, internal.SpecListItemJSON{
				Name:             s.Name,
				RequirementCount: s.RequirementCount,
			})
		}
	case "decisions":
		for _, dec := range d.decisions {
			item := internal.DecisionListItemJSON{
				Number:       dec.Number,
				Slug:         dec.Slug,
				Title:        dec.Title,
				Status:       string(dec.Status),
				Supersedes:   dec.Supersedes,
				SupersededBy: dec.SupersededBy,
			}
			if !dec.LastModified.IsZero() {
				item.LastModified = dec.LastModified.Format(time.RFC3339)
			}
			out.Decisions = append(out.Decisions, item)
		}
	case "backlog":
		for _, item := range d.backlog {
			out.Backlog = append(out.Backlog, internal.BacklogItemJSON{
				Section: item.Section,
				Title:   item.Title,
			})
		}
	case "changes":
		for _, c := range d.changes {
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

func renderListMinimalJSON(d *listData) error {
	switch d.kind {
	case "specs":
		type minimalSpecs struct {
			Specs []struct {
				Name             string `json:"name"`
				RequirementCount int    `json:"requirementCount"`
			} `json:"specs"`
		}
		min := minimalSpecs{}
		for _, s := range d.specs {
			min.Specs = append(min.Specs, struct {
				Name             string `json:"name"`
				RequirementCount int    `json:"requirementCount"`
			}{Name: s.Name, RequirementCount: s.RequirementCount})
		}
		return printJSON(min)

	case "decisions":
		type minimalDecisions struct {
			Decisions []struct {
				Number int    `json:"number"`
				Slug   string `json:"slug"`
			} `json:"decisions"`
		}
		min := minimalDecisions{}
		for _, dec := range d.decisions {
			min.Decisions = append(min.Decisions, struct {
				Number int    `json:"number"`
				Slug   string `json:"slug"`
			}{Number: dec.Number, Slug: dec.Slug})
		}
		return printJSON(min)

	case "backlog":
		type minimalBacklog struct {
			Backlog []struct {
				Section string `json:"section"`
				Title   string `json:"title"`
			} `json:"backlog"`
		}
		min := minimalBacklog{}
		for _, item := range d.backlog {
			min.Backlog = append(min.Backlog, struct {
				Section string `json:"section"`
				Title   string `json:"title"`
			}{Section: item.Section, Title: item.Title})
		}
		return printJSON(min)

	default: // "changes"
		type minimalChanges struct {
			Changes []struct {
				Name   string `json:"name"`
				Status string `json:"status"`
			} `json:"changes"`
		}
		min := minimalChanges{}
		for _, c := range d.changes {
			min.Changes = append(min.Changes, struct {
				Name   string `json:"name"`
				Status string `json:"status"`
			}{Name: c.Name, Status: internal.ChangeListStatus(c.CompletedTasks, c.TotalTasks)})
		}
		return printJSON(min)
	}
}

func renderListMinimalText(d *listData) error {
	switch d.kind {
	case "specs":
		for _, s := range d.specs {
			fmt.Printf("%s\t%d\n", s.Name, s.RequirementCount)
		}
	case "decisions":
		for _, dec := range d.decisions {
			fmt.Printf("%04d\t%s\n", dec.Number, dec.Slug)
		}
	case "backlog":
		for _, item := range d.backlog {
			fmt.Printf("%s\t%s\n", item.Section, item.Title)
		}
	case "changes":
		for _, c := range d.changes {
			fmt.Printf("%s\t%s\n", c.Name, internal.ChangeListStatus(c.CompletedTasks, c.TotalTasks))
		}
	}
	return nil
}

func renderListText(d *listData) error {
	switch d.kind {
	case "specs":
		fmt.Println("Specs:")
		if len(d.specs) == 0 {
			fmt.Println("  (none)")
		} else {
			fmt.Println()
			maxName := maxNameWidthSpecs(d.specs)
			nameHeaderWidth := max(maxName, 4)
			fmt.Printf("  %-*s  %s\n", nameHeaderWidth, "Name", "Requirements")
			for _, s := range d.specs {
				fmt.Printf("  %-*s  %d\n", nameHeaderWidth, s.Name, s.RequirementCount)
			}
		}

	case "decisions":
		fmt.Println("Decisions:")
		if len(d.decisions) == 0 {
			fmt.Println("  (none)")
			return nil
		}
		fmt.Println()
		for _, dec := range d.decisions {
			fmt.Printf("  %04d  %-30s  %-10s  %s\n", dec.Number, dec.Slug, dec.Status, dec.Title)
		}

	case "backlog":
		fmt.Println("Backlog:")
		if len(d.backlog) == 0 {
			fmt.Println("  (none)")
			return nil
		}
		fmt.Println()
		var currentSection string
		sectionLabels := map[string]string{
			"deferred":       "Deferred",
			"open-questions": "Open Questions",
			"future":         "Future",
			"other":          "Other",
		}
		for _, item := range d.backlog {
			if item.Section != currentSection {
				currentSection = item.Section
				label := sectionLabels[item.Section]
				if label == "" {
					label = item.Section
				}
				fmt.Printf("  %s:\n", label)
			}
			fmt.Printf("    ▪ %s\n", item.Title)
		}

	case "changes":
		fmt.Println("Changes:")
		if len(d.changes) == 0 {
			fmt.Println("  (none)")
		}
		maxName := maxNameWidthChanges(d.changes)
		for _, c := range d.changes {
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
	}
	return nil
}

func printJSON(v any) error {
	data, err := internal.MarshalJSON(v)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
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
