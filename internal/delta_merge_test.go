package internal

import (
	"testing"
)

func TestMergeDeltaAddedToEmptySpec(t *testing.T) {
	main := &Spec{
		Capability:   "newcap",
		Requirements: []SpecRequirement{},
	}
	delta := &DeltaSpec{
		Requirements: []DeltaRequirement{
			{
				Operation: DeltaAdded,
				Name:      "First Req",
				Content:   "The system SHALL do X.",
				Scenarios: []Scenario{{Name: "Happy path", Content: "- **WHEN** X"}},
			},
		},
	}

	result, err := MergeDelta(main, []*DeltaSpec{delta})
	if err != nil {
		t.Fatalf("MergeDelta: %v", err)
	}
	if len(result.Requirements) != 1 {
		t.Fatalf("Requirements count = %d, want 1", len(result.Requirements))
	}
	if result.Requirements[0].Name != "First Req" {
		t.Errorf("Name = %q, want %q", result.Requirements[0].Name, "First Req")
	}
	if result.Capability != "newcap" {
		t.Errorf("Capability = %q, want %q", result.Capability, "newcap")
	}
}

func TestMergeDeltaRemoved(t *testing.T) {
	main := &Spec{
		Capability: "auth",
		Requirements: []SpecRequirement{
			{Name: "Login", Content: "SHALL authenticate"},
			{Name: "Legacy", Content: "SHALL do legacy thing"},
			{Name: "Logout", Content: "SHALL invalidate"},
		},
	}
	delta := &DeltaSpec{
		Requirements: []DeltaRequirement{
			{Operation: DeltaRemoved, Name: "Legacy"},
		},
	}

	result, err := MergeDelta(main, []*DeltaSpec{delta})
	if err != nil {
		t.Fatalf("MergeDelta: %v", err)
	}
	if len(result.Requirements) != 2 {
		t.Fatalf("Requirements count = %d, want 2", len(result.Requirements))
	}
	for _, r := range result.Requirements {
		if r.Name == "Legacy" {
			t.Error("Legacy should have been removed")
		}
	}
}

func TestMergeDeltaRenameThenModify(t *testing.T) {
	main := &Spec{
		Capability: "auth",
		Requirements: []SpecRequirement{
			{Name: "Login", Content: "Old content", Scenarios: []Scenario{
				{Name: "Basic", Content: "when basic"},
			}},
		},
	}
	delta := &DeltaSpec{
		Requirements: []DeltaRequirement{
			{Operation: DeltaRenamed, OldName: "Login", Name: "Authenticate"},
			{Operation: DeltaModified, Name: "Authenticate", Content: "New content SHALL work.", Scenarios: []Scenario{
				{Name: "Updated", Content: "when updated"},
			}},
		},
	}

	result, err := MergeDelta(main, []*DeltaSpec{delta})
	if err != nil {
		t.Fatalf("MergeDelta: %v", err)
	}
	if len(result.Requirements) != 1 {
		t.Fatalf("Requirements count = %d, want 1", len(result.Requirements))
	}
	r := result.Requirements[0]
	if r.Name != "Authenticate" {
		t.Errorf("Name = %q, want %q", r.Name, "Authenticate")
	}
	if r.Content != "New content SHALL work." {
		t.Errorf("Content = %q, want updated content after rename+modify", r.Content)
	}
	if len(r.Scenarios) != 1 || r.Scenarios[0].Name != "Updated" {
		t.Errorf("Scenarios not replaced correctly after rename+modify: %+v", r.Scenarios)
	}
}

func TestMergeDeltaModifiedNotFound(t *testing.T) {
	main := &Spec{
		Capability:   "auth",
		Requirements: []SpecRequirement{},
	}
	delta := &DeltaSpec{
		Requirements: []DeltaRequirement{
			{Operation: DeltaModified, Name: "Ghost", Content: "stuff"},
		},
	}
	_, err := MergeDelta(main, []*DeltaSpec{delta})
	if err == nil {
		t.Fatal("expected error for MODIFIED nonexistent requirement")
	}
}

