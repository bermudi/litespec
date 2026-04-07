package internal

import (
	"testing"
)

func TestParseDeltaSpecAllFourOperations(t *testing.T) {
	input := `# auth

## ADDED Requirements

### Requirement: Rate Limiting
The system SHALL limit requests.

#### Scenario: Exceeds limit
- **WHEN** too many requests
- **THEN** return 429

## MODIFIED Requirements

### Requirement: Login
The system SHALL authenticate via OAuth.

#### Scenario: OAuth flow
- **WHEN** OAuth provider responds

## REMOVED Requirements

### Requirement: Legacy Auth

## RENAMED Requirements

### Requirement: Old Login → Authenticate
`
	delta, err := ParseDeltaSpec(input)
	if err != nil {
		t.Fatalf("ParseDeltaSpec: %v", err)
	}
	if delta.Capability != "auth" {
		t.Errorf("Capability = %q, want %q", delta.Capability, "auth")
	}
	if len(delta.Requirements) != 4 {
		t.Fatalf("Requirements count = %d, want 4", len(delta.Requirements))
	}

	wantOps := []DeltaOperation{DeltaAdded, DeltaModified, DeltaRemoved, DeltaRenamed}
	wantNames := []string{"Rate Limiting", "Login", "Legacy Auth", "Authenticate"}
	for i, req := range delta.Requirements {
		if req.Operation != wantOps[i] {
			t.Errorf("Req[%d].Operation = %q, want %q", i, req.Operation, wantOps[i])
		}
		if req.Name != wantNames[i] {
			t.Errorf("Req[%d].Name = %q, want %q", i, req.Name, wantNames[i])
		}
	}
	if delta.Requirements[3].OldName != "Old Login" {
		t.Errorf("RENAMED OldName = %q, want %q", delta.Requirements[3].OldName, "Old Login")
	}
}

func TestParseDeltaSpecRenamedWithUnicodeArrow(t *testing.T) {
	input := `## RENAMED Requirements

### Requirement: Foo → Bar
`
	delta, err := ParseDeltaSpec(input)
	if err != nil {
		t.Fatalf("ParseDeltaSpec: %v", err)
	}
	if len(delta.Requirements) != 1 {
		t.Fatalf("Requirements count = %d, want 1", len(delta.Requirements))
	}
	r := delta.Requirements[0]
	if r.OldName != "Foo" {
		t.Errorf("OldName = %q, want %q", r.OldName, "Foo")
	}
	if r.Name != "Bar" {
		t.Errorf("Name = %q, want %q", r.Name, "Bar")
	}
}

func TestParseDeltaSpecRenamedWithASCIIArrow(t *testing.T) {
	input := `## RENAMED Requirements

### Requirement: Old->New
`
	delta, err := ParseDeltaSpec(input)
	if err != nil {
		t.Fatalf("ParseDeltaSpec: %v", err)
	}
	r := delta.Requirements[0]
	if r.OldName != "Old" {
		t.Errorf("OldName = %q, want %q", r.OldName, "Old")
	}
	if r.Name != "New" {
		t.Errorf("Name = %q, want %q", r.Name, "New")
	}
}

func TestParseDeltaSpecRenamedMissingArrow(t *testing.T) {
	input := `## RENAMED Requirements

### Requirement: No Arrow Here
`
	_, err := ParseDeltaSpec(input)
	if err == nil {
		t.Fatal("expected error for RENAMED without arrow separator")
	}
}

func TestParseDeltaSpecEmpty(t *testing.T) {
	delta, err := ParseDeltaSpec("")
	if err != nil {
		t.Fatalf("ParseDeltaSpec empty: %v", err)
	}
	if delta.Capability != "" {
		t.Errorf("Capability = %q, want empty", delta.Capability)
	}
	if len(delta.Requirements) != 0 {
		t.Errorf("Requirements count = %d, want 0", len(delta.Requirements))
	}
}

func TestParseDeltaSpecNoH1(t *testing.T) {
	input := `## ADDED Requirements

### Requirement: R1
The system SHALL work.

#### Scenario: S1
- **WHEN** triggered
`
	delta, err := ParseDeltaSpec(input)
	if err != nil {
		t.Fatalf("ParseDeltaSpec: %v", err)
	}
	if delta.Capability != "" {
		t.Errorf("Capability = %q, want empty (no H1)", delta.Capability)
	}
	if len(delta.Requirements) != 1 {
		t.Fatalf("Requirements count = %d, want 1", len(delta.Requirements))
	}
}

func TestParseDeltaSpecMultipleInSameSection(t *testing.T) {
	input := `## REMOVED Requirements

### Requirement: Old A
### Requirement: Old B
### Requirement: Old C
`
	delta, err := ParseDeltaSpec(input)
	if err != nil {
		t.Fatalf("ParseDeltaSpec: %v", err)
	}
	if len(delta.Requirements) != 3 {
		t.Fatalf("Requirements count = %d, want 3", len(delta.Requirements))
	}
	for i, name := range []string{"Old A", "Old B", "Old C"} {
		if delta.Requirements[i].Name != name {
			t.Errorf("Req[%d].Name = %q, want %q", i, delta.Requirements[i].Name, name)
		}
		if delta.Requirements[i].Operation != DeltaRemoved {
			t.Errorf("Req[%d].Operation = %q, want REMOVED", i, delta.Requirements[i].Operation)
		}
	}
}

func TestParseDeltaSpecRejectsUnknownH2(t *testing.T) {
	input := `# Pipeline

## ADDED Requirements

### Requirement: Branch Creation
The system SHALL create a branch.

#### Scenario: Clean
- **WHEN** worktree is clean

## DELTA Requirements

### Requirement: Empty Commit Handling
The system MUST skip commits when no changes.
`
	_, err := ParseDeltaSpec(input)
	if err == nil {
		t.Fatal("expected error for unknown H2 section ## DELTA Requirements")
	}
	if !containsSubstr(err.Error(), "unexpected H2 section") {
		t.Errorf("error = %q, want mention of unexpected H2 section", err.Error())
	}
	if !containsSubstr(err.Error(), "## DELTA Requirements") {
		t.Errorf("error = %q, want mention of ## DELTA Requirements", err.Error())
	}
}
