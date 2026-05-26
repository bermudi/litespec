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
	if err := checkUnknownFlags(args, map[string]bool{"--allow-incomplete": true, "--json": true, "--minimal": true}); err != nil {
		return err
	}

	asJSON, asMinimal := parseOutputFlags(args)

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
		if a == jsonFlag || a == minimalFlag {
			continue
		}
		filtered = append(filtered, a)
	}
	name := filtered[0]
	if len(filtered) > 1 {
		return fmt.Errorf("unexpected arguments. Usage: litespec archive <name> [--allow-incomplete]")
	}

	root, err := requireProjectRoot()
	if err != nil {
		return err
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
	if !asJSON && !asMinimal {
		for _, issue := range result.Warnings {
			fmt.Printf("WARN   %s: %s\n", issue.File, issue.Message)
		}
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

	meta, metaErr := internal.ReadChangeMeta(root, name)
	if metaErr == nil && len(meta.DependsOn) > 0 {
		var unarchived []string
		for _, dep := range meta.DependsOn {
			if internal.ChangeExists(root, dep) {
				unarchived = append(unarchived, dep)
			}
		}
		if len(unarchived) > 0 {
			if allowIncomplete {
				fmt.Fprintf(os.Stderr, "WARN  unarchived dependencies: %s\n", strings.Join(unarchived, ", "))
			} else {
				return fmt.Errorf("unarchived dependencies: %s. Archive them first or use --allow-incomplete", strings.Join(unarchived, ", "))
			}
		}
	}

	writes, err := internal.PrepareArchiveWrites(root, name)
	if err != nil {
		return err
	}

	archiveDest, err := internal.ArchiveChange(root, name)
	if err != nil {
		return fmt.Errorf("archiving change: %w", err)
	}

	if err := internal.WritePendingSpecsAtomic(writes); err != nil {
		if restoreErr := internal.RestoreChange(root, archiveDest, name); restoreErr != nil {
			fmt.Fprintf(os.Stderr, "WARNING: failed to restore change after write failure: %v\n", restoreErr)
		}
		return err
	}

	for _, w := range writes {
		if !asJSON && !asMinimal {
			fmt.Printf("Updated spec: %s\n", w.Capability)
		}
	}

	// Build all representations for Render
	type archiveResultJSON struct {
		Change       string   `json:"change"`
		Capabilities []string `json:"capabilities"`
		ArchivedPath string   `json:"archivedPath"`
	}
	type archiveMinimalJSON struct {
		Archived     bool     `json:"archived"`
		Capabilities []string `json:"capabilities"`
	}

	caps := make([]string, len(writes))
	for i, w := range writes {
		caps[i] = w.Capability
	}

	var textSB strings.Builder
	for _, w := range writes {
		textSB.WriteString(fmt.Sprintf("Updated spec: %s\n", w.Capability))
	}
	textSB.WriteString(fmt.Sprintf("Change %q archived — deltas applied, change marked as implemented.\n", name))

	// Strip specs/ subtree from archived directory
	archivedSpecsDir := filepath.Join(archiveDest, internal.ChangeSpecsDirName)
	if err := os.RemoveAll(archivedSpecsDir); err != nil {
		fmt.Fprintf(os.Stderr, "WARN  could not remove specs/ from archived directory: %v\n", err)
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

	return Render(Response{
		Full:        archiveResultJSON{Change: name, Capabilities: caps, ArchivedPath: archiveDest},
		Minimal:     archiveMinimalJSON{Archived: true, Capabilities: caps},
		Text:        textSB.String(),
		MinimalText: "archived",
	}, asJSON, asMinimal)
}
