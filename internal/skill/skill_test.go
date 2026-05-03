package skill

import (
	"testing"
)

func TestGet_ReturnsNonEmptyForKnownIDs(t *testing.T) {
	knownIDs := []string{
		"explore", "grill", "propose", "review",
		"apply", "adopt", "glossary", "patch", "fix",
		"research", "workflow",
		"artifact-proposal", "artifact-specs",
		"artifact-design", "artifact-tasks",
	}
	for _, id := range knownIDs {
		tmpl := Get(id)
		if tmpl == "" {
			t.Errorf("Get(%q) returned empty string, expected non-empty template", id)
		}
	}
}

func TestValidateSkillTemplates_AllPresent(t *testing.T) {
	original := templates
	defer func() { templates = original }()

	templates = map[string]string{
		"alpha": "template-a",
		"beta":  "template-b",
	}

	missing := ValidateSkillTemplates([]string{"alpha", "beta"})
	if missing == nil {
		t.Error("expected empty slice, got nil")
	}
	if len(missing) != 0 {
		t.Errorf("expected 0 missing, got %d: %v", len(missing), missing)
	}
}

func TestValidateSkillTemplates_SomeMissing(t *testing.T) {
	original := templates
	defer func() { templates = original }()

	templates = map[string]string{
		"alpha": "template-a",
	}

	missing := ValidateSkillTemplates([]string{"alpha", "beta", "gamma"})
	if len(missing) != 2 {
		t.Fatalf("expected 2 missing, got %d: %v", len(missing), missing)
	}

	found := map[string]bool{}
	for _, id := range missing {
		found[id] = true
	}
	if !found["beta"] || !found["gamma"] {
		t.Errorf("expected missing beta and gamma, got %v", missing)
	}
}

func TestValidateSkillTemplates_EmptyInput(t *testing.T) {
	missing := ValidateSkillTemplates([]string{})
	if missing == nil {
		t.Error("expected empty slice, got nil")
	}
	if len(missing) != 0 {
		t.Errorf("expected 0 missing, got %d", len(missing))
	}
}
