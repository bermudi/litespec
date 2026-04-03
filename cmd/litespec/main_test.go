package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/bermudi/litespec/internal"
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
	specsDir := filepath.Join(root, "specs", "canon")
	changesDir := filepath.Join(root, "specs", "changes")
	archiveDir := filepath.Join(root, "specs", "changes", "archive")
	if err := os.MkdirAll(specsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(changesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(archiveDir, 0o755); err != nil {
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
- **THEN** expected result
`), 0o644)
}

func createSpec(t *testing.T, root, name string) {
	t.Helper()
	specDir := filepath.Join(root, "specs", "canon", name)
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(`# `+name+`

## Requirements

### Requirement: R1
The system SHALL work.

#### Scenario: S1
- **WHEN** triggered
- **THEN** expected result
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

func TestChangeStatusText(t *testing.T) {
	tests := []struct {
		completed int
		total     int
		want      string
	}{
		{0, 0, "No tasks"},
		{5, 5, "✓ Complete"},
		{3, 5, "3/5 tasks"},
		{0, 3, "0/3 tasks"},
	}
	for _, tt := range tests {
		c := internal.ChangeInfo{CompletedTasks: tt.completed, TotalTasks: tt.total}
		got := changeStatusText(c)
		if got != tt.want {
			t.Errorf("changeStatusText(%d/%d) = %q, want %q", tt.completed, tt.total, got, tt.want)
		}
	}
}

func TestSortChangesByRecent(t *testing.T) {
	now := time.Now()
	changes := []internal.ChangeInfo{
		{Name: "alpha", LastModified: now.Add(-2 * time.Hour)},
		{Name: "beta", LastModified: now},
		{Name: "gamma", LastModified: now.Add(-1 * time.Hour)},
	}
	sortChanges(changes, "recent", "")
	if changes[0].Name != "beta" {
		t.Errorf("first = %q, want %q", changes[0].Name, "beta")
	}
	if changes[1].Name != "gamma" {
		t.Errorf("second = %q, want %q", changes[1].Name, "gamma")
	}
	if changes[2].Name != "alpha" {
		t.Errorf("third = %q, want %q", changes[2].Name, "alpha")
	}
}

func TestSortChangesByName(t *testing.T) {
	now := time.Now()
	changes := []internal.ChangeInfo{
		{Name: "charlie", LastModified: now},
		{Name: "alpha", LastModified: now.Add(-1 * time.Hour)},
		{Name: "bravo", LastModified: now.Add(-2 * time.Hour)},
	}
	sortChanges(changes, "name", "")
	if changes[0].Name != "alpha" {
		t.Errorf("first = %q, want %q", changes[0].Name, "alpha")
	}
	if changes[1].Name != "bravo" {
		t.Errorf("second = %q, want %q", changes[1].Name, "bravo")
	}
	if changes[2].Name != "charlie" {
		t.Errorf("third = %q, want %q", changes[2].Name, "charlie")
	}
}

func TestMaxNameWidthChanges(t *testing.T) {
	changes := []internal.ChangeInfo{
		{Name: "short"},
		{Name: "a-very-long-change-name"},
		{Name: "medium-name"},
	}
	got := maxNameWidthChanges(changes)
	want := len("a-very-long-change-name")
	if got != want {
		t.Errorf("maxNameWidthChanges = %d, want %d", got, want)
	}
}

func TestMaxNameWidthSpecs(t *testing.T) {
	specs := []internal.SpecInfo{
		{Name: "ab"},
		{Name: "a"},
		{Name: "abc"},
	}
	got := maxNameWidthSpecs(specs)
	if got != 3 {
		t.Errorf("maxNameWidthSpecs = %d, want 3", got)
	}
}

func TestMaxNameWidthEmpty(t *testing.T) {
	if maxNameWidthChanges(nil) != 0 {
		t.Error("expected 0 for nil slice")
	}
	if maxNameWidthSpecs(nil) != 0 {
		t.Error("expected 0 for nil slice")
	}
}

func TestCLIHelpFlag(t *testing.T) {
	bin := buildBinary(t)
	root := t.TempDir()
	out, code := runCLI(t, bin, root, "--help")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	if !strings.Contains(out, "Commands:") {
		t.Error("expected Commands section in help output")
	}
}

func TestCLIHelpShortFlag(t *testing.T) {
	bin := buildBinary(t)
	root := t.TempDir()
	out, code := runCLI(t, bin, root, "-h")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	if !strings.Contains(out, "Commands:") {
		t.Error("expected Commands section in help output")
	}
}

func TestCLIInitHelp(t *testing.T) {
	bin := buildBinary(t)
	root := t.TempDir()
	out, code := runCLI(t, bin, root, "init", "--help")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	if !strings.Contains(out, "Usage: litespec init") {
		t.Error("expected init usage in help output")
	}
}

func TestCLIInitHelpShort(t *testing.T) {
	bin := buildBinary(t)
	root := t.TempDir()
	out, code := runCLI(t, bin, root, "init", "-h")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	if !strings.Contains(out, "Usage: litespec init") {
		t.Error("expected init usage in help output")
	}
}

func TestCLIUpdateHelp(t *testing.T) {
	bin := buildBinary(t)
	root := t.TempDir()
	out, code := runCLI(t, bin, root, "update", "--help")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	if !strings.Contains(out, "Usage: litespec update") {
		t.Error("expected update usage in help output")
	}
}

func TestCLINewHelp(t *testing.T) {
	bin := buildBinary(t)
	root := t.TempDir()
	out, code := runCLI(t, bin, root, "new", "--help")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	if !strings.Contains(out, "Usage: litespec new") {
		t.Error("expected new usage in help output")
	}
}

func TestCLIListHelp(t *testing.T) {
	bin := buildBinary(t)
	root := t.TempDir()
	out, code := runCLI(t, bin, root, "list", "--help")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	if !strings.Contains(out, "Usage: litespec list") {
		t.Error("expected list usage in help output")
	}
}

func TestCLIStatusHelp(t *testing.T) {
	bin := buildBinary(t)
	root := t.TempDir()
	out, code := runCLI(t, bin, root, "status", "--help")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	if !strings.Contains(out, "Usage: litespec status") {
		t.Error("expected status usage in help output")
	}
}

func TestCLIValidateHelp(t *testing.T) {
	bin := buildBinary(t)
	root := t.TempDir()
	out, code := runCLI(t, bin, root, "validate", "--help")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	if !strings.Contains(out, "Usage: litespec validate") {
		t.Error("expected validate usage in help output")
	}
}

func TestCLIInstructionsHelp(t *testing.T) {
	bin := buildBinary(t)
	root := t.TempDir()
	out, code := runCLI(t, bin, root, "instructions", "--help")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	if !strings.Contains(out, "Usage: litespec instructions") {
		t.Error("expected instructions usage in help output")
	}
}

func TestCLIArchiveHelp(t *testing.T) {
	bin := buildBinary(t)
	root := t.TempDir()
	out, code := runCLI(t, bin, root, "archive", "--help")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	if !strings.Contains(out, "Usage: litespec archive") {
		t.Error("expected archive usage in help output")
	}
}

func TestCLINewExtraArgs(t *testing.T) {
	bin, root := setupCLITest(t)
	_, code := runCLI(t, bin, root, "new", "foo", "bar")
	if code != 1 {
		t.Fatalf("expected exit 1 for extra args, got %d", code)
	}
}

func TestCLIArchiveExtraArgs(t *testing.T) {
	bin, root := setupCLITest(t)
	_, code := runCLI(t, bin, root, "archive", "foo", "bar")
	if code != 1 {
		t.Fatalf("expected exit 1 for extra args, got %d", code)
	}
}

func TestCLIInitUnknownFlag(t *testing.T) {
	bin := buildBinary(t)
	root := t.TempDir()
	_, code := runCLI(t, bin, root, "init", "--bogus")
	if code != 1 {
		t.Fatalf("expected exit 1 for unknown flag, got %d", code)
	}
}

func TestCLIListUnknownFlag(t *testing.T) {
	bin, root := setupCLITest(t)
	_, code := runCLI(t, bin, root, "list", "--bogus")
	if code != 1 {
		t.Fatalf("expected exit 1 for unknown flag, got %d", code)
	}
}

func TestCLIStatusUnknownFlag(t *testing.T) {
	bin, root := setupCLITest(t)
	_, code := runCLI(t, bin, root, "status", "--bogus")
	if code != 1 {
		t.Fatalf("expected exit 1 for unknown flag, got %d", code)
	}
}

func TestCLIValidateUnknownFlag(t *testing.T) {
	bin, root := setupCLITest(t)
	_, code := runCLI(t, bin, root, "validate", "--bogus")
	if code != 1 {
		t.Fatalf("expected exit 1 for unknown flag, got %d", code)
	}
}

func TestCLIInstructionsUnknownFlag(t *testing.T) {
	bin, root := setupCLITest(t)
	_, code := runCLI(t, bin, root, "instructions", "proposal", "--bogus")
	if code != 1 {
		t.Fatalf("expected exit 1 for unknown flag, got %d", code)
	}
}

func TestCLIArchiveUnknownFlag(t *testing.T) {
	bin, root := setupCLITest(t)
	_, code := runCLI(t, bin, root, "archive", "foo", "--bogus")
	if code != 1 {
		t.Fatalf("expected exit 1 for unknown flag, got %d", code)
	}
}

func TestCLIListSortMissingValue(t *testing.T) {
	bin, root := setupCLITest(t)
	_, code := runCLI(t, bin, root, "list", "--sort")
	if code != 1 {
		t.Fatalf("expected exit 1 for --sort without value, got %d", code)
	}
}

func TestCLIValidateTypeMissingValue(t *testing.T) {
	bin, root := setupCLITest(t)
	_, code := runCLI(t, bin, root, "validate", "foo", "--type")
	if code != 1 {
		t.Fatalf("expected exit 1 for --type without value, got %d", code)
	}
}

func TestCLIArchiveBlocksOnActiveDependent(t *testing.T) {
	bin, root := setupCLITest(t)

	changeDir := filepath.Join(root, "specs", "changes", "parent")
	if err := os.MkdirAll(changeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(changeDir, ".litespec.yaml"), []byte("schema: spec-driven\n"), 0o644)
	os.WriteFile(filepath.Join(changeDir, "proposal.md"), []byte("# Proposal\nTest."), 0o644)
	os.WriteFile(filepath.Join(changeDir, "design.md"), []byte("# Design\nTest."), 0o644)
	os.WriteFile(filepath.Join(changeDir, "tasks.md"), []byte("## Phase 1: Test\n- [x] Task one"), 0o644)
	specsDir := filepath.Join(changeDir, "specs", "cap")
	os.MkdirAll(specsDir, 0o755)
	os.WriteFile(filepath.Join(specsDir, "spec.md"), []byte(`## ADDED Requirements

### Requirement: R1
The system SHALL work.

#### Scenario: S1
- **WHEN** triggered
- **THEN** expected result
`), 0o644)

	childDir := filepath.Join(root, "specs", "changes", "child")
	if err := os.MkdirAll(childDir, 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(childDir, ".litespec.yaml"), []byte("schema: spec-driven\ndependsOn:\n  - parent\n"), 0o644)
	os.WriteFile(filepath.Join(childDir, "proposal.md"), []byte("# Proposal\nTest."), 0o644)
	os.WriteFile(filepath.Join(childDir, "design.md"), []byte("# Design\nTest."), 0o644)
	os.WriteFile(filepath.Join(childDir, "tasks.md"), []byte("## Phase 1: Test\n- [ ] Task one"), 0o644)
	childSpecsDir := filepath.Join(childDir, "specs", "cap2")
	os.MkdirAll(childSpecsDir, 0o755)
	os.WriteFile(filepath.Join(childSpecsDir, "spec.md"), []byte(`## ADDED Requirements

### Requirement: R1
The system SHALL work.

#### Scenario: S1
- **WHEN** triggered
- **THEN** expected result
`), 0o644)

	out, code := runCLI(t, bin, root, "archive", "parent")
	if code != 1 {
		t.Fatalf("expected exit 1 for active dependent, got %d: %s", code, out)
	}
	if !strings.Contains(out, "active changes depend on") {
		t.Errorf("expected dependent warning, got: %s", out)
	}
}

func TestCLIArchiveAllowsIncompleteWithDependents(t *testing.T) {
	bin, root := setupCLITest(t)

	changeDir := filepath.Join(root, "specs", "changes", "parent")
	if err := os.MkdirAll(changeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(changeDir, ".litespec.yaml"), []byte("schema: spec-driven\n"), 0o644)
	os.WriteFile(filepath.Join(changeDir, "proposal.md"), []byte("# Proposal\nTest."), 0o644)
	os.WriteFile(filepath.Join(changeDir, "design.md"), []byte("# Design\nTest."), 0o644)
	os.WriteFile(filepath.Join(changeDir, "tasks.md"), []byte("## Phase 1: Test\n- [x] Task one"), 0o644)
	specsDir := filepath.Join(changeDir, "specs", "cap")
	os.MkdirAll(specsDir, 0o755)
	os.WriteFile(filepath.Join(specsDir, "spec.md"), []byte(`## ADDED Requirements

### Requirement: R1
The system SHALL work.

#### Scenario: S1
- **WHEN** triggered
- **THEN** expected result
`), 0o644)

	childDir := filepath.Join(root, "specs", "changes", "child")
	if err := os.MkdirAll(childDir, 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(childDir, ".litespec.yaml"), []byte("schema: spec-driven\ndependsOn:\n  - parent\n"), 0o644)
	os.WriteFile(filepath.Join(childDir, "proposal.md"), []byte("# Proposal\nTest."), 0o644)
	os.WriteFile(filepath.Join(childDir, "design.md"), []byte("# Design\nTest."), 0o644)
	os.WriteFile(filepath.Join(childDir, "tasks.md"), []byte("## Phase 1: Test\n- [ ] Task one"), 0o644)
	childSpecsDir := filepath.Join(childDir, "specs", "cap2")
	os.MkdirAll(childSpecsDir, 0o755)
	os.WriteFile(filepath.Join(childSpecsDir, "spec.md"), []byte(`## ADDED Requirements

### Requirement: R1
The system SHALL work.

#### Scenario: S1
- **WHEN** triggered
- **THEN** expected result
`), 0o644)

	out, code := runCLI(t, bin, root, "archive", "parent", "--allow-incomplete")
	if code != 0 {
		t.Fatalf("expected success with --allow-incomplete, got %d: %s", code, out)
	}
	if !strings.Contains(out, "archived successfully") {
		t.Errorf("expected archive success, got: %s", out)
	}
	if !strings.Contains(out, "WARN") || !strings.Contains(out, "active changes depend on") {
		t.Errorf("expected warning about active dependents, got: %s", out)
	}
}

func TestCLIListSortDeps(t *testing.T) {
	bin, root := setupCLITest(t)

	parentDir := filepath.Join(root, "specs", "changes", "add-auth")
	if err := os.MkdirAll(parentDir, 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(parentDir, ".litespec.yaml"), []byte("schema: spec-driven\n"), 0o644)

	childDir := filepath.Join(root, "specs", "changes", "add-rate-limiting")
	if err := os.MkdirAll(childDir, 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(childDir, ".litespec.yaml"), []byte("schema: spec-driven\ndependsOn:\n  - add-auth\n"), 0o644)

	out, code := runCLI(t, bin, root, "list", "--sort", "deps", "--json")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("json: %v\n%s", err, out)
	}
	changes := result["changes"].([]interface{})
	if len(changes) != 2 {
		t.Fatalf("expected 2 changes, got %d", len(changes))
	}
	first := changes[0].(map[string]interface{})
	second := changes[1].(map[string]interface{})
	if first["name"] != "add-auth" {
		t.Errorf("first should be add-auth (dep), got %v", first["name"])
	}
	if second["name"] != "add-rate-limiting" {
		t.Errorf("second should be add-rate-limiting (dependent), got %v", second["name"])
	}
	deps, _ := second["dependsOn"].([]interface{})
	if len(deps) != 1 || deps[0] != "add-auth" {
		t.Errorf("expected dependsOn [add-auth], got %v", deps)
	}
}

func TestCLIViewHelp(t *testing.T) {
	bin := buildBinary(t)
	root := t.TempDir()
	out, code := runCLI(t, bin, root, "view", "--help")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	if !strings.Contains(out, "Usage: litespec view") {
		t.Error("expected view usage in help output")
	}
}

func TestCLIViewDashboard(t *testing.T) {
	bin, root := setupCLITest(t)

	createSpec(t, root, "auth")
	createSpec(t, root, "database")

	createChangeWithArtifacts(t, root, "add-auth")
	changeDir := filepath.Join(root, "specs", "changes", "add-auth")
	os.WriteFile(filepath.Join(changeDir, "tasks.md"), []byte("## Phase 1: Test\n- [x] Task one\n- [ ] Task two"), 0o644)

	createChange(t, root, "draft-change")

	out, code := runCLI(t, bin, root, "view")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}

	if !strings.Contains(out, "Summary:") {
		t.Error("expected Summary section")
	}
	if !strings.Contains(out, "Active Changes") {
		t.Error("expected Active Changes section")
	}
	if !strings.Contains(out, "Specifications") {
		t.Error("expected Specifications section")
	}
	if !strings.Contains(out, "Specifications: 2 specs, 2 requirements") {
		t.Error("expected '2 specs, 2 requirements' in summary")
	}
	if !strings.Contains(out, "Active Changes: 1 in progress") {
		t.Error("expected 1 active change")
	}
	if !strings.Contains(out, "Draft Changes: 1") {
		t.Error("expected 1 draft change")
	}
	if !strings.Contains(out, "Task Progress: 1/2 (50% complete)") {
		t.Error("expected 50% task progress")
	}
	if !strings.Contains(out, "auth") {
		t.Error("expected auth spec")
	}
	if !strings.Contains(out, "database") {
		t.Error("expected database spec")
	}
	if !strings.Contains(out, "add-auth") {
		t.Error("expected add-auth change")
	}
	if !strings.Contains(out, "draft-change") {
		t.Error("expected draft-change change")
	}
	if !strings.Contains(out, "█████") {
		t.Error("expected progress bar characters")
	}
	if !strings.Contains(out, "50%") {
		t.Error("expected 50% in progress bar")
	}
}

func TestCLIViewNoSpecs(t *testing.T) {
	bin := buildBinary(t)
	root := t.TempDir()

	specsDir := filepath.Join(root, "specs", "canon")
	if err := os.MkdirAll(specsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	changesDir := filepath.Join(root, "specs", "changes")
	if err := os.MkdirAll(changesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	archiveDir := filepath.Join(root, "specs", "changes", "archive")
	if err := os.MkdirAll(archiveDir, 0o755); err != nil {
		t.Fatal(err)
	}

	out, code := runCLI(t, bin, root, "view")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}

	if !strings.Contains(out, "Summary:") {
		t.Error("expected Summary section")
	}
	if !strings.Contains(out, "Specifications: 0 specs, 0 requirements") {
		t.Error("expected 0 specs in summary")
	}
	if !strings.Contains(out, "Active Changes: 0 in progress") {
		t.Error("expected 0 active changes")
	}
}

func TestCLIViewNoProjectRoot(t *testing.T) {
	bin := buildBinary(t)
	root := t.TempDir()

	out, code := runCLI(t, bin, root, "view")
	if code == 0 {
		t.Fatalf("expected non-zero exit code in dir without specs/: %s", out)
	}
	if !strings.Contains(out, "error") {
		t.Errorf("expected error message, got: %s", out)
	}
}

func TestCLIViewWithDependencyGraph(t *testing.T) {
	bin, root := setupCLITest(t)

	parentDir := filepath.Join(root, "specs", "changes", "parent-change")
	if err := os.MkdirAll(parentDir, 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(parentDir, ".litespec.yaml"), []byte("schema: spec-driven\n"), 0o644)

	childDir := filepath.Join(root, "specs", "changes", "child-change")
	if err := os.MkdirAll(childDir, 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(childDir, ".litespec.yaml"), []byte("schema: spec-driven\ndependsOn:\n  - parent-change\n"), 0o644)

	unrelatedDir := filepath.Join(root, "specs", "changes", "unrelated-change")
	if err := os.MkdirAll(unrelatedDir, 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(unrelatedDir, ".litespec.yaml"), []byte("schema: spec-driven\n"), 0o644)

	out, code := runCLI(t, bin, root, "view")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}

	if !strings.Contains(out, "Dependency Graph") {
		t.Error("expected Dependency Graph section when dependencies exist")
	}
	lines := strings.Split(out, "\n")
	parentIdx := -1
	childIdx := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "parent-change" || strings.HasSuffix(trimmed, "parent-change") {
			if !strings.Contains(line, "Active Changes") && !strings.Contains(line, "Draft Changes") {
				parentIdx = i
			}
		}
		if strings.Contains(line, "child-change") && !strings.Contains(line, "Active Changes") && !strings.Contains(line, "Draft Changes") {
			childIdx = i
		}
	}
	if parentIdx == -1 {
		t.Error("expected parent-change in graph")
	}
	if childIdx == -1 {
		t.Error("expected child-change in graph")
	}
	if parentIdx != -1 && childIdx != -1 && childIdx <= parentIdx {
		t.Errorf("expected parent-change (line %d) before child-change (line %d) in tree", parentIdx, childIdx)
	}
	if !strings.Contains(out, "└──") {
		t.Error("expected box-drawing characters in graph")
	}
	if !strings.Contains(out, "Unrelated:") {
		t.Error("expected Unrelated section for changes with no deps and no dependents")
	}
	if !strings.Contains(out, "unrelated-change") {
		t.Error("expected unrelated-change in unrelated list")
	}
}

