package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ValidateChange(root, name string) (*ValidationResult, error) {
	result := &ValidationResult{Valid: true}
	changeDir := ChangePath(root, name)

	if _, err := os.Stat(changeDir); err != nil {
		return nil, fmt.Errorf("change %q not found", name)
	}

	requiredFiles := []struct {
		id       string
		filename string
	}{
		{"proposal", "proposal.md"},
		{"design", "design.md"},
		{"tasks", "tasks.md"},
	}
	for _, rf := range requiredFiles {
		p := filepath.Join(changeDir, rf.filename)
		if _, err := os.Stat(p); err != nil {
			result.Errors = append(result.Errors, ValidationIssue{
				Severity: SeverityError,
				Message:  fmt.Sprintf("missing required artifact: %s", rf.id),
				File:     p,
			})
		}
	}

	specsDir := ChangeSpecsPath(root, name)
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		result.Errors = append(result.Errors, ValidationIssue{
			Severity: SeverityError,
			Message:  "missing specs directory",
			File:     specsDir,
		})
	} else {
		found := false
		for _, e := range entries {
			if e.IsDir() && hasMarkdownFiles(filepath.Join(specsDir, e.Name())) {
				found = true
				break
			}
		}
		if !found {
			result.Errors = append(result.Errors, ValidationIssue{
				Severity: SeverityError,
				Message:  "specs directory contains no delta spec files",
				File:     specsDir,
			})
		}
	}

	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			capDir := filepath.Join(specsDir, entry.Name())
			files, readErr := os.ReadDir(capDir)
			if readErr != nil {
				continue
			}
			for _, f := range files {
				if filepath.Ext(f.Name()) != ".md" {
					continue
				}
				specPath := filepath.Join(capDir, f.Name())
				data, readErr := os.ReadFile(specPath)
				if readErr != nil {
					continue
				}

				delta, parseErr := ParseDeltaSpec(string(data))
				if parseErr != nil {
					result.Errors = append(result.Errors, ValidationIssue{
						Severity: SeverityError,
						Message:  fmt.Sprintf("invalid delta spec: %s", parseErr),
						File:     specPath,
					})
					continue
				}

				for _, req := range delta.Requirements {
					if req.Operation == DeltaAdded || req.Operation == DeltaModified {
						if !strings.Contains(req.Content, "SHALL") && !strings.Contains(req.Content, "MUST") {
							result.Errors = append(result.Errors, ValidationIssue{
								Severity: SeverityError,
								Message:  fmt.Sprintf("%s requirement %q must contain SHALL or MUST", req.Operation, req.Name),
								File:     specPath,
							})
						}
						if len(req.Scenarios) == 0 {
							result.Errors = append(result.Errors, ValidationIssue{
								Severity: SeverityError,
								Message:  fmt.Sprintf("%s requirement %q must include at least one scenario", req.Operation, req.Name),
								File:     specPath,
							})
						}
					}
					if req.Operation == DeltaRemoved {
						if req.Content != "" {
							result.Errors = append(result.Errors, ValidationIssue{
								Severity: SeverityError,
								Message:  fmt.Sprintf("REMOVED requirement %q must not have body content", req.Name),
								File:     specPath,
							})
						}
						if len(req.Scenarios) > 0 {
							result.Errors = append(result.Errors, ValidationIssue{
								Severity: SeverityError,
								Message:  fmt.Sprintf("REMOVED requirement %q must not have scenarios", req.Name),
								File:     specPath,
							})
						}
					}
					if req.Operation == DeltaRenamed {
						if req.Content != "" {
							result.Errors = append(result.Errors, ValidationIssue{
								Severity: SeverityError,
								Message:  fmt.Sprintf("RENAMED requirement %q must not have body content", req.Name),
								File:     specPath,
							})
						}
						if len(req.Scenarios) > 0 {
							result.Errors = append(result.Errors, ValidationIssue{
								Severity: SeverityError,
								Message:  fmt.Sprintf("RENAMED requirement %q must not have scenarios", req.Name),
								File:     specPath,
							})
						}
						if req.OldName == req.Name {
							result.Warnings = append(result.Warnings, ValidationIssue{
								Severity: SeverityWarning,
								Message:  fmt.Sprintf("RENAMED requirement %q has same old and new name", req.Name),
								File:     specPath,
							})
						}
					}
				}

				needsMainSpec := false
				for _, req := range delta.Requirements {
					if req.Operation == DeltaModified || req.Operation == DeltaRemoved || req.Operation == DeltaRenamed || req.Operation == DeltaAdded {
						needsMainSpec = true
						break
					}
				}

				var existingNames map[string]bool
				if needsMainSpec {
					mainSpecPath := filepath.Join(SpecsPath(root), entry.Name(), "spec.md")
					mainData, readErr := os.ReadFile(mainSpecPath)
					if readErr != nil {
						hasModOrRenameOrRemove := false
						for _, req := range delta.Requirements {
							if req.Operation == DeltaModified || req.Operation == DeltaRemoved || req.Operation == DeltaRenamed {
								hasModOrRenameOrRemove = true
								break
							}
						}
						if hasModOrRenameOrRemove {
							result.Errors = append(result.Errors, ValidationIssue{
								Severity: SeverityError,
								Message:  fmt.Sprintf("main spec for capability %q does not exist", entry.Name()),
								File:     specPath,
							})
							continue
						}
					} else {
						mainSpec, parseErr := ParseMainSpec(string(mainData))
						if parseErr != nil {
							result.Errors = append(result.Errors, ValidationIssue{
								Severity: SeverityError,
								Message:  fmt.Sprintf("invalid main spec for capability %q: %s", entry.Name(), parseErr),
								File:     mainSpecPath,
							})
							continue
						}

						existingNames = make(map[string]bool)
						for _, r := range mainSpec.Requirements {
							existingNames[r.Name] = true
						}
					}
				}

				if existingNames != nil {
					for _, req := range delta.Requirements {
						switch req.Operation {
						case DeltaModified, DeltaRemoved:
							if !existingNames[req.Name] {
								result.Errors = append(result.Errors, ValidationIssue{
									Severity: SeverityError,
									Message:  fmt.Sprintf("%s requirement %q not found in main spec", req.Operation, req.Name),
									File:     specPath,
								})
							}
						case DeltaRenamed:
							if !existingNames[req.OldName] {
								result.Errors = append(result.Errors, ValidationIssue{
									Severity: SeverityError,
									Message:  fmt.Sprintf("RENAMED requirement %q not found in main spec", req.OldName),
									File:     specPath,
								})
							}
							if existingNames[req.Name] && req.Name != req.OldName {
								result.Errors = append(result.Errors, ValidationIssue{
									Severity: SeverityError,
									Message:  fmt.Sprintf("RENAMED requirement new name %q already exists in main spec", req.Name),
									File:     specPath,
								})
							}
						case DeltaAdded:
							if existingNames[req.Name] {
								result.Errors = append(result.Errors, ValidationIssue{
									Severity: SeverityError,
									Message:  fmt.Sprintf("ADDED requirement %q already exists in main spec", req.Name),
									File:     specPath,
								})
							}
						}
					}
				}
			}
		}
	}

	tasksPath := filepath.Join(changeDir, "tasks.md")
	tasksData, err := os.ReadFile(tasksPath)
	if err == nil {
		if !hasPhaseHeading(string(tasksData)) {
			result.Errors = append(result.Errors, ValidationIssue{
				Severity: SeverityError,
				Message:  "tasks.md has no phase headings (## Phase)",
				File:     tasksPath,
			})
		}
	}

	result.Valid = len(result.Errors) == 0
	return result, nil
}

func hasPhaseHeading(content string) bool {
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "## Phase") {
			return true
		}
	}
	return false
}

func ValidateSpecs(root string) (*ValidationResult, error) {
	result := &ValidationResult{Valid: true}
	specsDir := SpecsPath(root)

	entries, err := os.ReadDir(specsDir)
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
		}
	}

	result.Valid = len(result.Errors) == 0
	return result, nil
}

func ValidateAll(root string, strict bool) (*ValidationResult, error) {
	result := &ValidationResult{Valid: true}

	specResult, err := ValidateSpecs(root)
	if err != nil {
		return nil, err
	}
	result.Errors = append(result.Errors, specResult.Errors...)
	result.Warnings = append(result.Warnings, specResult.Warnings...)

	changes, err := ListChanges(root)
	if err != nil {
		return nil, err
	}

	for _, name := range changes {
		changeResult, err := ValidateChange(root, name)
		if err != nil {
			return nil, err
		}
		result.Errors = append(result.Errors, changeResult.Errors...)
		result.Warnings = append(result.Warnings, changeResult.Warnings...)
	}

	result.Valid = len(result.Errors) == 0
	if strict && len(result.Warnings) > 0 {
		result.Valid = false
	}

	return result, nil
}
