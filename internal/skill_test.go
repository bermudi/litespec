package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bermudi/litespec/internal/skill"
)

func registerAllTemplates(t *testing.T) {
	t.Helper()
	for _, s := range Skills {
		skill.Register(s.ID, "template content for "+s.ID)
	}
}

func resetTemplates() {
	for k := range skill.All() {
		delete(skill.All(), k)
	}
}

func TestGenerateSkills_CreatesAllSkillFiles(t *testing.T) {
	original := skill.All()
	defer func() {
		resetTemplates()
		for k, v := range original {
			skill.Register(k, v)
		}
	}()

	registerAllTemplates(t)

	root := t.TempDir()
	if err := GenerateSkills(root); err != nil {
		t.Fatalf("GenerateSkills: %v", err)
	}

	for _, s := range Skills {
		skillFile := filepath.Join(root, SkillsDir, s.Name, "SKILL.md")
		data, err := os.ReadFile(skillFile)
		if err != nil {
			t.Errorf("skill %s: reading SKILL.md: %v", s.Name, err)
			continue
		}

		content := string(data)

		if !strings.HasPrefix(content, "---\n") {
			t.Errorf("skill %s: missing opening frontmatter marker", s.Name)
		}
		if !strings.Contains(content, "\n---\n") {
			t.Errorf("skill %s: missing closing frontmatter marker", s.Name)
		}
		if !strings.Contains(content, s.Name) {
			t.Errorf("skill %s: file does not contain skill name", s.Name)
		}
		if !strings.Contains(content, "template content for "+s.ID) {
			t.Errorf("skill %s: file does not contain template content", s.Name)
		}
	}
}

func TestGenerateSkills_FrontmatterFormat(t *testing.T) {
	original := skill.All()
	defer func() {
		resetTemplates()
		for k, v := range original {
			skill.Register(k, v)
		}
	}()

	registerAllTemplates(t)

	root := t.TempDir()
	if err := GenerateSkills(root); err != nil {
		t.Fatalf("GenerateSkills: %v", err)
	}

	first := Skills[0]
	skillFile := filepath.Join(root, SkillsDir, first.Name, "SKILL.md")
	data, err := os.ReadFile(skillFile)
	if err != nil {
		t.Fatalf("reading SKILL.md: %v", err)
	}

	content := string(data)

	if !strings.HasPrefix(content, "---\n") {
		t.Fatal("missing opening frontmatter marker")
	}

	closingIdx := strings.Index(content[4:], "\n---\n")
	if closingIdx < 0 {
		t.Fatal("missing closing frontmatter marker")
	}

	fm := content[4 : closingIdx+4]

	if !strings.Contains(fm, "name: "+first.Name) {
		t.Errorf("frontmatter missing 'name:' key, got:\n%s", fm)
	}
	if !strings.Contains(fm, "description: ") {
		t.Errorf("frontmatter missing 'description:' key, got:\n%s", fm)
	}
}

func TestGenerateSkills_MissingTemplate(t *testing.T) {
	original := skill.All()
	defer func() {
		resetTemplates()
		for k, v := range original {
			skill.Register(k, v)
		}
	}()

	resetTemplates()

	root := t.TempDir()
	err := GenerateSkills(root)
	if err == nil {
		t.Fatal("expected error when templates are missing")
	}
	if !strings.Contains(err.Error(), "template not registered") {
		t.Errorf("expected 'template not registered' in error, got: %v", err)
	}
}

func TestGenerateSkills_ReadonlyDir(t *testing.T) {
	original := skill.All()
	defer func() {
		resetTemplates()
		for k, v := range original {
			skill.Register(k, v)
		}
	}()

	registerAllTemplates(t)

	root := t.TempDir()
	readonlyDir := filepath.Join(root, "readonly")
	if err := os.MkdirAll(readonlyDir, 0o555); err != nil {
		t.Fatal(err)
	}

	err := GenerateSkills(readonlyDir)
	if err == nil {
		t.Fatal("expected error for read-only root directory")
	}
}
