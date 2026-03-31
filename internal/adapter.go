package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func GetAdapter(toolID string) *ToolAdapter {
	for i := range Adapters {
		if Adapters[i].ID == toolID {
			return &Adapters[i]
		}
	}
	return nil
}

func GenerateAdapterCommands(root string, toolIDs []string) error {
	for _, toolID := range toolIDs {
		adapter := GetAdapter(toolID)
		if adapter == nil {
			return fmt.Errorf("unknown tool adapter: %s", toolID)
		}

		commandsDir := filepath.Join(root, adapter.CommandsDir)
		if err := os.MkdirAll(commandsDir, 0o755); err != nil {
			return fmt.Errorf("create commands directory for %s: %w", toolID, err)
		}

		for _, skill := range Skills {
			template := GetSkillTemplate(skill.ID)
			if template == "" {
				continue
			}

			fm := skillFrontmatter{
				Name:        skill.Name,
				Description: skill.Description,
			}

			fmBytes, err := yaml.Marshal(fm)
			if err != nil {
				return fmt.Errorf("marshal frontmatter for %s/%s: %w", toolID, skill.ID, err)
			}

			var sb strings.Builder
			sb.WriteString("---\n")
			sb.Write(fmBytes)
			sb.WriteString("---\n\n")
			sb.WriteString(template)
			sb.WriteString("\n")

			filename := skill.Name + adapter.FileExtension
			cmdFile := filepath.Join(commandsDir, filename)
			if err := os.WriteFile(cmdFile, []byte(sb.String()), 0o644); err != nil {
				return fmt.Errorf("write command file %s/%s: %w", toolID, filename, err)
			}
		}
	}

	return nil
}
