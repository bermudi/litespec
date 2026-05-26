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
	if err := checkUnknownFlags(args, map[string]bool{"--tools": true, "--json": true, "--minimal": true}); err != nil {
		return err
	}

	asJSON, asMinimal := parseOutputFlags(args)

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
	// Suppress text output during side-effect operations when JSON/minimal mode is active
	if !asJSON && !asMinimal {
		fmt.Println("Created specs/ directory structure")
	}

	if err := internal.GenerateSkills(root); err != nil {
		return err
	}
	if !asJSON && !asMinimal {
		fmt.Println("Generated .agents/skills/")
	}

	var toolIDs []string
	if tools != "" {
		toolIDs = splitCSV(tools)
		if err := validateToolIDs(toolIDs); err != nil {
			return err
		}
	} else {
		toolIDs = internal.DetectActiveAdapters(root)
	}

	// Suppress adapter text output during side-effect operations when JSON/minimal mode is active
	if len(toolIDs) > 0 {
		if err := internal.GenerateAdapterCommands(root, toolIDs); err != nil {
			return err
		}
		if !asJSON && !asMinimal {
			fmt.Printf("Generated adapter commands for: %s\n", strings.Join(toolIDs, ","))
		}
	}

	type initResultJSON struct {
		Initialized bool     `json:"initialized"`
		Directories []string `json:"directories"`
		Skills      []string `json:"skills"`
		Adapters    []string `json:"adapters"`
	}
	type initMinimalJSON struct {
		Initialized bool `json:"initialized"`
	}

	skillNames := make([]string, len(internal.Skills))
	for i, s := range internal.Skills {
		skillNames[i] = s.ID
	}

	return Render(Response{
		Full: initResultJSON{
			Initialized: true,
			Directories: []string{"specs/canon/", "specs/changes/"},
			Skills:      skillNames,
			Adapters:    toolIDs,
		},
		Minimal:     initMinimalJSON{Initialized: true},
		Text:        "Project initialized.\n\nTip: Create specs/backlog.md with ## Deferred, ## Open Questions, and ## Future Versions sections to surface backlog counts in `litespec view`.\n",
		MinimalText: "initialized",
	}, asJSON, asMinimal)
}