func TestCLIViewNoDependencyGraphWhenNoDeps(t *testing.T) {
	bin, root := setupCLITest(t)

	createChange(t, root, "change-a")
	createChange(t, root, "change-b")

	out, code := runCLI(t, bin, root, "view")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}

	if strings.Contains(out, "Dependency Graph") {
		t.Error("expected no Dependency Graph section when no dependencies exist")
	}
}

func TestCLIViewUnknownFlag(t *testing.T) {
	bin, root := setupCLITest(t)
	_, code := runCLI(t, bin, root, "view", "--bogus")
	if code != 1 {
		t.Fatalf("expected exit 1 for unknown flag, got %d", code)
	}
}

func TestCLIValidateChangesDetectsCycle(t *testing.T) {
	bin, root := setupCLITest(t)

	changeDirA := filepath.Join(root, "specs", "changes", "change-a")
	os.MkdirAll(changeDirA, 0o755)
	os.WriteFile(filepath.Join(changeDirA, ".litespec.yaml"), []byte("schema: spec-driven\ndependsOn:\n  - change-b\n"), 0o644)

	changeDirB := filepath.Join(root, "specs", "changes", "change-b")
	os.MkdirAll(changeDirB, 0o755)
	os.WriteFile(filepath.Join(changeDirB, ".litespec.yaml"), []byte("schema: spec-driven\ndependsOn:\n  - change-a\n"), 0o644)

	out, code := runCLI(t, bin, root, "validate", "--changes", "--json")
	if code != 1 {
		t.Fatalf("expected exit 1 for cycle, got %d: %s", code, out)
	}
	if !strings.Contains(out, "cycle") {
		t.Errorf("expected cycle error in output, got: %s", out)
	}
}

