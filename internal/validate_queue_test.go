package internal

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func ownedQueue(body string) string {
	return "Base: 1111111111111111111111111111111111111111\nBranch: litespec/test-change\n\n" + body
}

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
		_, issues := ValidateQueueBody(ownedQueue(wellFormed), source)
		if len(issues) > 0 {
			t.Fatalf("expected no issues, got %d: %v", len(issues), issues)
		}
	})

	t.Run("missing ownership lines fail", func(t *testing.T) {
		_, issues := ValidateQueueBody(wellFormed, source)
		if !containsIssue(issues, "exactly one Base:") || !containsIssue(issues, "exactly one Branch:") {
			t.Fatalf("expected missing ownership errors, got %v", issues)
		}
	})

	t.Run("duplicate ownership lines fail", func(t *testing.T) {
		body := ownedQueue(wellFormed)
		body = "Base: 2222222222222222222222222222222222222222\nBranch: litespec/other-change\n" + body
		_, issues := ValidateQueueBody(body, source)
		if !containsIssue(issues, "exactly one Base:") || !containsIssue(issues, "exactly one Branch:") {
			t.Fatalf("expected duplicate ownership errors, got %v", issues)
		}
	})

	t.Run("malformed ownership lines fail", func(t *testing.T) {
		body := "Base: short\nBranch: feature/not-litespec\n\n" + wellFormed
		_, issues := ValidateQueueBody(body, source)
		if !containsIssue(issues, "full 40- or 64-character") || !containsIssue(issues, "litespec/<kebab-change-name>") {
			t.Fatalf("expected malformed ownership errors, got %v", issues)
		}
	})

	t.Run("ownership after a heading fails", func(t *testing.T) {
		body := "## Proposal\n\nBase: 1111111111111111111111111111111111111111\nBranch: litespec/test-change\n\n" + wellFormed
		_, issues := ValidateQueueBody(body, source)
		if !containsIssue(issues, "exactly one Base:") || !containsIssue(issues, "exactly one Branch:") {
			t.Fatalf("expected ownership placement errors, got %v", issues)
		}
	})

	t.Run("missing Done means", func(t *testing.T) {
		_, issues := ValidateQueueBody(ownedQueue(missingDoneMeans), source)
		if !containsIssue(issues, "missing Done means:") {
			t.Fatalf("expected error containing 'missing Done means:', got %v", issues)
		}
	})

	t.Run("missing Verify", func(t *testing.T) {
		_, issues := ValidateQueueBody(ownedQueue(missingVerify), source)
		if !containsIssue(issues, "missing Verify:") {
			t.Fatalf("expected error containing 'missing Verify:', got %v", issues)
		}
	})

	t.Run("Verify without fenced block", func(t *testing.T) {
		_, issues := ValidateQueueBody(ownedQueue(missingFence), source)
		if !containsIssue(issues, "not followed by a command or fenced code block") {
			t.Fatalf("expected error containing 'not followed by a command or fenced code block', got %v", issues)
		}
	})

	t.Run("inline Verify command passes", func(t *testing.T) {
		body := "## My outcome\nDone means: something\nVerify: `go test ./internal/ -run TestX` and a fixture asserts errors\n- [ ] pending\n"
		units, issues := ValidateQueueBody(ownedQueue(body), source)
		if len(issues) > 0 {
			t.Fatalf("expected no issues, got %d: %v", len(issues), issues)
		}
		if len(units) != 1 {
			t.Fatalf("expected 1 unit, got %d", len(units))
		}
	})

	t.Run("inline Verify with empty content fails", func(t *testing.T) {
		body := "## My outcome\nDone means: something\nVerify:   \n- [ ] pending\n"
		_, issues := ValidateQueueBody(ownedQueue(body), source)
		if !containsIssue(issues, "not followed by a command or fenced code block") {
			t.Fatalf("expected error for empty Verify content, got %v", issues)
		}
	})

	t.Run("empty inline Verify command fails", func(t *testing.T) {
		body := "## My outcome\nDone means: something\nVerify: ``\n- [ ] pending\n"
		_, issues := ValidateQueueBody(ownedQueue(body), source)
		if !containsIssue(issues, "not followed by a command or fenced code block") {
			t.Fatalf("expected error for empty inline Verify command, got %v", issues)
		}
	})

	t.Run("unterminated Verify fence fails", func(t *testing.T) {
		body := "## My outcome\nDone means: something\nVerify:\n```bash\necho hi\n- [ ] pending\n"
		_, issues := ValidateQueueBody(ownedQueue(body), source)
		if !containsIssue(issues, "not followed by a command or fenced code block") {
			t.Fatalf("expected error for unterminated Verify fence, got %v", issues)
		}
	})

	t.Run("inline Verify without backtick command fails", func(t *testing.T) {
		body := "## My outcome\nDone means: something\nVerify: TODO write this\n- [ ] pending\n"
		_, issues := ValidateQueueBody(ownedQueue(body), source)
		if !containsIssue(issues, "not followed by a command or fenced code block") {
			t.Fatalf("expected error for non-backtick Verify content, got %v", issues)
		}
	})

	t.Run("missing checkbox", func(t *testing.T) {
		_, issues := ValidateQueueBody(ownedQueue(missingCheckbox), source)
		if !containsIssue(issues, "missing checkbox") {
			t.Fatalf("expected error containing 'missing checkbox', got %v", issues)
		}
	})

	t.Run("checkbox text in a Verify block does not count", func(t *testing.T) {
		body := "## My outcome\nDone means: something\nVerify:\n```bash\necho '- [ ]'\n```\n"
		_, issues := ValidateQueueBody(ownedQueue(body), source)
		if !containsIssue(issues, "missing checkbox") {
			t.Fatalf("expected missing checkbox error, got %v", issues)
		}
	})

	t.Run("empty heading", func(t *testing.T) {
		_, issues := ValidateQueueBody(ownedQueue(emptyHeading), source)
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
		_, issues := ValidateQueueBody(ownedQueue(body), source)
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
		_, issues := ValidateQueueBody(ownedQueue(body), source)
		if len(issues) > 0 {
			t.Fatalf("expected no issues, got %d: %v", len(issues), issues)
		}
	})

	t.Run("prose ## sections skipped, units validated", func(t *testing.T) {
		body := `## Proposal

We are doing X because Y.

## Design

- Some bullet about design.
- Another bullet.

## Not doing

Nothing here.

## Queue

## Real unit
Done means: it works
Verify:
` + "```bash\necho hi\n```\n" + `- [ ] pending

## Spec draft

The spec goes here. It mentions ` + "`Verify:`" + ` inline but no line starts with it.
`
		units, issues := ValidateQueueBody(ownedQueue(body), source)
		if len(issues) > 0 {
			t.Fatalf("expected no issues, got %d: %v", len(issues), issues)
		}
		if len(units) != 1 {
			t.Fatalf("expected 1 unit, got %d: %v", len(units), units)
		}
		if units[0].Heading != "Real unit" {
			t.Fatalf("expected unit heading 'Real unit', got %q", units[0].Heading)
		}
	})

	t.Run("prose section with checkbox but no Done means/Verify is skipped", func(t *testing.T) {
		body := `## Proposal

- [ ] a todo in the proposal, not a unit

` + wellFormed
		units, issues := ValidateQueueBody(ownedQueue(body), source)
		if len(issues) > 0 {
			t.Fatalf("expected no issues, got %d: %v", len(issues), issues)
		}
		if len(units) != 1 {
			t.Fatalf("expected 1 unit, got %d", len(units))
		}
	})
}

