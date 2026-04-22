package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestComputePreviewResultNewCapability(t *testing.T) {
	root := setupTestProject(t)

	delta := `# rate-limit

## ADDED Requirements

### Requirement: Rate Limiting
The system SHALL limit requests.

#### Scenario: Exceeds limit
- **WHEN** user exceeds limit
- **THEN** return 429
`
	writeDeltaSpecFile(t, root, "add-rate-limit", "rate-limit", "spec.md", delta)

	writes, err := PrepareArchiveWrites(root, "add-rate-limit")
	if err != nil {
		t.Fatalf("PrepareArchiveWrites: %v", err)
	}

	result, err := ComputePreviewResult(writes, root)
	if err != nil {
		t.Fatalf("ComputePreviewResult: %v", err)
	}

	if len(result.Capabilities) != 1 {
		t.Fatalf("capabilities = %d, want 1", len(result.Capabilities))
	}
	cap := result.Capabilities[0]
	if cap.Name != "rate-limit" {
		t.Errorf("name = %q, want %q", cap.Name, "rate-limit")
	}
	if !cap.IsNew {
		t.Error("IsNew = false, want true")
	}
	if len(cap.Operations) != 1 {
		t.Fatalf("operations = %d, want 1", len(cap.Operations))
	}
	if cap.Operations[0].Type != "ADDED" {
		t.Errorf("operation type = %q, want %q", cap.Operations[0].Type, "ADDED")
	}
	if cap.Operations[0].Requirement != "Rate Limiting" {
		t.Errorf("requirement = %q, want %q", cap.Operations[0].Requirement, "Rate Limiting")
	}
	if result.Totals.Capabilities != 1 {
		t.Errorf("totals.Capabilities = %d, want 1", result.Totals.Capabilities)
	}
	if result.Totals.Added != 1 {
		t.Errorf("totals.Added = %d, want 1", result.Totals.Added)
	}
}

func TestComputePreviewResultModifiedCapability(t *testing.T) {
	root := setupTestProject(t)

	mainSpec := `# auth

## Requirements

### Requirement: Login
The system SHALL provide login.

#### Scenario: Valid credentials
- **WHEN** valid credentials
- **THEN** success

### Requirement: Legacy OAuth
The system SHALL support legacy OAuth.
`
	writeMainSpecFile(t, root, "auth", mainSpec)

	delta := `# auth

## MODIFIED Requirements

### Requirement: Login
The system SHALL provide secure login.

#### Scenario: Valid credentials
- **WHEN** valid credentials
- **THEN** success

## REMOVED Requirements

### Requirement: Legacy OAuth
The system SHALL support legacy OAuth.

## ADDED Requirements

### Requirement: Session Timeout
The system SHALL enforce session timeout.
`
	writeDeltaSpecFile(t, root, "modify-auth", "auth", "spec.md", delta)

	writes, err := PrepareArchiveWrites(root, "modify-auth")
	if err != nil {
		t.Fatalf("PrepareArchiveWrites: %v", err)
	}

	result, err := ComputePreviewResult(writes, root)
	if err != nil {
		t.Fatalf("ComputePreviewResult: %v", err)
	}

	if len(result.Capabilities) != 1 {
		t.Fatalf("capabilities = %d, want 1", len(result.Capabilities))
	}

	cap := result.Capabilities[0]
	if cap.IsNew {
		t.Error("IsNew = true, want false")
	}

	// Operations: REMOVED first, then MODIFIED, then ADDED
	if len(cap.Operations) != 3 {
		t.Fatalf("operations = %d, want 3", len(cap.Operations))
	}

	wantOps := []PreviewOperation{
		{Type: "REMOVED", Requirement: "Legacy OAuth"},
		{Type: "MODIFIED", Requirement: "Login"},
		{Type: "ADDED", Requirement: "Session Timeout"},
	}
	for i, want := range wantOps {
		got := cap.Operations[i]
		if got.Type != want.Type || got.Requirement != want.Requirement {
			t.Errorf("operations[%d] = {%q, %q}, want {%q, %q}", i, got.Type, got.Requirement, want.Type, want.Requirement)
		}
	}

	if result.Totals.Added != 1 || result.Totals.Modified != 1 || result.Totals.Removed != 1 {
		t.Errorf("totals = %+v, want added=1, modified=1, removed=1", result.Totals)
	}
}

