package internal

import (
	"os"
	"path/filepath"
)

const (
	ProjectDirName = "specs"
	SkillsDir      = ".agents/skills"
	ChangesDirName = "changes"
)

var Skills = []SkillInfo{
	{
		ID:          "plan",
		Name:        "litespec-plan",
		Description: "Shape intent into a bounded GH issue (+ spec if load-bearing). Use fuzzy mode for half-baked ideas/questions/research and clear mode to nail the issue. Handles grilling ('grill-me'), codebase design, and glossary. Use when the user wants to plan, shape, explore, grill, or says 'plan', 'shape', 'grill-me', or 'let's think about'.",
	},
	{
		ID:          "build",
		Name:        "litespec-build",
		Description: "Implement one GH issue unit at a time, satisfying Done means and Verify. Use when the user wants to build, implement a unit, fix review findings, or says 'build', 'implement', or 'fix'.",
	},
	{
		ID:          "review",
		Name:        "litespec-review",
		Description: "Adversarial review of GH issue + spec vs implementation. Use when the user wants to review a change, check Verify strength, or says 'review' or 'check this'.",
	},
}

var Adapters = []ToolAdapter{
	{
		ID:        "claude",
		Name:      "Claude Code",
		SkillsDir: ".claude/skills",
	},
}

func FindProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, ProjectDirName)); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return cwd, nil
		}
		dir = parent
	}
}

func ProductPath(root string) string {
	return filepath.Join(root, ProjectDirName, "product.md")
}

func GlossaryPath(root string) string {
	return filepath.Join(root, ProjectDirName, "glossary.md")
}

func FeatureSpecPath(root, feature string) string {
	return filepath.Join(root, ProjectDirName, feature, "spec.md")
}

func ValidToolIDs() []string {
	ids := make([]string, len(Adapters))
	for i, a := range Adapters {
		ids[i] = a.ID
	}
	return ids
}
