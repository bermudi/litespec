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
			return fmt.Errorf("skill %s: template not registered", s.ID)
		}

		skillDir := filepath.Join(skillsDir, s.Name)

		// Clean stale files — only keep files we're about to write
		writtenPaths := map[string]bool{"SKILL.md": true}
		for relPath := range skill.GetResources(s.ID) {
			writtenPaths[relPath] = true
		}
		cleanSkillDir(skillDir, writtenPaths)

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

		// Write reference files
		for relPath, content := range skill.GetResources(s.ID) {
			absPath := filepath.Join(skillDir, relPath)
			if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
				return fmt.Errorf("create resource directory for %s/%s: %w", s.ID, relPath, err)
			}
			if err := os.WriteFile(absPath, []byte(content), 0o644); err != nil {
				return fmt.Errorf("write resource %s/%s: %w", s.ID, relPath, err)
			}
		}
	}

	return nil
}

// cleanSkillDir removes files not in keep. Preserves directories that contain kept files.
func cleanSkillDir(skillDir string, keep map[string]bool) {
	filepath.WalkDir(skillDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(skillDir, path)
		if err != nil {
			return nil
		}
		if !keep[rel] {
			os.Remove(path)
		}
		return nil
	})
}
