package main

import (
	"fmt"
	"strings"

	"github.com/bermudi/litespec/internal"
)


func cmdNew(args []string) error {
	if hasHelpFlag(args) {
		printNewHelp()
		return nil
	}
	if err := checkUnknownFlags(args, map[string]bool{"--json": true, "--minimal": true}); err != nil {
		return err
	}

	var name string
	var asJSON, asMinimal bool
	var positional int
	for _, arg := range args {
		switch arg {
		case jsonFlag:
			asJSON = true
		case minimalFlag:
			asMinimal = true
		default:
			if !strings.HasPrefix(arg, "-") {
				positional++
				if positional == 1 {
					name = arg
				}
			}
		}
	}
	if positional > 1 {
		return fmt.Errorf("unexpected arguments. Usage: litespec new <name>")
	}

	if name == "" {
		return fmt.Errorf("usage: litespec new <change-name>")
	}

	if err := validateChangeName(name); err != nil {
		return err
	}

	root, err := requireProjectRoot()
	if err != nil {
		return err
	}

	if err := internal.CreateChange(root, name); err != nil {
		return err
	}

	ctx, err := internal.LoadChangeContext(root, name)
	if err != nil {
		return err
	}

	status := internal.BuildChangeStatusJSON(ctx)

	type newMinimalJSON struct {
		ChangeName string `json:"changeName"`
		IsComplete  bool   `json:"isComplete"`
	}

	var textSB strings.Builder
	textSB.WriteString(fmt.Sprintf("Created: %s\n\n", internal.ChangePath(root, name)))
	textSB.WriteString("Artifacts:\n")
	for _, art := range internal.Artifacts {
		state := ctx.Artifacts[art.ID]
		var deps string
		if len(art.Requires) > 0 {
			deps = fmt.Sprintf(" (needs: %s)", strings.Join(art.Requires, ", "))
		}
		textSB.WriteString(fmt.Sprintf("  %-12s %-10s %s%s\n", art.ID, state, art.Filename, deps))
	}
	textSB.WriteString("\nCreate proposal.md first, then specs/ and design.md, then tasks.md.\n")
	textSB.WriteString("Delta specs go in specs/<capability>/spec.md using ADDED/MODIFIED/REMOVED/RENAMED markers.\n")
	textSB.WriteString("Use 'litespec instructions <artifact>' for per-artifact guidance.\n")

	return Render(Response{
		Full:        status,
		Minimal:     newMinimalJSON{ChangeName: status.ChangeName, IsComplete: status.IsComplete},
		Text:        textSB.String(),
		MinimalText: internal.ChangePath(root, name),
	}, asJSON, asMinimal)
}
