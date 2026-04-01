package internal

import (
	"os"
	"path/filepath"
)

const (
	SpecsDirName   = "specs"
	ChangesDirName = "changes"
	ArchiveDirName = "archive"
	MetaFileName   = ".litespec.yaml"
	ProjectDirName = "specs"
	SkillsDir      = ".agents/skills"
)

var Skills = []SkillInfo{
	{
		ID:          "explore",
		Name:        "litespec-explore",
		Description: "Think about the codebase and potential changes without creating artifacts",
	},
	{
		ID:          "grill",
		Name:        "litespec-grill",
		Description: "Relentlessly interview the user about a plan until reaching shared understanding",
	},
	{
		ID:          "propose",
		Name:        "litespec-propose",
		Description: "Create a change proposal with all planning artifacts (proposal, specs, design, tasks)",
	},
	{
		ID:          "continue",
		Name:        "litespec-continue",
		Description: "Create the next missing artifact for an existing change",
	},
	{
		ID:          "apply",
		Name:        "litespec-apply",
		Description: "Implement the next phase of tasks from a change proposal",
	},
	{
		ID:          "verify",
		Name:        "litespec-verify",
		Description: "Review implemented code against spec requirements",
	},
	{
		ID:          "adopt",
		Name:        "litespec-adopt",
		Description: "Generate a change proposal with specs from existing code",
	},
	{
		ID:          "archive",
		Name:        "litespec-archive",
		Description: "Apply delta operations and complete a change",
	},
}

var Adapters = []ToolAdapter{
	{
		ID:            "claude",
		Name:          "Claude Code",
		SkillsDir:     ".claude/skills",
		CommandsDir:   "",
		FileExtension: "",
		UsesSkillDir:  true,
	},
	{
		ID:            "cursor",
		Name:          "Cursor",
		SkillsDir:     ".cursor/skills",
		CommandsDir:   ".cursor/commands",
		FileExtension: ".md",
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

func SpecsPath(root string) string {
	return filepath.Join(root, ProjectDirName, SpecsDirName)
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
	return filepath.Join(ChangePath(root, name), SpecsDirName)
}
