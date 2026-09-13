package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bermudi/litespec/v2/internal"
)

const digestFixtureBody = `Base: 0000000000000000000000000000000000000001
Branch: litespec/demo

## First unit

Done means: it works

Verify:
` + "```bash\necho first\n```\n" + `
- [ ] pending

## Second unit

Read first: some context
Depends: First unit
Done means: it also works

Verify:
` + "```bash\necho second\n```\n" + `
- [ ] pending
`

var digestGoldenOutput = "1\tFirst unit\t0b704df028cb456f1cd69299c061ff6768433fe2be636748a3c5721000f493f9\n" +
	"1\tSecond unit\tb03977de6c940f1f948a1222728a63aaf930804c6dd3eb5a8ada7d9ea098ac80\n"

func writeDigestQueueFixture(t *testing.T, root string) string {
	t.Helper()
	dir := filepath.Join(root, "specs", "queues")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "demo.md")
	if err := os.WriteFile(path, []byte(digestFixtureBody), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeDuplicateHeadingQueueFixture(t *testing.T, root string) string {
	t.Helper()
	dir := filepath.Join(root, "specs", "queues")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "duplicates.md")
	body := `Base: 0000000000000000000000000000000000000001
Branch: litespec/demo

## Repeated unit

Done means: it works

Verify:
` + "```bash\necho one\n```\n" + `
- [ ] pending

## Other unit

Done means: it differs

Verify:
` + "```bash\necho two\n```\n" + `
- [ ] pending

## Repeated unit

Done means: it also works

Verify:
` + "```bash\necho three\n```\n" + `
- [ ] pending
`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestDigestHeadingFilter(t *testing.T) {
	bin, root := setupCLITest(t)
	path := writeDuplicateHeadingQueueFixture(t, root)

	t.Run("heading filter lists every occurrence of a duplicated heading", func(t *testing.T) {
		lines, err := internal.DigestQueueUnits(root, 0, path)
		if err != nil {
			t.Fatal(err)
		}
		var want []internal.UnitDigestLine
		for _, line := range lines {
			if line.Heading == "Repeated unit" {
				want = append(want, line)
			}
		}
		if len(want) != 2 {
			t.Fatalf("fixture must carry two Repeated unit occurrences, got %d", len(want))
		}
		if want[0].Occurrence != 1 || want[1].Occurrence != 2 {
			t.Fatalf("occurrences = %d,%d; want 1,2", want[0].Occurrence, want[1].Occurrence)
		}

		out, code := runCLI(t, bin, root, "digest", "--queue", path, "--heading", "Repeated unit")
		if code != 0 {
			t.Fatalf("exit %d: %s", code, out)
		}
		if wantOut := internal.FormatUnitDigestLines(want); out != wantOut {
			t.Errorf("output = %q, want %q", out, wantOut)
		}
	})

	t.Run("zero-match heading fails visibly and completion offers the flag", func(t *testing.T) {
		out, code := runCLI(t, bin, root, "digest", "--queue", path, "--heading", "Missing unit")
		if code == 0 {
			t.Fatal("zero-match heading must exit non-zero")
		}
		if !strings.Contains(out, "Missing unit") {
			t.Errorf("failure must name the heading; out=%s", out)
		}

		hasHeading := false
		for _, spec := range internal.CommandSpecs {
			if spec.Name != "digest" {
				continue
			}
			for _, f := range spec.Flags {
				if f.Name == "--heading" {
					hasHeading = true
				}
			}
		}
		if !hasHeading {
			t.Error("CommandSpecs does not register --heading on digest")
		}
		compOut, code := runCLI(t, bin, root, "completion", "bash")
		if code != 0 || !strings.Contains(compOut, "--heading") {
			t.Errorf("bash completion does not offer --heading; exit=%d", code)
		}
	})
}

func TestDigestCommandPrintsUnitDigests(t *testing.T) {
	bin, root := setupCLITest(t)

	t.Run("local queue file", func(t *testing.T) {
		path := writeDigestQueueFixture(t, root)
		out, code := runCLI(t, bin, root, "digest", "--queue", path)
		if code != 0 {
			t.Fatalf("exit %d: %s", code, out)
		}
		if out != digestGoldenOutput {
			t.Errorf("output = %q, want %q", out, digestGoldenOutput)
		}
	})

	t.Run("gh issue via fake gh on PATH", func(t *testing.T) {
		bodyJSON, err := json.Marshal(digestFixtureBody)
		if err != nil {
			t.Fatal(err)
		}
		payload := []byte(fmt.Sprintf(`{"number":42,"title":"t","body":%s,"url":"","comments":[]}`, bodyJSON))
		fakeBin := t.TempDir()
		script := "#!/bin/sh\ncat " + filepath.Join(fakeBin, "payload.json") + "\n"
		if err := os.WriteFile(filepath.Join(fakeBin, "gh"), []byte(script), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(fakeBin, "payload.json"), payload, 0o644); err != nil {
			t.Fatal(err)
		}

		cmd := exec.Command(bin, "digest", "--issue", "42")
		cmd.Dir = root
		cmd.Env = append(append(os.Environ(), "HOME="+root), "PATH="+fakeBin+":"+os.Getenv("PATH"))
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("exec: %v\n%s", err, out)
		}
		if string(out) != digestGoldenOutput {
			t.Errorf("output = %q, want %q", string(out), digestGoldenOutput)
		}

	})

	t.Run("requires exactly one source flag", func(t *testing.T) {
		if _, code := runCLI(t, bin, root, "digest"); code == 0 {
			t.Error("digest without flags must fail")
		}
		out, code := runCLI(t, bin, root, "digest", "--queue", "/tmp/x.md", "--issue", "42")
		if code == 0 || !strings.Contains(out, "mutually exclusive") {
			t.Errorf("both flags must fail naming exclusivity; exit=%d out=%s", code, out)
		}
	})

	t.Run("command registered for completion", func(t *testing.T) {
		found := false
		for _, spec := range internal.CommandSpecs {
			if spec.Name == "digest" {
				found = true
				var hasIssue, hasQueue bool
				for _, f := range spec.Flags {
					switch f.Name {
					case "--issue":
						hasIssue = true
					case "--queue":
						hasQueue = true
					}
				}
				if !hasIssue || !hasQueue {
					t.Errorf("digest spec missing flags: %+v", spec.Flags)
				}
			}
		}
		if !found {
			t.Error("CommandSpecs does not register digest command")
		}

		compOut, code := runCLI(t, bin, root, "completion", "bash")
		if code != 0 || !strings.Contains(compOut, "digest") {
			t.Errorf("bash completion does not offer digest; exit=%d", code)
		}
	})
}
