package internal

import (
	"encoding/json"
)

type ValidationResultJSON struct {
	Valid    bool                  `json:"valid"`
	Errors   []ValidationIssueJSON `json:"errors"`
	Warnings []ValidationIssueJSON `json:"warnings"`
	Summary  ValidationSummaryJSON `json:"summary"`
}

type ValidationIssueJSON struct {
	Severity string `json:"severity"`
	Message  string `json:"message"`
	File     string `json:"file"`
}

type ValidationSummaryJSON struct {
	Total        int `json:"total"`
	Invalid      int `json:"invalid"`
	Capabilities int `json:"capabilities"`
	Requirements int `json:"requirements"`
	Scenarios    int `json:"scenarios"`
	Decisions    int `json:"decisions,omitempty"`
	Units        int `json:"units,omitempty"`
}

func BuildValidationResultJSON(r *ValidationResult) ValidationResultJSON {
	errors := make([]ValidationIssueJSON, len(r.Errors))
	for i, e := range r.Errors {
		errors[i] = ValidationIssueJSON{Severity: "error", Message: e.Message, File: e.File}
	}
	warnings := make([]ValidationIssueJSON, len(r.Warnings))
	for i, w := range r.Warnings {
		warnings[i] = ValidationIssueJSON{Severity: "warning", Message: w.Message, File: w.File}
	}
	return ValidationResultJSON{
		Valid:    r.Valid,
		Errors:   errors,
		Warnings: warnings,
		Summary: ValidationSummaryJSON{
			Total:        len(errors) + len(warnings),
			Invalid:      len(errors),
			Capabilities: r.CapabilitiesCount,
			Requirements: r.RequirementsCount,
			Scenarios:    r.ScenariosCount,
			Decisions:    r.DecisionsCount,
			Units:        r.UnitsCount,
		},
	}
}

func MarshalJSON(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}
