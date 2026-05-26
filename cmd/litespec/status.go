package main

import (
	"fmt"
	"strings"

	"github.com/bermudi/litespec/internal"
)

func cmdStatus(args []string) error {
	fs := newFlagSet("status", printStatusHelp)
	var asJSON, asMinimal bool
	fs.BoolVar(&asJSON, "json", false, "output as JSON")
	fs.BoolVar(&asMinimal, "minimal", false, "minimal output")

	ok, err := parseFlagSet(fs, args)
	if !ok {
		return err
	}

	positional := fs.Args()
	var name string
	if len(positional) > 0 {
		name = positional[0]
	}

	root, err := requireProjectRoot()
	if err != nil {
		return err
	}

	if name != "" {
		if !internal.ChangeExists(root, name) {
			return fmt.Errorf("change %q not found", name)
		}

		ctx, err := internal.LoadChangeContext(root, name)
		if err != nil {
			return err
		}

		status := internal.BuildChangeStatusJSON(ctx)

		// Build minimal representation
		type statusMinimalJSON struct {
			ChangeName string `json:"changeName"`
			IsComplete bool   `json:"isComplete"`
			Artifacts  []struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			} `json:"artifacts"`
		}
		min := statusMinimalJSON{
			ChangeName: status.ChangeName,
			IsComplete:  status.IsComplete,
		}
		for _, a := range status.Artifacts {
			min.Artifacts = append(min.Artifacts, struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			}{ID: a.ID, Status: a.Status})
		}

		// Build text output
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Change: %s\n", name))
		if !ctx.Created.IsZero() {
			sb.WriteString(fmt.Sprintf("Created: %s\n", ctx.Created.Format("2006-01-02 15:04:05")))
		}
		if ctx.Mode == "patch" {
			sb.WriteString("\n")
			sb.WriteString(fmt.Sprintf("  %-12s %-10s %s\n", "specs", ctx.Artifacts["specs"], "(patch mode)"))
		} else {
			sb.WriteString("\n")
			for _, art := range internal.Artifacts {
				sb.WriteString(fmt.Sprintf("  %-12s %-10s %s\n", art.ID, ctx.Artifacts[art.ID], art.Description))
			}
		}

		return Render(Response{
			Full:        status,
			Minimal:     min,
			Text:        sb.String(),
			MinimalText: fmt.Sprintf("%s\tcomplete=%v", name, status.IsComplete),
		}, asJSON, asMinimal)
	}

	changes, err := internal.ListChanges(root)
	if err != nil {
		return err
	}

	// Check if this is a new project (no active or archived changes)
	isNew := len(changes) == 0
	if isNew {
		archived, _ := internal.ListArchivedChanges(root)
		if len(archived) > 0 {
			isNew = false
		}
	}

	// Load all contexts in one pass — keep contexts alongside statuses
	type loadedChange struct {
		ctx    *internal.Change
		status internal.ChangeStatusJSON
	}
	type statusAllOutput struct {
		Changes        []internal.ChangeStatusJSON `json:"changes"`
		IsNewProject   bool                        `json:"isNewProject"`
		Warnings       []string                    `json:"warnings,omitempty"`
	}
	var loaded []loadedChange
	var warnings []string
	for _, n := range changes {
		ctx, err := internal.LoadChangeContext(root, n.Name)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("error loading change %q: %v", n.Name, err))
			continue
		}
		loaded = append(loaded, loadedChange{
			ctx:    ctx,
			status: internal.BuildChangeStatusJSON(ctx),
		})
	}

	// Collect statuses for JSON output
	var statuses []internal.ChangeStatusJSON
	for _, l := range loaded {
		statuses = append(statuses, l.status)
	}
	allOut := statusAllOutput{Changes: statuses, IsNewProject: isNew, Warnings: warnings}

	// Build minimal all output
	type statusAllMinimal struct {
		Changes      []struct {
			ChangeName string `json:"changeName"`
			IsComplete bool   `json:"isComplete"`
		} `json:"changes"`
		IsNewProject bool `json:"isNewProject"`
	}
	allMin := statusAllMinimal{}
	for _, s := range statuses {
		allMin.Changes = append(allMin.Changes, struct {
			ChangeName string `json:"changeName"`
			IsComplete bool   `json:"isComplete"`
		}{ChangeName: s.ChangeName, IsComplete: s.IsComplete})
	}
	allMin.IsNewProject = isNew

	// Build text and minimal-text from loaded data
	var sb strings.Builder
	var minTextSB strings.Builder
	if len(loaded) == 0 && len(warnings) == 0 {
		sb.WriteString("No active changes.\n")
	} else {
		for _, l := range loaded {
			if l.ctx.Mode == "patch" {
				sb.WriteString(fmt.Sprintf("%s (patch mode)\n", l.status.ChangeName))
				sb.WriteString(fmt.Sprintf("  %-12s %s\n", "specs:", l.ctx.Artifacts["specs"]))
			} else {
				sb.WriteString(fmt.Sprintf("%s\n", l.status.ChangeName))
				for _, art := range internal.Artifacts {
					sb.WriteString(fmt.Sprintf("  %-12s %s\n", art.ID+":", l.ctx.Artifacts[art.ID]))
				}
			}
			minTextSB.WriteString(fmt.Sprintf("%s\tcomplete=%v\n", l.status.ChangeName, l.status.IsComplete))
		}
	}
	minText := strings.TrimRight(minTextSB.String(), "\n")

	return Render(Response{
		Full:        allOut,
		Minimal:     allMin,
		Text:        sb.String(),
		MinimalText: minText,
	}, asJSON, asMinimal)
}