func TestCLIValidateAllDetectsCycle(t *testing.T) {
	bin, root := setupCLITest(t)

	changeDirA := filepath.Join(root, "specs", "changes", "change-a")
	os.MkdirAll(changeDirA, 0o755)
	os.WriteFile(filepath.Join(changeDirA, ".litespec.yaml"), []byte("schema: spec-driven\ndependsOn:\n  - change-b\n"), 0o644)

	changeDirB := filepath.Join(root, "specs", "changes", "change-b")
	os.MkdirAll(changeDirB, 0o755)
	os.WriteFile(filepath.Join(changeDirB, ".litespec.yaml"), []byte("schema: spec-driven\ndependsOn:\n  - change-a\n"), 0o644)

	out, code := runCLI(t, bin, root, "validate", "--all", "--json")
	if code != 1 {
		t.Fatalf("expected exit 1 for cycle, got %d: %s", code, out)
	}
	if !strings.Contains(out, "cycle") {
		t.Errorf("expected cycle error in output, got: %s", out)
	}
}

func TestCLIListSortDepsWithCycle(t *testing.T) {
	bin, root := setupCLITest(t)

	changeDirA := filepath.Join(root, "specs", "changes", "change-a")
	os.MkdirAll(changeDirA, 0o755)
	os.WriteFile(filepath.Join(changeDirA, ".litespec.yaml"), []byte("schema: spec-driven\ndependsOn:\n  - change-b\n"), 0o644)

	changeDirB := filepath.Join(root, "specs", "changes", "change-b")
	os.MkdirAll(changeDirB, 0o755)
	os.WriteFile(filepath.Join(changeDirB, ".litespec.yaml"), []byte("schema: spec-driven\ndependsOn:\n  - change-a\n"), 0o644)

	out, code := runCLI(t, bin, root, "list", "--sort", "deps", "--json")
	if code != 0 {
		t.Fatalf("expected exit 0 for sort deps with cycle, got %d: %s", code, out)
	}
	if !strings.Contains(out, "WARN") {
		t.Errorf("expected cycle warning, got: %s", out)
	}

	jsonStart := strings.Index(out, "{")
	if jsonStart < 0 {
		t.Fatalf("no JSON found in output: %s", out)
	}
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(out[jsonStart:]), &result); err != nil {
		t.Fatalf("json: %v\n%s", err, out)
	}
	changes := result["changes"].([]interface{})
	if len(changes) != 2 {
		t.Fatalf("expected 2 changes, got %d", len(changes))
	}
	if changes[0].(map[string]interface{})["name"] != "change-a" {
		t.Errorf("expected alphabetical order, first = %v", changes[0].(map[string]interface{})["name"])
	}
	if changes[1].(map[string]interface{})["name"] != "change-b" {
		t.Errorf("expected alphabetical order, second = %v", changes[1].(map[string]interface{})["name"])
	}
}

