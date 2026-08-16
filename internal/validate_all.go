package internal

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bermudi/litespec/internal/skill"
)

func ValidateAll(root string, strict bool) (*ValidationResult, error) {
	result := &ValidationResult{Valid: true}

	specResult, err := ValidateSpecs(root)
	if err != nil {
		return nil, err
	}
	result.Errors = append(result.Errors, specResult.Errors...)
	result.Warnings = append(result.Warnings, specResult.Warnings...)
	result.CapabilitiesCount += specResult.CapabilitiesCount
	result.RequirementsCount += specResult.RequirementsCount
	result.ScenariosCount += specResult.ScenariosCount

	changes, err := ListChanges(root)
	if err != nil {
		changes = nil
	}

	if len(changes) > 0 {
		for _, ci := range changes {
			changeResult, err := ValidateChange(root, ci.Name)
			if err != nil {
				return nil, err
			}
			result.Errors = append(result.Errors, changeResult.Errors...)
			result.Warnings = append(result.Warnings, changeResult.Warnings...)
			result.ChangesCount += changeResult.ChangesCount
			result.CapabilitiesCount += changeResult.CapabilitiesCount
			result.RequirementsCount += changeResult.RequirementsCount
			result.ScenariosCount += changeResult.ScenariosCount
		}

		depMap, err := LoadDepMap(root)
		if err == nil {
			cycles := DetectCycles(depMap)
			for _, cycle := range cycles {
				path := strings.Join(cycle, " -> ")
				result.Errors = append(result.Errors, ValidationIssue{
					Severity: SeverityError,
					Message:  fmt.Sprintf("dependency cycle detected: %s", path),
				})
			}

			overlaps := DetectOverlaps(root, changes, depMap)
			result.Warnings = append(result.Warnings, overlaps...)
		}
	}

	skillIDs := make([]string, len(Skills))
	for i, s := range Skills {
		skillIDs[i] = s.ID
	}
	missingTemplates := skill.ValidateSkillTemplates(skillIDs)
	for _, id := range missingTemplates {
		result.Warnings = append(result.Warnings, ValidationIssue{
			Severity: SeverityWarning,
			Message:  fmt.Sprintf("skill %q has no registered template", id),
		})
	}

	queueResult := ValidateQueues(root)
	result.Errors = append(result.Errors, queueResult.Errors...)
	result.Warnings = append(result.Warnings, queueResult.Warnings...)

	decisions, decErr := ListDecisions(root)
	if decErr == nil && len(decisions) > 0 {
		decResult, decErr := ValidateDecisions(root)
		if decErr != nil {
			return nil, decErr
		}
		result.Errors = append(result.Errors, decResult.Errors...)
		result.Warnings = append(result.Warnings, decResult.Warnings...)
		result.DecisionsCount += decResult.DecisionsCount
	}

	backlogResult := ValidateBacklog(root)
	result.Warnings = append(result.Warnings, backlogResult.Warnings...)

	result.Valid = len(result.Errors) == 0
	if strict && len(result.Warnings) > 0 {
		result.Valid = false
	}

	return result, nil
}

func ValidateBacklog(root string) *ValidationResult {
	result := &ValidationResult{Valid: true}

	backlogPath := BacklogPath(root)
	backlog, _ := ParseBacklog(backlogPath)
	if backlog == nil {
		return result
	}

	backlogRel := filepath.Join(ProjectDirName, BacklogFileName)
	for _, section := range backlog.Unrecognized {
		result.Warnings = append(result.Warnings, ValidationIssue{
			Severity: SeverityWarning,
			Message:  fmt.Sprintf("%q is not a recognized section — use ## Deferred, ## Open Questions, ## Future Versions, or ## Other", section),
			File:     backlogRel,
		})
	}

	if len(result.Warnings) > 0 {
		result.Valid = false
	}

	return result
}
