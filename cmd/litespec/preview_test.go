package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bermudi/litespec/internal"
)

func setupPreviewTest(t *testing.T) string {
	t.Helper()
	root, err := os.MkdirTemp("", "TestPreviewCLI")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(root) })

	dirs := []string{
		filepath.Join(root, ".litespec"),
		filepath.Join(root, "specs", "canon"),
		filepath.Join(root, "specs", "changes"),
		filepath.Join(root, "specs", "changes", "archive"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func writeCLIChangeFile(t *testing.T, root, changeName, filename, content string) {
	t.Helper()
	dir := filepath.Join(root, "specs", "changes", changeName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, filename), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeCLIDeltaSpec(t *testing.T, root, changeName, capability, content string) {
	t.Helper()
	dir := filepath.Join(root, "specs", "changes", changeName, "specs", capability)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeCLICanonSpec(t *testing.T, root, capability, content string) {
	t.Helper()
	dir := filepath.Join(root, "specs", "canon", capability)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestCmdPreviewHappyPath(t *testing.T) {
	root := setupPreviewTest(t)
	origWd, _ := os.Getwd()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origWd)

	writeCLIChangeFile(t, root, "add-auth", "proposal.md", "# Proposal")
	writeCLIDeltaSpec(t, root, "add-auth", "auth", `# auth

## ADDED Requirements

### Requirement: Login
The system SHALL provide login.
`)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := cmdPreview([]string{"add-auth"})
	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("cmdPreview: %v", err)
	}

	outputBytes, _ := io.ReadAll(r)
	
	output := string(outputBytes)

	if !strings.Contains(output, "▸ auth (NEW SPEC)") {
		t.Errorf("missing NEW SPEC header, got:\n%s", output)
	}
	if !strings.Contains(output, "+ ADDED: Login") {
		t.Errorf("missing ADDED line, got:\n%s", output)
	}
}

func TestCmdPreviewEmptyChange(t *testing.T) {
	root := setupPreviewTest(t)
	origWd, _ := os.Getwd()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origWd)

	writeCLIChangeFile(t, root, "empty", "proposal.md", "# Proposal")
	// Ensure specs dir exists but empty
	os.MkdirAll(filepath.Join(root, "specs", "changes", "empty", "specs"), 0o755)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := cmdPreview([]string{"empty"})
	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("cmdPreview: %v", err)
	}

	outputBytes, _ := io.ReadAll(r)
	
	output := string(outputBytes)

	if !strings.Contains(output, "No changes to preview") {
		t.Errorf("expected empty message, got:\n%s", output)
	}
}

func TestCmdPreviewNonExistentChange(t *testing.T) {
	root := setupPreviewTest(t)
	origWd, _ := os.Getwd()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origWd)

	err := cmdPreview([]string{"no-such-change"})
	if err == nil {
		t.Fatal("expected error for non-existent change")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, want 'not found'", err.Error())
	}
}

func TestCmdPreviewArchivedChange(t *testing.T) {
	root := setupPreviewTest(t)
	origWd, _ := os.Getwd()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origWd)

	// Create an archived entry (YYYY-MM-DD-name format)
	archiveDir := filepath.Join(root, "specs", "changes", "archive", "2026-01-01-old-change")
	if err := os.MkdirAll(archiveDir, 0o755); err != nil {
		t.Fatal(err)
	}

	err := cmdPreview([]string{"old-change"})
	if err == nil {
		t.Fatal("expected error for archived change")
	}
	if !strings.Contains(err.Error(), "archived") {
		t.Errorf("error = %q, want 'archived'", err.Error())
	}
}

func TestCmdPreviewJSON(t *testing.T) {
	root := setupPreviewTest(t)
	origWd, _ := os.Getwd()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origWd)

	writeCLIChangeFile(t, root, "add-auth", "proposal.md", "# Proposal")
	writeCLIDeltaSpec(t, root, "add-auth", "auth", `# auth

## ADDED Requirements

### Requirement: Login
The system SHALL provide login.
`)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := cmdPreview([]string{"add-auth", "--json"})
	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("cmdPreview: %v", err)
	}

	outputBytes, _ := io.ReadAll(r)
	
	output := string(outputBytes)

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}

	caps, ok := parsed["capabilities"].([]interface{})
	if !ok || len(caps) != 1 {
		t.Fatalf("capabilities = %v, want 1 entry", caps)
	}
}