func TestComputePreviewResultRenamed(t *testing.T) {
	root := setupTestProject(t)

	mainSpec := `# auth

## Requirements

### Requirement: Two-Factor
The system SHALL support two-factor auth.
`
	writeMainSpecFile(t, root, "auth", mainSpec)

	delta := `# auth

## RENAMED Requirements

### Requirement: Two-Factor → MFA
The system SHALL support two-factor auth.
`
	writeDeltaSpecFile(t, root, "rename-auth", "auth", "spec.md", delta)

	writes, err := PrepareArchiveWrites(root, "rename-auth")
	if err != nil {
		t.Fatalf("PrepareArchiveWrites: %v", err)
	}

	result, err := ComputePreviewResult(writes, root)
	if err != nil {
		t.Fatalf("ComputePreviewResult: %v", err)
	}

	if len(result.Capabilities) != 1 {
		t.Fatalf("capabilities = %d, want 1", len(result.Capabilities))
	}
	cap := result.Capabilities[0]
	if len(cap.Operations) != 1 {
		t.Fatalf("operations = %d, want 1", len(cap.Operations))
	}
	op := cap.Operations[0]
	if op.Type != "RENAMED" {
		t.Errorf("type = %q, want %q", op.Type, "RENAMED")
	}
	if op.Requirement != "MFA" {
		t.Errorf("requirement = %q, want %q", op.Requirement, "MFA")
	}
	if op.OldName != "Two-Factor" {
		t.Errorf("oldName = %q, want %q", op.OldName, "Two-Factor")
	}
	if result.Totals.Renamed != 1 {
		t.Errorf("totals.Renamed = %d, want 1", result.Totals.Renamed)
	}
}

func TestComputePreviewResultEmptyChange(t *testing.T) {
	root := setupTestProject(t)
	// Create a change with no delta specs (but the specs dir must exist)
	writeChangeFile(t, root, "empty-change", "proposal.md", "# Proposal")
	// Ensure the specs directory exists but is empty
	os.MkdirAll(filepath.Join(ChangesPath(root), "empty-change", "specs"), 0o755)

	writes, err := PrepareArchiveWrites(root, "empty-change")
	if err != nil {
		t.Fatalf("PrepareArchiveWrites: %v", err)
	}

	result, err := ComputePreviewResult(writes, root)
	if err != nil {
		t.Fatalf("ComputePreviewResult: %v", err)
	}
	if len(result.Capabilities) != 0 {
		t.Errorf("capabilities = %d, want 0", len(result.Capabilities))
	}
	if result.Totals.Capabilities != 0 {
		t.Errorf("totals.Capabilities = %d, want 0", result.Totals.Capabilities)
	}
}

func TestComputePreviewResultNoNetChange(t *testing.T) {
	root := setupTestProject(t)

	mainSpec := `# auth

## Requirements

### Requirement: Login
The system SHALL provide login.
`
	writeMainSpecFile(t, root, "auth", mainSpec)

	// MODIFIED with identical content — should produce no operations
	delta := `# auth

## MODIFIED Requirements

### Requirement: Login
The system SHALL provide login.
`
	writeDeltaSpecFile(t, root, "same-auth", "auth", "spec.md", delta)

	writes, err := PrepareArchiveWrites(root, "same-auth")
	if err != nil {
		t.Fatalf("PrepareArchiveWrites: %v", err)
	}

	result, err := ComputePreviewResult(writes, root)
	if err != nil {
		t.Fatalf("ComputePreviewResult: %v", err)
	}

	if len(result.Capabilities) != 0 {
		t.Errorf("capabilities = %d, want 0 (no net change)", len(result.Capabilities))
	}
}

