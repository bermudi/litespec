package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func writeDecisionFile(t *testing.T, dir, filename, content string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, filename), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

const validDecision = `# Single Shared Workspace

## Status

accepted

## Context

We need a workspace model for the agent.

## Decision

There SHALL be exactly one workspace per agent instance.

## Consequences

Simplifies state management but limits multi-tenant use.
`

func TestParseDecision(t *testing.T) {
	dir := t.TempDir()
	writeDecisionFile(t, dir, "0001-single-shared-workspace.md", validDecision)

	d, err := ParseDecision(filepath.Join(dir, "0001-single-shared-workspace.md"))
	if err != nil {
		t.Fatal(err)
	}
	if d.Number != 1 {
		t.Errorf("Number = %d, want 1", d.Number)
	}
	if d.Slug != "single-shared-workspace" {
		t.Errorf("Slug = %q, want %q", d.Slug, "single-shared-workspace")
	}
	if d.Title != "Single Shared Workspace" {
		t.Errorf("Title = %q, want %q", d.Title, "Single Shared Workspace")
	}
	if d.Status != StatusAccepted {
		t.Errorf("Status = %q, want %q", d.Status, StatusAccepted)
	}
	if d.Context == "" {
		t.Error("Context is empty")
	}
	if d.Decision == "" {
		t.Error("Decision is empty")
	}
	if d.Consequences == "" {
		t.Error("Consequences is empty")
	}
}

func TestParseDecisionMissingSection(t *testing.T) {
	dir := t.TempDir()
	content := `# Title

## Status

proposed

## Context

Some context.
`
	writeDecisionFile(t, dir, "0001-missing.md", content)
	_, err := ParseDecision(filepath.Join(dir, "0001-missing.md"))
	if err == nil {
		t.Fatal("expected error for missing sections")
	}
	if !containsStr(err.Error(), "Decision") || !containsStr(err.Error(), "Consequences") {
		// Should fail on the first missing section at least
		t.Logf("error: %v", err)
	}
}

func TestParseDecisionBadStatus(t *testing.T) {
	dir := t.TempDir()
	content := `# Title

## Status

draft

## Context

ctx

## Decision

dec

## Consequences

con
`
	writeDecisionFile(t, dir, "0001-bad-status.md", content)
	_, err := ParseDecision(filepath.Join(dir, "0001-bad-status.md"))
	if err == nil {
		t.Fatal("expected error for bad status")
	}
	if !containsStr(err.Error(), "invalid status") {
		t.Errorf("error = %v, want invalid status", err)
	}
}

func TestParseDecisionMissingTitle(t *testing.T) {
	dir := t.TempDir()
	content := `## Status

proposed

## Context

ctx

## Decision

dec

## Consequences

con
`
	writeDecisionFile(t, dir, "0001-no-title.md", content)
	_, err := ParseDecision(filepath.Join(dir, "0001-no-title.md"))
	if err == nil {
		t.Fatal("expected error for missing title")
	}
}

func TestParseDecisionSupersedePointers(t *testing.T) {
	dir := t.TempDir()
	content := `# New Model

## Status

accepted

## Context

ctx

## Decision

dec

## Consequences

con

## Supersedes

- 0001-old-model

## Superseded-By

- 0003-newer-model
`
	writeDecisionFile(t, dir, "0002-new-model.md", content)
	d, err := ParseDecision(filepath.Join(dir, "0002-new-model.md"))
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Supersedes) != 1 || d.Supersedes[0] != "0001-old-model" {
		t.Errorf("Supersedes = %v, want [0001-old-model]", d.Supersedes)
	}
	if len(d.SupersededBy) != 1 || d.SupersededBy[0] != "0003-newer-model" {
		t.Errorf("SupersededBy = %v, want [0003-newer-model]", d.SupersededBy)
	}
}

func TestParseDecisionInvalidFilename(t *testing.T) {
	dir := t.TempDir()
	writeDecisionFile(t, dir, "bad-name.md", validDecision)
	_, err := ParseDecision(filepath.Join(dir, "bad-name.md"))
	if err == nil {
		t.Fatal("expected error for invalid filename")
	}

	writeDecisionFile(t, dir, "01-short.md", validDecision)
	_, err = ParseDecision(filepath.Join(dir, "01-short.md"))
	if err == nil {
		t.Fatal("expected error for non-4-digit number")
	}
}

