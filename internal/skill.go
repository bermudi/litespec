package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

var skillTemplates = map[string]string{
	"explore":  "Explore the codebase and discuss potential changes. Use `litespec status --json` to understand current state. No artifacts are created. Think freely about what could change and why.",
	"grill":    "Interview the user relentlessly about a plan or design until reaching shared understanding. Resolve each branch of the decision tree. Ask one question at a time. Provide your recommended answer for each question.",
	"propose":  "Create a change proposal with all planning artifacts. Use `litespec status --change <name> --json` to check state, then `litespec instructions <artifact> --change <name> --json` for guidance. Create proposal.md, specs/, design.md, and tasks.md in the change directory.",
	"continue": "Create the next missing artifact for a change. Use `litespec status --change <name> --json` to see what's ready. Use `litespec instructions <artifact> --change <name> --json` for guidance. Create only the next artifact.",
	"apply":    "Implement the next phase of tasks from a change. Read tasks.md to find the current phase (first phase with unchecked tasks). Implement each task in that phase sequentially, marking them complete with [x]. Use `litespec status --change <name> --json` to check progress.",
	"verify":   "Review implemented code against spec requirements. Read the change's specs/ directory for requirements. Examine the relevant code files. Report gaps between spec and implementation. Do NOT run tests or lint.",
	"adopt":    "Generate a change proposal from existing code. Given a file or directory path, read and analyze the code. Create a change directory with specs that describe what the code does. Use `litespec status --change <name> --json` and `litespec instructions` for guidance.",
	"archive":  "Complete a change by applying delta operations. Use `litespec validate --change <name>` to verify first. Then use `litespec archive <name>` to apply deltas and move to archive.",
}

type skillFrontmatter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

func GetSkillTemplate(skillID string) string {
	return skillTemplates[skillID]
}

func GenerateSkills(root string) error {
	skillsDir := filepath.Join(root, SkillsDir)
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		return fmt.Errorf("create skills directory: %w", err)
	}

	for _, skill := range Skills {
		template := GetSkillTemplate(skill.ID)
		if template == "" {
			continue
		}

		skillDir := filepath.Join(skillsDir, skill.Name)
		if err := os.MkdirAll(skillDir, 0o755); err != nil {
			return fmt.Errorf("create skill directory %s: %w", skill.Name, err)
		}

		fm := skillFrontmatter{
			Name:        skill.Name,
			Description: skill.Description,
		}

		fmBytes, err := yaml.Marshal(fm)
		if err != nil {
			return fmt.Errorf("marshal frontmatter for %s: %w", skill.ID, err)
		}

		var sb strings.Builder
		sb.WriteString("---\n")
		sb.Write(fmBytes)
		sb.WriteString("---\n\n")
		sb.WriteString(template)
		sb.WriteString("\n")

		skillFile := filepath.Join(skillDir, "SKILL.md")
		if err := os.WriteFile(skillFile, []byte(sb.String()), 0o644); err != nil {
			return fmt.Errorf("write skill file %s: %w", skill.ID, err)
		}
	}

	return nil
}
