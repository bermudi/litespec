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

	if asJSON {
		instr, err := internal.BuildArtifactInstructionsStandaloneJSON(artifactID)
		if err != nil {
			return err
		}
		if asMinimal {
			type instrMinimalJSON struct {
				ArtifactID  string `json:"artifactId"`
				Instruction string `json:"instruction"`
			}
			min := instrMinimalJSON{
				ArtifactID:  instr.ArtifactID,
				Instruction: instr.Instruction,
			}
			data, err := internal.MarshalJSON(min)
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(data))
		} else {
			data, err := internal.MarshalJSON(instr)
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(data))
		}
		return nil
	}

	instruction := internal.GetSkillTemplate(internal.ArtifactInstructionID(artifactID))
	fmt.Println(instruction)
	return nil
}
