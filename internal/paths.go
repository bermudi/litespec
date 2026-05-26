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
		Description: "Explore ideas, stress-test plans, and grill unresolved design decisions. Use when the user says 'grill me', 'let's think about', 'explore this', 'stress-test', 'help me decide', or 'what should I do next'. Covers exploration, grilling, and workflow routing modes.",
	},
	{
		ID:          "plan",
		Name:        "litespec-plan",
		Description: "Create or update litespec change proposals, patches, and adopt existing code. Use when the user wants to propose a new change, patch a small fix, adopt existing code into specs, or says 'propose', 'patch', or 'adopt'.",
	},
	{
		ID:          "build",
		Name:        "litespec-build",
		Description: "Implement litespec changes phase by phase, fix review findings, and research knowledge gaps. Use when the user wants to start coding, implement tasks, fix review feedback, research external dependencies, or says 'apply', 'fix', or 'research'.",
	},
	{
		ID:          "review",
		Name:        "litespec-review",
		Description: "Adversarial review of litespec artifacts or implementation. Use when the user wants to review a change, check completeness, stress-test implementation against specs, or says 'review'.",
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