func TestVerifyShellLint(t *testing.T) {
	source := "test-source"

	t.Run("bad syntax -> error", func(t *testing.T) {
		body := `## My outcome
Done means: something works
Verify:
` + "```bash\necho \"unclosed\n```\n" + `- [ ] pending
`
		_, issues := ValidateQueueBody(ownedQueue(body), source)
		if len(issues) == 0 {
			t.Fatalf("expected an error, got none")
		}
		found := false
		for _, issue := range issues {
			if strings.Contains(issue.Message, "Verify shell syntax error") {
				found = true
				if issue.Severity != SeverityError {
					t.Fatalf("expected error severity, got %s", issue.Severity)
				}
			}
		}
		if !found {
			t.Fatalf("expected error containing 'Verify shell syntax error', got %v", issues)
		}
	})

	t.Run("good syntax -> pass", func(t *testing.T) {
		body := `## My outcome
Done means: something works
Verify:
` + "```bash\necho hello\n```\n" + `- [ ] pending
`
		_, issues := ValidateQueueBody(ownedQueue(body), source)
		if len(issues) > 0 {
			t.Fatalf("expected no issues, got %d: %v", len(issues), issues)
		}
	})

	t.Run("bash absent -> warning", func(t *testing.T) {
		old := lookPathBash
		lookPathBash = func(string) (string, error) { return "", exec.ErrNotFound }
		defer func() { lookPathBash = old }()

		body := `## My outcome
Done means: something works
Verify:
` + "```bash\necho hello\n```\n" + `- [ ] pending
`
		_, issues := ValidateQueueBody(ownedQueue(body), source)
		if len(issues) == 0 {
			t.Fatalf("expected a warning, got none")
		}
		found := false
		for _, issue := range issues {
			if strings.Contains(issue.Message, "bash unavailable") {
				found = true
				if issue.Severity != SeverityWarning {
					t.Fatalf("expected warning severity, got %s", issue.Severity)
				}
			}
		}
		if !found {
			t.Fatalf("expected warning containing 'bash unavailable', got %v", issues)
		}

		emptyBody := `## My outcome
Done means: something works
Verify:
` + "```bash\n```\n" + `- [ ] pending
`
		_, issues = ValidateQueueBody(ownedQueue(emptyBody), source)
		if len(issues) == 0 {
			t.Fatalf("expected an error for empty block, got none")
		}
		found = false
		for _, issue := range issues {
			if strings.Contains(issue.Message, "Verify block is empty") {
				found = true
				if issue.Severity != SeverityError {
					t.Fatalf("expected error severity for empty block, got %s", issue.Severity)
				}
			}
		}
		if !found {
			t.Fatalf("expected error containing 'Verify block is empty', got %v", issues)
		}
	})
}

