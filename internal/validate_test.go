package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	dirs := []string{
		SpecsPath(root),
		ChangesPath(root),
		ArchivePath(root),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}
	return root
}

func writeChangeFile(t *testing.T, root, changeName, filename, content string) {
	t.Helper()
	dir := ChangePath(root, changeName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, filename), []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func writeDeltaSpecFile(t *testing.T, root, changeName, capability, filename, content string) {
	t.Helper()
	dir := filepath.Join(ChangeSpecsPath(root, changeName), capability)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, filename), []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func writeMainSpecFile(t *testing.T, root, capability, content string) {
	t.Helper()
	dir := filepath.Join(SpecsPath(root), capability)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func makeValidChange(t *testing.T, root, name string, deltaContent string) {
	t.Helper()
	writeChangeFile(t, root, name, "proposal.md", "# Proposal\nMotivation.")
	writeChangeFile(t, root, name, "design.md", "# Design\nApproach.")
	writeChangeFile(t, root, name, "tasks.md", "## Phase 1: Do stuff\n- [ ] Task one")
	writeDeltaSpecFile(t, root, name, "cap", "spec.md", deltaContent)
}

func TestValidateChangeValid(t *testing.T) {
	root := setupTestProject(t)
	makeValidChange(t, root, "test-change", `## ADDED Requirements

### Requirement: Login
The system SHALL authenticate.

#### Scenario: Valid
- **WHEN** valid creds
`)

	result, err := ValidateChange(root, "test-change")
	if err != nil {
		t.Fatalf("ValidateChange: %v", err)
	}
	if !result.Valid {
		for _, e := range result.Errors {
			t.Errorf("Unexpected error: %s: %s", e.File, e.Message)
		}
		t.Fatal("expected valid change")
	}
}

func TestValidateChangeMissingProposal(t *testing.T) {
	root := setupTestProject(t)
	writeChangeFile(t, root, "bad", "design.md", "# Design")
	writeChangeFile(t, root, "bad", "tasks.md", "## Phase 1\n- [ ] Task")
	writeDeltaSpecFile(t, root, "bad", "cap", "spec.md", `## ADDED Requirements

### Requirement: R1
The system SHALL work.

#### Scenario: S1
- **WHEN** triggered
`)

	result, err := ValidateChange(root, "bad")
	if err != nil {
		t.Fatalf("ValidateChange: %v", err)
	}
	if result.Valid {
		t.Fatal("expected invalid (missing proposal)")
	}
	found := false
	for _, e := range result.Errors {
		if e.Message == "missing required artifact: proposal" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'missing required artifact: proposal' error")
	}
}

func TestValidateChangeMissingSpecsDir(t *testing.T) {
	root := setupTestProject(t)
	writeChangeFile(t, root, "no-specs", "proposal.md", "# Proposal")
	writeChangeFile(t, root, "no-specs", "design.md", "# Design")
	writeChangeFile(t, root, "no-specs", "tasks.md", "## Phase 1\n- [ ] Task")

	result, err := ValidateChange(root, "no-specs")
	if err != nil {
		t.Fatalf("ValidateChange: %v", err)
	}
	if result.Valid {
		t.Fatal("expected invalid (missing specs)")
	}
	found := false
	for _, e := range result.Errors {
		if e.Message == "missing specs directory" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'missing specs directory' error")
	}
}

func TestValidateChangeADDEDWithoutSHALLOrMUST(t *testing.T) {
	root := setupTestProject(t)
	makeValidChange(t, root, "change", `## ADDED Requirements

### Requirement: R1
Some content without keywords.

#### Scenario: S1
- **WHEN** triggered
`)

	result, err := ValidateChange(root, "change")
	if err != nil {
		t.Fatalf("ValidateChange: %v", err)
	}
	if result.Valid {
		t.Fatal("expected invalid (no SHALL/MUST)")
	}
	found := false
	for _, e := range result.Errors {
		if e.Message == `ADDED requirement "R1" must contain SHALL or MUST` {
			found = true
		}
	}
	if !found {
		t.Error("expected SHALL/MUST error")
	}
}

func TestValidateChangeMODIFIEDWithoutSHALL(t *testing.T) {
	root := setupTestProject(t)
	writeMainSpecFile(t, root, "cap", `# cap

### Requirement: R1
The system SHALL work.

#### Scenario: S1
- **WHEN** triggered
`)
	writeChangeFile(t, root, "change", "proposal.md", "# Proposal")
	writeChangeFile(t, root, "change", "design.md", "# Design")
	writeChangeFile(t, root, "change", "tasks.md", "## Phase 1\n- [ ] Task")
	writeDeltaSpecFile(t, root, "change", "cap", "spec.md", `## MODIFIED Requirements

### Requirement: R1
Updated content without keywords.

#### Scenario: S2
- **WHEN** something
`)

	result, err := ValidateChange(root, "change")
	if err != nil {
		t.Fatalf("ValidateChange: %v", err)
	}
	if result.Valid {
		t.Fatal("expected invalid (MODIFIED without SHALL/MUST)")
	}
}

func TestValidateChangeADDEDWithoutScenarios(t *testing.T) {
	root := setupTestProject(t)
	makeValidChange(t, root, "change", `## ADDED Requirements

### Requirement: R1
The system SHALL work.
`)

	result, err := ValidateChange(root, "change")
	if err != nil {
		t.Fatalf("ValidateChange: %v", err)
	}
	if result.Valid {
		t.Fatal("expected invalid (no scenarios)")
	}
	found := false
	for _, e := range result.Errors {
		if e.Message == `ADDED requirement "R1" must include at least one scenario` {
			found = true
		}
	}
	if !found {
		t.Error("expected scenario error for ADDED requirement")
	}
}

func TestValidateChangeMODIFIEDWithoutScenarios(t *testing.T) {
	root := setupTestProject(t)
	writeMainSpecFile(t, root, "cap", `# cap

### Requirement: R1
The system SHALL work.

#### Scenario: S1
- **WHEN** triggered
`)
	writeChangeFile(t, root, "change", "proposal.md", "# Proposal")
	writeChangeFile(t, root, "change", "design.md", "# Design")
	writeChangeFile(t, root, "change", "tasks.md", "## Phase 1\n- [ ] Task")
	writeDeltaSpecFile(t, root, "change", "cap", "spec.md", `## MODIFIED Requirements

### Requirement: R1
The system SHALL work differently.
`)

	result, err := ValidateChange(root, "change")
	if err != nil {
		t.Fatalf("ValidateChange: %v", err)
	}
	if result.Valid {
		t.Fatal("expected invalid (MODIFIED without scenarios)")
	}
}

func TestValidateChangeDanglingDeltaNoMainSpec(t *testing.T) {
	root := setupTestProject(t)
	makeValidChange(t, root, "change", `## MODIFIED Requirements

### Requirement: Login
The system SHALL authenticate via SSO.

#### Scenario: SSO
- **WHEN** SSO token valid
`)

	result, err := ValidateChange(root, "change")
	if err != nil {
		t.Fatalf("ValidateChange: %v", err)
	}
	if result.Valid {
		t.Fatal("expected invalid (main spec missing for MODIFIED)")
	}
	found := false
	for _, e := range result.Errors {
		if e.Message == `main spec for capability "cap" does not exist` {
			found = true
		}
	}
	if !found {
		t.Error("expected dangling delta error (main spec missing)")
	}
}

func TestValidateChangeDanglingDeltaNonexistentRequirement(t *testing.T) {
	root := setupTestProject(t)
	writeMainSpecFile(t, root, "auth", `# auth

### Requirement: Login
The system SHALL authenticate.

#### Scenario: Valid
- **WHEN** valid creds
`)
	writeChangeFile(t, root, "change", "proposal.md", "# Proposal")
	writeChangeFile(t, root, "change", "design.md", "# Design")
	writeChangeFile(t, root, "change", "tasks.md", "## Phase 1\n- [ ] Task")
	writeDeltaSpecFile(t, root, "change", "auth", "spec.md", `## MODIFIED Requirements

### Requirement: Nonexistent
The system SHALL do something.

#### Scenario: S1
- **WHEN** triggered
`)

	result, err := ValidateChange(root, "change")
	if err != nil {
		t.Fatalf("ValidateChange: %v", err)
	}
	if result.Valid {
		t.Fatal("expected invalid (MODIFIED nonexistent requirement)")
	}
	found := false
	for _, e := range result.Errors {
		if e.Message == `MODIFIED requirement "Nonexistent" not found in main spec` {
			found = true
		}
	}
	if !found {
		t.Error("expected dangling delta error for nonexistent requirement")
	}
}

func TestValidateChangeDanglingDeltaRemovedNonexistent(t *testing.T) {
	root := setupTestProject(t)
	writeMainSpecFile(t, root, "auth", `# auth

### Requirement: Login
The system SHALL authenticate.

#### Scenario: Valid
- **WHEN** valid creds
`)
	writeChangeFile(t, root, "change", "proposal.md", "# Proposal")
	writeChangeFile(t, root, "change", "design.md", "# Design")
	writeChangeFile(t, root, "change", "tasks.md", "## Phase 1\n- [ ] Task")
	writeDeltaSpecFile(t, root, "change", "auth", "spec.md", `## REMOVED Requirements

### Requirement: Ghost
`)

	result, err := ValidateChange(root, "change")
	if err != nil {
		t.Fatalf("ValidateChange: %v", err)
	}
	if result.Valid {
		t.Fatal("expected invalid (REMOVED nonexistent requirement)")
	}
}

func TestValidateChangeTasksNoPhaseHeadings(t *testing.T) {
	root := setupTestProject(t)
	makeValidChange(t, root, "change", `## ADDED Requirements

### Requirement: R1
The system SHALL work.

#### Scenario: S1
- **WHEN** triggered
`)
	writeChangeFile(t, root, "change", "tasks.md", "- [ ] Task without phase heading")

	result, err := ValidateChange(root, "change")
	if err != nil {
		t.Fatalf("ValidateChange: %v", err)
	}
	if result.Valid {
		t.Fatal("expected invalid (no phase headings)")
	}
	found := false
	for _, e := range result.Errors {
		if e.Message == "tasks.md has no phase headings (## Phase)" {
			found = true
		}
	}
	if !found {
		t.Error("expected phase heading error")
	}
}

func TestValidateChangeNonexistent(t *testing.T) {
	root := setupTestProject(t)
	_, err := ValidateChange(root, "ghost")
	if err == nil {
		t.Fatal("expected error for nonexistent change")
	}
}

func TestValidateChangeREMOVEDNeedsNoSHALLOrScenarios(t *testing.T) {
	root := setupTestProject(t)
	writeMainSpecFile(t, root, "auth", `# auth

### Requirement: Legacy
The system SHALL do legacy thing.

#### Scenario: Old
- **WHEN** old thing
`)
	writeChangeFile(t, root, "change", "proposal.md", "# Proposal")
	writeChangeFile(t, root, "change", "design.md", "# Design")
	writeChangeFile(t, root, "change", "tasks.md", "## Phase 1\n- [ ] Task")
	writeDeltaSpecFile(t, root, "change", "auth", "spec.md", `## REMOVED Requirements

### Requirement: Legacy
`)

	result, err := ValidateChange(root, "change")
	if err != nil {
		t.Fatalf("ValidateChange: %v", err)
	}
	if !result.Valid {
		for _, e := range result.Errors {
			t.Errorf("Unexpected error: %s: %s", e.File, e.Message)
		}
		t.Fatal("REMOVED requirements should not need SHALL/MUST or scenarios")
	}
}

func TestValidateChangeMUSTKeywordAccepted(t *testing.T) {
	root := setupTestProject(t)
	makeValidChange(t, root, "change", `## ADDED Requirements

### Requirement: R1
The system MUST enforce limits.

#### Scenario: S1
- **WHEN** triggered
`)

	result, err := ValidateChange(root, "change")
	if err != nil {
		t.Fatalf("ValidateChange: %v", err)
	}
	if !result.Valid {
		for _, e := range result.Errors {
			t.Errorf("Unexpected error: %s: %s", e.File, e.Message)
		}
		t.Fatal("MUST should be accepted as a keyword")
	}
}
