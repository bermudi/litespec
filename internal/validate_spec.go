package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ValidateSpec(root, name string) (*ValidationResult, error) {
	result := &ValidationResult{Valid: true}
	specPath := FeatureSpecPath(root, name)

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
		if !containsKeyword(req.Content) {
			result.Errors = append(result.Errors, ValidationIssue{
				Severity: SeverityError,
				Message:  fmt.Sprintf("requirement %q in capability %q must contain SHALL or MUST", req.Name, name),
				File:     specPath,
			})
		}
		if len(req.Scenarios) == 0 {
			result.Errors = append(result.Errors, ValidationIssue{
				Severity: SeverityError,
				Message:  fmt.Sprintf("requirement %q in capability %q has no scenarios", req.Name, name),
				File:     specPath,
			})
		}
		for _, sc := range req.Scenarios {
			if !strings.Contains(sc.Content, "WHEN") || !strings.Contains(sc.Content, "THEN") {
				result.Errors = append(result.Errors, ValidationIssue{
					Severity: SeverityError,
					Message:  fmt.Sprintf("scenario %q in requirement %q must contain WHEN and THEN", sc.Name, req.Name),
					File:     specPath,
				})
			}
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

		data, err := os.ReadFile(specPath)
		if err != nil {
			result.Errors = append(result.Errors, ValidationIssue{
				Severity: SeverityError,
				Message:  fmt.Sprintf("spec file for capability %q not found", name),
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
				Message:  fmt.Sprintf("capability %q has no requirements", name),
				File:     specPath,
			})
		}

		for _, req := range spec.Requirements {
			if !containsKeyword(req.Content) {
				result.Errors = append(result.Errors, ValidationIssue{
					Severity: SeverityError,
					Message:  fmt.Sprintf("requirement %q in capability %q must contain SHALL or MUST", req.Name, name),
					File:     specPath,
				})
			}
			if len(req.Scenarios) == 0 {
				result.Errors = append(result.Errors, ValidationIssue{
					Severity: SeverityError,
					Message:  fmt.Sprintf("requirement %q in capability %q has no scenarios", req.Name, name),
					File:     specPath,
				})
			}
			for _, sc := range req.Scenarios {
				if !strings.Contains(sc.Content, "WHEN") || !strings.Contains(sc.Content, "THEN") {
					result.Errors = append(result.Errors, ValidationIssue{
						Severity: SeverityError,
						Message:  fmt.Sprintf("scenario %q in requirement %q must contain WHEN and THEN", sc.Name, req.Name),
						File:     specPath,
					})
				}
			}
			result.ScenariosCount += len(req.Scenarios)
		}
	}

	result.Valid = len(result.Errors) == 0
	return result, nil
}
