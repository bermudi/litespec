package skill_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bermudi/litespec/v2/internal"
)

func TestGeneratedSkillTemplatesDocumentOperations(t *testing.T) {
	root := t.TempDir()
	if err := internal.GenerateSkills(root); err != nil {
		t.Fatalf("GenerateSkills: %v", err)
	}

	t.Run("generated build skill documents red-pre shape and occurrence rule", func(t *testing.T) {
		data := readGeneratedSkillFile(t, root, "litespec-build", "SKILL.md")
		for _, want := range []string{
			"takes its red from the verifier-only commit carrying the new failing tests",
			"expected shape, not a smell",
			"Occurrence counts identical headings within one issue; it is not a unit index; different headings are each occurrence 1.",
		} {
			if !strings.Contains(data, want) {
				t.Errorf("generated litespec-build SKILL.md missing %q", want)
			}
		}
	})

	t.Run("generated plan clear reference states vocabulary and risk entry forms", func(t *testing.T) {
		data := readGeneratedSkillFile(t, root, "litespec-plan", filepath.Join("references", "clear.md"))
		for _, want := range []string{
			"Red-pre shape:",
			"takes exactly one of the closed vocabulary",
			"never mixed forms",
		} {
			if !strings.Contains(data, want) {
				t.Errorf("generated litespec-plan references/clear.md missing %q", want)
			}
		}
	})

	t.Run("generated review skill opens with runtime requirements", func(t *testing.T) {
		data := readGeneratedSkillFile(t, root, "litespec-review", "SKILL.md")
		body, ok := generatedSkillBody(data)
		if !ok {
			t.Fatalf("generated litespec-review SKILL.md has no frontmatter terminator")
		}
		if !strings.HasPrefix(body, "Runtime requirements:") {
			t.Errorf("generated litespec-review SKILL.md does not open with runtime requirements; body starts %q", firstLine(body))
		}
		for _, want := range []string{
			"git worktree support",
			"gh read plus comment access",
			"stops immediately and says so",
		} {
			if !strings.Contains(body, want) {
				t.Errorf("generated litespec-review SKILL.md missing %q", want)
			}
		}
	})
}

func readGeneratedSkillFile(t *testing.T, root, skillName, relPath string) string {
	t.Helper()
	path := filepath.Join(root, internal.SkillsDir, skillName, relPath)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func generatedSkillBody(data string) (string, bool) {
	rest, ok := strings.CutPrefix(data, "---\n")
	if !ok {
		return "", false
	}
	idx := strings.Index(rest, "\n---\n\n")
	if idx < 0 {
		return "", false
	}
	return rest[idx+len("\n---\n\n"):], true
}

func firstLine(s string) string {
	if idx := strings.Index(s, "\n"); idx >= 0 {
		return s[:idx]
	}
	return s
}
