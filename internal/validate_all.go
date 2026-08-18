package internal

import (
	"fmt"

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

	queueResult, err := ValidateGHIssueQueues(root)
	if err != nil {
		return nil, err
	}
	result.Errors = append(result.Errors, queueResult.Errors...)
	result.Warnings = append(result.Warnings, queueResult.Warnings...)
	result.UnitsCount += queueResult.UnitsCount

	result.Valid = len(result.Errors) == 0
	if strict && len(result.Warnings) > 0 {
		result.Valid = false
	}

	return result, nil
}
