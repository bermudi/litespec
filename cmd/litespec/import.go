package main

import (
	"fmt"
	"os"

	"github.com/bermudi/litespec/internal/importer"
)

func cmdImport(args []string) error {
	if hasHelpFlag(args) {
		printImportHelp()
		return nil
	}
	if err := checkUnknownFlags(args, map[string]bool{"--dry-run": true, "--source": true, "--force": true}); err != nil {
		return err
	}

	var dryRun bool
	var source string
	var force bool

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--dry-run":
			dryRun = true
		case "--source":
			if i+1 >= len(args) {
				return fmt.Errorf("--source requires a directory path")
			}
			source = args[i+1]
			i++
		case "--force":
			force = true
		}
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
		fmt.Println("Dry run - would import from:", source)
		fmt.Println("  Target:", target)
		fmt.Println()
		stats, err := importer.PreviewImport(source)
		if err != nil {
			return err
		}
		fmt.Printf("  Canon specs: %d\n", stats.CanonSpecs)
		fmt.Printf("  Active changes: %d\n", stats.ActiveChanges)
		fmt.Printf("  Archives: %d\n", stats.Archives)
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

	fmt.Println("Importing from:", source)
	fmt.Println("  Target:", target)
	fmt.Println()

	stats, err := importer.ImportOpenSpecProject(source, target)
	if err != nil {
		return fmt.Errorf("import failed: %w", err)
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
	return nil
}

func printImportHelp() {
	fmt.Print(`Usage: litespec import [options]

Import an OpenSpec project to litespec format.

Options:
  --source <dir>   Source OpenSpec project directory (default: current directory)
  --dry-run        Preview import without making changes
  --force          Overwrite existing files in target

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
`)
}
