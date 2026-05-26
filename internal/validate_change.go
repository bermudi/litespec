package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

func ValidateChange(root, name string) (*ValidationResult, error) {
	result := &ValidationResult{Valid: true}
	changeDir := ChangePath(root, name)

	if _, err := os.Stat(changeDir); err != nil {
		return nil, fmt.Errorf("change %q not found", name)
	}

	metaPath := filepath.Join(changeDir, MetaFileName)
	if _, err := os.Stat(metaPath); err != nil {
		result.Errors = append(result.Errors, ValidationIssue{
			Severity: SeverityError,
			Message:  fmt.Sprintf("missing %s — run 'litespec new <name>' or create it manually", MetaFileName),
			File:     changeDir,
		})
	}

	optionalFiles := []struct {
		id       string
		filename string
	}{
		{"proposal", "proposal.md"},
		{"design", "design.md"},
		{"tasks", "tasks.md"},
	}
	for _, of := range optionalFiles {
		p := filepath.Join(changeDir, of.filename)
		data, readErr := os.ReadFile(p)
		if readErr != nil {
			continue
		}
		switch of.id {
		case "proposal":
			for _, issue := range validateProposal(string(data)) {
				issue.File = p
				result.Errors = append(result.Errors, issue)
			}
		case "design":
			for _, issue := range validateDesign(string(data)) {
				issue.File = p
				result.Errors = append(result.Errors, issue)
			}
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
					existingNames = buildEffectiveNames(root, entry.Name(), name)
					if existingNames == nil {
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
		line = strings.TrimSuffix(line, "\r")
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
		trimmed := strings.TrimSuffix(line, "\r")
		trimmed = strings.TrimSpace(trimmed)
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

func validateProposal(content string) []ValidationIssue {
	var issues []ValidationIssue

	hasMotivation := false
	hasScope := false
	var currentHeading string
	var bodyLines []string

	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(strings.TrimSuffix(line, "\r"))
		if strings.HasPrefix(trimmed, "## ") {
			if currentHeading != "" && len(bodyLines) > 0 {
				switch currentHeading {
				case "Motivation", "Why":
					hasMotivation = true
				case "Scope", "What Changes":
					hasScope = true
				}
			}
			headingName := strings.TrimPrefix(trimmed, "## ")
			currentHeading = headingName
			bodyLines = nil
			continue
		}
		if currentHeading != "" && trimmed != "" {
			bodyLines = append(bodyLines, trimmed)
		}
	}
	if currentHeading != "" && len(bodyLines) > 0 {
		switch currentHeading {
		case "Motivation", "Why":
			hasMotivation = true
		case "Scope", "What Changes":
			hasScope = true
		}
	}

	if !hasMotivation {
		issues = append(issues, ValidationIssue{
			Severity: SeverityError,
			Message:  "proposal.md must contain ## Motivation (or ## Why) heading with non-blank body",
		})
	}
	if !hasScope {
		issues = append(issues, ValidationIssue{
			Severity: SeverityError,
			Message:  "proposal.md must contain ## Scope (or ## What Changes) heading with non-blank body",
		})
	}
	return issues
}

func validateDesign(content string) []ValidationIssue {
	var issues []ValidationIssue

	hasH2 := false
	nonBlankLines := 0
	inFence := false

	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(strings.TrimSuffix(line, "\r"))
		if strings.HasPrefix(trimmed, "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		if strings.HasPrefix(trimmed, "## ") {
			hasH2 = true
			continue
		}
		if strings.HasPrefix(trimmed, "# ") {
			continue
		}
		if trimmed != "" {
			nonBlankLines++
		}
	}

	if !hasH2 {
		issues = append(issues, ValidationIssue{
			Severity: SeverityError,
			Message:  "design.md must contain at least one ## heading",
		})
	}
	if nonBlankLines < 3 {
		issues = append(issues, ValidationIssue{
			Severity: SeverityError,
			Message:  "design.md must have at least 3 non-blank lines outside fenced code blocks",
		})
	}
	return issues
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

func buildEffectiveNames(root, capability, changeName string) map[string]bool {
	meta, metaErr := ReadChangeMeta(root, changeName)
	hasDeps := metaErr == nil && len(meta.DependsOn) > 0

	mainSpecPath := filepath.Join(CanonPath(root), capability, "spec.md")
	mainData, canonErr := os.ReadFile(mainSpecPath)

	if canonErr != nil && !hasDeps {
		return nil
	}

	var existingNames map[string]bool
	if canonErr == nil {
		mainSpec, parseErr := ParseMainSpec(string(mainData))
		if parseErr != nil {
			return nil
		}
		existingNames = make(map[string]bool)
		for _, r := range mainSpec.Requirements {
			existingNames[r.Name] = true
		}
	} else {
		existingNames = make(map[string]bool)
	}

	if !hasDeps {
		if len(existingNames) == 0 {
			return nil
		}
		return existingNames
	}

	var deltas []DeltaRequirement
	visited := make(map[string]bool)
	collectTransitiveDepDeltas(root, capability, meta.DependsOn, visited, &deltas)

	var renamed, removed, added []DeltaRequirement
	for _, r := range deltas {
		switch r.Operation {
		case DeltaRenamed:
			renamed = append(renamed, r)
		case DeltaRemoved:
			removed = append(removed, r)
		case DeltaAdded:
			added = append(added, r)
		}
	}

	for _, r := range renamed {
		if r.OldName != r.Name && existingNames[r.OldName] {
			delete(existingNames, r.OldName)
			existingNames[r.Name] = true
		}
	}
	for _, r := range removed {
		delete(existingNames, r.Name)
	}
	for _, r := range added {
		existingNames[r.Name] = true
	}

	if len(existingNames) == 0 {
		return nil
	}
	return existingNames
}

func collectTransitiveDepDeltas(root, capability string, deps []string, visited map[string]bool, out *[]DeltaRequirement) {
	for _, dep := range deps {
		if visited[dep] || !ChangeExists(root, dep) {
			continue
		}
		visited[dep] = true

		depMeta, err := ReadChangeMeta(root, dep)
		if err == nil && len(depMeta.DependsOn) > 0 {
			collectTransitiveDepDeltas(root, capability, depMeta.DependsOn, visited, out)
		}

		loadDepDeltas(root, dep, capability, out)
	}
}

func loadDepDeltas(root, dep, capability string, out *[]DeltaRequirement) {
	depSpecsDir := filepath.Join(ChangeSpecsPath(root, dep), capability)
	entries, err := os.ReadDir(depSpecsDir)
	if err != nil {
		return
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, f := range entries {
		if filepath.Ext(f.Name()) != ".md" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(depSpecsDir, f.Name()))
		if err != nil {
			continue
		}
		delta, parseErr := ParseDeltaSpec(string(data))
		if parseErr != nil {
			continue
		}
		*out = append(*out, delta.Requirements...)
	}
}
