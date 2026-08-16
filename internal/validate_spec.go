package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func resolveSpecPath(root, name string) string {
	featurePath := FeatureSpecPath(root, name)
	if _, err := os.Stat(featurePath); err == nil {
		return featurePath
	}
	return filepath.Join(CanonPath(root), name, "spec.md")
}

func ValidateSpec(root, name string) (*ValidationResult, error) {
	result := &ValidationResult{Valid: true}
	specPath := resolveSpecPath(root, name)

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
			result.Warnings = append(result.Warnings, ValidationIssue{
				Severity: SeverityWarning,
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

	seen := make(map[string]bool)
	var specFiles []struct {
		name string
		path string
	}

	canonDir := CanonPath(root)
	if entries, err := os.ReadDir(canonDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			name := entry.Name()
			specPath := filepath.Join(canonDir, name, "spec.md")
			if _, statErr := os.Stat(specPath); statErr != nil {
				continue
			}
			specFiles = append(specFiles, struct {
				name string
				path string
			}{name, specPath})
			seen[name] = true
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("read specs directory: %w", err)
	}

	projectSpecsDir := filepath.Join(root, ProjectDirName)
	if entries, err := os.ReadDir(projectSpecsDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			name := entry.Name()
			if name == CanonDirName || name == ChangesDirName || seen[name] {
				continue
			}
			specPath := filepath.Join(projectSpecsDir, name, "spec.md")
			if _, statErr := os.Stat(specPath); statErr != nil {
				continue
			}
			specFiles = append(specFiles, struct {
				name string
				path string
			}{name, specPath})
		}
	}

	if len(specFiles) == 0 {
		return result, nil
	}

	for _, sf := range specFiles {
		data, err := os.ReadFile(sf.path)
		if err != nil {
			result.Errors = append(result.Errors, ValidationIssue{
				Severity: SeverityError,
				Message:  fmt.Sprintf("spec file for capability %q not found", sf.name),
				File:     sf.path,
			})
			continue
		}

		spec, parseErr := ParseMainSpec(string(data))
		if parseErr != nil {
			result.Errors = append(result.Errors, ValidationIssue{
				Severity: SeverityError,
				Message:  fmt.Sprintf("invalid spec: %s", parseErr),
				File:     sf.path,
			})
			continue
		}

		result.CapabilitiesCount++
		result.RequirementsCount += len(spec.Requirements)

		if len(spec.Requirements) == 0 {
			result.Warnings = append(result.Warnings, ValidationIssue{
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("capability %q has no requirements", sf.name),
				File:     sf.path,
			})
		}

		for _, req := range spec.Requirements {
			if !containsKeyword(req.Content) {
				result.Errors = append(result.Errors, ValidationIssue{
					Severity: SeverityError,
					Message:  fmt.Sprintf("requirement %q in capability %q must contain SHALL or MUST", req.Name, sf.name),
					File:     sf.path,
				})
			}
			if len(req.Scenarios) == 0 {
				result.Warnings = append(result.Warnings, ValidationIssue{
					Severity: SeverityWarning,
					Message:  fmt.Sprintf("requirement %q in capability %q has no scenarios", req.Name, sf.name),
					File:     sf.path,
				})
			}
			for _, sc := range req.Scenarios {
				if !strings.Contains(sc.Content, "WHEN") || !strings.Contains(sc.Content, "THEN") {
					result.Errors = append(result.Errors, ValidationIssue{
						Severity: SeverityError,
						Message:  fmt.Sprintf("scenario %q in requirement %q must contain WHEN and THEN", sc.Name, req.Name),
						File:     sf.path,
					})
				}
			}
			result.ScenariosCount += len(req.Scenarios)
		}
	}

	result.Valid = len(result.Errors) == 0
	return result, nil
}
