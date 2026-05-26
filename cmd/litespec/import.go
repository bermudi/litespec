package main

import (
	"fmt"
	"os"

	"github.com/bermudi/litespec/internal"
	"github.com/bermudi/litespec/internal/importer"
)

func cmdImport(args []string) error {
	fs := newFlagSet("import", printImportHelp)
	var asJSON, asMinimal, dryRun, force bool
	var source string
	fs.BoolVar(&asJSON, "json", false, "output as JSON")
	fs.BoolVar(&asMinimal, "minimal", false, "minimal output")
	fs.BoolVar(&dryRun, "dry-run", false, "preview import without making changes")
	fs.StringVar(&source, "source", "", "source OpenSpec project directory (default: current directory)")
	fs.BoolVar(&force, "force", false, "overwrite existing files in target")

	ok, err := parseFlagSet(fs, args)
	if !ok {
		return err
	}

	if source == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory: %w", err)
		}
		source = cwd
	}

	sourceInfo, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("source directory does not exist: %s", source)
	}
	if !sourceInfo.IsDir() {
		return fmt.Errorf("source is not a directory: %s", source)
	}

	if !importer.DetectOpenSpecProject(source) {
		return fmt.Errorf("no OpenSpec project found at %s (expected openspec/specs/ or openspec/changes/)", source)
	}

	target, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	targetSpecs := fmt.Sprintf("%s/specs", target)
	if _, err := os.Stat(targetSpecs); err == nil && !force {
		targetCanon := fmt.Sprintf("%s/specs/canon", target)
		if _, err := os.Stat(targetCanon); err == nil {
			return fmt.Errorf("target directory already has specs/canon. Use --force to overwrite or choose a different target")
		}
	}

	conflicts, err := importer.CheckConflicts(source, target)
	if err != nil {
		return fmt.Errorf("check conflicts: %w", err)
	}

	if len(conflicts) > 0 && !force {
		fmt.Println("Conflicts detected:")
		for _, c := range conflicts {
			fmt.Printf("  %s\n", c)
		}
		fmt.Println("\nUse --force to overwrite existing files or choose a different target directory")
		return fmt.Errorf("conflicts detected (use --force to proceed)")
	}

	if dryRun {
		stats, err := importer.PreviewImport(source)
		if err != nil {
			return err
		}
		if asJSON {
			type importPreviewJSON struct {
				DryRun         bool     `json:"dryRun"`
				CanonSpecs     int      `json:"canonSpecs"`
				ActiveChanges  int      `json:"activeChanges"`
				Archives       int      `json:"archives"`
				Warnings       []string `json:"warnings"`
				SkippedFiles   []string `json:"skippedFiles"`
			}
			out := importPreviewJSON{
				DryRun:         true,
				CanonSpecs:     stats.CanonSpecs,
				ActiveChanges:  stats.ActiveChanges,
				Archives:       stats.Archives,
				Warnings:       stats.Warnings,
				SkippedFiles:   stats.SkippedFiles,
			}
			if asMinimal {
				type importMinimalJSON struct {
					DryRun        bool `json:"dryRun"`
					CanonSpecs    int  `json:"canonSpecs"`
					ActiveChanges int  `json:"activeChanges"`
					Archives      int  `json:"archives"`
				}
				data, err := internal.MarshalJSON(importMinimalJSON{DryRun: true, CanonSpecs: stats.CanonSpecs, ActiveChanges: stats.ActiveChanges, Archives: stats.Archives})
				if err != nil {
					return fmt.Errorf("failed to marshal JSON: %w", err)
				}
				fmt.Println(string(data))
				return nil
			}
			data, err := internal.MarshalJSON(out)
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}
		if asMinimal {
			fmt.Printf("preview\t%d specs\t%d changes\t%d archives\n", stats.CanonSpecs, stats.ActiveChanges, stats.Archives)
			return nil
		}
		fmt.Println("Dry run - would import from:", source)
		fmt.Println("  Target:", target)
		fmt.Println()
		printNameList := func(label string, count int, names []string) {
			if count == 0 {
				fmt.Printf("  %s: 0\n", label)
				return
			}
			fmt.Printf("  %s (%d):\n", label, count)
			for _, n := range names {
				fmt.Printf("    - %s\n", n)
			}
		}
		printNameList("Canon specs", stats.CanonSpecs, stats.CanonSpecNames)
		printNameList("Active changes", stats.ActiveChanges, stats.ActiveChangeNames)
		printNameList("Archives", stats.Archives, stats.ArchiveNames)
		if len(stats.Warnings) > 0 {
			fmt.Println("\nWarnings:")
			for _, w := range stats.Warnings {
				fmt.Printf("  - %s\n", w)
			}
		}
		if len(stats.SkippedFiles) > 0 {
			fmt.Printf("\nSkipped files: %d\n", len(stats.SkippedFiles))
		}
		fmt.Println("\nRun without --dry-run to perform the import")
		return nil
	}

	if !asJSON && !asMinimal {
		fmt.Println("Importing from:", source)
		fmt.Println("  Target:", target)
		fmt.Println()
	}

	stats, err := importer.ImportOpenSpecProject(source, target)
	if err != nil {
		return fmt.Errorf("import failed: %w", err)
	}

	if asJSON {
		type importResultJSON struct {
			CanonSpecs     int      `json:"canonSpecs"`
			ActiveChanges  int      `json:"activeChanges"`
			Archives       int      `json:"archives"`
			Warnings       []string `json:"warnings"`
			SkippedFiles   []string `json:"skippedFiles"`
		}
		if asMinimal {
			type importMinimalJSON struct {
				Imported      bool `json:"imported"`
				CanonSpecs    int  `json:"canonSpecs"`
				ActiveChanges int  `json:"activeChanges"`
				Archives      int  `json:"archives"`
			}
			data, err := internal.MarshalJSON(importMinimalJSON{Imported: true, CanonSpecs: stats.CanonSpecs, ActiveChanges: stats.ActiveChanges, Archives: stats.Archives})
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}
		data, err := internal.MarshalJSON(importResultJSON{
			CanonSpecs:     stats.CanonSpecs,
			ActiveChanges:  stats.ActiveChanges,
			Archives:       stats.Archives,
			Warnings:       stats.Warnings,
			SkippedFiles:   stats.SkippedFiles,
		})
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	if asMinimal {
		fmt.Printf("imported\t%d specs\t%d changes\t%d archives\n", stats.CanonSpecs, stats.ActiveChanges, stats.Archives)
		return nil
	}

	fmt.Printf("✓ Imported %d canon specs\n", stats.CanonSpecs)
	fmt.Printf("✓ Imported %d active changes\n", stats.ActiveChanges)
	fmt.Printf("✓ Imported %d archived changes\n", stats.Archives)

	if len(stats.Warnings) > 0 {
		fmt.Println("\nWarnings:")
		for _, w := range stats.Warnings {
			fmt.Printf("  - %s\n", w)
		}
	}

	if len(stats.SkippedFiles) > 0 {
		fmt.Printf("\nSkipped %d files (no litespec equivalent)\n", len(stats.SkippedFiles))
	}

	fmt.Println("\n✓ Import complete. Run 'litespec update' to generate skills.")
	fmt.Println("  (Import replaces init — no separate initialization needed.)")
	return nil
}

func printImportHelp() {
	fmt.Print(`Usage: litespec import [options]

Import an OpenSpec project to litespec format.

Options:
  --source <dir>   Source OpenSpec project directory (default: current directory)
  --dry-run        Preview import without making changes
  --force          Overwrite existing files in target
  --json           Output as JSON
  --minimal        Minimal output

The command:
  - Detects OpenSpec project structure (openspec/specs/ or openspec/changes/)
  - Moves canon specs to specs/canon/ (strips " Specification" from H1 titles)
  - Moves changes to specs/changes/ (converts .openspec.yaml to .litespec.yaml)
  - Strips specs/ subdirectories from archived changes
  - Synthesizes metadata for archives without .openspec.yaml
  - Normalizes task phase labels (## 1. Name → ## Phase 1: Name)
  - Warns about skipped files (config.yaml, project.md, explorations/, etc.)

Examples:
  litespec import                              Import from current directory
  litespec import --source /path/to/openspec   Import from specific directory
  litespec import --dry-run                    Preview import without changes
  litespec import --force                      Overwrite existing files
  litespec import --json                       Output as JSON
`)
}
