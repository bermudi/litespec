package skill_test

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/bermudi/litespec/internal"
)

func TestGeneratedSkillReferencesExist(t *testing.T) {
	root := t.TempDir()
	if err := internal.GenerateSkills(root); err != nil {
		t.Fatalf("GenerateSkills: %v", err)
	}

	referencePattern := regexp.MustCompile(`references/[A-Za-z0-9._/-]+`)
	for _, skillInfo := range internal.Skills {
		skillFile := filepath.Join(root, internal.SkillsDir, skillInfo.Name, "SKILL.md")
		data, err := os.ReadFile(skillFile)
		if err != nil {
			t.Fatalf("read %s: %v", skillFile, err)
		}

		for _, reference := range referencePattern.FindAllString(string(data), -1) {
			referencePath := filepath.Join(root, internal.SkillsDir, skillInfo.Name, filepath.FromSlash(reference))
			info, err := os.Stat(referencePath)
			if err != nil {
				t.Errorf("%s references missing file %s: %v", skillFile, reference, err)
				continue
			}
			if info.IsDir() {
				t.Errorf("%s references directory %s", skillFile, reference)
			}
		}
	}
}
