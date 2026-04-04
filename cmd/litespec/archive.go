package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bermudi/litespec/internal"
)

func cmdArchive(args []string) error {
	if hasHelpFlag(args) {
		printArchiveHelp()
		return nil
	}
	if err := checkUnknownFlags(args, map[string]bool{"--allow-incomplete": true}); err != nil {
		return err
	}

	if len(args) == 0 {
		return fmt.Errorf("usage: litespec archive <change-name> [--allow-incomplete]")
	}

	allowIncomplete := false
	filtered := args[:0]
	for _, a := range args {
		if a == "--allow-incomplete" {
			allowIncomplete = true
			continue
		}
		filtered = append(filtered, a)
	}
	name := filtered[0]
	if len(filtered) > 1 {
		return fmt.Errorf("unexpected arguments. Usage: litespec archive <name> [--allow-incomplete]")
	}

	root, err := internal.FindProjectRoot()
	if err != nil {
		return err
	}

	if _, err := os.Stat(filepath.Join(root, internal.ProjectDirName)); err != nil {
		return fmt.Errorf("not a litespec project. Run 'litespec init' first")
	}

	result, err := internal.ValidateChange(root, name)
	if err != nil {
		return err
	}
	if !result.Valid {
		for _, issue := range result.Errors {
			fmt.Fprintf(os.Stderr, "ERROR  %s: %s\n", issue.File, issue.Message)
		}
		return fmt.Errorf("validation failed. Fix errors before archiving")
	}
	for _, issue := range result.Warnings {
		fmt.Printf("WARN   %s: %s\n", issue.File, issue.Message)
	}

	if !allowIncomplete {
		tasksPath := filepath.Join(internal.ChangePath(root, name), "tasks.md")
		tasksData, tasksErr := os.ReadFile(tasksPath)
		if tasksErr == nil {
			completed, total := internal.TaskCompletion(string(tasksData))
			if completed < total {
				return fmt.Errorf("%d/%d tasks completed. Finish tasks or use --allow-incomplete", completed, total)
			}
		}
	}

	dependents, depErr := internal.GetDependents(root, name)
	if depErr != nil {
		return fmt.Errorf("checking dependents: %w", depErr)
	}
	if len(dependents) > 0 {
		if allowIncomplete {
			fmt.Fprintf(os.Stderr, "WARN  active changes depend on %q: %s\n", name, strings.Join(dependents, ", "))
		} else {
			return fmt.Errorf("active changes depend on %q: %s. Archive them first or use --allow-incomplete", name, strings.Join(dependents, ", "))
		}
	}

	writes, err := internal.PrepareArchiveWrites(root, name)
	if err != nil {
		return err
	}

	if err := internal.WritePendingSpecsAtomic(writes); err != nil {
		return err
	}

	for _, w := range writes {
		fmt.Printf("Updated spec: %s\n", w.Capability)
	}

	if err := internal.ArchiveChange(root, name); err != nil {
		return fmt.Errorf("archiving change: %w", err)
	}

	archiveEntries, archiveErr := os.ReadDir(internal.ArchivePath(root))
	if archiveErr != nil {
		return fmt.Errorf("post-archive verification failed: %w", archiveErr)
	}
	archiveFound := false
	for _, e := range archiveEntries {
		if strings.HasSuffix(e.Name(), "-"+name) {
			archiveFound = true
			break
		}
	}
	if !archiveFound {
		return fmt.Errorf("post-archive verification failed: archived directory for %q not found", name)
	}

	for _, w := range writes {
		data, readErr := os.ReadFile(w.Path)
		if readErr != nil {
			return fmt.Errorf("post-archive verification failed: cannot read spec %s: %w", w.Capability, readErr)
		}
		if _, parseErr := internal.ParseMainSpec(string(data)); parseErr != nil {
			return fmt.Errorf("post-archive verification failed: spec %s failed to parse: %w", w.Capability, parseErr)
		}
	}

	fmt.Printf("Change %q archived successfully.\n", name)
	return nil
}
