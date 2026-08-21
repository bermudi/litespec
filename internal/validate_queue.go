package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type ghIssue struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	URL    string `json:"url"`
}

type queueUnit struct {
	Heading string
	Body    []string
	Depends []string
}

func ValidateGHIssueQueues(root string) (*ValidationResult, error) {
	result := &ValidationResult{Valid: true}

	if _, err := lookPathGh("gh"); err != nil {
		result.Warnings = append(result.Warnings, ValidationIssue{
			Severity:     SeverityWarning,
			Message:      "gh not available — issue queue not validated",
			StrictExempt: true,
		})
		return result, nil
	}

	out, err := ghIssueList(root)
	if err != nil {
		result.Warnings = append(result.Warnings, ValidationIssue{
			Severity:     SeverityWarning,
			Message:      fmt.Sprintf("gh issue list failed — issue queue not validated: %v", err),
			StrictExempt: true,
		})
		return result, nil
	}

	if len(out) == 0 {
		result.Warnings = append(result.Warnings, ValidationIssue{
			Severity:     SeverityWarning,
			Message:      "gh issue list returned empty output — issue queue not validated",
			StrictExempt: true,
		})
	} else {
		var issues []ghIssue
		if err := json.Unmarshal(out, &issues); err != nil {
			result.Warnings = append(result.Warnings, ValidationIssue{
				Severity:     SeverityWarning,
				Message:      fmt.Sprintf("parse gh issue list output: %v — issue queue not validated", err),
				StrictExempt: true,
			})
			return result, nil
		}

		for _, issue := range issues {
			source := fmt.Sprintf("GH issue #%d", issue.Number)
			units, unitIssues := ValidateQueueBody(issue.Body, source)
			result.UnitsCount += len(units)
			for _, iss := range unitIssues {
				if iss.Severity == SeverityWarning {
					result.Warnings = append(result.Warnings, iss)
				} else {
					result.Errors = append(result.Errors, iss)
				}
			}
		}
	}

	result.Valid = len(result.Errors) == 0
	return result, nil
}

func parseQueueUnits(body string) []queueUnit {
	lines := strings.Split(body, "\n")
	var units []queueUnit
	var current *queueUnit

	for _, line := range lines {
		if strings.HasPrefix(line, "## ") {
			if current != nil {
				units = append(units, *current)
			}
			heading := strings.TrimSpace(line[3:])
			current = &queueUnit{Heading: heading}
			continue
		}
		if current != nil {
			current.Body = append(current.Body, line)
		}
	}

	if current != nil {
		units = append(units, *current)
	}

	return units
}

var lookPathBash = exec.LookPath
var lookPathGh = exec.LookPath
var queueBasePattern = regexp.MustCompile(`^[0-9a-fA-F]{40}([0-9a-fA-F]{24})?$`)
var queueBranchPattern = regexp.MustCompile(`^litespec/[a-z0-9]+(?:-[a-z0-9]+)*$`)

var ghIssueView = func(root string, number int) ([]byte, error) {
	cmd := exec.Command("gh", "issue", "view", strconv.Itoa(number),
		"--json", "number,title,body,url")
	cmd.Dir = root
	return cmd.Output()
}

var ghIssueList = func(root string) ([]byte, error) {
	cmd := exec.Command("gh", "issue", "list",
		"--label", "litespec",
		"--state", "open",
		"--json", "number,title,body,url",
		"--limit", "10000",
	)
	cmd.Dir = root
	return cmd.Output()
}

func lintVerifyShell(block string, source string, unitHeading string) []ValidationIssue {
	if strings.TrimSpace(block) == "" {
		return []ValidationIssue{{
			Severity: SeverityError,
			Message:  fmt.Sprintf("%s: unit %q Verify block is empty", source, unitHeading),
			File:     source,
		}}
	}

	bashPath, err := lookPathBash("bash")
	if err != nil {
		return []ValidationIssue{{
			Severity: SeverityWarning,
			Message:  fmt.Sprintf("%s: unit %q Verify block not syntax-checked (bash unavailable)", source, unitHeading),
			File:     source,
		}}
	}

	cmd := exec.Command(bashPath, "-n", "-c", block)
	out, err := cmd.CombinedOutput()
	if err != nil {
		trimmed := strings.TrimSpace(string(out))
		return []ValidationIssue{{
			Severity: SeverityError,
			Message:  fmt.Sprintf("%s: unit %q Verify shell syntax error: %s", source, unitHeading, trimmed),
			File:     source,
		}}
	}

	return nil
}

