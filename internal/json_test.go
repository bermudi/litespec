package internal

import (
	"testing"
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
