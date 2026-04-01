package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func GenerateAdapterCommands(root string, toolIDs []string) error {
	for _, toolID := range toolIDs {
		adapter := GetAdapter(toolID)
		if adapter == nil {
			return fmt.Errorf("unknown tool: %s", toolID)
		}

		if adapter.UsesSkillDir {
			if err := generateSkillsDir(root, adapter); err != nil {
				return err
			}
			continue
		}

		if adapter.CommandsDir != "" {
			if err := generateCommandFiles(root, adapter); err != nil {
				return err
			}
		}
	}
	return nil
}

func generateSkillsDir(root string, adapter *ToolAdapter) error {
	skillsDir := filepath.Join(root, adapter.SkillsDir)
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		return fmt.Errorf("create skills directory: %w", err)
	}

	for _, skill := range Skills {
		template := GetSkillTemplate(skill.ID)
		if template == "" {
			continue
		}

		skillDir := filepath.Join(skillsDir, skill.Name)
		if err := os.MkdirAll(skillDir, 0o755); err != nil {
			return fmt.Errorf("create skill directory %s: %w", skill.Name, err)
		}

		content := buildSkillFile(skill, template)
		skillFile := filepath.Join(skillDir, "SKILL.md")
		if err := os.WriteFile(skillFile, []byte(content), 0o644); err != nil {
			return fmt.Errorf("write skill file %s: %w", skill.ID, err)
		}
	}
	return nil
}

func generateCommandFiles(root string, adapter *ToolAdapter) error {
	commandsDir := filepath.Join(root, adapter.CommandsDir)
	if err := os.MkdirAll(commandsDir, 0o755); err != nil {
		return fmt.Errorf("create commands directory: %w", err)
	}

	for _, skill := range Skills {
		template := GetSkillTemplate(skill.ID)
		if template == "" {
			continue
		}

		content := buildSkillFile(skill, template)
		filename := skill.Name + adapter.FileExtension
		cmdFile := filepath.Join(commandsDir, filename)
		if err := os.WriteFile(cmdFile, []byte(content), 0o644); err != nil {
			return fmt.Errorf("write command file %s/%s: %w", adapter.ID, filename, err)
		}
	}
	return nil
}

func buildSkillFile(skill SkillInfo, template string) string {
	fm := skillFrontmatter{
		Name:        skill.Name,
		Description: skill.Description,
	}
	fmBytes, _ := yaml.Marshal(&fm)

	var sb strings.Builder
	sb.WriteString("---\n")
	sb.Write(fmBytes)
	sb.WriteString("---\n\n")
	sb.WriteString(template)
	sb.WriteString("\n")
	return sb.String()
}

func GetAdapter(toolID string) *ToolAdapter {
	for i := range Adapters {
		if Adapters[i].ID == toolID {
			return &Adapters[i]
		}
	}
	return nil
}
