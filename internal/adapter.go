package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bermudi/litespec/internal/skill"
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

	for _, si := range Skills {
		tmpl := skill.Get(si.ID)
		if tmpl == "" {
			continue
		}

		linkPath := filepath.Join(skillsDir, si.Name)

		if adapter.Symlink {
			target, err := filepath.Rel(skillsDir, filepath.Join(root, SkillsDir, si.Name))
			if err != nil {
				return fmt.Errorf("resolve symlink target for %s: %w", si.Name, err)
			}
			os.Remove(linkPath)
			if err := os.Symlink(target, linkPath); err != nil {
				return fmt.Errorf("symlink skill %s: %w", si.Name, err)
			}
			continue
		}

		if err := os.MkdirAll(linkPath, 0o755); err != nil {
			return fmt.Errorf("create skill directory %s: %w", si.Name, err)
		}

		content := buildSkillFile(si, tmpl)
		skillFile := filepath.Join(linkPath, "SKILL.md")
		if err := os.WriteFile(skillFile, []byte(content), 0o644); err != nil {
			return fmt.Errorf("write skill file %s: %w", si.ID, err)
		}
	}
	return nil
}

func generateCommandFiles(root string, adapter *ToolAdapter) error {
	commandsDir := filepath.Join(root, adapter.CommandsDir)
	if err := os.MkdirAll(commandsDir, 0o755); err != nil {
		return fmt.Errorf("create commands directory: %w", err)
	}

	for _, si := range Skills {
		tmpl := skill.Get(si.ID)
		if tmpl == "" {
			continue
		}

		content := buildSkillFile(si, tmpl)
		filename := si.Name + adapter.FileExtension
		cmdFile := filepath.Join(commandsDir, filename)
		if err := os.WriteFile(cmdFile, []byte(content), 0o644); err != nil {
			return fmt.Errorf("write command file %s/%s: %w", adapter.ID, filename, err)
		}
	}
	return nil
}

func buildSkillFile(si SkillInfo, tmpl string) string {
	fm := skillFrontmatter{
		Name:        si.Name,
		Description: si.Description,
	}
	fmBytes, _ := yaml.Marshal(&fm)

	var sb strings.Builder
	sb.WriteString("---\n")
	sb.Write(fmBytes)
	sb.WriteString("---\n\n")
	sb.WriteString(tmpl)
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
