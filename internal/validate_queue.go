package internal

import (
	"encoding/json"
	"fmt"
	"os/exec"
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
			Severity: SeverityWarning,
			Message:  "gh not available — issue queue not validated",
			File:     "",
		})
		return result, nil
	}

	gitCmd := exec.Command("git", "config", "--get", "remote.origin.url")
	gitCmd.Dir = root
	gitOut, err := gitCmd.Output()
	if err != nil || !strings.Contains(string(gitOut), "github.com") {
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
			Severity: SeverityWarning,
			Message:  fmt.Sprintf("gh issue list failed — issue queue not validated: %v", err),
			File:     "",
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
		result.Errors = append(result.Errors, ValidateQueueBody(issue.Body, source)...)
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

		for i, line := range unit.Body {
			if strings.HasPrefix(line, "Done means:") {
				doneFound = true
			}
			if !verifyFound && strings.HasPrefix(line, "Verify:") {
				verifyFound = true
				for j := i + 1; j < len(unit.Body); j++ {
					if strings.HasPrefix(unit.Body[j], "```") {
						fencedAfterVerify = true
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
