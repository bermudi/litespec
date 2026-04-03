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
	if err := checkUnknownFlags(args, map[string]bool{"--json": true}); err != nil {
		return err
	}

	if len(args) == 0 {
		return fmt.Errorf("usage: litespec instructions <artifact> [--json]")
	}

	artifactID := args[0]
	var asJSON bool
	for _, arg := range args[1:] {
		if arg == jsonFlag {
			asJSON = true
		}
	}

	artifactInfo := internal.GetArtifact(artifactID)
	if artifactInfo == nil {
		return fmt.Errorf("unknown artifact: %s (valid: proposal, specs, design, tasks)", artifactID)
	}

	if asJSON {
		instr, err := internal.BuildArtifactInstructionsStandaloneJSON(artifactID)
		if err != nil {
			return err
		}
		data, err := internal.MarshalJSON(instr)
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	instruction := internal.GetSkillTemplate(internal.ArtifactInstructionID(artifactID))
	fmt.Println(instruction)
	return nil
}
