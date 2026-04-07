package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bermudi/litespec/internal"
)

func cmdUpdate(args []string) error {
	if hasHelpFlag(args) {
		printUpdateHelp()
		return nil
	}
	if err := checkUnknownFlags(args, map[string]bool{"--tools": true}); err != nil {
		return err
	}

	var tools string
	for i := 0; i < len(args); i++ {
		if args[i] == "--tools" && i+1 < len(args) {
			tools = args[i+1]
			i++
		}
	}

	root, err := internal.FindProjectRoot()
	if err != nil {
		return err
	}

	if _, err := os.Stat(filepath.Join(root, internal.ProjectDirName)); err != nil {
		return fmt.Errorf("not a litespec project. Run 'litespec init' first")
	}

	if err := internal.GenerateSkills(root); err != nil {
		return err
	}
	fmt.Println("Updated .agents/skills/")

	var toolIDs []string
	if tools != "" {
		toolIDs = splitCSV(tools)
		if err := validateToolIDs(toolIDs); err != nil {
			return err
		}
	} else {
		toolIDs = internal.DetectActiveAdapters(root)
	}

	if len(toolIDs) > 0 {
		if err := internal.GenerateAdapterCommands(root, toolIDs); err != nil {
			return err
		}
		fmt.Printf("Updated adapter symlinks for: %s\n", strings.Join(toolIDs, ","))
	}
	return nil
}
