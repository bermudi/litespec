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
)

var Skills = []SkillInfo{
	{
		ID:          "explore",
		Name:        "litespec-explore",
		Description: "Enter explore mode - a thinking partner for exploring ideas, investigating problems, and clarifying requirements. Use when the user wants to think through something before or during a change.",
	},
	{
		ID:          "grill",
		Name:        "litespec-grill",
		Description: "Interview the user relentlessly about a plan or design until reaching shared understanding. Use when the user wants to stress-test a plan, get grilled on their design, or mentions \"grill me\".",
	},
	{
		ID:          "propose",
		Name:        "litespec-propose",
		Description: "Materialize a complete change proposal with all planning artifacts (proposal, specs, design, tasks). Use when the user wants to create a new change, start a feature, or says \"propose\".",
	},
	{
		ID:          "continue",
		Name:        "litespec-continue",
		Description: "Create exactly one missing artifact for an existing change, then stop. Use when the user wants to fill in the next missing piece of a change or says \"continue\".",
	},
	{
		ID:          "apply",
		Name:        "litespec-apply",
		Description: "Implement the next phase of tasks from a change proposal, one phase per session. Use when the user is ready to start coding, wants to execute tasks, or says \"apply\".",
	},
	{
		ID:          "verify",
		Name:        "litespec-verify",
		Description: "Context-aware review that adapts to change lifecycle: artifact review (pre-implementation), implementation review (during implementation), and pre-archive review (post-implementation). Use when the user wants to verify artifacts or implementation, check completeness, or says \"verify\".",
	},
	{
		ID:          "adopt",
		Name:        "litespec-adopt",
		Description: "Reverse-engineer specs from existing code. Use when the user provides a file or directory path to document, wants to spec existing code, or says \"adopt\".",
	},
	{
		ID:          "archive",
		Name:        "litespec-archive",
		Description: "Validate and archive a completed change, applying delta operations to merge specs. Use when a change is done and the user wants to finalize it or says \"archive\".",
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
