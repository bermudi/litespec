package internal

import (
	"testing"
	"time"
)

func TestBuildValidationResultJSON_ValidNoIssues(t *testing.T) {
	r := &ValidationResult{Valid: true, Errors: nil, Warnings: nil}
	got := BuildValidationResultJSON(r)
	if !got.Valid {
		t.Error("expected Valid=true")
	}
	if len(got.Errors) != 0 {
		t.Errorf("expected 0 errors, got %d", len(got.Errors))
	}
	if len(got.Warnings) != 0 {
		t.Errorf("expected 0 warnings, got %d", len(got.Warnings))
	}
	if got.Summary.Total != 0 {
		t.Errorf("expected Summary.Total=0, got %d", got.Summary.Total)
	}
	if got.Summary.Invalid != 0 {
		t.Errorf("expected Summary.Invalid=0, got %d", got.Summary.Invalid)
	}
}

func TestBuildValidationResultJSON_WithErrorsAndWarnings(t *testing.T) {
	r := &ValidationResult{
		Valid: false,
		Errors: []ValidationIssue{
			{Severity: SeverityError, Message: "err1", File: "a.md"},
			{Severity: SeverityError, Message: "err2", File: "b.md"},
		},
		Warnings: []ValidationIssue{
			{Severity: SeverityWarning, Message: "warn1", File: "c.md"},
		},
	}
	got := BuildValidationResultJSON(r)
	if got.Valid {
		t.Error("expected Valid=false")
	}
	if len(got.Errors) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(got.Errors))
	}
	if got.Errors[0].Severity != "error" || got.Errors[0].Message != "err1" || got.Errors[0].File != "a.md" {
		t.Errorf("unexpected error[0]: %+v", got.Errors[0])
	}
	if got.Errors[1].Severity != "error" || got.Errors[1].Message != "err2" || got.Errors[1].File != "b.md" {
		t.Errorf("unexpected error[1]: %+v", got.Errors[1])
	}
	if len(got.Warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(got.Warnings))
	}
	if got.Warnings[0].Severity != "warning" || got.Warnings[0].Message != "warn1" || got.Warnings[0].File != "c.md" {
		t.Errorf("unexpected warning[0]: %+v", got.Warnings[0])
	}
	if got.Summary.Total != 3 {
		t.Errorf("expected Summary.Total=3, got %d", got.Summary.Total)
	}
	if got.Summary.Invalid != 2 {
		t.Errorf("expected Summary.Invalid=2, got %d", got.Summary.Invalid)
	}
}

func TestBuildValidationResultJSON_NilSlicesProduceEmpty(t *testing.T) {
	r := &ValidationResult{Valid: true, Errors: nil, Warnings: nil}
	got := BuildValidationResultJSON(r)
	if got.Errors == nil {
		t.Error("Errors should not be nil")
	}
	if got.Warnings == nil {
		t.Error("Warnings should not be nil")
	}
}

func TestBuildChangeStatusJSON_AllDone(t *testing.T) {
	c := &Change{
		Name:    "my-change",
		Schema:  "spec-driven",
		Created: time.Now(),
		Artifacts: map[string]ArtifactState{
			"proposal": ArtifactDone,
			"specs":    ArtifactDone,
			"design":   ArtifactDone,
			"tasks":    ArtifactDone,
		},
	}
	got := BuildChangeStatusJSON(c)
	if !got.IsComplete {
		t.Error("expected IsComplete=true")
	}
	if got.ChangeName != "my-change" {
		t.Errorf("expected ChangeName=my-change, got %q", got.ChangeName)
	}
	if got.SchemaName != "spec-driven" {
		t.Errorf("expected SchemaName=spec-driven, got %q", got.SchemaName)
	}
	if len(got.Artifacts) != 4 {
		t.Fatalf("expected 4 artifacts, got %d", len(got.Artifacts))
	}
	for _, a := range got.Artifacts {
		if a.Status != "done" {
			t.Errorf("expected artifact %q status=done, got %q", a.ID, a.Status)
		}
	}
}

