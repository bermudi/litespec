package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
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

func receiptGitCommits(t *testing.T, root string) (string, string) {
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

func receiptOrphanCommit(t *testing.T, root string) string {
	t.Helper()
	git := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		cmd.Env = append(os.Environ(), "HOME="+root)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git("checkout", "-q", "--orphan", "isolated")
	if err := os.WriteFile(filepath.Join(root, "seed.txt"), []byte("unrelated\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git("add", ".")
	git("commit", "-q", "-m", "unrelated")
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "HOME="+root)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git rev-parse: %v\n%s", err, out)
	}
	return strings.TrimSpace(string(out))
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

// receiptPostingFakeGH installs a gh shim that serves the given issue body
// for `issue view`, records every other invocation (the comment writes) to
// logPath, and fails any invocation containing $FAIL_MARKER when set.
func receiptPostingFakeGH(t *testing.T, issueNumber int, body, logPath string) string {
	t.Helper()
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	fakeBin := t.TempDir()
	payload := fmt.Sprintf(`{"number":%d,"title":"t","body":%s,"url":"","comments":[]}`, issueNumber, bodyJSON)
	script := "#!/bin/sh\n" +
		"case \"$*\" in\n" +
		"*\"issue view\"*)\n" +
		"  cat " + filepath.Join(fakeBin, "payload.json") + "\n" +
		"  exit 0\n" +
		"  ;;\n" +
		"esac\n" +
		"echo \"gh $*\" >> " + logPath + "\n" +
		"if [ -n \"$FAIL_MARKER\" ] && echo \"$*\" | grep -q \"$FAIL_MARKER\"; then\n" +
		"  echo 'gh: simulated failure' >&2\n" +
		"  exit 1\n" +
		"fi\n" +
		"exit 0\n"
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

func TestReceiptCommandEmitsBoundedV2Comments(t *testing.T) {
	receiptSHA256 := func(s string) string {
		sum := sha256.Sum256([]byte(s))
		return hex.EncodeToString(sum[:])
	}
	// boundedFencePayload extracts the fenced payload that follows the given
	// anchor line, where the fence delimiter is a run of backticks.
	boundedFencePayload := func(t *testing.T, text, anchor string) string {
		t.Helper()
		i := strings.Index(text, anchor)
		if i < 0 {
			t.Fatalf("anchor %q not found in receipt:\n%s", anchor, text)
		}
		rest := text[i+len(anchor):]
		nl := strings.Index(rest, "\n")
		if nl < 0 || strings.Trim(rest[:nl], "`") != "" {
			t.Fatalf("no opening fence delimiter after anchor %q:\n%s", anchor, text)
		}
		delim := rest[:nl]
		body := rest[nl+1:]
		closeIdx := strings.Index(body, "\n"+delim+"\n")
		if closeIdx < 0 {
			if strings.HasSuffix(body, "\n"+delim) {
				closeIdx = len(body) - len(delim) - 1
			} else {
				t.Fatalf("no closing fence delimiter:\n%s", text)
			}
		}
		return body[:closeIdx]
	}
	assertBoundedV2Fields := func(t *testing.T, content, queuePath, pre, post, preOut, postOut string) {
		t.Helper()
		for _, want := range []string{
			"Protocol: evidence/v2",
			"Digest algorithm: unit-contract-sha256-v1",
			"unit digest: " + receiptUnitDigest(t, queuePath, "Only unit", 1),
			"pre sha: " + pre,
			"pre exit status: 1",
			"post sha: " + post,
			"post exit status: 0",
			"pre bytes: " + strconv.Itoa(len(preOut)),
			"pre output sha256: " + receiptSHA256(preOut),
			"post bytes: " + strconv.Itoa(len(postOut)),
			"post output sha256: " + receiptSHA256(postOut),
			"Pre-evidence scope: this command exited 1 at " + pre + "; nothing else is inferred.",
			"Post-evidence scope: this command exited 0 at " + post + "; nothing else is inferred.",
		} {
			if !strings.Contains(content, want) {
				t.Errorf("v2 receipt missing %q:\n%s", want, content)
			}
		}
		if !regexp.MustCompile(`(?m)^Receipt ID: receipt-sha256-v2:[0-9a-f]{64}$`).MatchString(content) {
			t.Errorf("v2 receipt missing a receipt-sha256-v2 Receipt ID:\n%s", content)
		}
		if strings.Contains(content, "Raw output chunk:") || strings.Contains(content, "Receipt continues in next comment") {
			t.Errorf("v2 receipt must not chunk or continue:\n%s", content)
		}
		if len(content) > 8192 {
			t.Errorf("v2 receipt is %d bytes, over the 8192-byte budget", len(content))
		}
	}

	t.Run("receipt command emits one bounded v2 comment file", func(t *testing.T) {
		preOut := "missing outcome\n"
		postOut := "outcome present\n"

		// Issue lane: exactly one gh comment command for exactly one file.
		bin, root := setupCLITest(t)
		queuePath := writeReceiptQueueFixture(t, root, receiptFixtureBody)
		fakePath := receiptFakeGH(t, 42, receiptFixtureBody)
		pre, post := receiptGitCommits(t, root)
		writeReceiptRunOutput(t, root, "pre.txt", preOut)
		writeReceiptRunOutput(t, root, "post.txt", postOut)

		cmd := exec.Command(bin, "receipt", "--issue", "42", "--heading", "Only unit",
			"--pre-sha", pre, "--pre-status", "1", "--pre-out", "pre.txt",
			"--post-sha", post, "--post-out", "post.txt")
		cmd.Dir = root
		cmd.Env = append(append(os.Environ(), "HOME="+root), "PATH="+fakePath)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("receipt failed: %v\n%s", err, out)
		}
		wantCommand := "gh issue comment 42 --body-file " + filepath.Join(root, "receipt-0001.md") + "\n"
		if string(out) != wantCommand {
			t.Fatalf("stdout = %q, want exactly one gh command %q", string(out), wantCommand)
		}
		files := receiptCommentFiles(t, root)
		if len(files) != 1 {
			t.Fatalf("want exactly one comment file, got %v", files)
		}
		content, err := os.ReadFile(files[0])
		if err != nil {
			t.Fatal(err)
		}
		assertBoundedV2Fields(t, string(content), queuePath, pre, post, preOut, postOut)
		if !strings.Contains(string(content), preOut) || !strings.Contains(string(content), postOut) {
			t.Errorf("outputs within the excerpt budget must appear verbatim:\n%s", content)
		}

		// Queue lane: the same single bounded file, no gh commands.
		bin, root = setupCLITest(t)
		queuePath = writeReceiptQueueFixture(t, root, receiptFixtureBody)
		pre, post = receiptGitCommits(t, root)
		writeReceiptRunOutput(t, root, "pre.txt", preOut)
		writeReceiptRunOutput(t, root, "post.txt", postOut)
		queueOut, code := runCLI(t, bin, root, "receipt", "--queue", queuePath, "--heading", "Only unit",
			"--pre-sha", pre, "--pre-status", "1", "--pre-out", "pre.txt",
			"--post-sha", post, "--post-out", "post.txt")
		if code != 0 {
			t.Fatalf("queue-lane receipt failed: %s", queueOut)
		}
		if strings.Contains(queueOut, "gh issue comment") {
			t.Errorf("queue lane must not print gh commands: %s", queueOut)
		}
		files = receiptCommentFiles(t, root)
		if len(files) != 1 {
			t.Fatalf("queue lane want exactly one comment file, got %v", files)
		}
		content, err = os.ReadFile(files[0])
		if err != nil {
			t.Fatal(err)
		}
		assertBoundedV2Fields(t, string(content), queuePath, pre, post, preOut, postOut)
	})

	t.Run("oversized output no longer chunks into multiple files", func(t *testing.T) {
		bin, root := setupCLITest(t)
		writeReceiptQueueFixture(t, root, receiptFixtureBody)
		fakePath := receiptFakeGH(t, 42, receiptFixtureBody)
		pre, post := receiptGitCommits(t, root)
		preOut := strings.Repeat("0123456789", 7000) + "\n"
		postOut := "outcome present\n"
		writeReceiptRunOutput(t, root, "pre.txt", preOut)
		writeReceiptRunOutput(t, root, "post.txt", postOut)

		cmd := exec.Command(bin, "receipt", "--issue", "42", "--heading", "Only unit",
			"--pre-sha", pre, "--pre-status", "1", "--pre-out", "pre.txt",
			"--post-sha", post, "--post-out", "post.txt")
		cmd.Dir = root
		cmd.Env = append(append(os.Environ(), "HOME="+root), "PATH="+fakePath)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("receipt failed: %v\n%s", err, out)
		}
		wantCommand := "gh issue comment 42 --body-file " + filepath.Join(root, "receipt-0001.md") + "\n"
		if string(out) != wantCommand {
			t.Fatalf("stdout = %q, want exactly one gh command for the single bounded file", string(out))
		}
		files := receiptCommentFiles(t, root)
		if len(files) != 1 {
			t.Fatalf("oversized output must emit exactly one comment file, got %v", files)
		}
		content, err := os.ReadFile(files[0])
		if err != nil {
			t.Fatal(err)
		}
		text := string(content)
		if len(text) > 8192 {
			t.Fatalf("bounded receipt is %d bytes, over the 8192-byte budget", len(text))
		}
		if !strings.Contains(text, "pre bytes: "+strconv.Itoa(len(preOut))) {
			t.Errorf("receipt must declare the full pre byte count %d:\n%s", len(preOut), text)
		}
		if !strings.Contains(text, "pre output sha256: "+receiptSHA256(preOut)) {
			t.Errorf("receipt must declare the full pre output SHA-256:\n%s", text)
		}
		markerRe := regexp.MustCompile(`(?m)^\.\.\. ([1-9][0-9]*) bytes elided \.\.\.$`)
		markers := markerRe.FindAllStringSubmatch(text, -1)
		if len(markers) != 1 {
			t.Fatalf("want exactly one elision marker, got %d:\n%s", len(markers), text)
		}
		elided, err := strconv.Atoi(markers[0][1])
		if err != nil {
			t.Fatal(err)
		}
		payload := boundedFencePayload(t, text, "pre exit status: 1\n")
		markerIdx := strings.Index(payload, markers[0][0])
		headLen := markerIdx
		rest := payload[markerIdx+len(markers[0][0]):]
		tailLen := 0
		if rest != "" {
			tailLen = len(rest) - 1 // the newline separating marker and tail
		}
		if headLen+elided+tailLen != len(preOut) {
			t.Errorf("reconstruction arithmetic broken: head %d + elided %d + tail %d != declared %d", headLen, elided, tailLen, len(preOut))
		}
	})

	t.Run("metadata over budget refuses before writing files", func(t *testing.T) {
		bin, root := setupCLITest(t)
		hugeHeading := "Big " + strings.Repeat("H", 8500)
		body := "Base: 0000000000000000000000000000000000000004\nBranch: litespec/receipt-fixture\n\n## " + hugeHeading + "\n\nDone means: it works\n\nVerify:\n```bash\necho over-budget\n```\n\n- [ ] pending\n"
		queuePath := writeReceiptQueueFixture(t, root, body)
		pre, post := receiptGitCommits(t, root)
		writeReceiptRunOutput(t, root, "pre.txt", "missing outcome\n")
		writeReceiptRunOutput(t, root, "post.txt", "outcome present\n")

		out, code := runCLI(t, bin, root, "receipt", "--queue", queuePath, "--heading", hugeHeading,
			"--pre-sha", pre, "--pre-status", "1", "--pre-out", "pre.txt",
			"--post-sha", post, "--post-out", "post.txt")
		if code == 0 {
			t.Fatalf("oversized identity fields must be refused, got exit 0: %s", out)
		}
		if !strings.Contains(out, "8192") || !strings.Contains(out, "Unit heading") {
			t.Errorf("refusal must name the byte budget and the oversized field: %s", out)
		}
		if files := receiptCommentFiles(t, root); len(files) != 0 {
			t.Fatalf("refusal wrote comment files: %v", files)
		}
	})
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

		pre, post := receiptGitCommits(t, root)
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
		pre, post := receiptGitCommits(t, root)
		unrelated := receiptOrphanCommit(t, root)

		args := func(preSHA, postSHA string, extra ...string) []string {
			all := []string{
				"receipt", "--queue", queuePath, "--heading", "Only unit",
				"--pre-sha", preSHA, "--pre-status", "1", "--pre-out", "pre.txt",
				"--post-sha", postSHA, "--post-out", "post.txt",
			}
			return append(all, extra...)
		}

		out, code := runCLI(t, bin, root, args(pre, unrelated)...)
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
		pre, post := receiptGitCommits(t, root)
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
		if string(out) != "gh issue comment 42 --body-file "+filepath.Join(root, "receipt-0001.md")+"\n" {
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
			"Protocol: evidence/v2",
			"Digest algorithm: unit-contract-sha256-v1",
			"Receipt ID: receipt-sha256-v2:",
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

		// Oversized output emits exactly one bounded comment file with one gh
		// command instead of chunking across multiple comments.
		bin, root = setupCLITest(t)
		writeReceiptQueueFixture(t, root, receiptFixtureBody)
		fakePath = receiptFakeGH(t, 42, receiptFixtureBody)
		pre, post = receiptGitCommits(t, root)
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
		if len(files) != 1 {
			t.Fatalf("oversized output must emit exactly one bounded comment file, got %v", files)
		}
		wantCommand := "gh issue comment 42 --body-file " + files[0] + "\n"
		if string(out) != wantCommand {
			t.Fatalf("stdout = %q, want the single exact gh command %q", string(out), wantCommand)
		}
		oversized, err := os.ReadFile(files[0])
		if err != nil {
			t.Fatal(err)
		}
		if len(oversized) > 8192 {
			t.Errorf("bounded comment file %s is %d bytes, over the 8192-byte receipt budget", files[0], len(oversized))
		}
		if !strings.Contains(string(oversized), " bytes elided ...") {
			t.Errorf("bounded receipt must elide the oversized output:\n%s", oversized)
		}

		// Queue mode stays emit-only: files without gh commands.
		bin, root = setupCLITest(t)
		queuePath = writeReceiptQueueFixture(t, root, receiptFixtureBody)
		pre, post = receiptGitCommits(t, root)
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

	t.Run("out directory receives files and failed writes leave no partial set", func(t *testing.T) {
		bin, root := setupCLITest(t)
		writeReceiptQueueFixture(t, root, receiptFixtureBody)
		fakePath := receiptFakeGH(t, 42, receiptFixtureBody)
		pre, post := receiptGitCommits(t, root)
		writeReceiptRunOutput(t, root, "pre.txt", strings.Repeat("0123456789", 14000)+"\n")
		writeReceiptRunOutput(t, root, "post.txt", "outcome present\n")

		runReceipt := func() (string, error) {
			cmd := exec.Command(bin, "receipt", "--issue", "42", "--heading", "Only unit",
				"--pre-sha", pre, "--pre-status", "1", "--pre-out", "pre.txt",
				"--post-sha", post, "--post-out", "post.txt", "--out", "out")
			cmd.Dir = root
			cmd.Env = append(append(os.Environ(), "HOME="+root), "PATH="+fakePath)
			out, err := cmd.CombinedOutput()
			return string(out), err
		}

		stdout, err := runReceipt()
		if err != nil {
			t.Fatalf("receipt with --out failed: %v\n%s", err, stdout)
		}
		wantCommand := fmt.Sprintf("gh issue comment 42 --body-file %s\n", filepath.Join(root, "out", "receipt-0001.md"))
		if stdout != wantCommand {
			t.Fatalf("stdout = %q, want the single gh command %q even for the oversized output", stdout, wantCommand)
		}
		if stray := receiptCommentFiles(t, root); len(stray) != 0 {
			t.Fatalf("--out must keep the working directory clean, found %v", stray)
		}

		// A failed comment-file write exits non-zero naming the file and
		// leaves no partial receipt set.
		bin, root = setupCLITest(t)
		writeReceiptQueueFixture(t, root, receiptFixtureBody)
		fakePath = receiptFakeGH(t, 42, receiptFixtureBody)
		pre, post = receiptGitCommits(t, root)
		writeReceiptRunOutput(t, root, "pre.txt", strings.Repeat("0123456789", 14000)+"\n")
		writeReceiptRunOutput(t, root, "post.txt", "outcome present\n")
		if err := os.MkdirAll(filepath.Join(root, "out", "receipt-0001.md"), 0o755); err != nil {
			t.Fatal(err)
		}
		stdout, err = runReceipt()
		if err == nil {
			t.Fatalf("failed comment-file write must exit non-zero: %s", stdout)
		}
		if !strings.Contains(stdout, "receipt-0001.md") {
			t.Fatalf("failure must name the failing file: %s", stdout)
		}
		if _, statErr := os.Stat(filepath.Join(root, "out", "receipt-0002.md")); !os.IsNotExist(statErr) {
			t.Fatalf("no partial set may remain beyond the failing name; receipt-0002.md stat err=%v", statErr)
		}
	})

	t.Run("completion registry carries the receipt command and flags", func(t *testing.T) {
		bin, root := setupCLITest(t)
		found := false
		wantFlags := []string{
			"--issue", "--queue", "--heading", "--occurrence",
			"--pre-sha", "--pre-status", "--pre-out",
			"--post-sha", "--post-status", "--post-out",
			"--rebuild", "--recovered-from", "--out",
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

func TestReceiptPostsCommentsWhenAsked(t *testing.T) {
	newScenario := func(t *testing.T) (bin, root, fakePath, logPath, pre, post string) {
		t.Helper()
		bin, root = setupCLITest(t)
		writeReceiptQueueFixture(t, root, receiptFixtureBody)
		logPath = filepath.Join(t.TempDir(), "gh-calls.log")
		fakePath = receiptPostingFakeGH(t, 42, receiptFixtureBody, logPath)
		pre, post = receiptGitCommits(t, root)
		writeReceiptRunOutput(t, root, "pre.txt", strings.Repeat("0123456789", 14000)+"\n")
		writeReceiptRunOutput(t, root, "post.txt", "outcome present\n")
		return bin, root, fakePath, logPath, pre, post
	}

	emitArgs := func(pre, post string, extra ...string) []string {
		args := []string{
			"receipt", "--issue", "42", "--heading", "Only unit",
			"--pre-sha", pre, "--pre-status", "1", "--pre-out", "pre.txt",
			"--post-sha", post, "--post-out", "post.txt",
		}
		return append(args, extra...)
	}

	runReceipt := func(t *testing.T, bin, root, fakePath string, args ...string) (string, int) {
		t.Helper()
		cmd := exec.Command(bin, args...)
		cmd.Dir = root
		cmd.Env = append(append(os.Environ(), "HOME="+root), "PATH="+fakePath)
		out, err := cmd.CombinedOutput()
		code := 0
		if err != nil {
			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) {
				t.Fatalf("receipt: %v\n%s", err, out)
			}
			code = exitErr.ExitCode()
		}
		return string(out), code
	}

	t.Run("opt-in flag posts the single bounded comment", func(t *testing.T) {
		bin, root, fakePath, logPath, pre, post := newScenario(t)

		out, code := runReceipt(t, bin, root, fakePath, emitArgs(pre, post, "--post")...)
		if code != 0 {
			t.Fatalf("posting receipt failed: exit %d\n%s", code, out)
		}
		files := receiptCommentFiles(t, root)
		if len(files) != 1 {
			t.Fatalf("expected exactly one bounded comment file even for the oversized output, got %v", files)
		}
		wantReport := fmt.Sprintf("posted %s (gh issue comment 42 --body-file %s)", files[0], files[0])
		if !strings.Contains(out, wantReport) {
			t.Fatalf("stdout must report the posted comment:\n%s", out)
		}
		if strings.Contains(out, "\ngh issue comment ") {
			t.Errorf("post mode must report posted comments, not print bare commands:\n%s", out)
		}
		logged, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatal(err)
		}
		wantCall := "gh issue comment 42 --body-file " + files[0]
		if strings.TrimSpace(string(logged)) != wantCall {
			t.Fatalf("gh calls = %q, want exactly %q", string(logged), wantCall)
		}
	})

	t.Run("default stays emit-only without any gh write", func(t *testing.T) {
		bin, root, fakePath, logPath, pre, post := newScenario(t)

		out, code := runReceipt(t, bin, root, fakePath, emitArgs(pre, post)...)
		if code != 0 {
			t.Fatalf("emit-only receipt failed: exit %d\n%s", code, out)
		}
		if !strings.Contains(out, "gh issue comment 42 --body-file "+filepath.Join(root, "receipt-0001.md")) {
			t.Fatalf("emit mode must print the exact gh commands:\n%s", out)
		}
		if _, err := os.Stat(logPath); !os.IsNotExist(err) {
			t.Errorf("emit mode must not invoke gh comment writes, log present: %v", err)
		}
	})

	t.Run("queue mode refuses --post", func(t *testing.T) {
		bin, root, fakePath, _, pre, post := newScenario(t)
		queuePath := filepath.Join(root, "specs", "queues", "receipt-fixture.md")
		args := emitArgs(pre, post, "--post")
		args[1] = "--queue"
		args[2] = queuePath

		out, code := runReceipt(t, bin, root, fakePath, args...)
		if code == 0 || !strings.Contains(out, "--post") {
			t.Errorf("--post with --queue must be refused naming the flag: exit=%d out=%s", code, out)
		}
	})

	t.Run("failing gh invocation stops visibly naming the command and the unposted file", func(t *testing.T) {
		bin, root, fakePath, logPath, pre, post := newScenario(t)
		failing := filepath.Join(root, "receipt-0001.md")
		cmd := exec.Command(bin, emitArgs(pre, post, "--post")...)
		cmd.Dir = root
		cmd.Env = append(append(os.Environ(), "HOME="+root), "PATH="+fakePath, "FAIL_MARKER="+failing)
		raw, err := cmd.CombinedOutput()
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) || exitErr.ExitCode() == 0 {
			t.Fatalf("gh failure must exit non-zero, got %v\n%s", err, raw)
		}
		out := string(raw)
		for _, want := range []string{
			"gh issue comment 42 --body-file " + failing,
			"posted: (none)",
			"not posted: receipt-0001.md",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("failure output must name %q:\n%s", want, out)
			}
		}
		logged, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatal(err)
		}
		wantCall := "gh issue comment 42 --body-file " + failing
		if strings.TrimSpace(string(logged)) != wantCall {
			t.Fatalf("gh calls = %q, want exactly one failing call %q with no retry", string(logged), wantCall)
		}
	})

	t.Run("posting from a subdirectory posts absolute body-file paths", func(t *testing.T) {
		bin, root, fakePath, logPath, pre, post := newScenario(t)
		sub := filepath.Join(root, "sub")
		if err := os.MkdirAll(sub, 0o755); err != nil {
			t.Fatal(err)
		}
		writeReceiptRunOutput(t, sub, "pre.txt", strings.Repeat("0123456789", 14000)+"\n")
		writeReceiptRunOutput(t, sub, "post.txt", "outcome present\n")

		cmd := exec.Command(bin, "receipt", "--issue", "42", "--heading", "Only unit",
			"--pre-sha", pre, "--pre-status", "1", "--pre-out", filepath.Join(sub, "pre.txt"),
			"--post-sha", post, "--post-out", filepath.Join(sub, "post.txt"), "--post")
		cmd.Dir = sub
		cmd.Env = append(append(os.Environ(), "HOME="+root), "PATH="+fakePath)
		raw, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("posting from a subdirectory must work with absolute paths: %v\n%s", err, raw)
		}
		files, err := filepath.Glob(filepath.Join(sub, "receipt-*.md"))
		if err != nil {
			t.Fatal(err)
		}
		sort.Strings(files)
		if len(files) != 1 {
			t.Fatalf("the oversized output must land as exactly one file in the invoking directory, got %v", files)
		}
		if stray := receiptCommentFiles(t, root); len(stray) != 0 {
			t.Fatalf("project root must stay clean when invoked from a subdirectory, found %v", stray)
		}
		logged, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatal(err)
		}
		wantCall := "gh issue comment 42 --body-file " + files[0]
		if strings.TrimSpace(string(logged)) != wantCall {
			t.Fatalf("gh calls = %q, want the absolute body-file path %q", string(logged), wantCall)
		}
	})
}
