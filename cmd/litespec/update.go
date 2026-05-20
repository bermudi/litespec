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

	if _, err := os.Stat(filepath.Join(root, internal.ProjectDirName)); err != nil {
		return fmt.Errorf("not a litespec project. Run 'litespec init' first")
	}

	if err := internal.GenerateSkills(root); err != nil {
		return err
	}
	if !asJSON && !asMinimal {
		fmt.Println("Updated .agents/skills/")
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
		if !asJSON && !asMinimal {
			fmt.Printf("Updated adapter symlinks for: %s\n", strings.Join(toolIDs, ","))
		}
	}

	if asJSON {
		type updateResultJSON struct {
			SkillsUpdated bool     `json:"skillsUpdated"`
			Adapters      []string `json:"adapters"`
		}
		if asMinimal {
			type updateMinimalJSON struct {
				Updated bool `json:"updated"`
			}
			data, err := internal.MarshalJSON(updateMinimalJSON{Updated: true})
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}
		data, err := internal.MarshalJSON(updateResultJSON{SkillsUpdated: true, Adapters: toolIDs})
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	if asMinimal {
		fmt.Println("updated")
		return nil
	}

	return nil
}
