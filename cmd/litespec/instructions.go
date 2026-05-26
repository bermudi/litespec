package main

import (
	"fmt"

	"github.com/bermudi/litespec/internal"
)

func cmdInstructions(args []string) error {
	if hasHelpFlag(args) {
		printInstructionsHelp()
		return nil
	}
	if err := checkUnknownFlags(args, map[string]bool{"--json": true, "--minimal": true}); err != nil {
		return err
	}

	if len(args) == 0 {
		return fmt.Errorf("usage: litespec instructions <artifact> [--json]")
	}

	artifactID := args[0]
	var asJSON, asMinimal bool
	for _, arg := range args[1:] {
		switch arg {
		case jsonFlag:
			asJSON = true
		case minimalFlag:
			asMinimal = true
		}
	}

	artifactInfo := internal.GetArtifact(artifactID)
	if artifactInfo == nil {
		return fmt.Errorf("unknown artifact: %s (valid: proposal, specs, design, tasks)", artifactID)
	}

	type instrMinimalJSON struct {
		ArtifactID  string `json:"artifactId"`
		Instruction string `json:"instruction"`
	}

	full, err := internal.BuildArtifactInstructionsStandaloneJSON(artifactID)
	if err != nil {
		return err
	}
	instruction := internal.GetSkillTemplate(internal.ArtifactInstructionID(artifactID))

	return Render(Response{
		Full:    full,
		Minimal: instrMinimalJSON{ArtifactID: full.ArtifactID, Instruction: full.Instruction},
		Text:    instruction + "\n",
	}, asJSON, asMinimal)
}