func TestValidateChangeName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty", "", true},
		{"path separator slash", "foo/bar", true},
		{"path separator backslash", "foo\\bar", true},
		{"traversal double dot", "..", true},
		{"traversal embedded", "foo..bar", true},
		{"leading whitespace", " foo", true},
		{"trailing whitespace", "foo ", true},
		{"reserved canon", "canon", true},
		{"reserved changes", "changes", true},
		{"reserved archive", "archive", true},
		{"too long", strings.Repeat("a", 101), true},
		{"valid simple", "add-auth", false},
		{"valid with numbers", "fix-123-issue", false},
		{"valid at limit", strings.Repeat("a", 100), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateChangeName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateChangeName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestCLINewInvalidName(t *testing.T) {
	bin, root := setupCLITest(t)
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"path separator", "foo/bar"},
		{"traversal", ".."},
		{"reserved", "canon"},
		{"whitespace padded", " foo "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []string{"new"}
			if tt.input != "" {
				args = append(args, tt.input)
			}
			_, code := runCLI(t, bin, root, args...)
			if code != 1 {
				t.Errorf("expected exit 1 for name %q, got %d", tt.input, code)
			}
		})
	}
}

func TestCLIInitUnknownTool(t *testing.T) {
	bin := buildBinary(t)
	root := t.TempDir()
	_, code := runCLI(t, bin, root, "init", "--tools", "unknown-tool")
	if code != 1 {
		t.Fatalf("expected exit 1 for unknown tool, got %d", code)
	}
}

func TestCLIInitKnownTool(t *testing.T) {
	bin := buildBinary(t)
	root := t.TempDir()
	out, code := runCLI(t, bin, root, "init", "--tools", "claude")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	if !strings.Contains(out, "adapter commands") {
		t.Errorf("expected adapter commands output, got: %s", out)
	}
}

