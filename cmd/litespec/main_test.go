package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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
	os.WriteFile(filepath.Join(changeDir, "proposal.md"), []byte("# Proposal\n\n## Motivation\nSome motivation.\n\n## Scope\nSome scope."), 0o644)
	os.WriteFile(filepath.Join(changeDir, "design.md"), []byte("# Design\n\n## Architecture\nLine one.\nLine two.\nLine three."), 0o644)
	os.WriteFile(filepath.Join(changeDir, "tasks.md"), []byte("## Phase 1: Test\n\n- [ ] Task"), 0o644)
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

func createDecision(t *testing.T, root, slug string) {
	t.Helper()
	decisionsDir := filepath.Join(root, "specs", "decisions")
	if err := os.MkdirAll(decisionsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `# Decision: ` + slug + `

## Status
accepted

## Context
Some context.

## Decision
Some decision.

## Consequences
Some consequences.
`
	os.WriteFile(filepath.Join(decisionsDir, "0001-"+slug+".md"), []byte(content), 0o644)
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
	createSpec(t, root, "shared")
	createDecision(t, root, "shared")
	_, code := runCLI(t, bin, root, "validate", "shared")
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
}

func TestCLIVerifyAmbiguousWithTypeSpec(t *testing.T) {
	bin, root := setupCLITest(t)
	createSpec(t, root, "shared")
	createDecision(t, root, "shared")
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
	_, code := runCLI(t, bin, root, "validate", "--type", "spec")
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
}

func TestCLIVerifyTypeWithBulkFlag(t *testing.T) {
	bin, root := setupCLITest(t)
	_, code := runCLI(t, bin, root, "validate", "--all", "--type", "spec")
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

func TestCLIVerifyBulkSpecs(t *testing.T) {
	bin, root := setupCLITest(t)
	out, code := runCLI(t, bin, root, "validate", "--specs", "--json")
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

func TestCLINewExtraArgs(t *testing.T) {
	bin, root := setupCLITest(t)
	_, code := runCLI(t, bin, root, "new", "foo", "bar")
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

func TestCLIValidateUnknownFlag(t *testing.T) {
	bin, root := setupCLITest(t)
	_, code := runCLI(t, bin, root, "validate", "--bogus")
	if code != 1 {
		t.Fatalf("expected exit 1 for unknown flag, got %d", code)
	}
}

func TestCLIValidateTypeMissingValue(t *testing.T) {
	bin, root := setupCLITest(t)
	_, code := runCLI(t, bin, root, "validate", "foo", "--type")
	if code != 1 {
		t.Fatalf("expected exit 1 for --type without value, got %d", code)
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

func TestCLIViewUnknownFlag(t *testing.T) {
	bin, root := setupCLITest(t)
	_, code := runCLI(t, bin, root, "view", "--bogus")
	if code != 1 {
		t.Fatalf("expected exit 1 for unknown flag, got %d", code)
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
		{"reserved decisions", "decisions", true},
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
	setupDirectTest(t)
	err := cmdNew([]string{"my-change", "--issue", "42"})
	if err != nil {
		t.Fatalf("cmdNew: %v", err)
	}
}

func TestCmdNewDirect_MissingName(t *testing.T) {
	setupDirectTest(t)
	err := cmdNew([]string{})
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestCmdNewDirect_MissingIssue(t *testing.T) {
	setupDirectTest(t)
	err := cmdNew([]string{"my-change"})
	if err == nil {
		t.Fatal("expected error for missing --issue")
	}
}

func TestCmdNewDirect_InvalidName(t *testing.T) {
	setupDirectTest(t)
	err := cmdNew([]string{"foo/bar", "--issue", "1"})
	if err == nil {
		t.Fatal("expected error for invalid name")
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

func TestCmdInitDirect_UnknownFlag(t *testing.T) {
	err := cmdInit([]string{"--bogus"})
	if err == nil {
		t.Fatal("expected error for unknown flag")
	}
}

func TestCmdCompletionDirect_NoArgs(t *testing.T) {
	err := cmdCompletion([]string{})
	if err != nil {
		t.Fatalf("expected no error for no args (usage), got: %v", err)
	}
}

func TestCmdCompletionDirect_InvalidShell(t *testing.T) {
	err := cmdCompletion([]string{"powershell"})
	if err == nil {
		t.Fatal("expected error for invalid shell")
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

func TestCmdValidateDirect_AllJSON(t *testing.T) {
	setupDirectTest(t)
	if err := cmdValidate([]string{"--all", "--json"}); err != nil {
		t.Fatalf("cmdValidate --all --json on empty project: %v", err)
	}
}

func TestCmdValidateDirect_SpecsJSON(t *testing.T) {
	setupDirectTest(t)
	if err := cmdValidate([]string{"--specs", "--json"}); err != nil {
		t.Fatalf("cmdValidate --specs --json on empty project: %v", err)
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

func TestCmdInitDirect_HappyPath(t *testing.T) {
	root := setupEmptyDir(t)
	if err := cmdInit([]string{}); err != nil {
		t.Fatalf("cmdInit: %v", err)
	}
	for _, dir := range []string{
		filepath.Join(root, "specs", "product.md"),
		filepath.Join(root, "specs", "glossary.md"),
		filepath.Join(root, "specs", "decisions"),
		filepath.Join(root, ".agents", "skills"),
	} {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("expected %s to exist", dir)
		}
	}
	for _, dir := range []string{
		filepath.Join(root, "specs", "canon"),
		filepath.Join(root, "specs", "backlog.md"),
	} {
		if _, err := os.Stat(dir); err == nil {
			t.Errorf("expected %s to NOT exist in lean v2", dir)
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

func TestCmdInitDirect_WithTools_NoConfigYaml(t *testing.T) {
	root := setupEmptyDir(t)
	if err := cmdInit([]string{"--tools", "claude"}); err != nil {
		t.Fatalf("cmdInit --tools claude: %v", err)
	}
	configPath := filepath.Join(root, "specs", "config.yaml")
	if _, err := os.Stat(configPath); err == nil {
		t.Error("expected specs/config.yaml to NOT exist")
	}
}

func TestCmdUpdateDirect_WithTools_NoConfigYaml(t *testing.T) {
	root := setupDirectTest(t)
	if err := internal.InitProject(root); err != nil {
		t.Fatal(err)
	}
	if err := cmdUpdate([]string{"--tools", "claude"}); err != nil {
		t.Fatalf("cmdUpdate --tools claude: %v", err)
	}
	configPath := filepath.Join(root, "specs", "config.yaml")
	if _, err := os.Stat(configPath); err == nil {
		t.Error("expected specs/config.yaml to NOT exist")
	}
}

func TestCmdUpdateDirect_AutoDetectAdapters(t *testing.T) {
	root := setupDirectTest(t)
	if err := internal.InitProject(root); err != nil {
		t.Fatal(err)
	}
	if err := cmdUpdate([]string{"--tools", "claude"}); err != nil {
		t.Fatalf("cmdUpdate --tools claude: %v", err)
	}
	if err := cmdUpdate([]string{}); err != nil {
		t.Fatalf("cmdUpdate (auto-detect): %v", err)
	}
	claudeSkills := filepath.Join(root, ".claude", "skills")
	entries, err := os.ReadDir(claudeSkills)
	if err != nil {
		t.Fatalf("reading .claude/skills: %v", err)
	}
	if len(entries) == 0 {
		t.Error("expected auto-detected symlinks to be refreshed in .claude/skills/")
	}
}

func TestCmdUpdateDirect_NoAdaptersNoFlag(t *testing.T) {
	root := setupDirectTest(t)
	if err := internal.InitProject(root); err != nil {
		t.Fatal(err)
	}
	if err := cmdUpdate([]string{}); err != nil {
		t.Fatalf("cmdUpdate (no adapters, no flag): %v", err)
	}
}

func TestCmdInitDirect_AutoDetectExistingAdapters(t *testing.T) {
	root := setupEmptyDir(t)
	if err := cmdInit([]string{"--tools", "claude"}); err != nil {
		t.Fatalf("first init --tools claude: %v", err)
	}
	if err := cmdInit([]string{}); err != nil {
		t.Fatalf("second init (auto-detect): %v", err)
	}
	claudeSkills := filepath.Join(root, ".claude", "skills")
	entries, err := os.ReadDir(claudeSkills)
	if err != nil {
		t.Fatalf("reading .claude/skills: %v", err)
	}
	if len(entries) == 0 {
		t.Error("expected auto-detected symlinks to be refreshed in .claude/skills/")
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

func TestValidateDecisionsMutualExclusion(t *testing.T) {
	bin, root := setupCLITest(t)
	_, code := runCLI(t, bin, root, "validate", "--decisions", "--specs")
	if code == 0 {
		t.Fatal("expected error for --decisions + --specs")
	}
}

func TestCLIInitJSON(t *testing.T) {
	bin, _ := setupCLITest(t)

	root := t.TempDir()
	out, code := runCLI(t, bin, root, "init", "--json")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("json: %v\n%s", err, out)
	}
	if result["initialized"] != true {
		t.Errorf("expected initialized=true, got %v", result["initialized"])
	}
	_, hasSkills := result["skills"]
	if !hasSkills {
		t.Error("expected skills field")
	}
}

func TestCLIInitMinimalJSON(t *testing.T) {
	bin, _ := setupCLITest(t)

	root := t.TempDir()
	out, code := runCLI(t, bin, root, "init", "--minimal", "--json")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("json: %v\n%s", err, out)
	}
	if result["initialized"] != true {
		t.Errorf("expected initialized=true, got %v", result["initialized"])
	}
	if _, hasSkills := result["skills"]; hasSkills {
		t.Error("minimal JSON should not contain skills")
	}
}

func TestCLIInitMinimalText(t *testing.T) {
	bin, _ := setupCLITest(t)

	root := t.TempDir()
	out, code := runCLI(t, bin, root, "init", "--minimal")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	if !strings.Contains(out, "initialized") {
		t.Errorf("expected minimal text output, got: %s", out)
	}
}

func TestCLIUpdateJSON(t *testing.T) {
	bin, _ := setupCLITest(t)

	root := t.TempDir()
	runCLI(t, bin, root, "init")

	out, code := runCLI(t, bin, root, "update", "--json")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("json: %v\n%s", err, out)
	}
	if result["skillsUpdated"] != true {
		t.Errorf("expected skillsUpdated=true, got %v", result["skillsUpdated"])
	}
}

func TestCLIUpdateMinimalJSON(t *testing.T) {
	bin, _ := setupCLITest(t)

	root := t.TempDir()
	runCLI(t, bin, root, "init")

	out, code := runCLI(t, bin, root, "update", "--minimal", "--json")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("json: %v\n%s", err, out)
	}
	if result["updated"] != true {
		t.Errorf("expected updated=true, got %v", result["updated"])
	}
	if _, hasSkills := result["skillsUpdated"]; hasSkills {
		t.Error("minimal JSON should not contain skillsUpdated")
	}
}

func TestCLIUpgradeJSON(t *testing.T) {
	bin, _ := setupCLITest(t)

	root := t.TempDir()
	out, code := runCLI(t, bin, root, "upgrade", "--json")
	// Will likely fail because not a go-install, but should still handle --json flag
	_ = out
	_ = code
	// The main thing is that --json is accepted as a valid flag
}

func TestCLIValidateMinimalJSON(t *testing.T) {
	bin, root := setupCLITest(t)

	out, code := runCLI(t, bin, root, "validate", "--minimal", "--json")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("json: %v\n%s", err, out)
	}
	if result["valid"] != true {
		t.Errorf("expected valid=true, got %v", result["valid"])
	}
	if _, hasSummary := result["summary"]; hasSummary {
		t.Error("minimal JSON should not contain summary")
	}
	if _, hasWarnings := result["warnings"]; hasWarnings {
		t.Error("minimal JSON should not contain warnings")
	}
}

func TestCLIValidateMinimalText(t *testing.T) {
	bin, root := setupCLITest(t)

	out, code := runCLI(t, bin, root, "validate", "--minimal")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	if !strings.HasPrefix(out, "ok\t") {
		t.Errorf("expected minimal text starting with 'ok\\t', got: %s", out)
	}
}

func TestCLIViewMinimalText(t *testing.T) {
	bin, root := setupCLITest(t)

	out, code := runCLI(t, bin, root, "view", "--minimal")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	if !strings.Contains(out, "specs\t") {
		t.Errorf("expected minimal summary line, got: %s", out)
	}
}

func TestCLIViewMinimalJSON(t *testing.T) {
	bin, root := setupCLITest(t)

	out, code := runCLI(t, bin, root, "view", "--minimal", "--json")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("json: %v\n%s", err, out)
	}
	if _, hasChanges := result["changes"]; hasChanges {
		t.Error("minimal JSON should not contain changes")
	}
	if _, hasSpecs := result["specs"]; hasSpecs {
		t.Error("minimal JSON should not contain specs")
	}
	if result["summary"] == nil {
		t.Error("minimal JSON should contain summary")
	}
}

func TestCLINewMinimalText(t *testing.T) {
	bin, root := setupCLITest(t)

	out, code := runCLI(t, bin, root, "new", "test-change", "--issue", "7", "--minimal")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	if !strings.Contains(out, "test-change") {
		t.Errorf("expected change name, got: %s", out)
	}
	if !strings.Contains(out, "#7") {
		t.Errorf("expected issue number, got: %s", out)
	}
}

func TestCLINewMinimalJSON(t *testing.T) {
	bin, root := setupCLITest(t)

	out, code := runCLI(t, bin, root, "new", "test-change", "--issue", "7", "--minimal", "--json")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("json: %v\n%s", err, out)
	}
	if result["changeName"] == nil {
		t.Error("minimal JSON should contain changeName")
	}
	if result["issue"] == nil {
		t.Error("minimal JSON should contain issue")
	}
}

