package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupTestProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ProjectDirName), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	return root
}

func writeMainSpecFile(t *testing.T, root, capability, content string) {
	t.Helper()
	dir := filepath.Join(root, ProjectDirName, capability)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func TestValidateSpecValid(t *testing.T) {
	root := setupTestProject(t)
	writeMainSpecFile(t, root, "auth", `# auth

## Requirements

### Requirement: Login
The system SHALL authenticate.

#### Scenario: Valid
	- **WHEN** valid creds
- **THEN** result
`)

	result, err := ValidateSpec(root, "auth")
	if err != nil {
		t.Fatalf("ValidateSpec: %v", err)
	}
	if !result.Valid {
		for _, e := range result.Errors {
			t.Errorf("Unexpected error: %s: %s", e.File, e.Message)
		}
		t.Fatal("expected valid spec")
	}
}

func TestValidateSpecNotFound(t *testing.T) {
	root := setupTestProject(t)
	_, err := ValidateSpec(root, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent spec")
	}
}

func TestValidateSpecInvalidContent(t *testing.T) {
	root := setupTestProject(t)
	writeMainSpecFile(t, root, "auth", `not a valid spec`)

	result, err := ValidateSpec(root, "auth")
	if err != nil {
		t.Fatalf("ValidateSpec: %v", err)
	}
	if result.Valid {
		t.Fatal("expected invalid for unparseable spec")
	}
}

func TestValidateSpecNoRequirements(t *testing.T) {
	root := setupTestProject(t)
	writeMainSpecFile(t, root, "auth", `# auth

## Requirements
`)

	result, err := ValidateSpec(root, "auth")
	if err != nil {
		t.Fatalf("ValidateSpec: %v", err)
	}
	if !result.Valid {
		t.Fatal("spec with no requirements should be valid")
	}
	found := false
	for _, w := range result.Warnings {
		if w.Message == `capability "auth" has no requirements` {
			found = true
		}
	}
	if !found {
		t.Error("expected warning for no requirements")
	}
}

func TestValidateDecisionValid(t *testing.T) {
	root := setupTestProject(t)
	decDir := DecisionsPath(root)
	writeDecisionFile(t, decDir, "0001-test-decision.md", makeDecision(1, "test-decision", "accepted"))

	result, err := ValidateDecision(root, "test-decision")
	if err != nil {
		t.Fatal(err)
	}
	if !result.Valid {
		t.Errorf("expected valid, got errors: %v", result.Errors)
	}
}

func TestValidateDecisionNotFound(t *testing.T) {
	root := setupTestProject(t)
	_, err := ValidateDecision(root, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent decision")
	}
}

func TestValidateDecisionsDuplicateNumber(t *testing.T) {
	root := setupTestProject(t)
	decDir := DecisionsPath(root)
	writeDecisionFile(t, decDir, "0001-foo.md", makeDecision(1, "foo", "accepted"))
	writeDecisionFile(t, decDir, "0001-bar.md", makeDecision(1, "bar", "accepted"))

	result, err := ValidateDecisions(root)
	if err != nil {
		t.Fatal(err)
	}
	if result.Valid {
		t.Error("expected invalid for duplicate numbers")
	}
	found := false
	for _, e := range result.Errors {
		if strings.Contains(e.Message, "duplicate decision number") {
			found = true
		}
	}
	if !found {
		t.Error("expected duplicate number error")
	}
}

func TestValidateDecisionsSupersedeDanglingPointer(t *testing.T) {
	root := setupTestProject(t)
	decDir := DecisionsPath(root)

	supersedesContent := `# New Model

## Status

accepted

## Context

ctx

## Decision

dec

## Consequences

con

## Supersedes

- 0001-nonexistent
`
	writeDecisionFile(t, decDir, "0002-new-model.md", supersedesContent)

	result, err := ValidateDecisions(root)
	if err != nil {
		t.Fatal(err)
	}
	if result.Valid {
		t.Error("expected invalid for dangling pointer")
	}
	found := false
	for _, e := range result.Errors {
		if strings.Contains(e.Message, "does not resolve") {
			found = true
		}
	}
	if !found {
		t.Error("expected dangling pointer error")
	}
}

func TestValidateDecisionsSupersededWithoutForwardPointer(t *testing.T) {
	root := setupTestProject(t)
	decDir := DecisionsPath(root)
	writeDecisionFile(t, decDir, "0001-old.md", makeDecision(1, "old", "superseded"))

	result, err := ValidateDecisions(root)
	if err != nil {
		t.Fatal(err)
	}
	if result.Valid {
		t.Error("expected invalid for superseded without forward pointer")
	}
	found := false
	for _, e := range result.Errors {
		if strings.Contains(e.Message, "no Superseded-By pointer") {
			found = true
		}
	}
	if !found {
		t.Error("expected missing Superseded-By error")
	}
}

func TestValidateDecisionsSupersedeResolves(t *testing.T) {
	root := setupTestProject(t)
	decDir := DecisionsPath(root)

	oldContent := `# Old Model

## Status

superseded

## Context

old ctx

## Decision

old dec

## Consequences

old con

## Superseded-By

- 0002-new-model
`
	writeDecisionFile(t, decDir, "0001-old-model.md", oldContent)

	newContent := `# New Model

## Status

accepted

## Context

new ctx

## Decision

new dec

## Consequences

new con

## Supersedes

- 0001-old-model
`
	writeDecisionFile(t, decDir, "0002-new-model.md", newContent)

	result, err := ValidateDecisions(root)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Valid {
		t.Errorf("expected valid, got errors: %v", result.Errors)
	}
}

func TestValidateDecisionsIncludesDecisionsCount(t *testing.T) {
	root := setupTestProject(t)
	decDir := DecisionsPath(root)
	writeDecisionFile(t, decDir, "0001-foo.md", makeDecision(1, "foo", "accepted"))
	writeDecisionFile(t, decDir, "0002-bar.md", makeDecision(2, "bar", "proposed"))

	result, err := ValidateDecisions(root)
	if err != nil {
		t.Fatal(err)
	}
	if result.DecisionsCount != 2 {
		t.Errorf("DecisionsCount = %d, want 2", result.DecisionsCount)
	}
}

func TestValidateAllIncludesDecisions(t *testing.T) {
	root := setupTestProject(t)
	decDir := DecisionsPath(root)
	writeDecisionFile(t, decDir, "0001-foo.md", makeDecision(1, "foo", "accepted"))

	result, err := ValidateAll(root, false)
	if err != nil {
		t.Fatal(err)
	}
	if result.DecisionsCount != 1 {
		t.Errorf("DecisionsCount = %d, want 1", result.DecisionsCount)
	}
}

func TestValidateDecisionsJSONShape(t *testing.T) {
	root := setupTestProject(t)
	decDir := DecisionsPath(root)
	writeDecisionFile(t, decDir, "0001-foo.md", makeDecision(1, "foo", "accepted"))

	result, err := ValidateDecisions(root)
	if err != nil {
		t.Fatal(err)
	}

	json := BuildValidationResultJSON(result)
	if json.Summary.Decisions != 1 {
		t.Errorf("Summary.Decisions = %d, want 1", json.Summary.Decisions)
	}
}

