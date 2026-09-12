package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
)

const issueCheckFixtureBody = `Base: 0000000000000000000000000000000000000004
Branch: litespec/issue-check-fixture

## First unit

Done means: it exists

Verify:
` + "```bash\necho first\n```\n" + `
- [x] pending

## Second unit

Done means: it ticks

Verify:
` + "```bash\necho second\n```\n" + `
- [ ] pending
`

const issueCheckTwinsFixtureBody = `Base: 0000000000000000000000000000000000000005
Branch: litespec/issue-check-fixture

## Twin unit

Done means: first twin

Verify:
` + "```bash\necho twin-one\n```\n" + `
- [ ] pending

## Twin unit

Done means: second twin

Verify:
` + "```bash\necho twin-two\n```\n" + `
- [ ] pending
`

const issueCheckAlreadyCheckedBody = `Base: 0000000000000000000000000000000000000006
Branch: litespec/issue-check-fixture

## Second unit

Done means: it ticks

Verify:
` + "```bash\necho second\n```\n" + `
- [x] pending
`

func issueCheckFakeView(t *testing.T, body string) func(string, int) ([]byte, error) {
	t.Helper()
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	payload := fmt.Sprintf(`{"number":42,"title":"t","body":%s,"url":"","comments":[]}`, bodyJSON)
	return func(string, int) ([]byte, error) { return []byte(payload), nil }
}

func issueCheckOwnershipLines(body string) []string {
	var out []string
	for _, line := range strings.Split(body, "\n") {
		if strings.HasPrefix(line, "Base:") || strings.HasPrefix(line, "Branch:") {
			out = append(out, line)
		}
	}
	return out
}

func issueCheckAssertOneCheckboxFlip(t *testing.T, before, after string) {
	t.Helper()
	beforeLines := strings.Split(before, "\n")
	afterLines := strings.Split(after, "\n")
	if len(beforeLines) != len(afterLines) {
		t.Fatalf("written body has %d lines, fetched body has %d — nothing but one checkbox flip may change", len(afterLines), len(beforeLines))
	}
	flipped := 0
	for i := range beforeLines {
		if beforeLines[i] == afterLines[i] {
			continue
		}
		flipped++
		want := strings.Replace(beforeLines[i], "- [ ]", "- [x]", 1)
		if want == beforeLines[i] || afterLines[i] != want {
			t.Fatalf("line %d changed to %q, want the single checkbox flip to %q", i+1, afterLines[i], want)
		}
	}
	if flipped != 1 {
		t.Fatalf("written body differs in %d lines, want exactly one checkbox flip", flipped)
	}
}

