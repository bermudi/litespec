package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func buildBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "litespec")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	return bin
}

func setupCLITest(t *testing.T) (string, string) {
	t.Helper()
	bin := buildBinary(t)
	root := t.TempDir()
	specsDir := filepath.Join(root, "specs", "specs")
	changesDir := filepath.Join(root, "specs", "changes")
	if err := os.MkdirAll(specsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(changesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	return bin, root
}

func runCLI(t *testing.T, bin, root string, args ...string) (string, int) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "HOME="+root)
	out, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("exec: %v\n%s", err, out)
		}
	}
	return string(out), exitCode
}

func createChange(t *testing.T, root, name string) {
	t.Helper()
	changeDir := filepath.Join(root, "specs", "changes", name)
	if err := os.MkdirAll(changeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	meta := []byte("schema: spec-driven\n")
	if err := os.WriteFile(filepath.Join(changeDir, ".litespec.yaml"), meta, 0o644); err != nil {
		t.Fatal(err)
	}
}

func createChangeWithArtifacts(t *testing.T, root, name string) {
	t.Helper()
	createChange(t, root, name)
	changeDir := filepath.Join(root, "specs", "changes", name)
	os.WriteFile(filepath.Join(changeDir, "proposal.md"), []byte("# Proposal\nTest."), 0o644)
	os.WriteFile(filepath.Join(changeDir, "design.md"), []byte("# Design\nTest."), 0o644)
	os.WriteFile(filepath.Join(changeDir, "tasks.md"), []byte("## Phase 1: Test\n- [ ] Task"), 0o644)
	specsSubdir := filepath.Join(changeDir, "specs", "cap")
	os.MkdirAll(specsSubdir, 0o755)
	os.WriteFile(filepath.Join(specsSubdir, "spec.md"), []byte(`## ADDED Requirements

### Requirement: R1
The system SHALL work.

#### Scenario: S1
- **WHEN** triggered
`), 0o644)
}

func createSpec(t *testing.T, root, name string) {
	t.Helper()
	specDir := filepath.Join(root, "specs", "specs", name)
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(`# `+name+`

### Requirement: R1
The system SHALL work.

#### Scenario: S1
- **WHEN** triggered
`), 0o644)
}

func TestCLIVerifyNoArgs(t *testing.T) {
	bin, root := setupCLITest(t)
	out, code := runCLI(t, bin, root, "validate", "--json")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("json: %v\n%s", err, out)
	}
	if result["valid"] != true {
		t.Error("expected valid")
	}
}

func TestCLIVerifyPositionalChange(t *testing.T) {
	bin, root := setupCLITest(t)
	createChangeWithArtifacts(t, root, "my-change")
	out, code := runCLI(t, bin, root, "validate", "my-change", "--json")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("json: %v\n%s", err, out)
	}
	if result["valid"] != true {
		t.Errorf("expected valid, got %v", result["valid"])
	}
}

func TestCLIVerifyPositionalSpec(t *testing.T) {
	bin, root := setupCLITest(t)
	createSpec(t, root, "auth")
	out, code := runCLI(t, bin, root, "validate", "auth", "--json")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("json: %v\n%s", err, out)
	}
	if result["valid"] != true {
		t.Errorf("expected valid, got %v", result["valid"])
	}
}

func TestCLIVerifyUnknownName(t *testing.T) {
	bin, root := setupCLITest(t)
	_, code := runCLI(t, bin, root, "validate", "nonexistent")
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
}

func TestCLIVerifyAmbiguousName(t *testing.T) {
	bin, root := setupCLITest(t)
	createChange(t, root, "shared")
	createSpec(t, root, "shared")
	_, code := runCLI(t, bin, root, "validate", "shared")
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
}

func TestCLIVerifyAmbiguousWithTypeChange(t *testing.T) {
	bin, root := setupCLITest(t)
	createChangeWithArtifacts(t, root, "shared")
	createSpec(t, root, "shared")
	out, code := runCLI(t, bin, root, "validate", "shared", "--type", "change", "--json")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("json: %v\n%s", err, out)
	}
	if result["valid"] != true {
		t.Error("expected valid")
	}
}

func TestCLIVerifyAmbiguousWithTypeSpec(t *testing.T) {
	bin, root := setupCLITest(t)
	createChange(t, root, "shared")
	createSpec(t, root, "shared")
	out, code := runCLI(t, bin, root, "validate", "shared", "--type", "spec", "--json")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("json: %v\n%s", err, out)
	}
	if result["valid"] != true {
		t.Error("expected valid")
	}
}

func TestCLIVerifyTypeWithoutName(t *testing.T) {
	bin, root := setupCLITest(t)
	_, code := runCLI(t, bin, root, "validate", "--type", "change")
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
}

func TestCLIVerifyTypeWithBulkFlag(t *testing.T) {
	bin, root := setupCLITest(t)
	_, code := runCLI(t, bin, root, "validate", "--all", "--type", "change")
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
}

func TestCLIVerifyBulkAll(t *testing.T) {
	bin, root := setupCLITest(t)
	out, code := runCLI(t, bin, root, "validate", "--all", "--json")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
}

func TestCLIVerifyBulkChanges(t *testing.T) {
	bin, root := setupCLITest(t)
	out, code := runCLI(t, bin, root, "validate", "--changes", "--json")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
}

func TestCLIVerifyBulkSpecs(t *testing.T) {
	bin, root := setupCLITest(t)
	out, code := runCLI(t, bin, root, "validate", "--specs", "--json")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
}

func TestCLIVerifyBulkCombined(t *testing.T) {
	bin, root := setupCLITest(t)
	out, code := runCLI(t, bin, root, "validate", "--changes", "--specs", "--json")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
}

func TestCLIVerifyNameWithBulkFlag(t *testing.T) {
	bin, root := setupCLITest(t)
	_, code := runCLI(t, bin, root, "validate", "my-change", "--all")
	if code != 1 {
		t.Fatalf("expected exit 1 for name + bulk, got %d", code)
	}
}

func TestCLIInstructionsArtifact(t *testing.T) {
	bin, root := setupCLITest(t)
	out, code := runCLI(t, bin, root, "instructions", "proposal")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	if strings.TrimSpace(out) == "" {
		t.Error("expected non-empty output for instructions proposal")
	}
}

func TestCLIInstructionsJSON(t *testing.T) {
	bin, root := setupCLITest(t)
	out, code := runCLI(t, bin, root, "instructions", "design", "--json")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("json: %v\n%s", err, out)
	}
	for _, field := range []string{"artifactId", "description", "instruction", "template", "outputPath"} {
		if _, ok := result[field]; !ok {
			t.Errorf("missing field %q in JSON output", field)
		}
	}
	if result["artifactId"] != "design" {
		t.Errorf("expected artifactId=design, got %v", result["artifactId"])
	}
}

func TestCLIInstructionsUnknownArtifact(t *testing.T) {
	bin, root := setupCLITest(t)
	out, code := runCLI(t, bin, root, "instructions", "unknown-artifact")
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(out, "valid:") {
		t.Errorf("expected valid artifact list in error, got: %s", out)
	}
}
