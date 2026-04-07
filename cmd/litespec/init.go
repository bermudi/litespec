package main

import (
	"fmt"
	"strings"

	"github.com/bermudi/litespec/internal"
)

func cmdInit(args []string) error {
	if hasHelpFlag(args) {
		printInitHelp()
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

	if err := internal.InitProject(root); err != nil {
		return err
	}
	fmt.Println("Created specs/ directory structure")

	if err := internal.GenerateSkills(root); err != nil {
		return err
	}
	fmt.Println("Generated .agents/skills/")

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
		fmt.Printf("Generated adapter commands for: %s\n", strings.Join(toolIDs, ","))
	}

	fmt.Println("Project initialized.")
	return nil
}
