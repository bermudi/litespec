package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bermudi/litespec/v2/internal/skill"
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

func TestGenerateSkills_RefusesSymlinkWithoutOverwritingTarget(t *testing.T) {
	original := skill.All()
	defer func() {
		resetTemplates()
		for k, v := range original {
			skill.Register(k, v)
		}
	}()

	registerAllTemplates(t)

	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "sentinel")
	const sentinelContent = "do not overwrite\n"
	if err := os.WriteFile(outside, []byte(sentinelContent), 0o644); err != nil {
		t.Fatal(err)
	}

	skillDir := filepath.Join(root, SkillsDir, Skills[0].Name)
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	skillFile := filepath.Join(skillDir, "SKILL.md")
	if err := os.Symlink(outside, skillFile); err != nil {
		t.Fatal(err)
	}

	err := GenerateSkills(root)
	if err == nil {
		t.Fatal("expected GenerateSkills to reject symlink")
	}
	if !strings.Contains(err.Error(), "refusing to generate skills through symlink") {
		t.Errorf("expected clear symlink error, got: %v", err)
	}

	data, err := os.ReadFile(outside)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != sentinelContent {
		t.Fatalf("sentinel changed: got %q, want %q", data, sentinelContent)
	}
}

func TestGenerateSkills_WritesResourceFiles(t *testing.T) {
	original := skill.All()
	originalResources := skill.GetResources("review")
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

	// If review has registered resources, verify they exist
	if originalResources != nil {
		for relPath := range originalResources {
			absPath := filepath.Join(root, SkillsDir, "litespec-review", relPath)
			if _, err := os.Stat(absPath); err != nil {
				t.Errorf("resource file %s: %v", relPath, err)
			}
		}
	}

	// If workflow has registered resources, verify they exist
	workflowResources := skill.GetResources("workflow")
	if workflowResources != nil {
		for relPath := range workflowResources {
			absPath := filepath.Join(root, SkillsDir, "litespec-workflow", relPath)
			if _, err := os.Stat(absPath); err != nil {
				t.Errorf("resource file %s: %v", relPath, err)
			}
		}
	}
}

func TestGenerateSkills_CleansStaleResources(t *testing.T) {
	original := skill.All()
	defer func() {
		resetTemplates()
		for k, v := range original {
			skill.Register(k, v)
		}
	}()

	registerAllTemplates(t)

	root := t.TempDir()

	// Pre-create a stale file in a skill directory
	staleDir := filepath.Join(root, SkillsDir, Skills[0].Name)
	if err := os.MkdirAll(staleDir, 0o755); err != nil {
		t.Fatal(err)
	}
	staleFile := filepath.Join(staleDir, "references", "old.md")
	if err := os.MkdirAll(filepath.Dir(staleFile), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(staleFile, []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := GenerateSkills(root); err != nil {
		t.Fatalf("GenerateSkills: %v", err)
	}

	// Stale file should have been removed
	if _, err := os.Stat(staleFile); !os.IsNotExist(err) {
		t.Error("stale resource file should have been removed")
	}

	// SKILL.md should still exist
	skillFile := filepath.Join(staleDir, "SKILL.md")
	if _, err := os.Stat(skillFile); err != nil {
		t.Error("SKILL.md should still exist")
	}
}

func TestCheckStaleSkills_NoStale(t *testing.T) {
	root := t.TempDir()
	result := CheckStaleSkills(root)
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestCheckStaleSkills_WithStaleDirs(t *testing.T) {
	root := t.TempDir()
	skillsDir := filepath.Join(root, SkillsDir)
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"litespec-explore", "litespec-grill", "litespec-propose"} {
		if err := os.MkdirAll(filepath.Join(skillsDir, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	result := CheckStaleSkills(root)
	if result == "" {
		t.Fatal("expected stale warning, got empty")
	}
	if !strings.Contains(result, "litespec-explore") {
		t.Errorf("expected litespec-explore in warning, got %q", result)
	}
	if !strings.Contains(result, "litespec update") {
		t.Errorf("expected 'litespec update' in warning, got %q", result)
	}
}

func TestCheckStaleSkills_IgnoresNonLitespecDirs(t *testing.T) {
	root := t.TempDir()
	skillsDir := filepath.Join(root, SkillsDir)
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"skill-creator", "the-drill", "research-vision"} {
		if err := os.MkdirAll(filepath.Join(skillsDir, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	result := CheckStaleSkills(root)
	if result != "" {
		t.Errorf("expected empty (no litespec-* stale dirs), got %q", result)
	}
}

func TestCheckStaleSkills_MixedCurrentAndStale(t *testing.T) {
	root := t.TempDir()
	skillsDir := filepath.Join(root, SkillsDir)
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, s := range Skills {
		if err := os.MkdirAll(filepath.Join(skillsDir, s.Name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"litespec-explore", "litespec-grill"} {
		if err := os.MkdirAll(filepath.Join(skillsDir, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	result := CheckStaleSkills(root)
	if result == "" {
		t.Fatal("expected stale warning, got empty")
	}
	for _, s := range Skills {
		if strings.Contains(result, s.Name) {
			t.Errorf("current skill %q should not appear in warning, got: %s", s.Name, result)
		}
	}
	if !strings.Contains(result, "litespec-explore") {
		t.Errorf("expected litespec-explore in warning, got %q", result)
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

func TestGeneratedSkillsUseRedGreenEvidence(t *testing.T) {
	root := t.TempDir()
	if err := GenerateSkills(root); err != nil {
		t.Fatalf("GenerateSkills: %v", err)
	}

	readFile := func(path string) string {
		t.Helper()
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		return string(data)
	}
	requireContains := func(name, content string, phrases ...string) {
		t.Helper()
		for _, phrase := range phrases {
			if !strings.Contains(content, phrase) {
				t.Errorf("%s missing %q", name, phrase)
			}
		}
	}

	build := readFile(filepath.Join(root, SkillsDir, "litespec-build", "SKILL.md"))
	requireContains(t.Name()+"/build", build,
		"Run the exact `Verify:` command on the clean starting commit before implementation.",
		"one verifier-only commit",
		"fails because the unit outcome is absent",
		"pre sha:",
		"pre exit status:",
		"Pre-evidence scope:",
		"post sha:",
		"post exit status: 0",
		"Post-evidence scope:",
		"Never amend either recorded evidence commit.",
	)

	review := readFile(filepath.Join(root, SkillsDir, "litespec-review", "SKILL.md"))
	requireContains(t.Name()+"/review", review,
		"pre is an ancestor of post and post is an ancestor of `HEAD`",
		"detached temporary Git worktree at pre",
		"detached temporary Git worktree at post",
		"Run the exact `Verify:` command again at `HEAD`",
		"must fail because the outcome is absent",
		"must exit 0 with the outcome present",
		"never check out an evidence SHA in the reviewer's current worktree",
		"does not prove that Verify targets the correct behavior",
	)

	for _, path := range []string{"../AGENTS.md", "../DESIGN.md", "../docs/workflow.md"} {
		content := readFile(path)
		requireContains(path, content,
			"pre",
			"post",
			"verifier-only commit",
			"detached temporary worktree",
			"does not prove",
		)
	}
}
