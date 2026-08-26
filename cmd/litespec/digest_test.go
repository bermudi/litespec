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