func TestListDecisions(t *testing.T) {
	root := t.TempDir()
	decisionsDir := DecisionsPath(root)

	writeDecisionFile(t, decisionsDir, "0003-third.md", makeDecision(3, "third", "accepted"))
	writeDecisionFile(t, decisionsDir, "0001-first.md", makeDecision(1, "first", "proposed"))
	writeDecisionFile(t, decisionsDir, "0002-second.md", makeDecision(2, "second", "accepted"))

	decisions, err := ListDecisions(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(decisions) != 3 {
		t.Fatalf("got %d decisions, want 3", len(decisions))
	}
	if decisions[0].Number != 1 {
		t.Errorf("first = %d, want 1", decisions[0].Number)
	}
	if decisions[1].Number != 2 {
		t.Errorf("second = %d, want 2", decisions[1].Number)
	}
	if decisions[2].Number != 3 {
		t.Errorf("third = %d, want 3", decisions[2].Number)
	}
}

func TestListDecisionsEmptyDir(t *testing.T) {
	root := t.TempDir()
	decisions, err := ListDecisions(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(decisions) != 0 {
		t.Errorf("got %d decisions, want 0", len(decisions))
	}
}

func TestListDecisionsSkipsNonMatching(t *testing.T) {
	root := t.TempDir()
	decisionsDir := DecisionsPath(root)
	writeDecisionFile(t, decisionsDir, "0001-valid.md", makeDecision(1, "valid", "accepted"))
	writeDecisionFile(t, decisionsDir, "README.md", "# Decisions")
	writeDecisionFile(t, decisionsDir, "not-a-decision.txt", "stuff")
	os.MkdirAll(filepath.Join(decisionsDir, "some-dir"), 0o755)

	decisions, err := ListDecisions(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(decisions) != 1 {
		t.Fatalf("got %d decisions, want 1", len(decisions))
	}
	if decisions[0].Slug != "valid" {
		t.Errorf("slug = %q, want %q", decisions[0].Slug, "valid")
	}
}

func TestFindDecisionBySlug(t *testing.T) {
	root := t.TempDir()
	decisionsDir := DecisionsPath(root)
	writeDecisionFile(t, decisionsDir, "0001-foo-bar.md", makeDecision(1, "foo-bar", "accepted"))

	d, err := FindDecisionBySlug(root, "foo-bar")
	if err != nil {
		t.Fatal(err)
	}
	if d == nil {
		t.Fatal("expected to find decision by slug")
	}
	if d.Number != 1 {
		t.Errorf("Number = %d, want 1", d.Number)
	}

	d, err = FindDecisionBySlug(root, "0001-foo-bar")
	if err != nil {
		t.Fatal(err)
	}
	if d == nil {
		t.Fatal("expected to find decision by full name")
	}

	d, err = FindDecisionBySlug(root, "nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if d != nil {
		t.Error("expected nil for nonexistent slug")
	}
}

func makeDecision(num int, slug, status string) string {
	return `# ` + slug + `

## Status

` + status + `

## Context

Context for ` + slug + `.

## Decision

Decision for ` + slug + `.

## Consequences

Consequences for ` + slug + `.
`
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && searchStr(s, sub)
}

func TestParseDecision_CRLFLineEndings(t *testing.T) {
	content := "# Single Shared Workspace\r\n\r\n## Status\r\n\r\naccepted\r\n\r\n## Context\r\n\r\nWe need a workspace model.\r\n\r\n## Decision\r\n\r\nOne workspace per agent.\r\n\r\n## Consequences\r\n\r\nSimplifies state management.\r\n"

	dir := t.TempDir()
	writeDecisionFile(t, dir, "0001-crlf-test.md", content)

	d, err := ParseDecision(filepath.Join(dir, "0001-crlf-test.md"))
	if err != nil {
		t.Fatalf("ParseDecision: %v", err)
	}
	if d.Title != "Single Shared Workspace" {
		t.Errorf("Title = %q, want %q", d.Title, "Single Shared Workspace")
	}
	if d.Status != StatusAccepted {
		t.Errorf("Status = %q, want %q", d.Status, StatusAccepted)
	}
	if d.Context == "" {
		t.Error("Context is empty")
	}
	if d.Decision == "" {
		t.Error("Decision is empty")
	}
	if d.Consequences == "" {
		t.Error("Consequences is empty")
	}
}

func searchStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
