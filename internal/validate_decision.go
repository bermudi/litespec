package internal

import (
	"fmt"
	"regexp"
	"strings"
)

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

	paths, err := decisionFiles(root)
	if err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		result.Valid = true
		return result, nil
	}

	var decisions []*Decision
	for _, path := range paths {
		d, err := ParseDecision(path)
		if err != nil {
			result.Errors = append(result.Errors, ValidationIssue{
				Severity: SeverityError,
				Message:  fmt.Sprintf("invalid decision: %v", err),
				File:     path,
			})
			continue
		}
		decisions = append(decisions, d)
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
