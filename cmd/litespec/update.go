package main

import (
	"fmt"
	"strings"

	"github.com/bermudi/litespec/internal"
)


func cmdUpdate(args []string) error {
	if hasHelpFlag(args) {
		printUpdateHelp()
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

	root, err := requireProjectRoot()
	if err != nil {
		return err
	}

	if err := internal.GenerateSkills(root); err != nil {
		return err
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

	if len(toolIDs) > 0 {
		if err := internal.GenerateAdapterCommands(root, toolIDs); err != nil {
			return err
		}
	}

	type updateResultJSON struct {
		SkillsUpdated bool     `json:"skillsUpdated"`
		Adapters      []string `json:"adapters"`
	}
	type updateMinimalJSON struct {
		Updated bool `json:"updated"`
	}

	var textSB strings.Builder
	textSB.WriteString("Updated .agents/skills/\n")
	if len(toolIDs) > 0 {
		textSB.WriteString(fmt.Sprintf("Updated adapter symlinks for: %s\n", strings.Join(toolIDs, ",")))
	}

	return Render(Response{
		Full:        updateResultJSON{SkillsUpdated: true, Adapters: toolIDs},
		Minimal:     updateMinimalJSON{Updated: true},
		Text:        textSB.String(),
		MinimalText: "updated",
	}, asJSON, asMinimal)
}
