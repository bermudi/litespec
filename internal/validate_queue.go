package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func ValidateQueues(root string) *ValidationResult {
	result := &ValidationResult{Valid: true}

	if _, err := exec.LookPath("gh"); err == nil {
		issues := fetchGHIssues(root)
		for _, iss := range issues {
			file := fmt.Sprintf("GH issue #%d", iss.Number)
			if iss.URL != "" {
				file = iss.URL
			}
			bodyIssues := validateQueueBody(iss.Body, file)
			for _, is := range bodyIssues {
				if is.Severity == SeverityError {
					result.Errors = append(result.Errors, is)
				} else {
					result.Warnings = append(result.Warnings, is)
				}
			}
		}
	}

	queueFiles, _ := filepath.Glob(filepath.Join(ChangesPath(root), "*", "QUEUE.md"))
	for _, qf := range queueFiles {
		data, err := os.ReadFile(qf)
		if err != nil {
			continue
		}
		rel, _ := filepath.Rel(root, qf)
		bodyIssues := validateQueueBody(string(data), rel)
		for _, is := range bodyIssues {
			if is.Severity == SeverityError {
				result.Errors = append(result.Errors, is)
			} else {
				result.Warnings = append(result.Warnings, is)
			}
		}
	}

	result.Valid = len(result.Errors) == 0
	return result
}

type ghIssue struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	URL    string `json:"url"`
}

func fetchGHIssues(root string) []ghIssue {
	cmd := exec.Command("gh", "issue", "list", "--json", "number,title,body,url", "--state", "open", "--limit", "100")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	var issues []ghIssue
	if err := json.Unmarshal(out, &issues); err != nil {
		return nil
	}
	return issues
}

func validateQueueBody(body, file string) []ValidationIssue {
	var issues []ValidationIssue
	if strings.TrimSpace(body) == "" {
		return issues
	}

	lines := strings.Split(body, "\n")
	queueIdx := -1
	for i, l := range lines {
		if strings.TrimSpace(strings.ToLower(l)) == "## queue" {
			queueIdx = i
			break
		}
	}
	inQueue := queueIdx == -1
	var currentHeading string
	var currentBlock []string

	flush := func() {
		if currentHeading == "" {
			return
		}
		if strings.TrimSpace(strings.ToLower(currentHeading)) == "## queue" {
			return
		}
		block := strings.Join(currentBlock, "\n")
		trimmedHeading := strings.TrimSpace(strings.TrimPrefix(currentHeading, "##"))
		if trimmedHeading == "" {
			return
		}
		if !strings.Contains(block, "Done means:") {
			issues = append(issues, ValidationIssue{
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("unit %q missing Done means:", trimmedHeading),
				File:     file,
			})
		}
		if !strings.Contains(block, "Verify:") {
			issues = append(issues, ValidationIssue{
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("unit %q missing Verify:", trimmedHeading),
				File:     file,
			})
		} else {
			snippet := extractVerifySnippet(block)
			if snippet != "" {
				if msg := validateShellSnippet(snippet); msg != "" {
					issues = append(issues, ValidationIssue{
						Severity: SeverityError,
						Message:  fmt.Sprintf("unit %q Verify shell syntax invalid: %s", trimmedHeading, msg),
						File:     file,
					})
				}
			}
		}
	}

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") {
			if i == queueIdx {
				flush()
				currentHeading = trimmed
				currentBlock = nil
				inQueue = true
				continue
			}
			if !inQueue {
				continue
			}
			flush()
			currentHeading = trimmed
			currentBlock = nil
			continue
		}
		if currentHeading != "" && inQueue {
			currentBlock = append(currentBlock, line)
		}
	}
	flush()
	return issues
}

func extractVerifySnippet(block string) string {
	idx := strings.Index(block, "Verify:")
	if idx == -1 {
		return ""
	}
	after := block[idx+len("Verify:"):]

	afterTrim := strings.TrimSpace(after)
	if afterTrim == "" {
		return ""
	}

	if fenceIdx := strings.Index(after, "```"); fenceIdx != -1 {
		remaining := after[fenceIdx+3:]
		// skip language tag line
		if nl := strings.Index(remaining, "\n"); nl != -1 {
			remaining = remaining[nl+1:]
		}
		if endIdx := strings.Index(remaining, "```"); endIdx != -1 {
			return strings.TrimSpace(remaining[:endIdx])
		}
		return strings.TrimSpace(remaining)
	}

	lines := strings.Split(after, "\n")
	for _, l := range lines {
		t := strings.TrimSpace(strings.Trim(l, "` \t"))
		if t != "" && !strings.HasPrefix(t, "- [") {
			return t
		}
	}
	return strings.TrimSpace(afterTrim)
}

func validateShellSnippet(snippet string) string {
	snippet = strings.TrimSpace(snippet)
	if snippet == "" {
		return ""
	}
	if _, err := exec.LookPath("bash"); err != nil {
		return ""
	}
	tmp, err := os.CreateTemp("", "verify-*.sh")
	if err != nil {
		return ""
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString(snippet + "\n"); err != nil {
		tmp.Close()
		return ""
	}
	tmp.Close()
	cmd := exec.Command("bash", "-n", tmp.Name())
	out, err := cmd.CombinedOutput()
	if err != nil {
		return strings.TrimSpace(string(out))
	}
	return ""
}
