package main

import (
	"fmt"
	"os"

	"github.com/bermudi/litespec/internal"
)

func cmdPreview(args []string) error {
	fs := newFlagSet("preview", printPreviewHelp)
	var asJSON, asMinimal bool
	fs.BoolVar(&asJSON, "json", false, "output as JSON")
	fs.BoolVar(&asMinimal, "minimal", false, "minimal output")

	ok, err := parseFlagSet(fs, args)
	if !ok {
		return err
	}

	positional := fs.Args()
	if len(positional) == 0 {
		return fmt.Errorf("change name is required. Usage: litespec preview <change-name> [--json]")
	}
	name := positional[0]
	if len(positional) > 1 {
		return fmt.Errorf("unexpected argument %q. Usage: litespec preview <change-name> [--json]", positional[1])
	}

	root, err := requireProjectRoot()
	if err != nil {
		return err
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

	// Build JSON representations
	type previewMinimalJSON struct {
		Totals struct {
			Capabilities      int `json:"capabilities"`
			Added             int `json:"added"`
			Modified          int `json:"modified"`
			Removed           int `json:"removed"`
			Renamed           int `json:"renamed"`
		} `json:"totals"`
	}
	min := previewMinimalJSON{}
	min.Totals.Capabilities = result.Totals.Capabilities
	min.Totals.Added = result.Totals.Added
	min.Totals.Modified = result.Totals.Modified
	min.Totals.Removed = result.Totals.Removed
	min.Totals.Renamed = result.Totals.Renamed

	minimalText := fmt.Sprintf("%d capabilities\t%d added\t%d modified\t%d removed\t%d renamed",
		result.Totals.Capabilities, result.Totals.Added, result.Totals.Modified, result.Totals.Removed, result.Totals.Renamed)

	return Render(Response{
		Full:        result,
		Minimal:     min,
		Text:        internal.FormatPreviewText(result),
		MinimalText: minimalText,
	}, asJSON, asMinimal)
}

func printPreviewHelp() {
	fmt.Print(`Usage: litespec preview <change-name> [--json] [--minimal]

Preview what archive would do to canonical specs without making changes.

Shows a structural summary of operations per capability:
  + ADDED requirements
  ~ MODIFIED requirements
  - REMOVED requirements
  → RENAMED requirements

Flags:
  --json    Output structured JSON instead of text
  --minimal Minimal output

Examples:
  litespec preview add-auth
  litespec preview add-auth --json
`)
}