func isUnit(unit queueUnit) bool {
	for _, line := range unit.Body {
		if strings.HasPrefix(line, "Done means:") || strings.HasPrefix(line, "Verify:") {
			return true
		}
	}
	return false
}

func isCheckboxLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	for _, checkbox := range []string{"- [ ]", "- [x]", "- [X]"} {
		if trimmed == checkbox || strings.HasPrefix(trimmed, checkbox+" ") {
			return true
		}
	}
	return false
}

func parseDepends(body []string) []string {
	var deps []string
	for _, line := range body {
		if !strings.HasPrefix(line, "Depends:") {
			continue
		}
		rest := strings.TrimSpace(line[len("Depends:"):])
		for _, part := range strings.Split(rest, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				deps = append(deps, part)
			}
		}
		break
	}
	return deps
}

func validateQueueOwnership(body string, source string) []ValidationIssue {
	type ownershipLine struct {
		value         string
		beforeHeading bool
	}

	var bases []ownershipLine
	var branches []ownershipLine
	beforeHeading := true
	for _, line := range strings.Split(body, "\n") {
		if strings.HasPrefix(line, "## ") {
			beforeHeading = false
		}
		if strings.HasPrefix(line, "Base:") {
			bases = append(bases, ownershipLine{
				value:         strings.TrimSpace(strings.TrimPrefix(line, "Base:")),
				beforeHeading: beforeHeading,
			})
		}
		if strings.HasPrefix(line, "Branch:") {
			branches = append(branches, ownershipLine{
				value:         strings.TrimSpace(strings.TrimPrefix(line, "Branch:")),
				beforeHeading: beforeHeading,
			})
		}
	}

	var issues []ValidationIssue
	if len(bases) != 1 {
		issues = append(issues, ValidationIssue{
			Severity: SeverityError,
			Message:  fmt.Sprintf("%s: expected exactly one Base: ownership line in the queue", source),
			File:     source,
		})
	} else if !bases[0].beforeHeading {
		issues = append(issues, ValidationIssue{
			Severity: SeverityError,
			Message:  fmt.Sprintf("%s: Base: ownership line must appear before the first ## heading", source),
			File:     source,
		})
	} else if !queueBasePattern.MatchString(bases[0].value) {
		issues = append(issues, ValidationIssue{
			Severity: SeverityError,
			Message:  fmt.Sprintf("%s: Base: must contain a full 40- or 64-character hexadecimal commit ID", source),
			File:     source,
		})
	}

	if len(branches) != 1 {
		issues = append(issues, ValidationIssue{
			Severity: SeverityError,
			Message:  fmt.Sprintf("%s: expected exactly one Branch: ownership line in the queue", source),
			File:     source,
		})
	} else if !branches[0].beforeHeading {
		issues = append(issues, ValidationIssue{
			Severity: SeverityError,
			Message:  fmt.Sprintf("%s: Branch: ownership line must appear before the first ## heading", source),
			File:     source,
		})
	} else if !queueBranchPattern.MatchString(branches[0].value) {
		issues = append(issues, ValidationIssue{
			Severity: SeverityError,
			Message:  fmt.Sprintf("%s: Branch: must match litespec/<kebab-change-name>", source),
			File:     source,
		})
	}

	return issues
}

