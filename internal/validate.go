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

func ValidateDecision(root, slug string) (*ValidationResult, error) {
	result := &ValidationResult{Valid: true}

	d, err := FindDecisionBySlug(root, slug)
	if err != nil {
		return nil, err
	}
	if d == nil {
		return nil, fmt.Errorf("decision %q not found", slug)
	}

	validateDecisionFields(d, result)

	result.Valid = len(result.Errors) == 0
	result.DecisionsCount = 1
	return result, nil
}

func ValidateDecisions(root string) (*ValidationResult, error) {
	result := &ValidationResult{Valid: true}

	decisions, err := ListDecisions(root)
	if err != nil {
		return nil, err
	}
	if len(decisions) == 0 {
		result.Valid = true
		return result, nil
	}

	// Duplicate number detection
	numSet := make(map[int]string)
	for _, d := range decisions {
		if prev, ok := numSet[d.Number]; ok {
			result.Errors = append(result.Errors, ValidationIssue{
				Severity: SeverityError,
				Message:  fmt.Sprintf("duplicate decision number %d in %q and %q", d.Number, prev, d.FilePath),
				File:     d.FilePath,
			})
		}
		numSet[d.Number] = d.FilePath
	}

	// Duplicate slug detection
	slugSet := make(map[string]string)
	for _, d := range decisions {
		if prev, ok := slugSet[d.Slug]; ok {
			result.Errors = append(result.Errors, ValidationIssue{
				Severity: SeverityError,
				Message:  fmt.Sprintf("duplicate decision slug %q in %q and %q", d.Slug, prev, d.FilePath),
				File:     d.FilePath,
			})
		}
		slugSet[d.Slug] = d.FilePath
	}

	// Build slug lookup for pointer resolution
	slugMap := make(map[string]*Decision)
	for _, d := range decisions {
		slugMap[d.Slug] = d
		slugMap[fmt.Sprintf("%04d-%s", d.Number, d.Slug)] = d
	}

	for _, d := range decisions {
		validateDecisionFields(d, result)

		// Supersede pointer resolution
		for _, ref := range d.Supersedes {
			target, ok := resolveDecisionRef(slugMap, ref)
			if !ok {
				result.Errors = append(result.Errors, ValidationIssue{
					Severity: SeverityError,
					Message:  fmt.Sprintf("supersedes pointer %q does not resolve to an existing decision", ref),
					File:     d.FilePath,
				})
			} else if target.Status != StatusSuperseded {
				result.Errors = append(result.Errors, ValidationIssue{
					Severity: SeverityError,
					Message:  fmt.Sprintf("supersedes pointer %q targets decision %q with status %q (expected superseded)", ref, target.Slug, target.Status),
					File:     d.FilePath,
				})
			}
		}

		for _, ref := range d.SupersededBy {
			_, ok := resolveDecisionRef(slugMap, ref)
			if !ok {
				result.Errors = append(result.Errors, ValidationIssue{
					Severity: SeverityError,
					Message:  fmt.Sprintf("superseded-by pointer %q does not resolve to an existing decision", ref),
					File:     d.FilePath,
				})
			}
		}
	}

	// Supersede cycle detection
	supDepMap := make(map[string][]string)
	for _, d := range decisions {
		if len(d.Supersedes) > 0 {
			supDepMap[d.Slug] = d.Supersedes
		}
	}
	for _, cycle := range DetectCycles(supDepMap) {
		path := strings.Join(cycle, " -> ")
		result.Errors = append(result.Errors, ValidationIssue{
			Severity: SeverityError,
			Message:  fmt.Sprintf("supersede cycle detected: %s", path),
		})
	}

	result.Valid = len(result.Errors) == 0
	result.DecisionsCount = len(decisions)
	return result, nil
}

func validateDecisionFields(d *Decision, result *ValidationResult) {
	if d.Title == "" {
		result.Errors = append(result.Errors, ValidationIssue{
			Severity: SeverityError,
			Message:  "decision has no title",
			File:     d.FilePath,
		})
	}
	if !validStatuses[d.Status] {
		result.Errors = append(result.Errors, ValidationIssue{
			Severity: SeverityError,
			Message:  fmt.Sprintf("invalid status %q (valid: proposed, accepted, superseded)", d.Status),
			File:     d.FilePath,
		})
	}
	if d.Status == StatusSuperseded && len(d.SupersededBy) == 0 {
		result.Errors = append(result.Errors, ValidationIssue{
			Severity: SeverityError,
			Message:  fmt.Sprintf("decision %q is superseded but has no Superseded-By pointer", d.Slug),
			File:     d.FilePath,
		})
	}
	if d.Context == "" {
		result.Errors = append(result.Errors, ValidationIssue{
			Severity: SeverityError,
			Message:  fmt.Sprintf("decision %q missing Context section", d.Slug),
			File:     d.FilePath,
		})
	}
	if d.Decision == "" {
		result.Errors = append(result.Errors, ValidationIssue{
			Severity: SeverityError,
			Message:  fmt.Sprintf("decision %q missing Decision section", d.Slug),
			File:     d.FilePath,
		})
	}
	if d.Consequences == "" {
		result.Errors = append(result.Errors, ValidationIssue{
			Severity: SeverityError,
			Message:  fmt.Sprintf("decision %q missing Consequences section", d.Slug),
			File:     d.FilePath,
		})
	}
}

// resolveDecisionRef tries to find a decision by slug reference.
// Accepts either a slug, a NNNN-slug, or the slug portion after stripping a number prefix.
var numberPrefixRe = regexp.MustCompile(`^\d{4}-`)

func resolveDecisionRef(slugMap map[string]*Decision, ref string) (*Decision, bool) {
	if d, ok := slugMap[ref]; ok {
		return d, true
	}
	trimmed := numberPrefixRe.ReplaceAllString(ref, "")
	if trimmed != ref {
		if d, ok := slugMap[trimmed]; ok {
			return d, true
		}
	}
	return nil, false
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

	// Include decisions if present
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

	// Validate backlog if present
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
