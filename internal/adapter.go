package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bermudi/litespec/internal/skill"
)

func GenerateAdapterCommands(root string, toolIDs []string) error {
	for _, toolID := range toolIDs {
		adapter := GetAdapter(toolID)
		if adapter == nil {
			return fmt.Errorf("unknown tool: %s (supported: %s)", toolID, strings.Join(ValidToolIDs(), ", "))
		}

		skillsDir := filepath.Join(root, adapter.SkillsDir)
		if err := os.MkdirAll(skillsDir, 0o755); err != nil {
			return fmt.Errorf("create %s: %w", adapter.SkillsDir, err)
		}

		for _, si := range Skills {
			tmpl := skill.Get(si.ID)
			if tmpl == "" {
				return fmt.Errorf("skill %s: template not registered for adapter %s", si.ID, toolID)
			}

			linkPath := filepath.Join(skillsDir, si.Name)
			target, err := filepath.Rel(skillsDir, filepath.Join(root, SkillsDir, si.Name))
			if err != nil {
				return fmt.Errorf("resolve symlink target for %s: %w", si.Name, err)
			}
			os.Remove(linkPath)
			if err := os.Symlink(target, linkPath); err != nil {
				return fmt.Errorf("symlink skill %s: %w", si.Name, err)
			}
		}
	}
	return nil
}

func GetAdapter(toolID string) *ToolAdapter {
	for i := range Adapters {
		if Adapters[i].ID == toolID {
			return &Adapters[i]
		}
	}
	return nil
}
