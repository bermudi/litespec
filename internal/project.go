package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type SpecInfo struct {
	Name             string
	RequirementCount int
}

func InitProject(root string) error {
	specsDir := filepath.Join(root, ProjectDirName)
	if err := os.MkdirAll(specsDir, 0o755); err != nil {
		return fmt.Errorf("create specs directory: %w", err)
	}

	productPath := ProductPath(root)
	if _, err := os.Stat(productPath); os.IsNotExist(err) {
		if writeErr := os.WriteFile(productPath, []byte(productTemplate), 0o644); writeErr != nil {
			return fmt.Errorf("write product.md: %w", writeErr)
		}
	}

	glossaryPath := GlossaryPath(root)
	if _, err := os.Stat(glossaryPath); os.IsNotExist(err) {
		if writeErr := os.WriteFile(glossaryPath, []byte(glossaryTemplate), 0o644); writeErr != nil {
			return fmt.Errorf("write glossary.md: %w", writeErr)
		}
	}

	decisionsDir := DecisionsPath(root)
	if err := os.MkdirAll(decisionsDir, 0o755); err != nil {
		return fmt.Errorf("create decisions directory: %w", err)
	}

	return nil
}

const productTemplate = `# Product

> What we are, what we aren't, how we think.

## Mental Models

- **Model 1**: describe here

## Flows

1. **Flow 1**: describe here
2. **Flow 2**: describe here
`

const glossaryTemplate = "# Glossary\n\nProject-wide ubiquitous language. Curated, optional but recommended.\n\n- **Spec**: A load-bearing contract in specs/<feature>/spec.md with SHALL/MUST and WHEN/THEN scenarios.\n- **Decision**: A durable ruling in specs/decisions/NNNN-slug.md with spine: true for load-bearing.\n- **Unit**: One demo-able outcome per labeled GH issue ## with Done means:, a Verify: that fails without the outcome, and a - [ ] checkbox; optional Read first: (context) and Constraints: (boundaries, never an edit list). Offline, it lives in specs/queues/<name>.md.\n"

func ListSpecs(root string) ([]SpecInfo, error) {
	var result []SpecInfo

	projectSpecsDir := filepath.Join(root, ProjectDirName)
	entries, err := os.ReadDir(projectSpecsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil
		}
		return nil, fmt.Errorf("read specs directory: %w", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if name == ChangesDirName || name == "decisions" {
			continue
		}
		specPath := filepath.Join(projectSpecsDir, name, "spec.md")
		if _, statErr := os.Stat(specPath); statErr != nil {
			continue
		}
		var reqCount int
		data, readErr := os.ReadFile(specPath)
		if readErr == nil {
			spec, parseErr := ParseMainSpec(string(data))
			if parseErr == nil {
				reqCount = len(spec.Requirements)
			}
		}
		result = append(result, SpecInfo{
			Name:             name,
			RequirementCount: reqCount,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result, nil
}