func TestCmdPreviewNoName(t *testing.T) {
	root := setupPreviewTest(t)
	origWd, _ := os.Getwd()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origWd)

	err := cmdPreview([]string{})
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if !strings.Contains(err.Error(), "usage:") {
		t.Errorf("error = %q, want usage message", err.Error())
	}
}

func TestCmdPreviewIntegration(t *testing.T) {
	root := setupPreviewTest(t)
	origWd, _ := os.Getwd()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origWd)

	// Set up canon spec
	writeCLICanonSpec(t, root, "auth", `# auth

## Requirements

### Requirement: Login
The system SHALL provide login.

#### Scenario: Valid
- **WHEN** valid creds
- **THEN** success
`)

	// Set up change delta
	writeCLIChangeFile(t, root, "modify-auth", "proposal.md", "# Proposal")
	writeCLIDeltaSpec(t, root, "modify-auth", "auth", `# auth

## MODIFIED Requirements

### Requirement: Login
The system SHALL provide secure login.

#### Scenario: Valid
- **WHEN** valid creds
- **THEN** success

## ADDED Requirements

### Requirement: Logout
The system SHALL provide logout.
`)

	// Preview
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := cmdPreview([]string{"modify-auth"})
	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("cmdPreview: %v", err)
	}

	outputBytes, _ := io.ReadAll(r)
	
	output := string(outputBytes)

	if !strings.Contains(output, "▸ auth (MODIFIED)") {
		t.Errorf("missing MODIFIED header, got:\n%s", output)
	}
	if !strings.Contains(output, "~ MODIFIED: Login") {
		t.Errorf("missing MODIFIED Login, got:\n%s", output)
	}
	if !strings.Contains(output, "+ ADDED: Logout") {
		t.Errorf("missing ADDED Logout, got:\n%s", output)
	}

	// Verify canon was NOT modified
	data, _ := os.ReadFile(filepath.Join(root, "specs", "canon", "auth", "spec.md"))
	if !strings.Contains(string(data), "The system SHALL provide login.") {
		t.Error("canon spec was modified — preview must be read-only")
	}
}

// Verify preview output matches archive merge result
func TestCmdPreviewMatchesArchive(t *testing.T) {
	root := setupPreviewTest(t)
	origWd, _ := os.Getwd()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origWd)

	writeCLICanonSpec(t, root, "api", `# api

## Requirements

### Requirement: List Items
The system SHALL list items.
`)

	writeCLIChangeFile(t, root, "add-api", "proposal.md", "# Proposal")
	writeCLIChangeFile(t, root, "add-api", "tasks.md", "## Phase 1\n- [x] Done")
	writeCLIDeltaSpec(t, root, "add-api", "api", `# api

## ADDED Requirements

### Requirement: Create Item
The system SHALL create items.
`)

	// Get preview result
	writes, err := internal.PrepareArchiveWrites(root, "add-api")
	if err != nil {
		t.Fatal(err)
	}
	previewResult, err := internal.ComputePreviewResult(writes, root)
	if err != nil {
		t.Fatal(err)
	}

	// The merged content from PrepareArchiveWrites should match what preview describes
	if len(previewResult.Capabilities) != 1 {
		t.Fatalf("capabilities = %d, want 1", len(previewResult.Capabilities))
	}
	cap := previewResult.Capabilities[0]
	if cap.Name != "api" {
		t.Errorf("name = %q, want api", cap.Name)
	}
	if len(cap.Operations) != 1 || cap.Operations[0].Type != "ADDED" || cap.Operations[0].Requirement != "Create Item" {
		t.Errorf("operations = %+v, want 1 ADDED 'Create Item'", cap.Operations)
	}

	// Verify the merged content is the same
	merged, err := internal.ParseMainSpec(writes[0].Content)
	if err != nil {
		t.Fatal(err)
	}
	if len(merged.Requirements) != 2 {
		t.Fatalf("merged requirements = %d, want 2", len(merged.Requirements))
	}
	foundCreate := false
	for _, r := range merged.Requirements {
		if r.Name == "Create Item" {
			foundCreate = true
		}
	}
	if !foundCreate {
		t.Error("merged spec missing 'Create Item' requirement")
	}
}
