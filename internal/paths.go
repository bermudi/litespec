package internal

import (
	"os"
	"path/filepath"
	"regexp"
)

const (
	CanonDirName       = "canon"
	ChangeSpecsDirName = "specs"
	ChangesDirName     = "changes"
	ArchiveDirName     = "archive"
	MetaFileName       = ".litespec.yaml"
	ProjectDirName     = "specs"
	SkillsDir          = ".agents/skills"
	BacklogFileName    = "backlog.md"
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

func CanonPath(root string) string {
	return filepath.Join(root, ProjectDirName, CanonDirName)
}

func ChangesPath(root string) string {
	return filepath.Join(root, ProjectDirName, ChangesDirName)
}

func ArchivePath(root string) string {
	return filepath.Join(root, ProjectDirName, ChangesDirName, ArchiveDirName)
}

func ChangePath(root, name string) string {
	return filepath.Join(root, ProjectDirName, ChangesDirName, name)
}

func ChangeSpecsPath(root, name string) string {
	return filepath.Join(ChangePath(root, name), ChangeSpecsDirName)
}

func ValidToolIDs() []string {
	ids := make([]string, len(Adapters))
	for i, a := range Adapters {
		ids[i] = a.ID
	}
	return ids
}

var ArchivedNameRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}-(.+)$`)

func ParseArchivedName(name string) string {
	m := ArchivedNameRe.FindStringSubmatch(name)
	if len(m) == 2 {
		return m[1]
	}
	return name
}
