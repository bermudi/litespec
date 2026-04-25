package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPatch_HappyPath(t *testing.T) {
	bin, root := setupCLITest(t)

	out, code := runCLI(t, bin, root, "patch", "add-verbose", "cli")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}

	changeDir := filepath.Join(root, "specs", "changes", "add-verbose")

	metaData, err := os.ReadFile(filepath.Join(changeDir, ".litespec.yaml"))
	if err != nil {
		t.Fatalf("read meta: %v", err)
	}
	if !strings.Contains(string(metaData), "mode: patch") {
		t.Errorf("expected .litespec.yaml to contain 'mode: patch', got:\n%s", string(metaData))
	}

	specPath := filepath.Join(changeDir, "specs", "cli", "spec.md")
	specData, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read spec: %v", err)
	}
	if !strings.Contains(string(specData), "# cli") {
		t.Errorf("expected spec stub to contain '# cli', got:\n%s", string(specData))
	}
	if !strings.Contains(string(specData), "## ADDED Requirements") {
		t.Errorf("expected spec stub to contain '## ADDED Requirements', got:\n%s", string(specData))
	}

	for _, f := range []string{"proposal.md", "design.md", "tasks.md"} {
		if _, err := os.Stat(filepath.Join(changeDir, f)); err == nil {
			t.Errorf("expected %s to not exist in patch-mode change", f)
		}
	}
}

func TestPatch_MissingArgs(t *testing.T) {
	bin, root := setupCLITest(t)

	out, code := runCLI(t, bin, root, "patch")
	if code == 0 {
		t.Fatalf("expected non-zero exit, got: %s", out)
	}
	if !strings.Contains(out, "usage:") {
		t.Errorf("expected usage message, got: %s", out)
	}
}

func TestPatch_OneArg(t *testing.T) {
	bin, root := setupCLITest(t)

	out, code := runCLI(t, bin, root, "patch", "add-verbose")
	if code == 0 {
		t.Fatalf("expected non-zero exit, got: %s", out)
	}
	if !strings.Contains(out, "usage:") {
		t.Errorf("expected usage message, got: %s", out)
	}
}

func TestPatch_ExistingChange(t *testing.T) {
	bin, root := setupCLITest(t)
	createChange(t, root, "existing")

	out, code := runCLI(t, bin, root, "patch", "existing", "cli")
	if code == 0 {
		t.Fatalf("expected non-zero exit, got: %s", out)
	}
	if !strings.Contains(out, "already exists") {
		t.Errorf("expected 'already exists' error, got: %s", out)
	}
}

func TestPatch_InvalidName(t *testing.T) {
	bin, root := setupCLITest(t)

	out, code := runCLI(t, bin, root, "patch", "../evil", "cli")
	if code == 0 {
		t.Fatalf("expected non-zero exit, got: %s", out)
	}
	if !strings.Contains(out, "path separators") && !strings.Contains(out, "path traversal") {
		t.Errorf("expected path-related error, got: %s", out)
	}
}

func TestPatch_InvalidCapability(t *testing.T) {
	bin, root := setupCLITest(t)

	out, code := runCLI(t, bin, root, "patch", "mychange", "../../evil")
	if code == 0 {
		t.Fatalf("expected non-zero exit, got: %s", out)
	}
	if !strings.Contains(out, "invalid capability name") {
		t.Errorf("expected invalid capability error, got: %s", out)
	}
}

func TestPatch_ThenValidate(t *testing.T) {
	bin, root := setupCLITest(t)

	_, code := runCLI(t, bin, root, "patch", "fix-flag", "cli")
	if code != 0 {
		t.Fatal("patch failed")
	}

	specPath := filepath.Join(root, "specs", "changes", "fix-flag", "specs", "cli", "spec.md")
	os.WriteFile(specPath, []byte(`# cli

## ADDED Requirements

### Requirement: Add verbose flag
The CLI SHALL support a --verbose flag.

#### Scenario: Verbose output
- **WHEN** the user passes --verbose
- **THEN** the CLI outputs detailed log lines
`), 0o644)

	out, code := runCLI(t, bin, root, "validate", "fix-flag", "--json")
	if code != 0 {
		t.Fatalf("validate exit %d: %s", code, out)
	}
	if !strings.Contains(out, `"valid": true`) {
		t.Errorf("expected valid=true, got: %s", out)
	}
}
