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
		ID:          "think",
		Name:        "litespec-think",
		Description: "Explore ideas and stress-test plans for litespec changes.",
	},
	{
		ID:          "plan",
		Name:        "litespec-plan",
		Description: "Create or update litespec change proposals and patches.",
	},
	{
		ID:          "build",
		Name:        "litespec-build",
		Description: "Implement litespec changes, fix review findings, and research knowledge gaps.",
	},
	{
		ID:          "review",
		Name:        "litespec-review",
		Description: "Adversarial review of litespec artifacts or implementation.",
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
