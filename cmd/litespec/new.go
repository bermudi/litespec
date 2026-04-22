package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bermudi/litespec/internal"
)

func cmdNew(args []string) error {
	if hasHelpFlag(args) {
		printNewHelp()
		return nil
	}
	if err := checkUnknownFlags(args, map[string]bool{"--json": true}); err != nil {
		return err
	}

	var name string
	var asJSON bool
	var positional int
	for _, arg := range args {
		switch arg {
		case jsonFlag:
			asJSON = true
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

	root, err := internal.FindProjectRoot()
	if err != nil {
		return err
	}

	if _, err := os.Stat(filepath.Join(root, internal.ProjectDirName)); err != nil {
		return fmt.Errorf("not a litespec project. Run 'litespec init' first")
	}

	if err := internal.CreateChange(root, name); err != nil {
		return err
	}

	ctx, err := internal.LoadChangeContext(root, name)
	if err != nil {
		return err
	}

	if asJSON {
		status := internal.BuildChangeStatusJSON(ctx)
		data, err := internal.MarshalJSON(status)
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	fmt.Printf("Created: %s\n\n", internal.ChangePath(root, name))
	fmt.Println("Artifacts:")
	for _, art := range internal.Artifacts {
		state := ctx.Artifacts[art.ID]
		var deps string
		if len(art.Requires) > 0 {
			deps = fmt.Sprintf(" (needs: %s)", strings.Join(art.Requires, ", "))
		}
		fmt.Printf("  %-12s %-10s %s%s\n", art.ID, state, art.Filename, deps)
	}
	fmt.Println("\nCreate proposal.md first, then specs/ and design.md, then tasks.md.")
	fmt.Println("Delta specs go in specs/<capability>/spec.md using ADDED/MODIFIED/REMOVED/RENAMED markers.")
	fmt.Println("Use 'litespec instructions <artifact>' for per-artifact guidance.")
	return nil
}
