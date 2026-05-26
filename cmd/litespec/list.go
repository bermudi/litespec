package main

import (
	"fmt"
	"sort"
	"strings"
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

	root, err := requireProjectRoot()
	if err != nil {
		return err
	}

	data, err := fetchListData(root, specsOnly, decisionsOnly, backlogOnly, sortBy, statusFilter)
	if err != nil {
		return err
	}

	return renderList(data, asJSON, asMinimal)
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

func renderList(d *listData, asJSON, asMinimal bool) error {
	return Render(Response{
		Full:        buildListFullJSON(d),
		Minimal:     buildListMinimalJSON(d),
		Text:        buildListText(d),
		MinimalText: buildListMinimalText(d),
	}, asJSON, asMinimal)
}

func buildListFullJSON(d *listData) any {
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
	return out
}

func buildListMinimalJSON(d *listData) any {
	switch d.kind {
	case "specs":
		type minimalSpecs struct {
			Specs []internal.SpecListItemJSON `json:"specs"`
		}
		min := minimalSpecs{}
		for _, s := range d.specs {
			min.Specs = append(min.Specs, internal.SpecListItemJSON{
				Name:             s.Name,
				RequirementCount: s.RequirementCount,
			})
		}
		return min

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
		return min

	case "backlog":
		type minimalBacklog struct {
			Backlog []internal.BacklogItemJSON `json:"backlog"`
		}
		min := minimalBacklog{}
		for _, item := range d.backlog {
			min.Backlog = append(min.Backlog, internal.BacklogItemJSON{
				Section: item.Section,
				Title:   item.Title,
			})
		}
		return min

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
		return min
	}
}

func buildListMinimalText(d *listData) string {
	var b strings.Builder
	switch d.kind {
	case "specs":
		for _, s := range d.specs {
			b.WriteString(fmt.Sprintf("%s\t%d\n", s.Name, s.RequirementCount))
		}
	case "decisions":
		for _, dec := range d.decisions {
			b.WriteString(fmt.Sprintf("%04d\t%s\n", dec.Number, dec.Slug))
		}
	case "backlog":
		for _, item := range d.backlog {
			b.WriteString(fmt.Sprintf("%s\t%s\n", item.Section, item.Title))
		}
	case "changes":
		for _, c := range d.changes {
			b.WriteString(fmt.Sprintf("%s\t%s\n", c.Name, internal.ChangeListStatus(c.CompletedTasks, c.TotalTasks)))
		}
	}
	return b.String()
}

func buildListText(d *listData) string {
	var b strings.Builder
	switch d.kind {
	case "specs":
		b.WriteString("Specs:\n")
		if len(d.specs) == 0 {
			b.WriteString("  (none)\n")
		} else {
			b.WriteString("\n")
			maxName := maxNameWidthSpecs(d.specs)
			nameHeaderWidth := max(maxName, 4)
			b.WriteString(fmt.Sprintf("  %-*s  %s\n", nameHeaderWidth, "Name", "Requirements"))
			for _, s := range d.specs {
				b.WriteString(fmt.Sprintf("  %-*s  %d\n", nameHeaderWidth, s.Name, s.RequirementCount))
			}
		}

	case "decisions":
		b.WriteString("Decisions:\n")
		if len(d.decisions) == 0 {
			b.WriteString("  (none)\n")
			break
		}
		b.WriteString("\n")
		for _, dec := range d.decisions {
			b.WriteString(fmt.Sprintf("  %04d  %-30s  %-10s  %s\n", dec.Number, dec.Slug, dec.Status, dec.Title))
		}

	case "backlog":
		b.WriteString("Backlog:\n")
		if len(d.backlog) == 0 {
			b.WriteString("  (none)\n")
			break
		}
		b.WriteString("\n")
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
				b.WriteString(fmt.Sprintf("  %s:\n", label))
			}
			b.WriteString(fmt.Sprintf("    ▪ %s\n", item.Title))
		}

	case "changes":
		b.WriteString("Changes:\n")
		if len(d.changes) == 0 {
			b.WriteString("  (none)\n")
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
			b.WriteString(fmt.Sprintf("  %-*s  %-16s %-12s %s\n", maxName, c.Name, status, born, relTime))
		}
	}
	return b.String()
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
