package internal

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func ownedQueue(body string) string {
	lines := strings.Split(body, "\n")
	var normalized []string
	openFence := ""
	for _, line := range lines {
		if consumeMarkdownFenceLine(&openFence, line) {
			normalized = append(normalized, line)
			continue
		}
		if strings.HasPrefix(line, "Done means:") {
			value := strings.TrimSpace(strings.TrimPrefix(line, "Done means:"))
			if value != "" {
				normalized = append(normalized,
					"Done means:",
					"- [outcome] "+value,
					"Scenarios:",
					"- [outcome] TestOutcome",
				)
				continue
			}
		}
		normalized = append(normalized, line)
	}
	return "Base: 1111111111111111111111111111111111111111\nBranch: litespec/test-change\n\n" + strings.Join(normalized, "\n")
}

const evidenceTestSHA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
const evidencePostTestSHA = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

func legacyEvidenceReceipt(verifyCmd string) string {
	return "Evidence:\n" +
		verifyCmd + "\n" +
		"sha: " + evidenceTestSHA + "\n" +
		"exit status: 0\n" +
		"```\n" +
		verifyCmd + " output\n" +
		"```\n" +
		"Evidence scope: this command exited 0 at " + evidenceTestSHA + "; nothing else is inferred.\n"
}

func evidenceReceipt(verifyCmd string) string {
	return redGreenEvidenceReceipt(
		verifyCmd,
		evidenceTestSHA,
		"1",
		verifyCmd+" missing outcome",
		evidencePostTestSHA,
		"0",
		verifyCmd+" output",
	)
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
		if !containsIssue(issues, "Base: ownership line must appear before") || !containsIssue(issues, "Branch: ownership line must appear before") {
			t.Fatalf("expected ownership placement errors, got %v", issues)
		}
	})

	t.Run("duplicate ownership after a heading fails", func(t *testing.T) {
		body := ownedQueue(wellFormed) + "\n## Notes\nBase: 2222222222222222222222222222222222222222\nBranch: litespec/other-change\n"
		_, issues := ValidateQueueBody(body, source)
		if !containsIssue(issues, "exactly one Base:") || !containsIssue(issues, "exactly one Branch:") {
			t.Fatalf("expected duplicate ownership errors, got %v", issues)
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
` + "```\necho two\n```\n" + strings.Replace(
		evidenceReceipt("echo two"),
		fixtureUnitDigest("echo two"),
		unitDigestFor("Good Two", "two", "echo two"),
		1,
	) + `- [x] done

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

func checkedUnit(verify, evidence string) string {
	return "## My outcome\nDone means: something\nVerify:\n```\n" + verify + "\n```\n" + evidence + "- [x] done\n"
}

func fixtureUnitDigest(verifyCmd string) string {
	return unitDigestFor("My outcome", "something", verifyCmd)
}

func unitDigestFor(heading, doneMeans, verifyCmd string) string {
	return unitContractDigest(queueUnit{
		Heading: heading,
		Body: []string{
			"Done means:",
			"- [outcome] " + doneMeans,
			"Scenarios:",
			"- [outcome] TestOutcome",
			"Verify:",
			"```",
			verifyCmd,
			"```",
		},
	})
}

func redGreenEvidenceReceipt(verifyCmd, preSHA, preStatus, preOutput, postSHA, postStatus, postOutput string) string {
	return "Evidence:\n" +
		verifyCmd + "\n" +
		"unit digest: " + fixtureUnitDigest(verifyCmd) + "\n" +
		"pre sha: " + preSHA + "\n" +
		"pre exit status: " + preStatus + "\n" +
		"```\n" + preOutput + "\n```\n" +
		"Pre-evidence scope: this command exited " + preStatus + " at " + preSHA + "; nothing else is inferred.\n" +
		"post sha: " + postSHA + "\n" +
		"post exit status: " + postStatus + "\n" +
		"```\n" + postOutput + "\n```\n" +
		"Post-evidence scope: this command exited " + postStatus + " at " + postSHA + "; nothing else is inferred.\n"
}

func TestValidateQueueRedGreenReceipt(t *testing.T) {
	source := "GH issue #1"
	valid := redGreenEvidenceReceipt(
		"echo hi",
		evidenceTestSHA,
		"1",
		"missing outcome",
		evidencePostTestSHA,
		"0",
		"hi",
	)

	t.Run("complete receipt passes", func(t *testing.T) {
		_, issues := ValidateQueueBody(ownedQueue(checkedUnit("echo hi", valid)), source)
		if len(issues) > 0 {
			t.Fatalf("expected no issues, got %v", issues)
		}
	})

	t.Run("receipt label in raw output passes", func(t *testing.T) {
		evidence := redGreenEvidenceReceipt(
			"echo hi",
			evidenceTestSHA,
			"1",
			"Evidence:\nmissing outcome",
			evidencePostTestSHA,
			"0",
			"hi",
		)
		_, issues := ValidateQueueBody(ownedQueue(checkedUnit("echo hi", evidence)), source)
		if len(issues) > 0 {
			t.Fatalf("expected no issues, got %v", issues)
		}
	})

	t.Run("checkbox in raw output passes", func(t *testing.T) {
		evidence := redGreenEvidenceReceipt(
			"echo hi",
			evidenceTestSHA,
			"1",
			"- [x] pre output",
			evidencePostTestSHA,
			"0",
			"- [x] post output",
		)
		_, issues := ValidateQueueBody(ownedQueue(checkedUnit("echo hi", evidence)), source)
		if len(issues) > 0 {
			t.Fatalf("expected no issues, got %v", issues)
		}
	})

	t.Run("longer fences preserve shorter fence and checkbox output", func(t *testing.T) {
		evidence := strings.Replace(
			valid,
			"```\nmissing outcome\n```",
			"````text\nbefore\n```\n- [x] pre raw output\n```\nafter\n````",
			1,
		)
		evidence = strings.Replace(
			evidence,
			"```\nhi\n```",
			"````text\nbefore\n```\n- [x] post raw output\n```\nafter\n````",
			1,
		)
		_, issues := ValidateQueueBody(ownedQueue(checkedUnit("echo hi", evidence)), source)
		if len(issues) > 0 {
			t.Fatalf("expected no issues, got %v", issues)
		}
	})

	t.Run("fenced Verify takes precedence over inline Verify", func(t *testing.T) {
		body := "## My outcome\nDone means: something\nVerify: `echo inline`\n```\necho fenced\n```\n" +
			evidenceReceipt("echo fenced") +
			"- [x] done\n"
		_, issues := ValidateQueueBody(ownedQueue(body), source)
		if len(issues) > 0 {
			t.Fatalf("expected no issues, got %v", issues)
		}
	})

	t.Run("comment heading followed by Verify command needs no Evidence label", func(t *testing.T) {
		body := ownedQueue(checkedUnit("echo hi", ""))
		units, issues := ValidateQueueBody(body, source)
		comment := "## My outcome\n" + strings.TrimPrefix(evidenceReceipt("echo hi"), "Evidence:\n")
		result := &ValidationResult{Valid: true}
		applyQueueIssues(result, "GitHub comments", units, issues, []string{comment})
		if len(result.Errors) > 0 {
			t.Fatalf("expected heading and command comment to satisfy receipt, got %v", result.Errors)
		}
	})

	t.Run("one comment receipt satisfies only one unit with the same Verify", func(t *testing.T) {
		body := ownedQueue(
			"## First outcome\nDone means: first thing\nVerify:\n```\necho hi\n```\n- [x] done\n\n" +
				"## Second outcome\nDone means: second thing\nVerify:\n```\necho hi\n```\n- [x] done\n",
		)
		units, issues := ValidateQueueBody(body, source)
		comment := "First outcome\nSecond outcome\n" + strings.Replace(
			evidenceReceipt("echo hi"),
			fixtureUnitDigest("echo hi"),
			unitDigestFor("First outcome", "first thing", "echo hi"),
			1,
		)
		result := &ValidationResult{Valid: true}
		applyQueueIssues(result, "GitHub comments", units, issues, []string{comment})
		if len(result.Errors) != 1 {
			t.Fatalf("expected one unmatched unit receipt, got %v", result.Errors)
		}
	})

	t.Run("one comment receipt satisfies only one duplicate heading occurrence", func(t *testing.T) {
		body := ownedQueue(
			"## Same outcome\nDone means: first thing\nVerify:\n```\necho hi\n```\n- [x] done\n\n" +
				"## Same outcome\nDone means: second thing\nVerify:\n```\necho hi\n```\n- [x] done\n",
		)
		units, issues := ValidateQueueBody(body, source)
		comment := "## Same outcome\n" + strings.Replace(
			strings.TrimPrefix(evidenceReceipt("echo hi"), "Evidence:\n"),
			fixtureUnitDigest("echo hi"),
			unitDigestFor("Same outcome", "first thing", "echo hi"),
			1,
		)
		result := &ValidationResult{Valid: true}
		applyQueueIssues(result, "GitHub comments", units, issues, []string{comment})
		if len(result.Errors) != 1 {
			t.Fatalf("expected one unmatched duplicate unit receipt, got %v", result.Errors)
		}
	})

	t.Run("receipt output structural tokens are ignored", func(t *testing.T) {
		evidence := redGreenEvidenceReceipt(
			"echo hi",
			evidenceTestSHA,
			"1",
			"## Phantom unit\nDone means: phantom\nVerify:\nBase: 2222222222222222222222222222222222222222\nBranch: litespec/phantom\n- [x] done",
			evidencePostTestSHA,
			"0",
			"## Another phantom\nDepends: Missing outcome\nBase: 3333333333333333333333333333333333333333\nBranch: litespec/another-phantom",
		)
		units, issues := ValidateQueueBody(ownedQueue(checkedUnit("echo hi", evidence)), source)
		if len(units) != 1 {
			t.Fatalf("expected one unit, got %d: %v", len(units), units)
		}
		if len(issues) > 0 {
			t.Fatalf("expected no issues, got %v", issues)
		}
	})

	tests := []struct {
		name        string
		evidence    string
		wantMessage string
	}{
		{
			name:        "legacy green-only receipt fails",
			evidence:    legacyEvidenceReceipt("echo hi"),
			wantMessage: "must include a `unit digest:` field",
		},
		{
			name: "green pre fails",
			evidence: redGreenEvidenceReceipt(
				"echo hi", evidenceTestSHA, "0", "hi", evidencePostTestSHA, "0", "hi",
			),
			wantMessage: "pre exit status must be non-zero",
		},
		{
			name: "failed post fails",
			evidence: redGreenEvidenceReceipt(
				"echo hi", evidenceTestSHA, "1", "missing", evidencePostTestSHA, "2", "broken",
			),
			wantMessage: "post exit status must be 0",
		},
		{
			name: "same shas fail",
			evidence: redGreenEvidenceReceipt(
				"echo hi", evidenceTestSHA, "1", "missing", evidenceTestSHA, "0", "hi",
			),
			wantMessage: "pre and post sha must differ",
		},
		{
			name: "edited command fails",
			evidence: redGreenEvidenceReceipt(
				"echo other", evidenceTestSHA, "1", "missing", evidencePostTestSHA, "0", "hi",
			),
			wantMessage: "quote the Verify command verbatim",
		},
		{
			name: "empty pre output fails",
			evidence: redGreenEvidenceReceipt(
				"echo hi", evidenceTestSHA, "1", "", evidencePostTestSHA, "0", "hi",
			),
			wantMessage: "pre raw command output",
		},
		{
			name: "empty post output fails",
			evidence: redGreenEvidenceReceipt(
				"echo hi", evidenceTestSHA, "1", "missing", evidencePostTestSHA, "0", "",
			),
			wantMessage: "post raw command output",
		},
		{
			name: "pre scope mismatch fails",
			evidence: strings.Replace(
				valid,
				"Pre-evidence scope: this command exited 1",
				"Pre-evidence scope: this command exited 2",
				1,
			),
			wantMessage: "pre scope line status must match",
		},
		{
			name: "post scope mismatch fails",
			evidence: strings.Replace(
				valid,
				"Post-evidence scope: this command exited 0 at "+evidencePostTestSHA,
				"Post-evidence scope: this command exited 0 at "+evidenceTestSHA,
				1,
			),
			wantMessage: "post scope line sha must match",
		},
		{
			name: "out of order fields fail",
			evidence: strings.Replace(
				valid,
				"pre sha: "+evidenceTestSHA+"\npre exit status: 1",
				"pre exit status: 1\npre sha: "+evidenceTestSHA,
				1,
			),
			wantMessage: "fields must appear in order",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, issues := ValidateQueueBody(ownedQueue(checkedUnit("echo hi", tt.evidence)), source)
			if !containsIssue(issues, tt.wantMessage) {
				t.Fatalf("expected %q error, got %v", tt.wantMessage, issues)
			}
		})
	}
}

func TestCheckedUnitEvidence(t *testing.T) {
	source := "GH issue #1"

	t.Run("unchecked unit needs no evidence", func(t *testing.T) {
		body := "## My outcome\nDone means: something\nVerify:\n```\necho hi\n```\n- [ ] pending\n"
		_, issues := ValidateQueueBody(ownedQueue(body), source)
		if len(issues) > 0 {
			t.Fatalf("expected no issues, got %v", issues)
		}
	})

	t.Run("checked unit with full receipt passes", func(t *testing.T) {
		body := checkedUnit("echo hi", evidenceReceipt("echo hi"))
		_, issues := ValidateQueueBody(ownedQueue(body), source)
		if len(issues) > 0 {
			t.Fatalf("expected no issues, got %v", issues)
		}
	})

	t.Run("checked unit missing evidence fails", func(t *testing.T) {
		body := checkedUnit("echo hi", "")
		_, issues := ValidateQueueBody(ownedQueue(body), source)
		if !containsIssue(issues, "missing Evidence receipt") {
			t.Fatalf("expected missing Evidence receipt, got %v", issues)
		}
	})

	t.Run("prose sticker fails", func(t *testing.T) {
		body := checkedUnit("echo hi", "Evidence: verified at abc123\n")
		_, issues := ValidateQueueBody(ownedQueue(body), source)
		if !containsIssue(issues, "must quote the Verify command verbatim") {
			t.Fatalf("expected verbatim command error, got %v", issues)
		}
	})

	t.Run("short sha fails", func(t *testing.T) {
		body := checkedUnit("echo hi", redGreenEvidenceReceipt(
			"echo hi", "abc123", "1", "missing", evidencePostTestSHA, "0", "hi",
		))
		_, issues := ValidateQueueBody(ownedQueue(body), source)
		if !containsIssue(issues, "pre sha must be a full") {
			t.Fatalf("expected sha error, got %v", issues)
		}
	})

	t.Run("empty fence fails", func(t *testing.T) {
		body := checkedUnit("echo hi", redGreenEvidenceReceipt(
			"echo hi", evidenceTestSHA, "1", "", evidencePostTestSHA, "0", "hi",
		))
		_, issues := ValidateQueueBody(ownedQueue(body), source)
		if !containsIssue(issues, "pre raw command output") {
			t.Fatalf("expected empty fence error, got %v", issues)
		}
	})

	t.Run("checked unit with failed evidence fails", func(t *testing.T) {
		body := checkedUnit("echo hi", redGreenEvidenceReceipt(
			"echo hi", evidenceTestSHA, "1", "missing", evidencePostTestSHA, "1", "FAIL",
		))
		_, issues := ValidateQueueBody(ownedQueue(body), source)
		if !containsIssue(issues, "post exit status must be 0") {
			t.Fatalf("expected non-zero exit status error, got %v", issues)
		}
	})

	t.Run("checked unit with explicit no output passes", func(t *testing.T) {
		body := checkedUnit("echo hi", redGreenEvidenceReceipt(
			"echo hi", evidenceTestSHA, "1", "<no output>", evidencePostTestSHA, "0", "<no output>",
		))
		_, issues := ValidateQueueBody(ownedQueue(body), source)
		if len(issues) > 0 {
			t.Fatalf("expected valid evidence, got %v", issues)
		}
	})

	t.Run("edited command fails", func(t *testing.T) {
		body := checkedUnit("echo hi", evidenceReceipt("echo other"))
		_, issues := ValidateQueueBody(ownedQueue(body), source)
		if !containsIssue(issues, "must quote the Verify command verbatim") {
			t.Fatalf("expected verbatim command error, got %v", issues)
		}
	})

	t.Run("scope sha mismatch fails", func(t *testing.T) {
		evidence := strings.Replace(
			evidenceReceipt("echo hi"),
			"Pre-evidence scope: this command exited 1 at "+evidenceTestSHA,
			"Pre-evidence scope: this command exited 1 at "+evidencePostTestSHA,
			1,
		)
		body := checkedUnit("echo hi", evidence)
		_, issues := ValidateQueueBody(ownedQueue(body), source)
		if !containsIssue(issues, "pre scope line sha must match pre sha") {
			t.Fatalf("expected sha mismatch, got %v", issues)
		}
	})

	t.Run("scope status mismatch fails", func(t *testing.T) {
		evidence := strings.Replace(
			evidenceReceipt("echo hi"),
			"Post-evidence scope: this command exited 0",
			"Post-evidence scope: this command exited 1",
			1,
		)
		body := checkedUnit("echo hi", evidence)
		_, issues := ValidateQueueBody(ownedQueue(body), source)
		if !containsIssue(issues, "post scope line status must match post exit status") {
			t.Fatalf("expected status mismatch, got %v", issues)
		}
	})

	t.Run("inline verify quoted in receipt passes", func(t *testing.T) {
		body := "## My outcome\nDone means: something\nVerify: `go test ./internal/ -run TestX`\n" + evidenceReceipt("go test ./internal/ -run TestX") + "- [x] done\n"
		_, issues := ValidateQueueBody(ownedQueue(body), source)
		if len(issues) > 0 {
			t.Fatalf("expected no issues, got %v", issues)
		}
	})

	t.Run("evidence label inside Verify does not satisfy", func(t *testing.T) {
		body := checkedUnit("printf 'Evidence:\\n'", "")
		_, issues := ValidateQueueBody(ownedQueue(body), source)
		if !containsIssue(issues, "missing Evidence receipt") {
			t.Fatalf("expected missing receipt, got %v", issues)
		}
	})
}

func TestGHCommentEvidenceReceipt(t *testing.T) {
	root := t.TempDir()
	body := ownedQueue("## My outcome\nDone means: something\nVerify:\n```\necho hi\n```\n- [x] done\n")

	t.Run("sticker comment does not satisfy", func(t *testing.T) {
		issue := ghIssue{
			Number: 9,
			Title:  "Add auth",
			Body:   body,
			URL:    "https://github.com/bermudi/litespec/issues/9",
			Comments: []struct {
				Body string `json:"body"`
			}{{Body: "My outcome\nEvidence: verified at abc123"}},
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

		result, err := ValidateGHIssueByNumber(root, 9)
		if err != nil {
			t.Fatalf("ValidateGHIssueByNumber: %v", err)
		}
		if result.Valid {
			t.Fatalf("expected Valid false, got true")
		}
		if !containsIssue(result.Errors, "missing Evidence receipt") && !containsIssue(result.Errors, "Evidence receipt") {
			t.Fatalf("expected receipt error, got %v", result.Errors)
		}
	})

	t.Run("full receipt comment satisfies", func(t *testing.T) {
		comment := "My outcome\n" + evidenceReceipt("echo hi")
		issue := ghIssue{
			Number: 9,
			Title:  "Add auth",
			Body:   body,
			URL:    "https://github.com/bermudi/litespec/issues/9",
			Comments: []struct {
				Body string `json:"body"`
			}{{Body: comment}},
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

		result, err := ValidateGHIssueByNumber(root, 9)
		if err != nil {
			t.Fatalf("ValidateGHIssueByNumber: %v", err)
		}
		if !result.Valid {
			t.Fatalf("expected Valid true, got false: %v", result.Errors)
		}
	})

	t.Run("legacy receipt comment does not satisfy", func(t *testing.T) {
		if commentSatisfiesEvidence("My outcome", "echo hi", fixtureUnitDigest("echo hi"), []string{
			"My outcome\n" + legacyEvidenceReceipt("echo hi"),
		}) {
			t.Fatal("expected legacy comment receipt to fail")
		}
	})

	t.Run("overlapping heading does not satisfy", func(t *testing.T) {
		if commentSatisfiesEvidence("My", "echo hi", fixtureUnitDigest("echo hi"), []string{
			"My outcome\n" + evidenceReceipt("echo hi"),
		}) {
			t.Fatal("expected non-exact heading mention to fail")
		}
	})

	t.Run("heading in raw output does not satisfy", func(t *testing.T) {
		comment := "Other outcome\n" + redGreenEvidenceReceipt(
			"echo hi",
			evidenceTestSHA,
			"1",
			"My outcome",
			evidencePostTestSHA,
			"0",
			"hi",
		)
		if commentSatisfiesEvidence("My outcome", "echo hi", fixtureUnitDigest("echo hi"), []string{comment}) {
			t.Fatal("expected heading inside raw output not to associate the comment")
		}
	})
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

func TestQueueUnitScenarioMapping(t *testing.T) {
	source := "GH issue #1"
	unit := func(doneMeans, scenarios string) string {
		return ownedQueue("## My outcome\nDone means:\n" + doneMeans + "Scenarios:\n" + scenarios + "Verify: `go test ./internal -run TestMyOutcome`\n- [ ] pending\n")
	}

	t.Run("complete mapping passes", func(t *testing.T) {
		_, issues := ValidateQueueBody(unit(
			"- [timeout] returns on timeout\n- [cleanup] removes temporary state\n",
			"- [timeout] TestTimeout\n- [cleanup] TestCleanup\n",
		), source)
		if len(issues) > 0 {
			t.Fatalf("expected complete mapping to pass, got %v", issues)
		}
	})

	tests := []struct {
		name        string
		doneMeans   string
		scenarios   string
		wantMessage string
	}{
		{
			name:        "clause without identifier fails",
			doneMeans:   "- returns on timeout\n",
			scenarios:   "- [timeout] TestTimeout\n",
			wantMessage: "identified Done means clause",
		},
		{
			name:        "duplicate clause identifier fails",
			doneMeans:   "- [timeout] returns on timeout\n- [timeout] stops work\n",
			scenarios:   "- [timeout] TestTimeout\n",
			wantMessage: `duplicate Done means clause ID "timeout"`,
		},
		{
			name:        "unmapped clause fails",
			doneMeans:   "- [timeout] returns on timeout\n- [cleanup] removes temporary state\n",
			scenarios:   "- [timeout] TestTimeout\n",
			wantMessage: `Done means clause ID "cleanup" has no scenario mapping`,
		},
		{
			name:        "unknown clause mapping fails",
			doneMeans:   "- [timeout] returns on timeout\n",
			scenarios:   "- [timeout] TestTimeout\n- [cleanup] TestCleanup\n",
			wantMessage: `scenario mapping references unknown Done means clause ID "cleanup"`,
		},
		{
			name:        "unnamed scenario fails",
			doneMeans:   "- [timeout] returns on timeout\n",
			scenarios:   "- [timeout]\n",
			wantMessage: `scenario mapping for clause ID "timeout" must name a test scenario`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, issues := ValidateQueueBody(unit(tt.doneMeans, tt.scenarios), source)
			if !containsIssue(issues, tt.wantMessage) {
				t.Fatalf("expected %q, got %v", tt.wantMessage, issues)
			}
		})
	}
}

func TestScenarioMappingAffectsUnitDigest(t *testing.T) {
	unit := queueUnit{
		Heading: "My outcome",
		Body: []string{
			"Done means:",
			"- [timeout] returns on timeout",
			"Scenarios:",
			"- [timeout] TestTimeout",
			"Verify: `go test ./internal -run TestTimeout`",
		},
	}
	changedClause := unit
	changedClause.Body = append([]string(nil), unit.Body...)
	changedClause.Body[1] = "- [timeout] reports timeout"
	changedScenario := unit
	changedScenario.Body = append([]string(nil), unit.Body...)
	changedScenario.Body[3] = "- [timeout] TestTimeoutResult"

	baseDigest := unitContractDigest(unit)
	if got := unitContractDigest(changedClause); got == baseDigest {
		t.Fatal("changing an identified Done means clause did not change the unit digest")
	}
	if got := unitContractDigest(changedScenario); got == baseDigest {
		t.Fatal("changing a scenario mapping did not change the unit digest")
	}
}

func TestValidateUnitContractDigest(t *testing.T) {
	source := "GH issue #1"

	// Hand-computed canonical serialization of the fixture unit: heading,
	// Done means, Scenarios, Verify content — each length-prefixed, in that order.
	doneMeans := "- [outcome] something"
	scenarios := "- [outcome] TestOutcome"
	canonical := "10:My outcome" +
		fmt.Sprintf("%d:%s", len(doneMeans), doneMeans) +
		fmt.Sprintf("%d:%s", len(scenarios), scenarios) +
		fmt.Sprintf("%d:echo hi", len("echo hi"))
	sum := sha256.Sum256([]byte(canonical))
	goodDigest := hex.EncodeToString(sum[:])

	if computed := fixtureUnitDigest("echo hi"); computed != goodDigest {
		t.Fatalf("unitContractDigest = %s, want hand-computed %s", computed, goodDigest)
	}

	digestReceipt := func(digest string) string {
		return strings.Replace(
			evidenceReceipt("echo hi"),
			fixtureUnitDigest("echo hi"),
			digest,
			1,
		)
	}

	t.Run("complete receipt with matching digest passes", func(t *testing.T) {
		_, issues := ValidateQueueBody(ownedQueue(checkedUnit("echo hi", digestReceipt(goodDigest))), source)
		if len(issues) > 0 {
			t.Fatalf("expected no issues, got %v", issues)
		}
	})

	t.Run("missing digest fails naming the unit", func(t *testing.T) {
		evidence := strings.Replace(evidenceReceipt("echo hi"), "unit digest: "+fixtureUnitDigest("echo hi")+"\n", "", 1)
		_, issues := ValidateQueueBody(ownedQueue(checkedUnit("echo hi", evidence)), source)
		if !containsIssue(issues, `checked unit "My outcome" Evidence receipt must include a `+"`unit digest:`") {
			t.Fatalf("expected missing unit digest error naming the unit, got %v", issues)
		}
	})

	t.Run("malformed digest fails", func(t *testing.T) {
		_, issues := ValidateQueueBody(ownedQueue(checkedUnit("echo hi", digestReceipt("XYZ"))), source)
		if !containsIssue(issues, "unit digest must be 64 lowercase hexadecimal characters") {
			t.Fatalf("expected malformed digest error, got %v", issues)
		}
	})

	t.Run("digest mismatch names expected and actual digests", func(t *testing.T) {
		otherDigest := unitDigestFor("My outcome", "something", "echo other")
		_, issues := ValidateQueueBody(ownedQueue(checkedUnit("echo hi", digestReceipt(otherDigest))), source)
		if !containsIssue(issues, "expected "+goodDigest+", actual "+otherDigest) {
			t.Fatalf("expected mismatch error naming both digests, got %v", issues)
		}
	})

	t.Run("editing Done means after evidence invalidates the receipt", func(t *testing.T) {
		body := checkedUnit("echo hi", digestReceipt(goodDigest))
		tampered := strings.Replace(body, "Done means: something", "Done means: moved goalpost", 1)
		_, issues := ValidateQueueBody(ownedQueue(tampered), source)
		if !containsIssue(issues, "unit digest mismatch") || !containsIssue(issues, "actual "+goodDigest) {
			t.Fatalf("expected digest mismatch after contract edit, got %v", issues)
		}
	})

	t.Run("optional contract fields are covered by the digest", func(t *testing.T) {
		withConstraints := strings.Replace(
			checkedUnit("echo hi", digestReceipt(goodDigest)),
			"Done means: something",
			"Done means: something\nConstraints: keep boundaries",
			1,
		)
		constraintsCanonical := "10:My outcome15:keep boundaries" + canonical[len("10:My outcome"):]
		constraintsSum := sha256.Sum256([]byte(constraintsCanonical))
		constraintsDigest := hex.EncodeToString(constraintsSum[:])
		_, issues := ValidateQueueBody(ownedQueue(withConstraints), source)
		if !containsIssue(issues, "expected "+constraintsDigest+", actual "+goodDigest) {
			t.Fatalf("expected Constraints to be covered by digest, got %v", issues)
		}
		updated := strings.Replace(withConstraints, "unit digest: "+goodDigest, "unit digest: "+constraintsDigest, 1)
		_, issues = ValidateQueueBody(ownedQueue(updated), source)
		if len(issues) > 0 {
			t.Fatalf("expected updated digest to pass, got %v", issues)
		}
	})

	t.Run("digest excludes checkbox and evidence content", func(t *testing.T) {
		base := queueUnit{
			Heading: "My outcome",
			Body:    []string{"Done means: something", "Verify:", "```", "echo hi", "```"},
		}
		mutable := queueUnit{
			Heading: "My outcome",
			Body: []string{
				"Done means: something",
				"Verify:",
				"```",
				"echo hi",
				"```",
				"- [x] done",
				"Evidence:\nwhatever mutable trail",
			},
		}
		if unitContractDigest(base) != unitContractDigest(mutable) {
			t.Fatalf("digest changed for mutable-only differences")
		}
	})

	t.Run("identity-bearing comment receipt requires matching digest", func(t *testing.T) {
		units, issues := ValidateQueueBody(ownedQueue(checkedUnit("echo hi", "")), source)
		comment := "## My outcome\n" + digestReceipt(goodDigest)
		result := &ValidationResult{Valid: true}
		applyQueueIssues(result, "GitHub comments", units, issues, []string{comment})
		if len(result.Errors) != 0 {
			t.Fatalf("expected matching comment receipt to satisfy validation, got %v", result.Errors)
		}

		staleComment := "## My outcome\n" + digestReceipt(strings.Repeat("a", 64))
		result = &ValidationResult{Valid: true}
		applyQueueIssues(result, "GitHub comments", units, issues, []string{staleComment})
		if len(result.Errors) == 0 {
			t.Fatalf("expected stale comment digest to leave receipt unsatisfied")
		}
	})

	t.Run("local queue file uses the same canonicalization", func(t *testing.T) {
		root := t.TempDir()
		path := filepath.Join(root, "specs", "queues", "digest-fixture.md")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		writeFile := func(content string) {
			t.Helper()
			if err := os.WriteFile(path, []byte(ownedQueue(content)), 0o644); err != nil {
				t.Fatalf("write queue fixture: %v", err)
			}
		}

		writeFile(strings.Replace(
			checkedUnit("echo hi", digestReceipt(goodDigest)),
			"Done means: something", "Done means: moved goalpost", 1,
		))
		result, err := ValidateQueueFile(path)
		if err != nil {
			t.Fatalf("ValidateQueueFile: %v", err)
		}
		if result.Valid || !containsIssue(result.Errors, "unit digest mismatch") {
			t.Fatalf("expected local queue to reject tampered contract, got %+v", result)
		}
	})
}