func TestIssueCheckTicksExactlyOneCheckbox(t *testing.T) {
	t.Run("managed tick changes exactly one checkbox and preserves ownership lines", func(t *testing.T) {
		var writtenBodies []string
		var writtenFiles []string
		originalView, originalEdit := ghIssueView, ghIssueEdit
		ghIssueView = issueCheckFakeView(t, issueCheckFixtureBody)
		ghIssueEdit = func(root string, number int, bodyFile string) ([]byte, error) {
			if root == "" {
				t.Error("the managed tick must run gh inside the project root")
			}
			if number != 42 {
				t.Errorf("gh issue edit targeted issue %d, want 42", number)
			}
			data, err := os.ReadFile(bodyFile)
			if err != nil {
				t.Fatalf("read written body file: %v", err)
			}
			writtenBodies = append(writtenBodies, string(data))
			writtenFiles = append(writtenFiles, bodyFile)
			return []byte("https://example.invalid/issues/42\n"), nil
		}
		defer func() { ghIssueView, ghIssueEdit = originalView, originalEdit }()

		result, err := IssueCheckTicksOneUnit("/repo", 42, "Second unit", 0)
		if err != nil {
			t.Fatalf("managed tick failed: %v", err)
		}
		if result.Number != 42 || result.Heading != "Second unit" || result.Occurrence != 1 {
			t.Fatalf("result = %+v, want issue 42, heading \"Second unit\", occurrence 1", result)
		}
		if len(writtenBodies) != 1 {
			t.Fatalf("issued %d writes, want exactly one", len(writtenBodies))
		}
		issueCheckAssertOneCheckboxFlip(t, issueCheckFixtureBody, writtenBodies[0])
		if !strings.Contains(writtenBodies[0], "- [x] pending") || strings.Contains(writtenBodies[0], "- [ ] pending") {
			t.Errorf("written body does not carry the flipped checkbox:\n%s", writtenBodies[0])
		}
		before, after := issueCheckOwnershipLines(issueCheckFixtureBody), issueCheckOwnershipLines(writtenBodies[0])
		if strings.Join(before, "\n") != strings.Join(after, "\n") {
			t.Errorf("ownership lines changed: %v -> %v", before, after)
		}
		for _, file := range writtenFiles {
			if _, err := os.Stat(file); err == nil {
				t.Errorf("body file %s survived the write; the tick must clean it up", file)
			}
		}

		writtenBodies = nil
		writtenFiles = nil
		ghIssueView = issueCheckFakeView(t, issueCheckTwinsFixtureBody)
		result, err = IssueCheckTicksOneUnit("/repo", 42, "Twin unit", 2)
		if err != nil {
			t.Fatalf("managed tick with explicit occurrence failed: %v", err)
		}
		if result.Occurrence != 2 {
			t.Fatalf("resolved occurrence %d, want 2", result.Occurrence)
		}
		if len(writtenBodies) != 1 {
			t.Fatalf("issued %d writes, want exactly one", len(writtenBodies))
		}
		issueCheckAssertOneCheckboxFlip(t, issueCheckTwinsFixtureBody, writtenBodies[0])
		if !strings.Contains(writtenBodies[0], "Done means: second twin") || !strings.Contains(writtenBodies[0], "- [x] pending") {
			t.Errorf("occurrence 2 tick must flip the second twin's checkbox:\n%s", writtenBodies[0])
		}
		_, rest, found := strings.Cut(writtenBodies[0], "## Twin unit\n")
		if !found {
			t.Fatalf("written body lost the twin headings:\n%s", writtenBodies[0])
		}
		firstTwin, secondTwin, found := strings.Cut(rest, "## Twin unit\n")
		if !found {
			t.Fatalf("written body lost the second twin:\n%s", writtenBodies[0])
		}
		if !strings.Contains(firstTwin, "- [ ] pending") || strings.Contains(firstTwin, "- [x] pending") {
			t.Errorf("the first twin must stay unchecked:\n%s", firstTwin)
		}
		if !strings.Contains(secondTwin, "- [x] pending") {
			t.Errorf("only the second twin may flip:\n%s", secondTwin)
		}
	})

	t.Run("managed tick round trip does not grow trailing newlines", func(t *testing.T) {
		stored := issueCheckTwinsFixtureBody
		var writtenBodies []string
		originalView, originalEdit := ghIssueView, ghIssueEdit
		defer func() { ghIssueView, ghIssueEdit = originalView, originalEdit }()
		// GitHub stores every written issue body with exactly one appended
		// trailing newline (measured on the live API: writing a body with
		// N trailing newlines stores N+1). Each uncompensated managed tick
		// therefore grows the stored body by one newline.
		ghIssueView = func(root string, number int) ([]byte, error) {
			bodyJSON, err := json.Marshal(stored)
			if err != nil {
				t.Fatal(err)
			}
			payload := fmt.Sprintf(`{"number":42,"title":"t","body":%s,"url":"","comments":[]}`, bodyJSON)
			return []byte(payload), nil
		}
		ghIssueEdit = func(root string, number int, bodyFile string) ([]byte, error) {
			data, err := os.ReadFile(bodyFile)
			if err != nil {
				t.Fatalf("read written body file: %v", err)
			}
			writtenBodies = append(writtenBodies, string(data))
			stored = string(data) + "\n"
			return []byte("https://example.invalid/issues/42\n"), nil
		}

		for tick := 1; tick <= 2; tick++ {
			fetched := stored
			writtenBodies = nil
			if _, err := IssueCheckTicksOneUnit("/repo", 42, "Twin unit", tick); err != nil {
				t.Fatalf("managed tick %d failed: %v", tick, err)
			}
			if len(writtenBodies) != 1 {
				t.Fatalf("managed tick %d issued %d writes, want exactly one", tick, len(writtenBodies))
			}
			issueCheckAssertOneCheckboxFlip(t, strings.TrimRight(fetched, "\n"), strings.TrimRight(writtenBodies[0], "\n"))
			if trailing := len(writtenBodies[0]) - len(strings.TrimRight(writtenBodies[0], "\n")); trailing != 1 {
				t.Fatalf("managed tick %d wrote %d trailing newlines; the platform appends one per write, so the written body must normalize to exactly one", tick, trailing)
			}
			if trailing := len(stored) - len(strings.TrimRight(stored, "\n")); trailing != 2 {
				t.Fatalf("after managed tick %d the stored body carries %d trailing newlines; the round trip must stabilize at the written one plus the platform's appended one", tick, trailing)
			}
		}
	})

	t.Run("refusals issue no write and leave the remote body untouched", func(t *testing.T) {
		var editCalls int
		originalView, originalEdit := ghIssueView, ghIssueEdit
		ghIssueEdit = func(root string, number int, bodyFile string) ([]byte, error) {
			editCalls++
			return []byte("https://example.invalid/issues/42\n"), nil
		}
		defer func() { ghIssueView, ghIssueEdit = originalView, originalEdit }()

		refuse := func(name, body, heading string, occurrence int, wantParts ...string) {
			t.Helper()
			editCalls = 0
			ghIssueView = issueCheckFakeView(t, body)
			_, err := IssueCheckTicksOneUnit("/repo", 42, heading, occurrence)
			if err == nil {
				t.Fatalf("%s: must refuse, got success", name)
			}
			for _, want := range wantParts {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("%s: refusal %q does not name %q", name, err.Error(), want)
				}
			}
			if editCalls != 0 {
				t.Errorf("%s: refusal issued %d writes, want none", name, editCalls)
			}
		}

		refuse("already-checked target", issueCheckAlreadyCheckedBody, "Second unit", 0, "already checked")
		refuse("unknown heading", issueCheckFixtureBody, "Missing unit", 0, "Missing unit")
		refuse("ambiguous heading", issueCheckTwinsFixtureBody, "Twin unit", 0, "occurrence")
		refuse("out-of-range occurrence", issueCheckTwinsFixtureBody, "Twin unit", 3, "occurrence")

		editCalls = 0
		ghIssueView = issueCheckFakeView(t, issueCheckFixtureBody)
		ghIssueEdit = func(root string, number int, bodyFile string) ([]byte, error) {
			editCalls++
			return []byte("gh: simulated failure\n"), fmt.Errorf("exit status 1")
		}
		_, err := IssueCheckTicksOneUnit("/repo", 42, "Second unit", 0)
		if err == nil {
			t.Fatalf("gh failure must be a visible refusal")
		}
		if !strings.Contains(err.Error(), "gh issue edit") {
			t.Errorf("gh-failure refusal must name the command: %v", err)
		}
		if editCalls != 1 {
			t.Errorf("gh failure issued %d write attempts, want exactly one with no retry", editCalls)
		}
	})
}
