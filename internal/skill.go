package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

var skillTemplates = map[string]string{
	"explore": `Think deeply about the codebase and potential changes.

Run ` + "`litespec list --json`" + ` to see active changes.

No artifacts are created. Think freely.

Investigate codebase: read files, search code, challenge assumptions.

When insights crystallize, offer to proceed to grill.`,

	"grill": `Interview the user relentlessly about a plan or design until reaching shared understanding.

Resolve each branch of the decision tree, one question at a time.

Provide your recommended answer for each question.

If a question can be answered by exploring the codebase, explore it instead of asking.

When the plan is fully resolved, offer to proceed to propose.`,

	"propose": `Ask the user what they want to build. Derive a kebab-case change name.

Run ` + "`litespec new <name>`" + ` to create the change directory.

Then loop through artifacts in dependency order:

1. Run ` + "`litespec status --change <name> --json`" + ` to get artifact states. Response: {changeName, schemaName, isComplete, artifacts: [{id, outputPath, status, missingDeps}]}
2. For each "ready" artifact, run ` + "`litespec instructions <artifact-id> --change <name> --json`" + ` to get template + context. Response: {changeName, artifactId, changeDir, outputPath, description, instruction, template, dependencies: [{id, done, path}], unlocks}
3. Read dependency files listed in dependencies, create the artifact file using the template structure.

Continue until proposal, specs, design, and tasks are all created.`,

	"continue": `Run ` + "`litespec list --json`" + ` to see changes. If no name given, prompt user to select.

Run ` + "`litespec status --change <name> --json`" + ` to see which artifacts are ready.

Run ` + "`litespec instructions <artifact-id> --change <name> --json`" + ` for the first ready artifact.

Read dependency files, create exactly ONE artifact, then STOP.

Report which artifact was created and what's now unlocked.`,

	"apply": `Run ` + "`litespec status --change <name> --json`" + ` to verify all artifacts are done.

Run ` + "`litespec instructions apply --change <name> --json`" + ` to get context. Response: {changeName, changeDir, contextFiles: {proposal: "path", ...}, progress: {total, complete, remaining}, phases: [{name, tasks: [{id, description, done}], complete, total}], currentPhase, state, instruction}

If state is "blocked", tell user to create missing artifacts first.

Read all contextFiles (proposal.md, design.md, specs/, tasks.md).

Focus on the current phase (currentPhase index in phases array).

Implement each task in that phase sequentially.

After completing each task, mark it [x] in tasks.md.

After completing all tasks in the phase, commit with message: "phase N: <phase name>"

Stop after one phase. User can re-invoke apply for the next phase.`,

	"verify": `Run ` + "`litespec instructions apply --change <name> --json`" + ` to load context files.

Read all artifacts: proposal, specs, design, tasks.

Search the codebase for implementation evidence.

Check three dimensions: Completeness (all tasks done, all spec requirements covered), Correctness (requirement-to-implementation mapping), Coherence (design adherence).

Generate a verification report with CRITICAL/WARNING/SUGGESTION issues.

Produce a summary scorecard.`,

	"adopt": `Given a file or directory path from the user, read and analyze the code.

Run ` + "`litespec new <name>`" + ` to create a change directory.

Generate specs that describe what the code does using ADDED Requirements markers.

Create proposal.md explaining what was adopted and why.

Create design.md documenting the existing architecture discovered.

Run ` + "`litespec status --change <name> --json`" + ` to verify all artifacts.`,

	"archive": `Run ` + "`litespec validate --change <name>`" + ` to verify the change.

Review validation output. If errors exist, fix them before proceeding.

Run ` + "`litespec archive <name>`" + ` to apply delta operations and move to archive.

The CLI handles: RENAMED → REMOVED → MODIFIED → ADDED delta merge, then moves to archive/.

Optionally offer to create a branch and PR for the completed change.`,
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
