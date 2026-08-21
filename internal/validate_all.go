package internal

import (
	"fmt"

	"github.com/bermudi/litespec/v2/internal/skill"
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

	decResult, err := ValidateDecisions(root)
	if err != nil {
		return nil, err
	}
	result.Errors = append(result.Errors, decResult.Errors...)
	result.Warnings = append(result.Warnings, decResult.Warnings...)
	result.DecisionsCount += decResult.DecisionsCount

	queueResult, err := ValidateGHIssueQueues(root)
	if err != nil {
		return nil, err
	}
	result.Errors = append(result.Errors, queueResult.Errors...)
	result.Warnings = append(result.Warnings, queueResult.Warnings...)
	result.UnitsCount += queueResult.UnitsCount

	localQueueResult, err := ValidateLocalQueues(root)
	if err != nil {
		return nil, err
	}
	result.Errors = append(result.Errors, localQueueResult.Errors...)
	result.Warnings = append(result.Warnings, localQueueResult.Warnings...)
	result.UnitsCount += localQueueResult.UnitsCount

	result.Valid = len(result.Errors) == 0
	if strict {
		for _, w := range result.Warnings {
			if !w.StrictExempt {
				result.Valid = false
				break
			}
		}
	}

	return result, nil
}
