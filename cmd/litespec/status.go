package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bermudi/litespec/internal"
)

func cmdStatus(args []string) error {
	if hasHelpFlag(args) {
		printStatusHelp()
		return nil
	}
	if err := checkUnknownFlags(args, map[string]bool{"--json": true}); err != nil {
		return err
	}

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
		return err
	}

	if _, err := os.Stat(filepath.Join(root, internal.ProjectDirName)); err != nil {
		return fmt.Errorf("not a litespec project. Run 'litespec init' first")
	}

	if name != "" {
		if !internal.ChangeExists(root, name) {
			return fmt.Errorf("change %q not found", name)
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

		fmt.Printf("Change: %s\n", name)
		if !ctx.Created.IsZero() {
			fmt.Printf("Created: %s\n", ctx.Created.Format("2006-01-02 15:04:05"))
		}
		fmt.Println()
		for _, art := range internal.Artifacts {
			fmt.Printf("  %-12s %-10s %s\n", art.ID, ctx.Artifacts[art.ID], art.Description)
		}
		return nil
	}

	changes, err := internal.ListChanges(root)
	if err != nil {
		return err
	}

	if asJSON {
		type statusAllOutput struct {
			Changes  []internal.ChangeStatusJSON `json:"changes"`
			Warnings []string                    `json:"warnings,omitempty"`
		}
		var statuses []internal.ChangeStatusJSON
		var warnings []string
		for _, n := range changes {
			ctx, err := internal.LoadChangeContext(root, n.Name)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("error loading change %q: %v", n.Name, err))
				continue
			}
			statuses = append(statuses, internal.BuildChangeStatusJSON(ctx))
		}
		data, err := internal.MarshalJSON(statusAllOutput{Changes: statuses, Warnings: warnings})
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	if len(changes) == 0 {
		fmt.Println("No active changes.")
		return nil
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
	return nil
}