func TestLocalQueueFallback(t *testing.T) {
	wellFormed := `## Add auth
Done means: authentication is wired
Verify:
` + "```bash\necho ok\n```\n" + `- [ ] pending
`

	missingDoneMeans := `## Add auth
Verify:
` + "```bash\necho ok\n```\n" + `- [ ] pending
`
	wellFormed = ownedQueue(wellFormed)
	missingDoneMeans = ownedQueue(missingDoneMeans)

	t.Run("local queue validated when gh absent", func(t *testing.T) {
		root := t.TempDir()
		queuePath := filepath.Join(root, "specs", "queues", "add-auth.md")
		if err := os.MkdirAll(filepath.Dir(queuePath), 0755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(queuePath, []byte(wellFormed), 0644); err != nil {
			t.Fatalf("write queue: %v", err)
		}

		result, err := ValidateLocalQueues(root)
		if err != nil {
			t.Fatalf("ValidateLocalQueues: %v", err)
		}
		if len(result.Errors) > 0 {
			t.Fatalf("expected no errors, got %v", result.Errors)
		}
		if result.UnitsCount != 1 {
			t.Fatalf("expected UnitsCount 1, got %d", result.UnitsCount)
		}
	})

	t.Run("local queue with malformed unit", func(t *testing.T) {
		root := t.TempDir()
		queuePath := filepath.Join(root, "specs", "queues", "add-auth.md")
		if err := os.MkdirAll(filepath.Dir(queuePath), 0755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(queuePath, []byte(missingDoneMeans), 0644); err != nil {
			t.Fatalf("write queue: %v", err)
		}

		result, err := ValidateLocalQueues(root)
		if err != nil {
			t.Fatalf("ValidateLocalQueues: %v", err)
		}
		if !containsIssue(result.Errors, "missing Done means:") {
			t.Fatalf("expected error containing 'missing Done means:', got %v", result.Errors)
		}
	})

	t.Run("no specs/queues/ directory", func(t *testing.T) {
		root := t.TempDir()
		result, err := ValidateLocalQueues(root)
		if err != nil {
			t.Fatalf("ValidateLocalQueues: %v", err)
		}
		if len(result.Errors) > 0 {
			t.Fatalf("expected no errors, got %v", result.Errors)
		}
		if len(result.Warnings) > 0 {
			t.Fatalf("expected no warnings, got %v", result.Warnings)
		}
		if result.UnitsCount != 0 {
			t.Fatalf("expected UnitsCount 0, got %d", result.UnitsCount)
		}
	})

	t.Run("--queue path via ValidateQueueFile", func(t *testing.T) {
		root := t.TempDir()
		queuePath := filepath.Join(root, "queue.md")
		if err := os.WriteFile(queuePath, []byte(wellFormed), 0644); err != nil {
			t.Fatalf("write queue: %v", err)
		}

		result, err := ValidateQueueFile(queuePath)
		if err != nil {
			t.Fatalf("ValidateQueueFile: %v", err)
		}
		if len(result.Errors) > 0 {
			t.Fatalf("expected no errors, got %v", result.Errors)
		}
		if result.UnitsCount != 1 {
			t.Fatalf("expected UnitsCount 1, got %d", result.UnitsCount)
		}
	})

	t.Run("warning routing", func(t *testing.T) {
		root := t.TempDir()
		queuePath := filepath.Join(root, "queue.md")
		if err := os.WriteFile(queuePath, []byte(wellFormed), 0644); err != nil {
			t.Fatalf("write queue: %v", err)
		}

		old := lookPathBash
		lookPathBash = func(string) (string, error) { return "", exec.ErrNotFound }
		defer func() { lookPathBash = old }()

		result, err := ValidateQueueFile(queuePath)
		if err != nil {
			t.Fatalf("ValidateQueueFile: %v", err)
		}
		if !result.Valid {
			t.Fatalf("expected Valid true, got false")
		}
		if len(result.Errors) > 0 {
			t.Fatalf("expected no errors, got %v", result.Errors)
		}
		if !containsIssue(result.Warnings, "bash unavailable") {
			t.Fatalf("expected warning in Warnings containing 'bash unavailable', got %v", result.Warnings)
		}
	})

	t.Run("--issue N via ValidateGHIssueByNumber", func(t *testing.T) {
		root := t.TempDir()

		issue := ghIssue{
			Number: 42,
			Title:  "Add auth",
			Body:   wellFormed,
			URL:    "https://github.com/bermudi/litespec/issues/42",
		}
		data, err := json.Marshal(issue)
		if err != nil {
			t.Fatalf("marshal issue: %v", err)
		}

		oldLook := lookPathGh
		lookPathGh = func(string) (string, error) { return "gh", nil }
		defer func() { lookPathGh = oldLook }()

		oldView := ghIssueView
		ghIssueView = func(string, int) ([]byte, error) { return data, nil }
		defer func() { ghIssueView = oldView }()

		result, err := ValidateGHIssueByNumber(root, 42)
		if err != nil {
			t.Fatalf("ValidateGHIssueByNumber: %v", err)
		}
		if !result.Valid {
			t.Fatalf("expected Valid true, got false")
		}
		if len(result.Errors) > 0 {
			t.Fatalf("expected no errors, got %v", result.Errors)
		}
		if result.UnitsCount != 1 {
			t.Fatalf("expected UnitsCount 1, got %d", result.UnitsCount)
		}
	})

	t.Run("--issue N with malformed unit", func(t *testing.T) {
		root := t.TempDir()

		issue := ghIssue{
			Number: 7,
			Title:  "Add auth",
			Body:   missingDoneMeans,
			URL:    "https://github.com/bermudi/litespec/issues/7",
		}
		data, err := json.Marshal(issue)
		if err != nil {
			t.Fatalf("marshal issue: %v", err)
		}

		oldLook := lookPathGh
		lookPathGh = func(string) (string, error) { return "gh", nil }
		defer func() { lookPathGh = oldLook }()

		oldView := ghIssueView
		ghIssueView = func(string, int) ([]byte, error) { return data, nil }
		defer func() { ghIssueView = oldView }()

		result, err := ValidateGHIssueByNumber(root, 7)
		if err != nil {
			t.Fatalf("ValidateGHIssueByNumber: %v", err)
		}
		if result.Valid {
			t.Fatalf("expected Valid false, got true")
		}
		if !containsIssue(result.Errors, "missing Done means:") {
			t.Fatalf("expected error containing 'missing Done means:', got %v", result.Errors)
		}
		if result.UnitsCount != 1 {
			t.Fatalf("expected UnitsCount 1, got %d", result.UnitsCount)
		}
	})

	t.Run("--issue N gh absent returns error", func(t *testing.T) {
		root := t.TempDir()

		old := lookPathGh
		lookPathGh = func(string) (string, error) { return "", exec.ErrNotFound }
		defer func() { lookPathGh = old }()

		_, err := ValidateGHIssueByNumber(root, 42)
		if err == nil {
			t.Fatalf("expected error when gh is absent, got nil")
		}
		if !strings.Contains(err.Error(), "gh not available") {
			t.Fatalf("expected error containing 'gh not available', got %v", err)
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

func TestValidateGHIssueQueues_NoGitHubRemote(t *testing.T) {
	root := t.TempDir()
	result, err := ValidateGHIssueQueues(root)
	if err != nil {
		t.Fatalf("ValidateGHIssueQueues: %v", err)
	}
	if !result.Valid {
		t.Fatalf("expected Valid true, got false")
	}
	if len(result.Errors) > 0 {
		t.Fatalf("expected no errors, got %v", result.Errors)
	}
	if !containsIssue(result.Warnings, "issue queue not validated") {
		t.Fatalf("expected a warning about queue not validated, got %v", result.Warnings)
	}
	for _, w := range result.Warnings {
		if !w.StrictExempt {
			t.Fatalf("expected all queue-absence warnings to be StrictExempt, got %q", w.Message)
		}
	}
}

func TestValidateGHIssueQueues_JSONParseFailureWarns(t *testing.T) {
	root := t.TempDir()

	oldLook := lookPathGh
	lookPathGh = func(string) (string, error) { return "gh", nil }
	defer func() { lookPathGh = oldLook }()

	oldList := ghIssueList
	ghIssueList = func(string) ([]byte, error) { return []byte("not valid json"), nil }
	defer func() { ghIssueList = oldList }()

	result, err := ValidateGHIssueQueues(root)
	if err != nil {
		t.Fatalf("ValidateGHIssueQueues: %v", err)
	}
	if !result.Valid {
		t.Fatalf("expected Valid true, got false")
	}
	if len(result.Errors) > 0 {
		t.Fatalf("expected no errors, got %v", result.Errors)
	}
	if !containsIssue(result.Warnings, "parse gh issue list output") {
		t.Fatalf("expected warning about parse failure, got %v", result.Warnings)
	}
	for _, w := range result.Warnings {
		if !w.StrictExempt {
			t.Fatalf("expected parse-failure warning to be StrictExempt, got %q", w.Message)
		}
	}
}

func TestQueueDepends(t *testing.T) {
	source := "GH issue #1"

	unit := func(heading, depends, done string) string {
		dep := ""
		if depends != "" {
			dep = "Depends: " + depends + "\n"
		}
		return "## " + heading + "\n" +
			dep +
			"Done means: " + done + "\n" +
			"Verify:\n" + "```bash\necho " + heading + "\n```\n" +
			"- [ ] pending\n"
	}

	t.Run("Depends parsing", func(t *testing.T) {
		body := unit("Unit A", "Foo, Bar", "something") + "\n" +
			unit("Foo", "", "foo works") + "\n" +
			unit("Bar", "", "bar works")
		units, issues := ValidateQueueBody(ownedQueue(body), source)
		if len(issues) > 0 {
			t.Fatalf("expected no issues, got %d: %v", len(issues), issues)
		}
		if len(units) != 3 {
			t.Fatalf("expected 3 units, got %d", len(units))
		}
		if len(units[0].Depends) != 2 || units[0].Depends[0] != "Foo" || units[0].Depends[1] != "Bar" {
			t.Fatalf("expected Depends [Foo Bar], got %v", units[0].Depends)
		}
	})

	t.Run("dangling reference", func(t *testing.T) {
		body := unit("Unit A", "Nonexistent", "something") + "\n" + unit("Real Unit", "", "real works")
		_, issues := ValidateQueueBody(ownedQueue(body), source)
		if !containsIssue(issues, "depends on non-existent unit") {
			t.Fatalf("expected error for dangling dependency, got %v", issues)
		}
	})

	t.Run("valid reference", func(t *testing.T) {
		body := unit("Unit A", "Existing Unit", "something") + "\n" + unit("Existing Unit", "", "existing works")
		_, issues := ValidateQueueBody(ownedQueue(body), source)
		if len(issues) > 0 {
			t.Fatalf("expected no issues, got %d: %v", len(issues), issues)
		}
	})

	t.Run("no Depends passes", func(t *testing.T) {
		_, issues := ValidateQueueBody(ownedQueue(unit("Unit A", "", "something")), source)
		if len(issues) > 0 {
			t.Fatalf("expected no issues, got %d: %v", len(issues), issues)
		}
	})

	t.Run("multiple valid", func(t *testing.T) {
		body := unit("Unit A", "First Unit, Second Unit", "something") + "\n" +
			unit("First Unit", "", "first works") + "\n" +
			unit("Second Unit", "", "second works")
		_, issues := ValidateQueueBody(ownedQueue(body), source)
		if len(issues) > 0 {
			t.Fatalf("expected no issues, got %d: %v", len(issues), issues)
		}
	})
}
