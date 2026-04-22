package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bermudi/litespec/internal/skill"
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
		for _, e := range entries {
			if e.IsDir() && hasMarkdownFiles(filepath.Join(specsDir, e.Name())) {
				result.CapabilitiesCount++
			}
		}
		found := result.CapabilitiesCount > 0
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

				result.RequirementsCount += len(delta.Requirements)
				for _, req := range delta.Requirements {
					result.ScenariosCount += len(req.Scenarios)
				}

				for _, req := range delta.Requirements {
					if req.Operation == DeltaAdded || req.Operation == DeltaModified {
						if !containsKeyword(req.Content) {
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
						for _, sc := range req.Scenarios {
							if !strings.Contains(sc.Content, "WHEN") || !strings.Contains(sc.Content, "THEN") {
								result.Errors = append(result.Errors, ValidationIssue{
									Severity: SeverityError,
									Message:  fmt.Sprintf("scenario %q in requirement %q must contain WHEN and THEN", sc.Name, req.Name),
									File:     specPath,
								})
							}
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

				seenReqNames := make(map[string]bool)
				for _, req := range delta.Requirements {
					if seenReqNames[req.Name] {
						result.Errors = append(result.Errors, ValidationIssue{
							Severity: SeverityError,
							Message:  fmt.Sprintf("duplicate requirement name %q", req.Name),
							File:     specPath,
						})
					}
					seenReqNames[req.Name] = true

					seenScenarioNames := make(map[string]bool)
					for _, sc := range req.Scenarios {
						if seenScenarioNames[sc.Name] {
							result.Errors = append(result.Errors, ValidationIssue{
								Severity: SeverityError,
								Message:  fmt.Sprintf("duplicate scenario name %q in requirement %q", sc.Name, req.Name),
								File:     specPath,
							})
						}
						seenScenarioNames[sc.Name] = true
					}
				}

				nameOps := make(map[string][]DeltaOperation)
				for _, req := range delta.Requirements {
					switch req.Operation {
					case DeltaRenamed:
						nameOps[req.OldName] = append(nameOps[req.OldName], DeltaRenamed)
						if req.OldName != req.Name {
							nameOps[req.Name] = append(nameOps[req.Name], DeltaAdded)
						}
					default:
						nameOps[req.Name] = append(nameOps[req.Name], req.Operation)
					}
				}
				for name, ops := range nameOps {
					if len(ops) > 1 {
						result.Errors = append(result.Errors, ValidationIssue{
							Severity: SeverityError,
							Message:  fmt.Sprintf("conflicting operations on requirement %q", name),
							File:     specPath,
						})
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
					mainSpecPath := filepath.Join(CanonPath(root), entry.Name(), "spec.md")
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
		} else {
			for _, prob := range validateTasksChecklist(string(tasksData)) {
				result.Errors = append(result.Errors, ValidationIssue{
					Severity: SeverityError,
					Message:  prob,
					File:     tasksPath,
				})
			}
		}
	}

	meta, metaErr := ReadChangeMeta(root, name)
	if metaErr == nil && len(meta.DependsOn) > 0 {
		metaPath := filepath.Join(changeDir, MetaFileName)
		for _, dep := range meta.DependsOn {
			_, found := ResolveDep(root, dep)
			if !found {
				result.Errors = append(result.Errors, ValidationIssue{
					Severity: SeverityError,
					Message:  fmt.Sprintf("dependency %q not found", dep),
					File:     metaPath,
				})
			}
		}
	}

	result.Valid = len(result.Errors) == 0
	result.ChangesCount = 1
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

var checkboxLineRe = regexp.MustCompile(`(?i)^\s*- \[[ xX]\]`)

func validateTasksChecklist(content string) []string {
	var problems []string
	inPhase := false
	hasCheckbox := false
	phaseName := ""

	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## Phase") {
			if inPhase && !hasCheckbox {
				problems = append(problems, fmt.Sprintf("phase %q has no checklist items (- [ ])", phaseName))
			}
			inPhase = true
			hasCheckbox = false
			phaseName = trimmed
		} else if strings.HasPrefix(trimmed, "## ") && !strings.HasPrefix(trimmed, "## Phase") {
			if inPhase && !hasCheckbox {
				problems = append(problems, fmt.Sprintf("phase %q has no checklist items (- [ ])", phaseName))
			}
			inPhase = false
		} else if inPhase && checkboxLineRe.MatchString(line) {
			hasCheckbox = true
		}
	}
	if inPhase && !hasCheckbox {
		problems = append(problems, fmt.Sprintf("phase %q has no checklist items (- [ ])", phaseName))
	}
	return problems
}

var keywordRe = regexp.MustCompile(`\b(SHALL|MUST)\b`)

func stripCodeBlocks(content string) string {
	noFenced := regexp.MustCompile("(?s)```.*?```").ReplaceAllString(content, "")
	noInline := regexp.MustCompile("`[^`]*`").ReplaceAllString(noFenced, "")
	return noInline
}

func containsKeyword(content string) bool {
	return keywordRe.MatchString(stripCodeBlocks(content))
}

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
		return nil, err
	}

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
	if err != nil {
		return nil, fmt.Errorf("load dependency map: %w", err)
	}

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

	result.Valid = len(result.Errors) == 0
	if strict && len(result.Warnings) > 0 {
		result.Valid = false
	}

	return result, nil
}
