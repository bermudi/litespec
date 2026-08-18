package internal

import (
	"strings"
	"testing"
)

func TestValidateGHIssueQueue(t *testing.T) {
	source := "GH issue #1"

	wellFormed := `## My outcome
Done means: something
Verify:
` + "```bash\necho hi\n```\n" + `- [ ] pending
`

	missingDoneMeans := `## My outcome
Verify:
` + "```bash\necho hi\n```\n" + `- [ ] pending
`

	missingVerify := `## My outcome
Done means: something
- [ ] pending
`

	missingFence := `## My outcome
Done means: something
Verify:
- [ ] pending
`

	missingCheckbox := `## My outcome
Done means: something
Verify:
` + "```bash\necho hi\n```\n"

	emptyHeading := `## 
Done means: something
Verify:
` + "```bash\necho hi\n```\n" + `- [ ] pending
`

	t.Run("well-formed fixture passes", func(t *testing.T) {
		issues := ValidateQueueBody(wellFormed, source)
		if len(issues) > 0 {
			t.Fatalf("expected no issues, got %d: %v", len(issues), issues)
		}
	})

	t.Run("missing Done means", func(t *testing.T) {
		issues := ValidateQueueBody(missingDoneMeans, source)
		if !containsIssue(issues, "missing Done means:") {
			t.Fatalf("expected error containing 'missing Done means:', got %v", issues)
		}
	})

	t.Run("missing Verify", func(t *testing.T) {
		issues := ValidateQueueBody(missingVerify, source)
		if !containsIssue(issues, "missing Verify:") {
			t.Fatalf("expected error containing 'missing Verify:', got %v", issues)
		}
	})

	t.Run("Verify without fenced block", func(t *testing.T) {
		issues := ValidateQueueBody(missingFence, source)
		if !containsIssue(issues, "not followed by fenced code block") {
			t.Fatalf("expected error containing 'not followed by fenced code block', got %v", issues)
		}
	})

	t.Run("missing checkbox", func(t *testing.T) {
		issues := ValidateQueueBody(missingCheckbox, source)
		if !containsIssue(issues, "missing checkbox") {
			t.Fatalf("expected error containing 'missing checkbox', got %v", issues)
		}
	})

	t.Run("empty heading", func(t *testing.T) {
		issues := ValidateQueueBody(emptyHeading, source)
		if !containsIssue(issues, "empty unit heading") {
			t.Fatalf("expected error containing 'empty unit heading', got %v", issues)
		}
	})

	t.Run("multiple units", func(t *testing.T) {
		body := `## Good One
Done means: one
Verify:
` + "```\necho one\n```\n" + `- [ ] pending

## Good Two
Done means: two
Verify:
` + "```\necho two\n```\n" + `- [x] done

## Bad One
Verify:
` + "```\necho bad\n```\n" + `- [ ] pending
`
		issues := ValidateQueueBody(body, source)
		if len(issues) != 1 {
			t.Fatalf("expected 1 error for the malformed unit, got %d: %v", len(issues), issues)
		}
		if !strings.Contains(issues[0].Message, "missing Done means:") {
			t.Fatalf("expected missing Done means: error for Bad One, got %v", issues)
		}
	})

	t.Run("non-unit content ignored", func(t *testing.T) {
		body := `This is the proposal.
More design text.

` + wellFormed
		issues := ValidateQueueBody(body, source)
		if len(issues) > 0 {
			t.Fatalf("expected no issues, got %d: %v", len(issues), issues)
		}
	})
}

func containsIssue(issues []ValidationIssue, substr string) bool {
	for _, issue := range issues {
		if strings.Contains(issue.Message, substr) {
			return true
		}
	}
	return false
}