func TestCLIUpdateUnknownTool(t *testing.T) {
	bin, root := setupCLITest(t)
	_, code := runCLI(t, bin, root, "update", "--tools", "bogus")
	if code != 1 {
		t.Fatalf("expected exit 1 for unknown tool, got %d", code)
	}
}

func setupDirectTest(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	dirs := []string{
		filepath.Join(root, "specs", "canon"),
		filepath.Join(root, "specs", "changes"),
		filepath.Join(root, "specs", "changes", "archive"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(root)
	return root
}

func TestCmdNewDirect_HappyPath(t *testing.T) {
	root := setupDirectTest(t)
	err := cmdNew([]string{"my-change"})
	if err != nil {
		t.Fatalf("cmdNew: %v", err)
	}
	changeDir := filepath.Join(root, "specs", "changes", "my-change")
	if _, statErr := os.Stat(changeDir); os.IsNotExist(statErr) {
		t.Error("expected change directory to exist")
	}
	meta, metaErr := os.ReadFile(filepath.Join(changeDir, ".litespec.yaml"))
	if metaErr != nil {
		t.Fatalf("reading metadata: %v", metaErr)
	}
	if !strings.Contains(string(meta), "spec-driven") {
		t.Errorf("expected spec-driven schema in metadata, got: %s", string(meta))
	}
}

func TestCmdNewDirect_MissingName(t *testing.T) {
	setupDirectTest(t)
	err := cmdNew([]string{})
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestCmdNewDirect_InvalidName(t *testing.T) {
	setupDirectTest(t)
	err := cmdNew([]string{"foo/bar"})
	if err == nil {
		t.Fatal("expected error for invalid name")
	}
}

func TestCmdStatusDirect_NonexistentChange(t *testing.T) {
	setupDirectTest(t)
	err := cmdStatus([]string{"nonexistent"})
	if err == nil {
		t.Fatal("expected error for nonexistent change")
	}
}

func TestCmdStatusDirect_AllChangesEmpty(t *testing.T) {
	setupDirectTest(t)
	err := cmdStatus([]string{})
	if err != nil {
		t.Fatalf("cmdStatus (all): %v", err)
	}
}

func TestCmdStatusDirect_SpecificChange(t *testing.T) {
	root := setupDirectTest(t)
	createChange(t, root, "test-change")
	err := cmdStatus([]string{"test-change"})
	if err != nil {
		t.Fatalf("cmdStatus: %v", err)
	}
}

func TestCmdListDirect_EmptyChanges(t *testing.T) {
	setupDirectTest(t)
	err := cmdList([]string{})
	if err != nil {
		t.Fatalf("cmdList: %v", err)
	}
}

func TestCmdListDirect_Specs(t *testing.T) {
	root := setupDirectTest(t)
	createSpec(t, root, "auth")
	err := cmdList([]string{"--specs"})
	if err != nil {
		t.Fatalf("cmdList --specs: %v", err)
	}
}

func TestCmdListDirect_InvalidSort(t *testing.T) {
	setupDirectTest(t)
	err := cmdList([]string{"--sort", "invalid"})
	if err == nil {
		t.Fatal("expected error for invalid sort value")
	}
}

func TestCmdValidateDirect_InvalidType(t *testing.T) {
	setupDirectTest(t)
	err := cmdValidate([]string{"foo", "--type", "invalid"})
	if err == nil {
		t.Fatal("expected error for invalid --type value")
	}
}

func TestCmdValidateDirect_NameWithBulk(t *testing.T) {
	setupDirectTest(t)
	err := cmdValidate([]string{"foo", "--all"})
	if err == nil {
		t.Fatal("expected error for name + bulk flag")
	}
}

func TestCmdValidateDirect_TypeWithoutName(t *testing.T) {
	setupDirectTest(t)
	err := cmdValidate([]string{"--type", "change"})
	if err == nil {
		t.Fatal("expected error for --type without name")
	}
}

func TestCmdInstructionsDirect_NoArgs(t *testing.T) {
	setupDirectTest(t)
	err := cmdInstructions([]string{})
	if err == nil {
		t.Fatal("expected error for no args")
	}
}

func TestCmdInstructionsDirect_UnknownArtifact(t *testing.T) {
	setupDirectTest(t)
	err := cmdInstructions([]string{"bogus"})
	if err == nil {
		t.Fatal("expected error for unknown artifact")
	}
}

func TestCmdArchiveDirect_NoArgs(t *testing.T) {
	setupDirectTest(t)
	err := cmdArchive([]string{})
	if err == nil {
		t.Fatal("expected error for no args")
	}
}

func TestCmdInitDirect_UnknownFlag(t *testing.T) {
	err := cmdInit([]string{"--bogus"})
	if err == nil {
		t.Fatal("expected error for unknown flag")
	}
}

func TestCmdCompletionDirect_NoArgs(t *testing.T) {
	err := cmdCompletion([]string{})
	if err == nil {
		t.Fatal("expected error for no args")
	}
}

func TestCmdCompletionDirect_InvalidShell(t *testing.T) {
	err := cmdCompletion([]string{"powershell"})
	if err == nil {
		t.Fatal("expected error for invalid shell")
	}
}

func TestCLIStatusJSONWarningOnError(t *testing.T) {
	bin, root := setupCLITest(t)

	goodDir := filepath.Join(root, "specs", "changes", "good-change")
	if err := os.MkdirAll(goodDir, 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(goodDir, ".litespec.yaml"), []byte("schema: spec-driven\n"), 0o644)

	badDir := filepath.Join(root, "specs", "changes", "bad-change")
	if err := os.MkdirAll(badDir, 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(badDir, ".litespec.yaml"), []byte("key: [unclosed\n"), 0o644)

	out, code := runCLI(t, bin, root, "status", "--json")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("json: %v\n%s", err, out)
	}
	warnings, ok := result["warnings"].([]interface{})
	if !ok || len(warnings) == 0 {
		t.Errorf("expected warnings in output, got: %v", result["warnings"])
	}
}

func TestCmdViewDirect_HappyPath(t *testing.T) {
	root := setupDirectTest(t)
	createSpec(t, root, "auth")
	createChangeWithArtifacts(t, root, "add-auth")
	if err := cmdView([]string{}); err != nil {
		t.Fatalf("cmdView: %v", err)
	}
}

func TestCmdViewDirect_NoProjectRoot(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	err := cmdView([]string{})
	if err == nil {
		t.Fatal("expected error when no project root")
	}
}

func TestCmdUpdateDirect_HappyPath(t *testing.T) {
	root := setupDirectTest(t)
	if err := internal.InitProject(root); err != nil {
		t.Fatal(err)
	}
	if err := cmdUpdate([]string{}); err != nil {
		t.Fatalf("cmdUpdate: %v", err)
	}
}

func TestCmdUpdateDirect_WithTools(t *testing.T) {
	root := setupDirectTest(t)
	if err := internal.InitProject(root); err != nil {
		t.Fatal(err)
	}
	if err := cmdUpdate([]string{"--tools", "claude"}); err != nil {
		t.Fatalf("cmdUpdate --tools claude: %v", err)
	}
}

func TestCmdUpdateDirect_UnknownTool(t *testing.T) {
	root := setupDirectTest(t)
	if err := internal.InitProject(root); err != nil {
		t.Fatal(err)
	}
	err := cmdUpdate([]string{"--tools", "bogus"})
	if err == nil {
		t.Fatal("expected error for unknown tool")
	}
}

func TestCreateProgressBar_Half(t *testing.T) {
	got := createProgressBar(5, 10, 10)
	want := "[█████░░░░░]"
	if got != want {
		t.Errorf("createProgressBar(5,10,10) = %q, want %q", got, want)
	}
}

func TestCreateProgressBar_Zero(t *testing.T) {
	got := createProgressBar(0, 5, 10)
	want := "[░░░░░░░░░░]"
	if got != want {
		t.Errorf("createProgressBar(0,5,10) = %q, want %q", got, want)
	}
}

func TestCreateProgressBar_Complete(t *testing.T) {
	got := createProgressBar(5, 5, 10)
	want := "[██████████]"
	if got != want {
		t.Errorf("createProgressBar(5,5,10) = %q, want %q", got, want)
	}
}

func TestCreateProgressBar_ZeroTotal(t *testing.T) {
	got := createProgressBar(0, 0, 10)
	want := "──────────"
	if got != want {
		t.Errorf("createProgressBar(0,0,10) = %q, want %q", got, want)
	}
}

func TestCreateProgressBar_Width(t *testing.T) {
	got := createProgressBar(3, 6, 20)
	if utf8.RuneCountInString(got) != 22 {
		t.Errorf("rune count of createProgressBar(3,6,20) = %d, want 22", utf8.RuneCountInString(got))
	}
}

func captureStdout(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestRenderDependencyGraph_SimpleTree(t *testing.T) {
	depMap := map[string][]string{
		"child": {"parent"},
	}
	changes := []internal.ChangeInfo{
		{Name: "parent"},
		{Name: "child"},
	}

	output := captureStdout(func() {
		renderDependencyGraph(depMap, changes)
	})

	parentIdx := strings.Index(output, "parent")
	childIdx := strings.Index(output, "child")
	if parentIdx == -1 {
		t.Error("expected parent in output")
	}
	if childIdx == -1 {
		t.Error("expected child in output")
	}
	if parentIdx != -1 && childIdx != -1 && childIdx <= parentIdx {
		t.Error("expected parent before child in tree")
	}
	if !strings.Contains(output, "└──") {
		t.Error("expected └── connector in output")
	}
}

func TestRenderDependencyGraph_MultipleChildren(t *testing.T) {
	depMap := map[string][]string{
		"child-a": {"parent"},
		"child-b": {"parent"},
	}
	changes := []internal.ChangeInfo{
		{Name: "parent"},
		{Name: "child-a"},
		{Name: "child-b"},
	}

	output := captureStdout(func() {
		renderDependencyGraph(depMap, changes)
	})

	if !strings.Contains(output, "child-a") {
		t.Error("expected child-a in output")
	}
	if !strings.Contains(output, "child-b") {
		t.Error("expected child-b in output")
	}
	if !strings.Contains(output, "parent") {
		t.Error("expected parent in output")
	}
}

func TestRenderDependencyGraph_UnrelatedChanges(t *testing.T) {
	depMap := map[string][]string{
		"child": {"parent"},
	}
	changes := []internal.ChangeInfo{
		{Name: "parent"},
		{Name: "child"},
		{Name: "lonely"},
	}

	output := captureStdout(func() {
		renderDependencyGraph(depMap, changes)
	})

	if !strings.Contains(output, "Unrelated:") {
		t.Error("expected Unrelated section in output")
	}
	if !strings.Contains(output, "lonely") {
		t.Error("expected lonely change in Unrelated section")
	}
}

func TestRenderDependencyGraph_DeepChain(t *testing.T) {
	depMap := map[string][]string{
		"B": {"A"},
		"C": {"B"},
	}
	changes := []internal.ChangeInfo{
		{Name: "A"},
		{Name: "B"},
		{Name: "C"},
	}

	output := captureStdout(func() {
		renderDependencyGraph(depMap, changes)
	})

	lines := strings.Split(strings.TrimSpace(output), "\n")

	aIdx := -1
	bIdx := -1
	cIdx := -1
	for i, line := range lines {
		if strings.Contains(line, "A") && !strings.Contains(line, "Unrelated") {
			aIdx = i
		}
		if strings.Contains(line, "B") && !strings.Contains(line, "Unrelated") {
			bIdx = i
		}
		if strings.Contains(line, "C") && !strings.Contains(line, "Unrelated") {
			cIdx = i
		}
	}

	if aIdx == -1 || bIdx == -1 || cIdx == -1 {
		t.Fatalf("expected A, B, C in output:\n%s", output)
	}
	if !(aIdx < bIdx && bIdx < cIdx) {
		t.Errorf("expected A < B < C line order, got A=%d B=%d C=%d", aIdx, bIdx, cIdx)
	}

	cLine := lines[cIdx]
	if !strings.HasPrefix(strings.TrimSpace(cLine), "└── C") && !strings.HasPrefix(strings.TrimSpace(cLine), "├── C") {
		t.Errorf("expected C to have tree connector, got: %q", cLine)
	}

	indentA := len(lines[aIdx]) - len(strings.TrimLeft(lines[aIdx], " │├└─"))
	indentC := len(lines[cIdx]) - len(strings.TrimLeft(lines[cIdx], " │├└─"))
	if indentC <= indentA {
		t.Errorf("expected C to be indented deeper than A, got A indent=%d C indent=%d", indentA, indentC)
	}
}

func TestCmdValidateDirect_AllJSON(t *testing.T) {
	setupDirectTest(t)
	if err := cmdValidate([]string{"--all", "--json"}); err != nil {
		t.Fatalf("cmdValidate --all --json on empty project: %v", err)
	}
}

func TestCmdValidateDirect_ChangesJSON(t *testing.T) {
	setupDirectTest(t)
	if err := cmdValidate([]string{"--changes", "--json"}); err != nil {
		t.Fatalf("cmdValidate --changes --json on empty project: %v", err)
	}
}

func TestCmdValidateDirect_SpecsJSON(t *testing.T) {
	setupDirectTest(t)
	if err := cmdValidate([]string{"--specs", "--json"}); err != nil {
		t.Fatalf("cmdValidate --specs --json on empty project: %v", err)
	}
}

func TestCmdValidateDirect_SpecificChange(t *testing.T) {
	root := setupDirectTest(t)
	createChangeWithArtifacts(t, root, "my-change")
	if err := cmdValidate([]string{"my-change"}); err != nil {
		t.Fatalf("cmdValidate my-change: %v", err)
	}
}

func TestCmdValidateDirect_SpecificSpec(t *testing.T) {
	root := setupDirectTest(t)
	createSpec(t, root, "auth")
	if err := cmdValidate([]string{"auth"}); err != nil {
		t.Fatalf("cmdValidate auth: %v", err)
	}
}

func TestCmdValidateDirect_StrictOnEmpty(t *testing.T) {
	setupDirectTest(t)
	if err := cmdValidate([]string{"--strict", "--all"}); err != nil {
		t.Fatalf("cmdValidate --strict --all on empty project: %v", err)
	}
}

func TestCmdValidateDirect_DefaultBulk(t *testing.T) {
	setupDirectTest(t)
	if err := cmdValidate([]string{}); err != nil {
		t.Fatalf("cmdValidate with no args on empty project: %v", err)
	}
}

func setupEmptyDir(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	t.Chdir(root)
	return root
}

func TestCmdArchiveDirect_HappyPath(t *testing.T) {
	root := setupDirectTest(t)
	createChangeWithArtifacts(t, root, "my-change")
	tasksPath := filepath.Join(root, "specs", "changes", "my-change", "tasks.md")
	os.WriteFile(tasksPath, []byte("## Phase 1\n- [x] Done"), 0o644)
	err := cmdArchive([]string{"my-change"})
	if err != nil {
		t.Fatalf("cmdArchive: %v", err)
	}
}

func TestCmdArchiveDirect_UnknownFlag(t *testing.T) {
	setupDirectTest(t)
	err := cmdArchive([]string{"foo", "--bogus"})
	if err == nil {
		t.Fatal("expected error for unknown flag")
	}
}

func TestCmdArchiveDirect_NonexistentChange(t *testing.T) {
	setupDirectTest(t)
	err := cmdArchive([]string{"nonexistent"})
	if err == nil {
		t.Fatal("expected error for nonexistent change")
	}
}

func TestCmdArchiveDirect_IncompleteTasks(t *testing.T) {
	root := setupDirectTest(t)
	createChangeWithArtifacts(t, root, "my-change")
	err := cmdArchive([]string{"my-change"})
	if err == nil {
		t.Fatal("expected error for incomplete tasks")
	}
	if !strings.Contains(err.Error(), "tasks") {
		t.Errorf("expected tasks error, got: %v", err)
	}
}

func TestCmdArchiveDirect_AllowIncomplete(t *testing.T) {
	root := setupDirectTest(t)
	createChangeWithArtifacts(t, root, "my-change")
	err := cmdArchive([]string{"my-change", "--allow-incomplete"})
	if err != nil {
		t.Fatalf("cmdArchive --allow-incomplete: %v", err)
	}
}

func TestCmdInitDirect_HappyPath(t *testing.T) {
	root := setupEmptyDir(t)
	if err := cmdInit([]string{}); err != nil {
		t.Fatalf("cmdInit: %v", err)
	}
	for _, dir := range []string{
		filepath.Join(root, "specs", "canon"),
		filepath.Join(root, "specs", "changes"),
		filepath.Join(root, "specs", "changes", "archive"),
		filepath.Join(root, ".agents", "skills"),
	} {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("expected %s to exist", dir)
		}
	}
}

func TestCmdInitDirect_WithTools(t *testing.T) {
	root := setupEmptyDir(t)
	if err := cmdInit([]string{"--tools", "claude"}); err != nil {
		t.Fatalf("cmdInit --tools claude: %v", err)
	}
	claudeSkills := filepath.Join(root, ".claude", "skills")
	if _, err := os.Stat(claudeSkills); os.IsNotExist(err) {
		t.Fatal("expected .claude/skills/ to exist")
	}
	entries, err := os.ReadDir(claudeSkills)
	if err != nil {
		t.Fatalf("reading .claude/skills: %v", err)
	}
	if len(entries) == 0 {
		t.Error("expected symlinks in .claude/skills/")
	}
	for _, e := range entries {
		linkPath := filepath.Join(claudeSkills, e.Name())
		target, err := os.Readlink(linkPath)
		if err != nil {
			t.Errorf("expected %s to be a symlink: %v", e.Name(), err)
		}
		resolved := filepath.Join(claudeSkills, target)
		if _, statErr := os.Stat(resolved); os.IsNotExist(statErr) {
			t.Errorf("symlink %s target %s does not exist", e.Name(), target)
		}
	}
}

func TestCmdInitDirect_UnknownTool(t *testing.T) {
	setupEmptyDir(t)
	err := cmdInit([]string{"--tools", "bogus"})
	if err == nil {
		t.Fatal("expected error for unknown tool")
	}
}

func TestCmdInitDirect_HelpFlag(t *testing.T) {
	setupEmptyDir(t)
	if err := cmdInit([]string{"--help"}); err != nil {
		t.Fatalf("cmdInit --help: %v", err)
	}
}

func TestMarshalJSONErrorPropagation(t *testing.T) {
	_, err := internal.MarshalJSON(map[string]chan int{"ch": make(chan int)})
	if err == nil {
		t.Fatal("expected error for unmarshallable value")
	}
}

func TestFindProjectRoot_InProjectRoot(t *testing.T) {
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "specs"), 0o755); err != nil {
		t.Fatal(err)
	}
	os.Chdir(root)

	got, err := internal.FindProjectRoot()
	if err != nil {
		t.Fatalf("FindProjectRoot: %v", err)
	}
	if got != root {
		t.Errorf("got %q, want %q", got, root)
	}
}