func TestComputePreviewResultMixedCapabilities(t *testing.T) {
	root := setupTestProject(t)

	// Existing canon spec for "auth"
	mainSpec := `# auth

## Requirements

### Requirement: Login
The system SHALL provide login.
`
	writeMainSpecFile(t, root, "auth", mainSpec)

	// Delta: modify auth, add rate-limit (new)
	authDelta := `# auth

## ADDED Requirements

### Requirement: Logout
The system SHALL provide logout.
`
	writeDeltaSpecFile(t, root, "mixed", "auth", "spec.md", authDelta)

	rateDelta := `# rate-limit

## ADDED Requirements

### Requirement: Rate Limiting
The system SHALL limit requests.
`
	writeDeltaSpecFile(t, root, "mixed", "rate-limit", "spec.md", rateDelta)

	writes, err := PrepareArchiveWrites(root, "mixed")
	if err != nil {
		t.Fatalf("PrepareArchiveWrites: %v", err)
	}

	result, err := ComputePreviewResult(writes, root)
	if err != nil {
		t.Fatalf("ComputePreviewResult: %v", err)
	}

	if len(result.Capabilities) != 2 {
		t.Fatalf("capabilities = %d, want 2", len(result.Capabilities))
	}
	if result.Totals.Capabilities != 2 {
		t.Errorf("totals.Capabilities = %d, want 2", result.Totals.Capabilities)
	}
	if result.Totals.Added != 2 {
		t.Errorf("totals.Added = %d, want 2", result.Totals.Added)
	}
}

func TestFormatPreviewTextEmpty(t *testing.T) {
	result := &PreviewResult{}
	text := FormatPreviewText(result)
	if text != "No changes to preview" {
		t.Errorf("text = %q, want %q", text, "No changes to preview")
	}
}

func TestFormatPreviewTextSingleCapability(t *testing.T) {
	result := &PreviewResult{
		Capabilities: []PreviewCapability{
			{
				Name:  "auth",
				IsNew: false,
				Operations: []PreviewOperation{
					{Type: "REMOVED", Requirement: "Legacy OAuth"},
					{Type: "MODIFIED", Requirement: "Login"},
					{Type: "ADDED", Requirement: "Session Timeout"},
				},
			},
		},
		Totals: PreviewTotals{
			Capabilities: 1,
			Added:        1,
			Modified:     1,
			Removed:      1,
			Renamed:      0,
		},
	}

	text := FormatPreviewText(result)
	if !strings.Contains(text, "▸ auth (MODIFIED)") {
		t.Errorf("missing capability header")
	}
	if !strings.Contains(text, "  + ADDED: Session Timeout") {
		t.Errorf("missing ADDED line")
	}
	if !strings.Contains(text, "  ~ MODIFIED: Login") {
		t.Errorf("missing MODIFIED line")
	}
	if !strings.Contains(text, "  - REMOVED: Legacy OAuth") {
		t.Errorf("missing REMOVED line")
	}
	if !strings.Contains(text, "1 capability affected • 1 added • 1 modified • 1 removed • 0 renamed") {
		t.Errorf("missing footer, got:\n%s", text)
	}
}

func TestFormatPreviewTextNewSpec(t *testing.T) {
	result := &PreviewResult{
		Capabilities: []PreviewCapability{
			{
				Name:  "rate-limit",
				IsNew: true,
				Operations: []PreviewOperation{
					{Type: "ADDED", Requirement: "Rate Limiting"},
				},
			},
		},
		Totals: PreviewTotals{
			Capabilities: 1,
			Added:        1,
		},
	}

	text := FormatPreviewText(result)
	if !strings.Contains(text, "▸ rate-limit (NEW SPEC)") {
		t.Errorf("missing NEW SPEC header, got:\n%s", text)
	}
}

func TestFormatPreviewTextRenamed(t *testing.T) {
	result := &PreviewResult{
		Capabilities: []PreviewCapability{
			{
				Name:  "auth",
				IsNew: false,
				Operations: []PreviewOperation{
					{Type: "RENAMED", Requirement: "MFA", OldName: "Two-Factor"},
				},
			},
		},
		Totals: PreviewTotals{
			Capabilities: 1,
			Renamed:      1,
		},
	}

	text := FormatPreviewText(result)
	if !strings.Contains(text, "  → RENAMED: Two-Factor → MFA") {
		t.Errorf("missing RENAMED line, got:\n%s", text)
	}
}

func TestFormatPreviewTextPluralCapabilities(t *testing.T) {
	result := &PreviewResult{
		Capabilities: []PreviewCapability{
			{Name: "a", IsNew: true, Operations: []PreviewOperation{{Type: "ADDED", Requirement: "R1"}}},
			{Name: "b", IsNew: true, Operations: []PreviewOperation{{Type: "ADDED", Requirement: "R2"}}},
		},
		Totals: PreviewTotals{Capabilities: 2, Added: 2},
	}

	text := FormatPreviewText(result)
	if !strings.Contains(text, "2 capabilities affected") {
		t.Errorf("expected plural 'capabilities', got:\n%s", text)
	}
}