func TestMergeDeltaRemovedNotFound(t *testing.T) {
	main := &Spec{
		Capability:   "auth",
		Requirements: []SpecRequirement{},
	}
	delta := &DeltaSpec{
		Requirements: []DeltaRequirement{
			{Operation: DeltaRemoved, Name: "Ghost"},
		},
	}
	_, err := MergeDelta(main, []*DeltaSpec{delta})
	if err == nil {
		t.Fatal("expected error for REMOVED nonexistent requirement")
	}
}

func TestMergeDeltaRenamedNotFound(t *testing.T) {
	main := &Spec{
		Capability:   "auth",
		Requirements: []SpecRequirement{},
	}
	delta := &DeltaSpec{
		Requirements: []DeltaRequirement{
			{Operation: DeltaRenamed, OldName: "Ghost", Name: "NewGhost"},
		},
	}
	_, err := MergeDelta(main, []*DeltaSpec{delta})
	if err == nil {
		t.Fatal("expected error for RENAMED nonexistent requirement")
	}
}

func TestMergeDeltaMultipleDeltasAppliedInOrder(t *testing.T) {
	main := &Spec{
		Capability: "cap",
		Requirements: []SpecRequirement{
			{Name: "A", Content: "original A"},
			{Name: "B", Content: "original B"},
		},
	}
	d1 := &DeltaSpec{
		Requirements: []DeltaRequirement{
			{Operation: DeltaRenamed, OldName: "A", Name: "Alpha"},
			{Operation: DeltaRemoved, Name: "B"},
		},
	}
	d2 := &DeltaSpec{
		Requirements: []DeltaRequirement{
			{Operation: DeltaAdded, Name: "C", Content: "new C"},
		},
	}

	result, err := MergeDelta(main, []*DeltaSpec{d1, d2})
	if err != nil {
		t.Fatalf("MergeDelta: %v", err)
	}
	if len(result.Requirements) != 2 {
		t.Fatalf("Requirements count = %d, want 2", len(result.Requirements))
	}
	if result.Requirements[0].Name != "Alpha" {
		t.Errorf("Req[0].Name = %q, want %q", result.Requirements[0].Name, "Alpha")
	}
	if result.Requirements[0].Content != "original A" {
		t.Errorf("Req[0].Content = %q, want preserved original", result.Requirements[0].Content)
	}
	if result.Requirements[1].Name != "C" {
		t.Errorf("Req[1].Name = %q, want %q", result.Requirements[1].Name, "C")
	}
}

func TestMergeDeltaNoRequirements(t *testing.T) {
	main := &Spec{
		Capability:   "cap",
		Requirements: []SpecRequirement{},
	}
	delta := &DeltaSpec{}

	result, err := MergeDelta(main, []*DeltaSpec{delta})
	if err != nil {
		t.Fatalf("MergeDelta: %v", err)
	}
	if len(result.Requirements) != 0 {
		t.Errorf("Requirements count = %d, want 0", len(result.Requirements))
	}
}

func TestMergeDeltaRemoveThenAddSameName(t *testing.T) {
	main := &Spec{
		Capability: "auth",
		Requirements: []SpecRequirement{
			{Name: "Login", Content: "old content"},
		},
	}
	delta := &DeltaSpec{
		Requirements: []DeltaRequirement{
			{Operation: DeltaRemoved, Name: "Login"},
			{Operation: DeltaAdded, Name: "Login", Content: "new content", Scenarios: []Scenario{
				{Name: "Works", Content: "when works"},
			}},
		},
	}

	result, err := MergeDelta(main, []*DeltaSpec{delta})
	if err != nil {
		t.Fatalf("MergeDelta: %v", err)
	}
	if len(result.Requirements) != 1 {
		t.Fatalf("Requirements count = %d, want 1", len(result.Requirements))
	}
	if result.Requirements[0].Content != "new content" {
		t.Errorf("Content = %q, want new content after remove+add", result.Requirements[0].Content)
	}
}
