package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/bermudi/litespec/v2/internal"
)

const receiptFixtureBody = `Base: 0000000000000000000000000000000000000002
Branch: litespec/receipt-fixture

## Only unit

Done means: it works

Verify:
` + "```bash\necho receipt-fixture\n```\n" + `
- [ ] pending
`

const receiptDuplicateFixtureBody = `Base: 0000000000000000000000000000000000000003
Branch: litespec/receipt-fixture

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

func writeReceiptQueueFixture(t *testing.T, root, body string) string {
	t.Helper()
	dir := filepath.Join(root, "specs", "queues")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "receipt-fixture.md")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeReceiptRunOutput(t *testing.T, root, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func receiptGitCommits(t *testing.T, root string, unrelated bool) (string, string) {
	t.Helper()
	git := func(args ...string) string {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		cmd.Env = append(os.Environ(), "HOME="+root)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	git("init", "-q", "-b", "main")
	git("config", "user.email", "receipt@example.com")
	git("config", "user.name", "receipt test")
	if err := os.WriteFile(filepath.Join(root, "seed.txt"), []byte("pre\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git("add", ".")
	git("commit", "-q", "-m", "pre")
	pre := git("rev-parse", "HEAD")
	if unrelated {
		git("checkout", "-q", "--orphan", "isolated")
	}
	if err := os.WriteFile(filepath.Join(root, "seed.txt"), []byte("post\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git("add", ".")
	git("commit", "-q", "-m", "post")
	post := git("rev-parse", "HEAD")
	if pre == post {
		t.Fatal("expected two distinct commits")
	}
	return pre, post
}

func receiptCommentFiles(t *testing.T, root string) []string {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(root, "receipt-*.md"))
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(matches)
	return matches
}

// receiptFakeGH installs a gh shim on a custom PATH that serves the given
// issue body; it returns the PATH value to use for the CLI invocation.
func receiptFakeGH(t *testing.T, issueNumber int, body string) string {
	t.Helper()
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	payload := fmt.Sprintf(`{"number":%d,"title":"t","body":%s,"url":"","comments":[]}`, issueNumber, bodyJSON)
	fakeBin := t.TempDir()
	script := "#!/bin/sh\ncat " + filepath.Join(fakeBin, "payload.json") + "\n"
	if err := os.WriteFile(filepath.Join(fakeBin, "gh"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fakeBin, "payload.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	return fakeBin + ":" + os.Getenv("PATH")
}

func receiptUnitDigest(t *testing.T, queuePath, heading string, occurrence int) string {
	t.Helper()
	lines, err := internal.DigestQueueUnits("", 0, queuePath)
	if err != nil {
		t.Fatalf("digest queue fixture: %v", err)
	}
	for _, line := range lines {
		if line.Heading == heading && line.Occurrence == occurrence {
			return line.Digest
		}
	}
	t.Fatalf("fixture has no unit %q occurrence %d", heading, occurrence)
	return ""
}

func TestReceiptCommandEmitsCommentFiles(t *testing.T) {
	t.Run("ambiguous heading refused without explicit occurrence", func(t *testing.T) {
		bin, root := setupCLITest(t)
		queuePath := writeReceiptQueueFixture(t, root, receiptDuplicateFixtureBody)

		base := func(heading string, extra ...string) []string {
			args := []string{
				"receipt", "--queue", queuePath, "--heading", heading,
				"--pre-sha", strings.Repeat("a", 40), "--pre-status", "1",
				"--pre-out", "pre.txt", "--post-sha", strings.Repeat("b", 40),
				"--post-out", "post.txt",
			}
			return append(args, extra...)
		}

		out, code := runCLI(t, bin, root, base("Twin unit")...)
		if code == 0 {
			t.Fatalf("ambiguous heading must be refused, got exit 0: %s", out)
		}
		if !strings.Contains(out, "occurrence") {
			t.Errorf("refusal does not name the ambiguity: %s", out)
		}
		if files := receiptCommentFiles(t, root); len(files) != 0 {
			t.Fatalf("refusal wrote comment files: %v", files)
		}

		out, code = runCLI(t, bin, root, base("Missing unit")...)
		if code == 0 || !strings.Contains(out, "Missing unit") {
			t.Errorf("unknown heading must fail naming the heading: exit=%d out=%s", code, out)
		}

		out, code = runCLI(t, bin, root, base("Twin unit", "--occurrence", "3")...)
		if code == 0 || !strings.Contains(out, "occurrence") {
			t.Errorf("out-of-range occurrence must fail: exit=%d out=%s", code, out)
		}
		if files := receiptCommentFiles(t, root); len(files) != 0 {
			t.Fatalf("refusals wrote comment files: %v", files)
		}

		pre, post := receiptGitCommits(t, root, false)
		writeReceiptRunOutput(t, root, "pre.txt", "missing outcome\n")
		writeReceiptRunOutput(t, root, "post.txt", "outcome present\n")
		out, code = runCLI(t, bin, root, base("Twin unit", "--occurrence", "2", "--pre-sha", pre, "--post-sha", post)...)
		if code != 0 {
			t.Fatalf("explicit occurrence must resolve the second twin, exit %d: %s", code, out)
		}
		files := receiptCommentFiles(t, root)
		if len(files) != 1 {
			t.Fatalf("expected one comment file, got %v", files)
		}
		content, err := os.ReadFile(files[0])
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(content), "echo twin-two") {
			t.Errorf("receipt does not quote the second twin's Verify: %s", content)
		}
		wantDigest := receiptUnitDigest(t, queuePath, "Twin unit", 2)
		if !strings.Contains(string(content), "unit digest: "+wantDigest) {
			t.Errorf("receipt does not carry the resolved unit digest %s: %s", wantDigest, content)
		}
	})

	t.Run("non-ancestor pre refused before any comment file is written", func(t *testing.T) {
		bin, root := setupCLITest(t)
		queuePath := writeReceiptQueueFixture(t, root, receiptFixtureBody)
		writeReceiptRunOutput(t, root, "pre.txt", "missing outcome\n")
		writeReceiptRunOutput(t, root, "post.txt", "outcome present\n")
		pre, post := receiptGitCommits(t, root, true)

		args := func(preSHA, postSHA string, extra ...string) []string {
			all := []string{
				"receipt", "--queue", queuePath, "--heading", "Only unit",
				"--pre-sha", preSHA, "--pre-status", "1", "--pre-out", "pre.txt",
				"--post-sha", postSHA, "--post-out", "post.txt",
			}
			return append(all, extra...)
		}

		out, code := runCLI(t, bin, root, args(pre, post)...)
		if code == 0 {
			t.Fatalf("non-ancestor pre must be refused, got exit 0: %s", out)
		}
		if !strings.Contains(out, "not an ancestor") {
			t.Errorf("refusal does not name the ancestry violation: %s", out)
		}
		if files := receiptCommentFiles(t, root); len(files) != 0 {
			t.Fatalf("refusal wrote comment files: %v", files)
		}

		out, code = runCLI(t, bin, root, args(pre, pre)...)
		if code == 0 || !strings.Contains(out, "pre and post") {
			t.Errorf("equal SHAs must be refused: exit=%d out=%s", code, out)
		}

		out, code = runCLI(t, bin, root, args(pre, post, "--pre-status", "0")...)
		if code == 0 || !strings.Contains(out, "pre exit status") {
			t.Errorf("zero pre status must be refused: exit=%d out=%s", code, out)
		}

		out, code = runCLI(t, bin, root, args(pre, post, "--post-status", "3")...)
		if code == 0 || !strings.Contains(out, "post exit status") {
			t.Errorf("nonzero post status must be refused: exit=%d out=%s", code, out)
		}

		missingArgs := []string{
			"receipt", "--queue", queuePath, "--heading", "Only unit",
			"--pre-sha", pre, "--pre-status", "1", "--pre-out", "missing-pre.txt",
			"--post-sha", post, "--post-out", "post.txt",
		}
		out, code = runCLI(t, bin, root, missingArgs...)
		if code == 0 || !strings.Contains(out, "missing-pre.txt") {
			t.Errorf("unreadable pre output file must be refused: exit=%d out=%s", code, out)
		}
		if files := receiptCommentFiles(t, root); len(files) != 0 {
			t.Fatalf("refusals wrote comment files: %v", files)
		}
	})

	t.Run("numbered comment files and exact gh commands emitted in order", func(t *testing.T) {
		bin, root := setupCLITest(t)
		queuePath := writeReceiptQueueFixture(t, root, receiptFixtureBody)
		fakePath := receiptFakeGH(t, 42, receiptFixtureBody)
		pre, post := receiptGitCommits(t, root, false)
		writeReceiptRunOutput(t, root, "pre.txt", "missing outcome\n")
		writeReceiptRunOutput(t, root, "post.txt", "outcome present\n")

		cmd := exec.Command(bin, "receipt", "--issue", "42", "--heading", "Only unit",
			"--pre-sha", pre, "--pre-status", "1", "--pre-out", "pre.txt",
			"--post-sha", post, "--post-out", "post.txt")
		cmd.Dir = root
		cmd.Env = append(append(os.Environ(), "HOME="+root), "PATH="+fakePath)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("receipt failed: %v\n%s", err, out)
		}
		if string(out) != "gh issue comment 42 --body-file receipt-0001.md\n" {
			t.Fatalf("stdout = %q, want the single exact gh command", string(out))
		}
		content, err := os.ReadFile(filepath.Join(root, "receipt-0001.md"))
		if err != nil {
			t.Fatal(err)
		}
		receipt := string(content)
		for _, want := range []string{
			"## Only unit",
			"Evidence:",
			"Protocol: evidence/v1",
			"Digest algorithm: unit-contract-sha256-v1",
			"Receipt ID: receipt-sha256-v1:",
			"echo receipt-fixture",
			"unit digest: " + receiptUnitDigest(t, queuePath, "Only unit", 1),
			"pre sha: " + pre,
			"pre exit status: 1",
			"post sha: " + post,
			"post exit status: 0",
			"Pre-evidence scope: this command exited 1 at " + pre + "; nothing else is inferred.",
			"Post-evidence scope: this command exited 0 at " + post + "; nothing else is inferred.",
		} {
			if !strings.Contains(receipt, want) {
				t.Errorf("receipt missing %q:\n%s", want, receipt)
			}
		}

		// Oversized output splits into numbered files with gh commands in
		// posting order.
		bin, root = setupCLITest(t)
		writeReceiptQueueFixture(t, root, receiptFixtureBody)
		fakePath = receiptFakeGH(t, 42, receiptFixtureBody)
		pre, post = receiptGitCommits(t, root, false)
		writeReceiptRunOutput(t, root, "pre.txt", strings.Repeat("0123456789", 7000)+"\n")
		writeReceiptRunOutput(t, root, "post.txt", "outcome present\n")

		cmd = exec.Command(bin, "receipt", "--issue", "42", "--heading", "Only unit",
			"--pre-sha", pre, "--pre-status", "1", "--pre-out", "pre.txt",
			"--post-sha", post, "--post-out", "post.txt")
		cmd.Dir = root
		cmd.Env = append(append(os.Environ(), "HOME="+root), "PATH="+fakePath)
		out, err = cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("oversized receipt failed: %v\n%s", err, out)
		}
		files := receiptCommentFiles(t, root)
		if len(files) < 2 {
			t.Fatalf("expected oversized output to split across files, got %v", files)
		}
		var wantCommands []string
		for i := range files {
			name := fmt.Sprintf("receipt-%04d.md", i+1)
			wantCommands = append(wantCommands, "gh issue comment 42 --body-file "+name)
		}
		if string(out) != strings.Join(wantCommands, "\n")+"\n" {
			t.Fatalf("stdout = %q, want gh commands in posting order %v", string(out), wantCommands)
		}
		for _, name := range files {
			info, err := os.Stat(filepath.Join(root, name))
			if err != nil {
				t.Fatal(err)
			}
			if info.Size() > 65536 {
				t.Errorf("comment file %s is %d bytes, over the comment cap", name, info.Size())
			}
		}

		// Queue mode stays emit-only: files without gh commands.
		bin, root = setupCLITest(t)
		queuePath = writeReceiptQueueFixture(t, root, receiptFixtureBody)
		pre, post = receiptGitCommits(t, root, false)
		writeReceiptRunOutput(t, root, "pre.txt", "missing outcome\n")
		writeReceiptRunOutput(t, root, "post.txt", "outcome present\n")
		queueOut, code := runCLI(t, bin, root, "receipt", "--queue", queuePath, "--heading", "Only unit",
			"--pre-sha", pre, "--pre-status", "1", "--pre-out", "pre.txt",
			"--post-sha", post, "--post-out", "post.txt")
		if code != 0 {
			t.Fatalf("queue-mode receipt failed: %s", queueOut)
		}
		if strings.Contains(queueOut, "gh issue comment") {
			t.Errorf("queue mode must not print gh commands: %s", queueOut)
		}
		if files := receiptCommentFiles(t, root); len(files) != 1 {
			t.Fatalf("queue mode must still write the comment file, got %v", files)
		}
	})

	t.Run("completion registry carries the receipt command and flags", func(t *testing.T) {
		bin, root := setupCLITest(t)
		found := false
		wantFlags := []string{
			"--issue", "--queue", "--heading", "--occurrence",
			"--pre-sha", "--pre-status", "--pre-out",
			"--post-sha", "--post-status", "--post-out",
			"--rebuild", "--recovered-from",
		}
		for _, spec := range internal.CommandSpecs {
			if spec.Name != "receipt" {
				continue
			}
			found = true
			have := map[string]bool{}
			for _, f := range spec.Flags {
				have[f.Name] = true
			}
			for _, flag := range wantFlags {
				if !have[flag] {
					t.Errorf("receipt spec is missing flag %s", flag)
				}
			}
		}
		if !found {
			t.Fatal("CommandSpecs does not register the receipt command")
		}

		compOut, code := runCLI(t, bin, root, "completion", "bash")
		if code != 0 || !strings.Contains(compOut, "receipt") {
			t.Errorf("bash completion does not offer receipt; exit=%d", code)
		}
	})
}