func TestIsGoInstall_InGOBIN(t *testing.T) {
	dir := t.TempDir()
	binPath := filepath.Join(dir, "litespec")
	if err := os.WriteFile(binPath, nil, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOBIN", dir)
	t.Setenv("GOPATH", "")
	if !isGoInstallCheck(t, binPath) {
		t.Error("expected true for binary in GOBIN")
	}
}

func TestIsGoInstall_InGOPATHBin(t *testing.T) {
	dir := t.TempDir()
	gobinDir := filepath.Join(dir, "bin")
	if err := os.MkdirAll(gobinDir, 0o755); err != nil {
		t.Fatal(err)
	}
	binPath := filepath.Join(gobinDir, "litespec")
	if err := os.WriteFile(binPath, nil, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOBIN", "")
	t.Setenv("GOPATH", dir)
	if !isGoInstallCheck(t, binPath) {
		t.Error("expected true for binary in GOPATH/bin")
	}
}

func TestIsGoInstall_Elsewhere(t *testing.T) {
	dir := t.TempDir()
	binPath := filepath.Join(dir, "litespec")
	if err := os.WriteFile(binPath, nil, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOBIN", "/nonexistent")
	t.Setenv("GOPATH", "/nonexistent")
	if isGoInstallCheck(t, binPath) {
		t.Error("expected false for binary outside GOBIN/GOPATH/bin")
	}
}

func TestIsGoInstall_DefaultGOPATH(t *testing.T) {
	home := t.TempDir()
	gobinDir := filepath.Join(home, "go", "bin")
	if err := os.MkdirAll(gobinDir, 0o755); err != nil {
		t.Fatal(err)
	}
	binPath := filepath.Join(gobinDir, "litespec")
	if err := os.WriteFile(binPath, nil, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOBIN", "")
	t.Setenv("GOPATH", "")
	t.Setenv("HOME", home)
	if !isGoInstallCheck(t, binPath) {
		t.Error("expected true for binary in ~/go/bin with empty GOPATH")
	}
}

func isGoInstallCheck(t *testing.T, exePath string) bool {
	t.Helper()
	if gobin := os.Getenv("GOBIN"); gobin != "" {
		if strings.HasPrefix(exePath, filepath.Clean(gobin)+string(os.PathSeparator)) {
			return true
		}
	}
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = filepath.Join(os.Getenv("HOME"), "go")
	}
	gobinDefault := filepath.Join(gopath, "bin")
	return strings.HasPrefix(exePath, filepath.Clean(gobinDefault)+string(os.PathSeparator))
}

func TestParseSemver_Valid(t *testing.T) {
	tests := []struct {
		input    string
		majorExp int
		minorExp int
		patchExp int
	}{
		{"v1.2.3", 1, 2, 3},
		{"0.1.0", 0, 1, 0},
		{"v10.20.30", 10, 20, 30},
		{"v1.2.3-alpha", 1, 2, 3},
		{"v1.2.3-beta.1+build", 1, 2, 3},
	}
	for _, tt := range tests {
		major, minor, patch, err := parseSemver(tt.input)
		if err != nil {
			t.Errorf("parseSemver(%q): %v", tt.input, err)
			continue
		}
		if major != tt.majorExp || minor != tt.minorExp || patch != tt.patchExp {
			t.Errorf("parseSemver(%q) = %d.%d.%d, want %d.%d.%d", tt.input, major, minor, patch, tt.majorExp, tt.minorExp, tt.patchExp)
		}
	}
}

func TestParseSemver_Invalid(t *testing.T) {
	tests := []string{"", "1", "1.2", "a.b.c", "v1.2.x"}
	for _, input := range tests {
		_, _, _, err := parseSemver(input)
		if err == nil {
			t.Errorf("parseSemver(%q): expected error", input)
		}
	}
}

func TestCompareSemver(t *testing.T) {
	tests := []struct {
		local, remote string
		want          int
	}{
		{"0.1.0", "0.1.0", 0},
		{"0.1.0", "0.2.0", -1},
		{"0.2.0", "0.1.0", 1},
		{"1.0.0", "0.9.9", 1},
		{"0.9.9", "1.0.0", -1},
		{"0.1.0", "0.1.1", -1},
	}
	for _, tt := range tests {
		got, err := compareSemver(tt.local, tt.remote)
		if err != nil {
			t.Errorf("compareSemver(%q, %q): %v", tt.local, tt.remote, err)
			continue
		}
		if got != tt.want {
			t.Errorf("compareSemver(%q, %q) = %d, want %d", tt.local, tt.remote, got, tt.want)
		}
	}
}

func TestGetModulePath(t *testing.T) {
	path, err := getModulePath()
	if err != nil {
		t.Fatalf("getModulePath(): %v", err)
	}
	if path == "" {
		t.Error("expected non-empty module path")
	}
	if !strings.HasPrefix(path, "github.com/bermudi/litespec") {
		t.Errorf("got %q, want path starting with github.com/bermudi/litespec", path)
	}
}

func TestFetchLatestVersion_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"tag_name": "v0.2.0"}`)
	}))
	defer server.Close()

	tag, err := fetchLatestVersionFromURL(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag != "v0.2.0" {
		t.Errorf("got %q, want v0.2.0", tag)
	}
}

func TestFetchLatestVersion_Non200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	_, err := fetchLatestVersionFromURL(server.URL)
	if err == nil {
		t.Error("expected error for non-200 response")
	}
}

func TestFetchLatestVersion_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `not json`)
	}))
	defer server.Close()

	_, err := fetchLatestVersionFromURL(server.URL)
	if err == nil {
		t.Error("expected error for malformed JSON")
	}
}

func TestFetchLatestVersion_EmptyTag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"tag_name": ""}`)
	}))
	defer server.Close()

	_, err := fetchLatestVersionFromURL(server.URL)
	if err == nil {
		t.Error("expected error for empty tag")
	}
}

