package internal

import (
	"fmt"
	"os"
	"path/filepath"
)

func ValidateSpec(root, name string) (*ValidationResult, error) {
	result := &ValidationResult{Valid: true}
	specPath := filepath.Join(CanonPath(root), name, "spec.md")

	data, err := os.ReadFile(specPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("spec %q not found", name)
		}
		return nil, fmt.Errorf("read spec %q: %w", name, err)
	}

	spec, parseErr := ParseMainSpec(string(data))
	if parseErr != nil {
		result.Errors = append(result.Errors, ValidationIssue{
			Severity: SeverityError,
			Message:  fmt.Sprintf("invalid spec: %s", parseErr),
			File:     specPath,
		})
		result.Valid = false
		return result, nil
	}

	if len(spec.Requirements) == 0 {
		result.Warnings = append(result.Warnings, ValidationIssue{
			Severity: SeverityWarning,
			Message:  fmt.Sprintf("capability %q has no requirements", name),
			File:     specPath,
		})
	}

	for _, req := range spec.Requirements {
		if len(req.Scenarios) == 0 {
			result.Warnings = append(result.Warnings, ValidationIssue{
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("requirement %q in capability %q has no scenarios", req.Name, name),
				File:     specPath,
			})
		}
		result.ScenariosCount += len(req.Scenarios)
	}

	result.CapabilitiesCount = 1
	result.RequirementsCount = len(spec.Requirements)
	result.Valid = len(result.Errors) == 0
	return result, nil
}

func ValidateSpecs(root string) (*ValidationResult, error) {
	result := &ValidationResult{Valid: true}
	specsDir := CanonPath(root)

	entries, err := os.ReadDir(specsDir)
	if err != nil {
		if os.IsNotExist(err) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationIssue{
				Severity: SeverityError,
				Message:  "specs directory does not exist",
				File:     specsDir,
			})
			return result, nil
		}
		return nil, fmt.Errorf("read specs directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		specPath := filepath.Join(specsDir, entry.Name(), "spec.md")
		data, err := os.ReadFile(specPath)
		if err != nil {
			result.Errors = append(result.Errors, ValidationIssue{
				Severity: SeverityError,
				Message:  fmt.Sprintf("spec file for capability %q not found", entry.Name()),
				File:     specPath,
			})
			continue
		}

		spec, parseErr := ParseMainSpec(string(data))
		if parseErr != nil {
			result.Errors = append(result.Errors, ValidationIssue{
				Severity: SeverityError,
				Message:  fmt.Sprintf("invalid spec: %s", parseErr),
				File:     specPath,
			})
			continue
		}

		result.CapabilitiesCount++
		result.RequirementsCount += len(spec.Requirements)

		if len(spec.Requirements) == 0 {
			result.Warnings = append(result.Warnings, ValidationIssue{
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("capability %q has no requirements", entry.Name()),
				File:     specPath,
			})
		}

		for _, req := range spec.Requirements {
			if len(req.Scenarios) == 0 {
				result.Warnings = append(result.Warnings, ValidationIssue{
					Severity: SeverityWarning,
					Message:  fmt.Sprintf("requirement %q in capability %q has no scenarios", req.Name, entry.Name()),
					File:     specPath,
				})
			}
			result.ScenariosCount += len(req.Scenarios)
		}
	}

	result.Valid = len(result.Errors) == 0
	return result, nil
}
