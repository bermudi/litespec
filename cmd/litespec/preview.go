package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bermudi/litespec/internal"
)

func cmdPreview(args []string) error {
	if hasHelpFlag(args) {
		printPreviewHelp()
		return nil
	}
	if err := checkUnknownFlags(args, map[string]bool{"--json": true}); err != nil {
		return err
	}

	if len(args) == 0 {
		return fmt.Errorf("usage: litespec preview <change-name> [--json]")
	}

	useJSON := false
	name := ""
	for _, a := range args {
		if a == jsonFlag {
			useJSON = true
			continue
		}
		if strings.HasPrefix(a, "--") {
			continue
		}
		if name == "" {
			name = a
		} else {
			return fmt.Errorf("unexpected argument %q. Usage: litespec preview <change-name> [--json]", a)
		}
	}

	if name == "" {
		return fmt.Errorf("change name is required. Usage: litespec preview <change-name> [--json]")
	}

	root, err := internal.FindProjectRoot()
	if err != nil {
		return err
	}

	if _, err := os.Stat(filepath.Join(root, internal.ProjectDirName)); err != nil {
		return fmt.Errorf("not a litespec project. Run 'litespec init' first")
	}

	changeDir := internal.ChangePath(root, name)
	if _, err := os.Stat(changeDir); err != nil {
		// Check if archived
		archived, _ := internal.ListArchivedChanges(root)
		for _, a := range archived {
			if a == name {
				return fmt.Errorf("change %q is archived and cannot be previewed", name)
			}
		}
		return fmt.Errorf("change %q not found", name)
	}

	writes, err := internal.PrepareArchiveWrites(root, name)
	if err != nil {
		return err
	}

	result, err := internal.ComputePreviewResult(writes, root)
	if err != nil {
		return err
	}

	if useJSON {
		data, err := internal.FormatPreviewJSON(result)
		if err != nil {
			return fmt.Errorf("formatting JSON: %w", err)
		}
		fmt.Print(string(data))
	} else {
		fmt.Print(internal.FormatPreviewText(result))
	}

	return nil
}

func printPreviewHelp() {
	fmt.Print(`Usage: litespec preview <change-name> [--json]

Preview what archive would do to canonical specs without making changes.

Shows a structural summary of operations per capability:
  + ADDED requirements
  ~ MODIFIED requirements
  - REMOVED requirements
  → RENAMED requirements

Flags:
  --json    Output structured JSON instead of text

Examples:
  litespec preview add-auth
  litespec preview add-auth --json
`)
}
