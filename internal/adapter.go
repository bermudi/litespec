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

		if err := cleanStaleSymlinks(skillsDir); err != nil {
			return fmt.Errorf("clean stale symlinks in %s: %w", adapter.SkillsDir, err)
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

func cleanStaleSymlinks(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read adapter dir: %w", err)
	}
	valid := make(map[string]bool, len(Skills))
	for _, si := range Skills {
		valid[si.Name] = true
	}
	for _, entry := range entries {
		if valid[entry.Name()] {
			continue
		}
		if entry.Type()&os.ModeSymlink == 0 {
			continue
		}
		if err := os.Remove(filepath.Join(dir, entry.Name())); err != nil {
			return fmt.Errorf("remove stale symlink %s: %w", entry.Name(), err)
		}
	}
	return nil
}

func DetectActiveAdapters(root string) []string {
	canonicalSkills := filepath.Join(root, SkillsDir)
	var active []string
	for _, adapter := range Adapters {
		adapterDir := filepath.Join(root, adapter.SkillsDir)
		entries, err := os.ReadDir(adapterDir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.Type()&os.ModeSymlink == 0 {
				continue
			}
			linkPath := filepath.Join(adapterDir, entry.Name())
			target, err := os.Readlink(linkPath)
			if err != nil {
				continue
			}
			resolved := target
			if !filepath.IsAbs(resolved) {
				resolved = filepath.Join(adapterDir, resolved)
			}
			rel, err := filepath.Rel(canonicalSkills, resolved)
			if err != nil {
				continue
			}
			if !strings.HasPrefix(rel, "..") {
				active = append(active, adapter.ID)
				break
			}
		}
	}
	return active
}

func GetAdapter(toolID string) *ToolAdapter {
	for i := range Adapters {
		if Adapters[i].ID == toolID {
			return &Adapters[i]
		}
	}
	return nil
}