func ValidateQueueBody(body string, source string) ([]queueUnit, []ValidationIssue) {
	all := parseQueueUnits(body)
	units := make([]queueUnit, 0, len(all))
	for _, u := range all {
		if isUnit(u) {
			units = append(units, u)
		}
	}
	issues := validateQueueOwnership(body, source)

	for _, unit := range units {
		if strings.TrimSpace(unit.Heading) == "" {
			issues = append(issues, ValidationIssue{
				Severity: SeverityError,
				Message:  fmt.Sprintf("%s: empty unit heading", source),
				File:     source,
			})
		}

		doneFound := false
		verifyFound := false
		checkboxFound := false
		inlineVerify := false
		hasFencedBlock := false
		var verifyBlock string
		inFencedBlock := false

		for i, line := range unit.Body {
			if strings.HasPrefix(strings.TrimSpace(line), "```") {
				inFencedBlock = !inFencedBlock
				continue
			}
			if inFencedBlock {
				continue
			}
			if strings.HasPrefix(line, "Done means:") {
				doneFound = true
			}
			if !verifyFound && strings.HasPrefix(line, "Verify:") {
				verifyFound = true
				rest := strings.TrimSpace(line[len("Verify:"):])
				firstBacktick := strings.Index(rest, "`")
				lastBacktick := strings.LastIndex(rest, "`")
				if firstBacktick >= 0 && lastBacktick > firstBacktick &&
					strings.TrimSpace(rest[firstBacktick+1:lastBacktick]) != "" {
					inlineVerify = true
				}
				for j := i + 1; j < len(unit.Body); j++ {
					if strings.HasPrefix(unit.Body[j], "```") {
						var blockLines []string
						for k := j + 1; k < len(unit.Body); k++ {
							if strings.HasPrefix(unit.Body[k], "```") {
								hasFencedBlock = true
								break
							}
							blockLines = append(blockLines, unit.Body[k])
						}
						if hasFencedBlock {
							verifyBlock = strings.Join(blockLines, "\n")
						}
						break
					}
				}
			}
			if isCheckboxLine(line) {
				checkboxFound = true
			}
		}

		if !doneFound {
			issues = append(issues, ValidationIssue{
				Severity: SeverityError,
				Message:  fmt.Sprintf("%s: unit %q missing Done means:", source, unit.Heading),
				File:     source,
			})
		}
		if !verifyFound {
			issues = append(issues, ValidationIssue{
				Severity: SeverityError,
				Message:  fmt.Sprintf("%s: unit %q missing Verify:", source, unit.Heading),
				File:     source,
			})
		} else if !inlineVerify && !hasFencedBlock {
			issues = append(issues, ValidationIssue{
				Severity: SeverityError,
				Message:  fmt.Sprintf("%s: unit %q Verify: not followed by a command or fenced code block", source, unit.Heading),
				File:     source,
			})
		} else if hasFencedBlock {
			issues = append(issues, lintVerifyShell(verifyBlock, source, unit.Heading)...)
		}
		if !checkboxFound {
			issues = append(issues, ValidationIssue{
				Severity: SeverityError,
				Message:  fmt.Sprintf("%s: unit %q missing checkbox", source, unit.Heading),
				File:     source,
			})
		}
	}

	headings := make(map[string]bool, len(units))
	for _, u := range units {
		headings[u.Heading] = true
	}

	for i := range units {
		deps := parseDepends(units[i].Body)
		units[i].Depends = deps
		seen := make(map[string]bool, len(deps))
		for _, dep := range deps {
			if seen[dep] {
				continue
			}
			seen[dep] = true
			if !headings[dep] {
				issues = append(issues, ValidationIssue{
					Severity: SeverityError,
					Message:  fmt.Sprintf("%s: unit %q depends on non-existent unit %q", source, units[i].Heading, dep),
					File:     source,
				})
			}
		}
	}

	return units, issues
}

func ValidateGHIssueByNumber(root string, number int) (*ValidationResult, error) {
	result := &ValidationResult{Valid: true}

	if _, err := lookPathGh("gh"); err != nil {
		return nil, fmt.Errorf("gh not available")
	}

	out, err := ghIssueView(root, number)
	if err != nil {
		return nil, fmt.Errorf("gh issue view %d failed: %w", number, err)
	}

	var issue ghIssue
	if err := json.Unmarshal(out, &issue); err != nil {
		return nil, fmt.Errorf("parse gh issue: %w", err)
	}

	source := fmt.Sprintf("GH issue #%d", issue.Number)
	units, unitIssues := ValidateQueueBody(issue.Body, source)
	result.UnitsCount += len(units)
	for _, iss := range unitIssues {
		if iss.Severity == SeverityWarning {
			result.Warnings = append(result.Warnings, iss)
		} else {
			result.Errors = append(result.Errors, iss)
		}
	}
	result.Valid = len(result.Errors) == 0
	return result, nil
}

func ValidateQueueFile(path string) (*ValidationResult, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	result := &ValidationResult{Valid: true}
	source := fmt.Sprintf("queue file %s", path)
	units, unitIssues := ValidateQueueBody(string(body), source)
	result.UnitsCount += len(units)
	for _, iss := range unitIssues {
		if iss.Severity == SeverityWarning {
			result.Warnings = append(result.Warnings, iss)
		} else {
			result.Errors = append(result.Errors, iss)
		}
	}
	result.Valid = len(result.Errors) == 0
	return result, nil
}

func ValidateLocalQueues(root string) (*ValidationResult, error) {
	result := &ValidationResult{Valid: true}

	dir := filepath.Join(root, "specs", "queues")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		sub, err := ValidateQueueFile(path)
		if err != nil {
			return nil, err
		}
		result.Errors = append(result.Errors, sub.Errors...)
		result.Warnings = append(result.Warnings, sub.Warnings...)
		result.UnitsCount += sub.UnitsCount
	}

	result.Valid = len(result.Errors) == 0
	return result, nil
}