func TestCLIUpgradeHelp(t *testing.T) {
	bin := buildBinary(t)
	root := t.TempDir()
	out, code := runCLI(t, bin, root, "upgrade", "--help")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	if !strings.Contains(out, "Usage: litespec upgrade") {
		t.Error("expected upgrade usage in help output")
	}
}

func TestCLIUpgrade_NotGoInstall(t *testing.T) {
	bin := buildBinary(t)
	root := t.TempDir()
	out, code := runCLI(t, bin, root, "upgrade")
	if code == 0 {
		t.Fatal("expected non-zero exit for non-go-install binary")
	}
	if !strings.Contains(out, "go install") {
		t.Errorf("expected go install error message, got: %s", out)
	}
}

func TestMaybeBackgroundUpgrade_SkipsWhenNotGoInstall(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("GOBIN", "/nonexistent")
	t.Setenv("GOPATH", "/nonexistent")

	cacheDir := filepath.Join(home, ".cache", "litespec")
	stampFile := filepath.Join(cacheDir, "last-update-check")
	if _, err := os.Stat(stampFile); !os.IsNotExist(err) {
		t.Error("expected no stamp file when not go install")
	}
}

func TestMaybeBackgroundUpgrade_SkipsWhenRecent(t *testing.T) {
	home := t.TempDir()
	cacheDir := filepath.Join(home, ".cache", "litespec")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		t.Fatal(err)
	}
	stampFile := filepath.Join(cacheDir, "last-update-check")
	if err := os.WriteFile(stampFile, nil, 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", home)

	info, err := os.Stat(stampFile)
	if err != nil {
		t.Fatal(err)
	}
	if time.Since(info.ModTime()) >= backgroundUpgradeInterval {
		t.Error("stamp should be recent")
	}
}

