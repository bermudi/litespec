package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bermudi/litespec/internal/skill"
	"gopkg.in/yaml.v3"
)

type skillFrontmatter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

func GetSkillTemplate(skillID string) string {
	return skill.Get(skillID)
}

func GenerateSkills(root string) error {
	skillsDir := filepath.Join(root, SkillsDir)
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		return fmt.Errorf("create skills directory: %w", err)
	}

	for _, s := range Skills {
		template := skill.Get(s.ID)
		if template == "" {
			continue
		}

		skillDir := filepath.Join(skillsDir, s.Name)
		if err := os.MkdirAll(skillDir, 0o755); err != nil {
			return fmt.Errorf("create skill directory %s: %w", s.Name, err)
		}

		fm := skillFrontmatter{
			Name:        s.Name,
			Description: s.Description,
		}

		fmBytes, err := yaml.Marshal(fm)
		if err != nil {
			return fmt.Errorf("marshal frontmatter for %s: %w", s.ID, err)
		}

		var sb strings.Builder
		sb.WriteString("---\n")
		sb.Write(fmBytes)
		sb.WriteString("---\n\n")
		sb.WriteString(template)
		sb.WriteString("\n")

		skillFile := filepath.Join(skillDir, "SKILL.md")
		if err := os.WriteFile(skillFile, []byte(sb.String()), 0o644); err != nil {
			return fmt.Errorf("write skill file %s: %w", s.ID, err)
		}
	}

	return nil
}
