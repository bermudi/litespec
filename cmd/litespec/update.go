package main

import (
	"fmt"
	"strings"

	"github.com/bermudi/litespec/internal"
)


func cmdUpdate(args []string) error {
	fs := newFlagSet("update", printUpdateHelp)
	var asJSON, asMinimal bool
	var tools string
	fs.BoolVar(&asJSON, "json", false, "output as JSON")
	fs.BoolVar(&asMinimal, "minimal", false, "minimal output")
	fs.StringVar(&tools, "tools", "", "comma-separated tool IDs (e.g., claude)")

	ok, err := parseFlagSet(fs, args)
	if !ok {
		return err
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
