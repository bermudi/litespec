package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

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
