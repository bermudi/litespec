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

func GenerateSkills(root string) error {
	skillsDir := filepath.Join(root, SkillsDir)
	if err := rejectSkillPathSymlinks(skillsDir); err != nil {
		return err
	}
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		return fmt.Errorf("create skills directory: %w", err)
	}

	// Build set of current skill directory names
	activeSkillNames := make(map[string]bool, len(Skills))
	for _, s := range Skills {
		activeSkillNames[s.Name] = true
	}

	// Remove legacy litespec skill directories
	if err := cleanLegacySkillDirs(skillsDir); err != nil {
		return fmt.Errorf("clean legacy skill dirs: %w", err)
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

func rejectSkillPathSymlinks(skillsDir string) error {
	for path := filepath.Clean(skillsDir); ; path = filepath.Dir(path) {
		info, err := os.Lstat(path)
		if err == nil {
			if info.Mode()&os.ModeSymlink != 0 {
				return fmt.Errorf("refusing to generate skills through symlink %s", path)
			}
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("inspect skills path %s: %w", path, err)
		}

		parent := filepath.Dir(path)
		if parent == path {
			break
		}
	}

	info, err := os.Lstat(skillsDir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect skills path %s: %w", skillsDir, err)
	}
	if !info.IsDir() {
		return nil
	}

	err = filepath.WalkDir(skillsDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("inspect skills path %s: %w", path, walkErr)
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("refusing to generate skills through symlink %s", path)
		}
		return nil
	})
	if err != nil {
		return err
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

// cleanLegacySkillDirs removes directories under skillsDir that start with
// the skill name prefix but are not in the active set.
const skillNamePrefix = "litespec-"

func findStaleSkillDirs(skillsDir string) ([]string, error) {
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil, fmt.Errorf("read skills dir: %w", err)
	}

	active := make(map[string]bool, len(Skills))
	for _, s := range Skills {
		active[s.Name] = true
	}

	var stale []string
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), skillNamePrefix) {
			continue
		}
		if !active[entry.Name()] {
			stale = append(stale, entry.Name())
		}
	}
	return stale, nil
}

func CheckStaleSkills(root string) string {
	stale, err := findStaleSkillDirs(filepath.Join(root, SkillsDir))
	if err != nil || len(stale) == 0 {
		return ""
	}
	return fmt.Sprintf("stale skill directories detected: %s. Run 'litespec update' to regenerate.", strings.Join(stale, ", "))
}

func cleanLegacySkillDirs(skillsDir string) error {
	stale, err := findStaleSkillDirs(skillsDir)
	if err != nil {
		return err
	}
	for _, name := range stale {
		if err := os.RemoveAll(filepath.Join(skillsDir, name)); err != nil {
			return fmt.Errorf("remove legacy skill dir %s: %w", name, err)
		}
	}
	return nil
}