func TestMaybeBackgroundUpgrade_FiresWhenExpired(t *testing.T) {
	home := t.TempDir()
	cacheDir := filepath.Join(home, ".cache", "litespec")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		t.Fatal(err)
	}
	stampFile := filepath.Join(cacheDir, "last-update-check")
	oldTime := time.Now().Add(-8 * 24 * time.Hour)
	if err := os.WriteFile(stampFile, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(stampFile, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", home)

	info, _ := os.Stat(stampFile)
	if info != nil && time.Since(info.ModTime()) < backgroundUpgradeInterval {
		t.Error("stamp should be expired")
	}
}

func TestMaybeBackgroundUpgrade_FiresWhenNoStamp(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	stampFile := filepath.Join(home, ".cache", "litespec", "last-update-check")
	if _, err := os.Stat(stampFile); !os.IsNotExist(err) {
		t.Error("expected no stamp file")
	}
}

func TestFindProjectRoot_SymlinkedSpecs(t *testing.T) {
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)

	root := t.TempDir()
	realSpecs := t.TempDir()
	if err := os.MkdirAll(filepath.Join(realSpecs, "specs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(realSpecs, "specs"), filepath.Join(root, "specs")); err != nil {
		t.Fatal(err)
	}
	os.Chdir(root)

	got, err := internal.FindProjectRoot()
	if err != nil {
		t.Fatalf("FindProjectRoot: %v", err)
	}
	if got != root {
		t.Errorf("got %q, want %q", got, root)
	}
}
