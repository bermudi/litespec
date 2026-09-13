package main

import (
	"fmt"
	"strings"

	"github.com/bermudi/litespec/v2/internal"
)

func cmdInit(args []string) error {
	fs := newFlagSet("init", printInitHelp)
	var asJSON, asMinimal bool
	var tools string
	fs.BoolVar(&asJSON, "json", false, "output as JSON")
	fs.BoolVar(&asMinimal, "minimal", false, "minimal output")
	fs.StringVar(&tools, "tools", "", "comma-separated tool IDs (e.g., claude)")

	ok, err := parseFlagSet(fs, args)
	if !ok {
		return err
	}

	var toolIDs []string
	if tools != "" {
		toolIDs = splitCSV(tools)
		if err := validateToolIDs(toolIDs); err != nil {
			return err
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

	if tools == "" {
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
			Directories: []string{"specs/product.md", "specs/glossary.md", "specs/decisions/"},
			Skills:      skillNames,
			Adapters:    toolIDs,
		},
		Minimal:     initMinimalJSON{Initialized: true},
		Text:        "Project initialized.\n\nGH issue is the queue — proposal + design + queue live in the GH issue body. Run `litespec view` to see product + specs + GH issues.\n",
		MinimalText: "initialized",
	}, asJSON, asMinimal)
}
