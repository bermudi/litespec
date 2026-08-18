package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
}

func ValidateGHIssueQueues(root string) (*ValidationResult, error) {
	result := &ValidationResult{Valid: true}

	if _, err := exec.LookPath("gh"); err != nil {
		result.Warnings = append(result.Warnings, ValidationIssue{
			Severity:     SeverityWarning,
			Message:      "gh not available — issue queue not validated",
			StrictExempt: true,
		})
		return result, nil
	}

	gitCmd := exec.Command("git", "config", "--get", "remote.origin.url")
	gitCmd.Dir = root
	gitOut, err := gitCmd.Output()
	if err != nil || !strings.Contains(string(gitOut), "github.com") {
		result.Warnings = append(result.Warnings, ValidationIssue{
			Severity:     SeverityWarning,
			Message:      "no GitHub remote configured — issue queue not validated",
			StrictExempt: true,
		})
		return result, nil
	}

	cmd := exec.Command("gh", "issue", "list",
		"--label", "litespec",
		"--state", "open",
		"--json", "number,title,body,url",
		"--limit", "50",
	)
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		result.Warnings = append(result.Warnings, ValidationIssue{
			Severity:     SeverityWarning,
			Message:      fmt.Sprintf("gh issue list failed — issue queue not validated: %v", err),
			StrictExempt: true,
		})
		return result, nil
	}

	var issues []ghIssue
	if err := json.Unmarshal(out, &issues); err != nil {
		return nil, fmt.Errorf("parse gh issue list output: %w", err)
	}

	for _, issue := range issues {
		source := fmt.Sprintf("GH issue #%d", issue.Number)
		result.UnitsCount += countQueueUnits(issue.Body)
		for _, iss := range ValidateQueueBody(issue.Body, source) {
			if iss.Severity == SeverityWarning {
				result.Warnings = append(result.Warnings, iss)
			} else {
				result.Errors = append(result.Errors, iss)
			}
		}
	}

	result.Valid = len(result.Errors) == 0
	return result, nil
}

func countQueueUnits(body string) int {
	return len(parseQueueUnits(body))
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

func ValidateQueueBody(body string, source string) []ValidationIssue {
	units := parseQueueUnits(body)
	var issues []ValidationIssue

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
		fencedAfterVerify := false
		checkboxFound := false
		var verifyBlock string

		for i, line := range unit.Body {
			if strings.HasPrefix(line, "Done means:") {
				doneFound = true
			}
			if !verifyFound && strings.HasPrefix(line, "Verify:") {
				verifyFound = true
				for j := i + 1; j < len(unit.Body); j++ {
					if strings.HasPrefix(unit.Body[j], "```") {
						fencedAfterVerify = true
						var blockLines []string
						for k := j + 1; k < len(unit.Body); k++ {
							if strings.HasPrefix(unit.Body[k], "```") {
								break
							}
							blockLines = append(blockLines, unit.Body[k])
						}
						verifyBlock = strings.Join(blockLines, "\n")
						break
					}
				}
			}
			if strings.Contains(line, "- [ ]") ||
				strings.Contains(line, "- [x]") ||
				strings.Contains(line, "- [X]") {
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
		} else if !fencedAfterVerify {
			issues = append(issues, ValidationIssue{
				Severity: SeverityError,
				Message:  fmt.Sprintf("%s: unit %q Verify: not followed by fenced code block", source, unit.Heading),
				File:     source,
			})
		} else {
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

	return issues
}

func ValidateGHIssueByNumber(root string, number int) (*ValidationResult, error) {
	result := &ValidationResult{Valid: true}

	if _, err := exec.LookPath("gh"); err != nil {
		return nil, fmt.Errorf("gh not available")
	}

	gitCmd := exec.Command("git", "config", "--get", "remote.origin.url")
	gitCmd.Dir = root
	gitOut, err := gitCmd.Output()
	if err != nil || !strings.Contains(string(gitOut), "github.com") {
		return nil, fmt.Errorf("not a GitHub repository")
	}

	cmd := exec.Command("gh", "issue", "view", strconv.Itoa(number),
		"--json", "number,title,body,url",
	)
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("gh issue view %d failed: %w", number, err)
	}

	var issue ghIssue
	if err := json.Unmarshal(out, &issue); err != nil {
		return nil, fmt.Errorf("parse gh issue: %w", err)
	}

	source := fmt.Sprintf("GH issue #%d", issue.Number)
	result.UnitsCount += countQueueUnits(issue.Body)
	for _, iss := range ValidateQueueBody(issue.Body, source) {
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
	result.UnitsCount += countQueueUnits(string(body))
	for _, iss := range ValidateQueueBody(string(body), source) {
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