func TestBuildChangeStatusJSON_PartialArtifacts(t *testing.T) {
	c := &Change{
		Name:    "partial",
		Schema:  "spec-driven",
		Created: time.Now(),
		Artifacts: map[string]ArtifactState{
			"proposal": ArtifactReady,
			"specs":    ArtifactBlocked,
			"design":   ArtifactBlocked,
			"tasks":    ArtifactBlocked,
		},
	}
	got := BuildChangeStatusJSON(c)
	if got.IsComplete {
		t.Error("expected IsComplete=false")
	}
	statusMap := map[string]ArtifactStatusJSON{}
	for _, a := range got.Artifacts {
		statusMap[a.ID] = a
	}
	if statusMap["proposal"].Status != "ready" {
		t.Errorf("proposal: expected ready, got %q", statusMap["proposal"].Status)
	}
	if statusMap["specs"].Status != "blocked" {
		t.Errorf("specs: expected blocked, got %q", statusMap["specs"].Status)
	}
	if statusMap["design"].Status != "blocked" {
		t.Errorf("design: expected blocked, got %q", statusMap["design"].Status)
	}
	if statusMap["tasks"].Status != "blocked" {
		t.Errorf("tasks: expected blocked, got %q", statusMap["tasks"].Status)
	}
	if len(statusMap["specs"].MissingDeps) != 1 || statusMap["specs"].MissingDeps[0] != "proposal" {
		t.Errorf("specs MissingDeps: expected [proposal], got %v", statusMap["specs"].MissingDeps)
	}
	if len(statusMap["design"].MissingDeps) != 1 || statusMap["design"].MissingDeps[0] != "proposal" {
		t.Errorf("design MissingDeps: expected [proposal], got %v", statusMap["design"].MissingDeps)
	}
	if len(statusMap["tasks"].MissingDeps) != 3 {
		t.Errorf("tasks MissingDeps: expected 3, got %v", statusMap["tasks"].MissingDeps)
	}
}

func TestBuildChangeStatusJSON_EmptyArtifacts(t *testing.T) {
	c := &Change{
		Name:      "empty",
		Schema:    "spec-driven",
		Created:   time.Now(),
		Artifacts: map[string]ArtifactState{},
	}
	got := BuildChangeStatusJSON(c)
	if got.IsComplete {
		t.Error("expected IsComplete=false")
	}
	for _, a := range got.Artifacts {
		if a.Status != "blocked" {
			t.Errorf("artifact %q: expected blocked, got %q", a.ID, a.Status)
		}
	}
}

func TestArtifactStateToString(t *testing.T) {
	if got := artifactStateToString(ArtifactBlocked); got != "blocked" {
		t.Errorf("expected blocked, got %q", got)
	}
	if got := artifactStateToString(ArtifactReady); got != "ready" {
		t.Errorf("expected ready, got %q", got)
	}
	if got := artifactStateToString(ArtifactDone); got != "done" {
		t.Errorf("expected done, got %q", got)
	}
}

func TestBuildChangeStatusJSON_ArtifactOrderMatchesArtifacts(t *testing.T) {
	c := &Change{
		Name:    "ordered",
		Schema:  "spec-driven",
		Created: time.Now(),
		Artifacts: map[string]ArtifactState{
			"proposal": ArtifactDone,
			"specs":    ArtifactDone,
			"design":   ArtifactDone,
			"tasks":    ArtifactDone,
		},
	}
	got := BuildChangeStatusJSON(c)
	expected := []string{"proposal", "specs", "design", "tasks"}
	for i, a := range got.Artifacts {
		if a.ID != expected[i] {
			t.Errorf("artifact[%d].ID = %q, want %q", i, a.ID, expected[i])
		}
		if a.OutputPath != Artifacts[i].Filename {
			t.Errorf("artifact[%d].OutputPath = %q, want %q", i, a.OutputPath, Artifacts[i].Filename)
		}
	}
}

func TestBuildChangeStatusJSON_SpecsReadyHasNoMissingDeps(t *testing.T) {
	c := &Change{
		Name:    "deps-check",
		Schema:  "spec-driven",
		Created: time.Now(),
		Artifacts: map[string]ArtifactState{
			"proposal": ArtifactDone,
			"specs":    ArtifactReady,
		},
	}
	got := BuildChangeStatusJSON(c)
	for _, a := range got.Artifacts {
		if a.ID == "specs" {
			if len(a.MissingDeps) != 0 {
				t.Errorf("specs MissingDeps should be empty when proposal is done, got %v", a.MissingDeps)
			}
		}
	}
}