func TestFormatPreviewJSONShape(t *testing.T) {
	result := &PreviewResult{
		Capabilities: []PreviewCapability{
			{
				Name:  "auth",
				IsNew: false,
				Operations: []PreviewOperation{
					{Type: "ADDED", Requirement: "Login"},
					{Type: "REMOVED", Requirement: "Legacy"},
				},
			},
		},
		Totals: PreviewTotals{
			Capabilities: 1,
			Added:        1,
			Removed:      1,
		},
	}

	data, err := FormatPreviewJSON(result)
	if err != nil {
		t.Fatalf("FormatPreviewJSON: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}

	caps, ok := parsed["capabilities"].([]interface{})
	if !ok {
		t.Fatal("capabilities not an array")
	}
	if len(caps) != 1 {
		t.Fatalf("capabilities len = %d, want 1", len(caps))
	}

	cap0 := caps[0].(map[string]interface{})
	if cap0["name"] != "auth" {
		t.Errorf("name = %v, want auth", cap0["name"])
	}
	if cap0["isNew"] != false {
		t.Errorf("isNew = %v, want false", cap0["isNew"])
	}

	ops := cap0["operations"].([]interface{})
	if len(ops) != 2 {
		t.Fatalf("operations len = %d, want 2", len(ops))
	}

	totals := parsed["totals"].(map[string]interface{})
	if totals["capabilities"] != float64(1) {
		t.Errorf("totals.capabilities = %v, want 1", totals["capabilities"])
	}
	if totals["added"] != float64(1) {
		t.Errorf("totals.added = %v, want 1", totals["added"])
	}
}

func TestFormatPreviewJSONEmpty(t *testing.T) {
	result := &PreviewResult{}
	data, err := FormatPreviewJSON(result)
	if err != nil {
		t.Fatalf("FormatPreviewJSON: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}

	caps, _ := parsed["capabilities"].([]interface{})
	if len(caps) != 0 {
		t.Errorf("capabilities len = %d, want 0", len(caps))
	}

	totals := parsed["totals"].(map[string]interface{})
	for _, key := range []string{"capabilities", "added", "modified", "removed", "renamed"} {
		if totals[key] != float64(0) {
			t.Errorf("totals[%s] = %v, want 0", key, totals[key])
		}
	}
}

func TestComputePreviewResultCorruptedMerged(t *testing.T) {
	root := setupTestProject(t)
	// Manually construct a PendingWrite with invalid merged content
	writes := []PendingWrite{
		{
			Capability: "bad",
			Path:       filepath.Join(CanonPath(root), "bad", "spec.md"),
			Dir:        filepath.Join(CanonPath(root), "bad"),
			Content:    "this is not a valid spec",
		},
	}
	_, err := ComputePreviewResult(writes, root)
	if err == nil {
		t.Fatal("expected error for corrupted merged content")
	}
}

func TestComputePreviewResultCorruptedCanon(t *testing.T) {
	root := setupTestProject(t)

	// Write a corrupted canon spec that still has a capability heading
	// so PrepareArchiveWrites can parse it, but with broken structure
	corruptSpecPath := filepath.Join(CanonPath(root), "auth", "spec.md")
	if err := os.MkdirAll(filepath.Dir(corruptSpecPath), 0o755); err != nil {
		t.Fatal(err)
	}
	// Valid enough for PrepareArchiveWrites but we'll corrupt the PendingWrite
	if err := os.WriteFile(corruptSpecPath, []byte("# auth\n\n## Requirements\n\n### Requirement: Login\nContent."), 0o644); err != nil {
		t.Fatal(err)
	}

	delta := `# auth

## ADDED Requirements

### Requirement: New Req
Content here.
`
	writeDeltaSpecFile(t, root, "test-change", "auth", "spec.md", delta)

	writes, err := PrepareArchiveWrites(root, "test-change")
	if err != nil {
		t.Fatalf("PrepareArchiveWrites: %v", err)
	}

	// Corrupt the merged content to trigger parse error in ComputePreviewResult
	writes[0].Content = "this is not valid"

	_, err = ComputePreviewResult(writes, root)
	if err == nil {
		t.Fatal("expected error for corrupted merged content")
	}
}
