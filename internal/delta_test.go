package internal

import (
	"testing"
)

func TestParseMainSpecWithScenarios(t *testing.T) {
	input := `# auth

### Requirement: Login
The system SHALL authenticate users.

#### Scenario: Valid credentials
- **WHEN** user submits correct email and password
- **THEN** the system returns a session token

#### Scenario: Invalid credentials
- **WHEN** user submits wrong password
- **THEN** the system returns 401

### Requirement: Logout
The system SHALL invalidate sessions on logout.

#### Scenario: Session invalidation
- **WHEN** user calls logout
- **THEN** the session token is no longer valid
`

	spec, err := ParseMainSpec(input)
	if err != nil {
		t.Fatalf("ParseMainSpec: %v", err)
	}
	if spec.Capability != "auth" {
		t.Errorf("Capability = %q, want %q", spec.Capability, "auth")
	}
	if len(spec.Requirements) != 2 {
		t.Fatalf("Requirements count = %d, want 2", len(spec.Requirements))
	}

	r0 := spec.Requirements[0]
	if r0.Name != "Login" {
		t.Errorf("Req[0].Name = %q, want %q", r0.Name, "Login")
	}
	if r0.Content != "The system SHALL authenticate users." {
		t.Errorf("Req[0].Content = %q", r0.Content)
	}
	if len(r0.Scenarios) != 2 {
		t.Fatalf("Req[0] Scenarios count = %d, want 2", len(r0.Scenarios))
	}
	if r0.Scenarios[0].Name != "Valid credentials" {
		t.Errorf("Scenario[0].Name = %q, want %q", r0.Scenarios[0].Name, "Valid credentials")
	}
	if r0.Scenarios[1].Name != "Invalid credentials" {
		t.Errorf("Scenario[1].Name = %q, want %q", r0.Scenarios[1].Name, "Invalid credentials")
	}

	r1 := spec.Requirements[1]
	if r1.Name != "Logout" {
		t.Errorf("Req[1].Name = %q, want %q", r1.Name, "Logout")
	}
	if len(r1.Scenarios) != 1 {
		t.Fatalf("Req[1] Scenarios count = %d, want 1", len(r1.Scenarios))
	}
}

func TestParseDeltaSpecWithScenarios(t *testing.T) {
	input := `## ADDED Requirements

### Requirement: Rate Limiting
The system SHALL limit API requests per user.

#### Scenario: Exceeds limit
- **WHEN** user sends more than 100 requests per minute
- **THEN** the system returns 429 Too Many Requests
`

	delta, err := ParseDeltaSpec(input)
	if err != nil {
		t.Fatalf("ParseDeltaSpec: %v", err)
	}
	if len(delta.Requirements) != 1 {
		t.Fatalf("Requirements count = %d, want 1", len(delta.Requirements))
	}
	r := delta.Requirements[0]
	if r.Operation != DeltaAdded {
		t.Errorf("Operation = %q, want %q", r.Operation, DeltaAdded)
	}
	if r.Name != "Rate Limiting" {
		t.Errorf("Name = %q, want %q", r.Name, "Rate Limiting")
	}
	if r.Content != "The system SHALL limit API requests per user." {
		t.Errorf("Content = %q", r.Content)
	}
	if len(r.Scenarios) != 1 {
		t.Fatalf("Scenarios count = %d, want 1", len(r.Scenarios))
	}
	if r.Scenarios[0].Name != "Exceeds limit" {
		t.Errorf("Scenario.Name = %q, want %q", r.Scenarios[0].Name, "Exceeds limit")
	}
}

func TestParseRequirementWithNoScenarios(t *testing.T) {
	input := `# test

### Requirement: Simple
The system SHALL do something.
`

	spec, err := ParseMainSpec(input)
	if err != nil {
		t.Fatalf("ParseMainSpec: %v", err)
	}
	if len(spec.Requirements) != 1 {
		t.Fatalf("Requirements count = %d, want 1", len(spec.Requirements))
	}
	r := spec.Requirements[0]
	if r.Content != "The system SHALL do something." {
		t.Errorf("Content = %q", r.Content)
	}
	if len(r.Scenarios) != 0 {
		t.Errorf("Scenarios count = %d, want 0", len(r.Scenarios))
	}
}

func TestParseMultipleScenariosUnderOneRequirement(t *testing.T) {
	input := `# cap

### Requirement: Multi
The system SHALL support multiple scenarios.

#### Scenario: First
- **WHEN** condition A
- **THEN** result A

#### Scenario: Second
- **WHEN** condition B
- **THEN** result B

#### Scenario: Third
- **WHEN** condition C
- **THEN** result C
`

	spec, err := ParseMainSpec(input)
	if err != nil {
		t.Fatalf("ParseMainSpec: %v", err)
	}
	if len(spec.Requirements) != 1 {
		t.Fatalf("Requirements count = %d, want 1", len(spec.Requirements))
	}
	scenarios := spec.Requirements[0].Scenarios
	if len(scenarios) != 3 {
		t.Fatalf("Scenarios count = %d, want 3", len(scenarios))
	}
	names := []string{"First", "Second", "Third"}
	for i, want := range names {
		if scenarios[i].Name != want {
			t.Errorf("Scenario[%d].Name = %q, want %q", i, scenarios[i].Name, want)
		}
	}
}

func TestSerializeRoundTrip(t *testing.T) {
	original := `# auth

### Requirement: Login
The system SHALL authenticate users.

#### Scenario: Valid credentials
- **WHEN** user submits correct email and password
- **THEN** the system returns a session token

### Requirement: Logout
The system SHALL invalidate sessions.

#### Scenario: Session cleared
- **WHEN** user calls logout
- **THEN** the session is gone
`

	spec, err := ParseMainSpec(original)
	if err != nil {
		t.Fatalf("ParseMainSpec: %v", err)
	}

	serialized := SerializeSpec(spec)
	spec2, err := ParseMainSpec(serialized)
	if err != nil {
		t.Fatalf("ParseMainSpec(round-trip): %v", err)
	}

	if spec2.Capability != spec.Capability {
		t.Errorf("Capability: got %q, want %q", spec2.Capability, spec.Capability)
	}
	if len(spec2.Requirements) != len(spec.Requirements) {
		t.Fatalf("Requirements count: got %d, want %d", len(spec2.Requirements), len(spec.Requirements))
	}
	for i, orig := range spec.Requirements {
		got := spec2.Requirements[i]
		if got.Name != orig.Name {
			t.Errorf("Req[%d].Name: got %q, want %q", i, got.Name, orig.Name)
		}
		if got.Content != orig.Content {
			t.Errorf("Req[%d].Content: got %q, want %q", i, got.Content, orig.Content)
		}
		if len(got.Scenarios) != len(orig.Scenarios) {
			t.Errorf("Req[%d].Scenarios count: got %d, want %d", i, len(got.Scenarios), len(orig.Scenarios))
			continue
		}
		for j, sc := range orig.Scenarios {
			if got.Scenarios[j].Name != sc.Name {
				t.Errorf("Req[%d].Scenarios[%d].Name: got %q, want %q", i, j, got.Scenarios[j].Name, sc.Name)
			}
		}
	}
}

func TestMergeDeltaAddedWithScenarios(t *testing.T) {
	main := &Spec{
		Capability: "auth",
		Requirements: []SpecRequirement{
			{Name: "Login", Content: "The system SHALL authenticate.", Scenarios: []Scenario{
				{Name: "Basic", Content: "- **WHEN** creds"},
			}},
		},
	}

	delta := &DeltaSpec{
		Requirements: []DeltaRequirement{
			{
				Operation: DeltaAdded,
				Name:      "Register",
				Content:   "The system SHALL register new users.",
				Scenarios: []Scenario{
					{Name: "New user", Content: "- **WHEN** valid registration"},
					{Name: "Duplicate", Content: "- **WHEN** existing email"},
				},
			},
		},
	}

	result, err := MergeDelta(main, []*DeltaSpec{delta})
	if err != nil {
		t.Fatalf("MergeDelta: %v", err)
	}
	if len(result.Requirements) != 2 {
		t.Fatalf("Requirements count = %d, want 2", len(result.Requirements))
	}
	added := result.Requirements[1]
	if added.Name != "Register" {
		t.Errorf("Added req Name = %q, want %q", added.Name, "Register")
	}
	if len(added.Scenarios) != 2 {
		t.Errorf("Added req Scenarios count = %d, want 2", len(added.Scenarios))
	}
}

func TestMergeDeltaModifiedWithScenarios(t *testing.T) {
	main := &Spec{
		Capability: "auth",
		Requirements: []SpecRequirement{
			{Name: "Login", Content: "Old content.", Scenarios: []Scenario{
				{Name: "Old scenario", Content: "old"},
			}},
		},
	}

	delta := &DeltaSpec{
		Requirements: []DeltaRequirement{
			{
				Operation: DeltaModified,
				Name:      "Login",
				Content:   "The system SHALL authenticate via SSO.",
				Scenarios: []Scenario{
					{Name: "SSO login", Content: "- **WHEN** SSO token valid"},
					{Name: "SSO failed", Content: "- **WHEN** SSO token invalid"},
				},
			},
		},
	}

	result, err := MergeDelta(main, []*DeltaSpec{delta})
	if err != nil {
		t.Fatalf("MergeDelta: %v", err)
	}
	r := result.Requirements[0]
	if r.Content != "The system SHALL authenticate via SSO." {
		t.Errorf("Content = %q", r.Content)
	}
	if len(r.Scenarios) != 2 {
		t.Fatalf("Scenarios count = %d, want 2", len(r.Scenarios))
	}
	if r.Scenarios[0].Name != "SSO login" {
		t.Errorf("Scenario[0].Name = %q", r.Scenarios[0].Name)
	}
}

func TestMergeDeltaRenamedPreservesScenarios(t *testing.T) {
	main := &Spec{
		Capability: "auth",
		Requirements: []SpecRequirement{
			{
				Name:    "Old Login",
				Content: "The system SHALL authenticate.",
				Scenarios: []Scenario{
					{Name: "Valid", Content: "- **WHEN** valid creds"},
				},
			},
		},
	}

	delta := &DeltaSpec{
		Requirements: []DeltaRequirement{
			{
				Operation: DeltaRenamed,
				OldName:   "Old Login",
				Name:      "Authenticate User",
			},
		},
	}

	result, err := MergeDelta(main, []*DeltaSpec{delta})
	if err != nil {
		t.Fatalf("MergeDelta: %v", err)
	}
	r := result.Requirements[0]
	if r.Name != "Authenticate User" {
		t.Errorf("Name = %q, want %q", r.Name, "Authenticate User")
	}
	if r.Content != "The system SHALL authenticate." {
		t.Errorf("Content = %q, want preserved original", r.Content)
	}
	if len(r.Scenarios) != 1 {
		t.Fatalf("Scenarios count = %d, want 1 (preserved)", len(r.Scenarios))
	}
	if r.Scenarios[0].Name != "Valid" {
		t.Errorf("Scenario[0].Name = %q, want %q", r.Scenarios[0].Name, "Valid")
	}
}

func TestParseDeltaSpecRemovedNoScenarios(t *testing.T) {
	input := `## REMOVED Requirements

### Requirement: Legacy Login
### Requirement: Old Feature
`

	delta, err := ParseDeltaSpec(input)
	if err != nil {
		t.Fatalf("ParseDeltaSpec: %v", err)
	}
	if len(delta.Requirements) != 2 {
		t.Fatalf("Requirements count = %d, want 2", len(delta.Requirements))
	}
	for _, r := range delta.Requirements {
		if r.Operation != DeltaRemoved {
			t.Errorf("Operation = %q, want %q", r.Operation, DeltaRemoved)
		}
		if len(r.Scenarios) != 0 {
			t.Errorf("Removed req %q should have 0 scenarios, got %d", r.Name, len(r.Scenarios))
		}
	}
}

func TestSerializeSpecWithNoScenarios(t *testing.T) {
	spec := &Spec{
		Capability: "test",
		Requirements: []SpecRequirement{
			{Name: "Simple", Content: "The system SHALL work."},
		},
	}

	out := SerializeSpec(spec)
	spec2, err := ParseMainSpec(out)
	if err != nil {
		t.Fatalf("ParseMainSpec: %v", err)
	}
	if len(spec2.Requirements[0].Scenarios) != 0 {
		t.Errorf("Expected 0 scenarios, got %d", len(spec2.Requirements[0].Scenarios))
	}
}

func TestParseDeltaSpecModifiedReplacesScenarios(t *testing.T) {
	input := `## MODIFIED Requirements

### Requirement: Login
The system SHALL authenticate via OAuth.

#### Scenario: OAuth success
- **WHEN** OAuth provider returns valid token
- **THEN** user is authenticated
`

	delta, err := ParseDeltaSpec(input)
	if err != nil {
		t.Fatalf("ParseDeltaSpec: %v", err)
	}
	if len(delta.Requirements) != 1 {
		t.Fatalf("Requirements count = %d, want 1", len(delta.Requirements))
	}
	r := delta.Requirements[0]
	if r.Operation != DeltaModified {
		t.Errorf("Operation = %q, want %q", r.Operation, DeltaModified)
	}
	if len(r.Scenarios) != 1 {
		t.Errorf("Scenarios count = %d, want 1", len(r.Scenarios))
	}
	if r.Scenarios[0].Name != "OAuth success" {
		t.Errorf("Scenario.Name = %q, want %q", r.Scenarios[0].Name, "OAuth success")
	}
}

func TestDeltaSpecCapabilityFallbackInMerge(t *testing.T) {
	main := &Spec{
		Capability:   "auth",
		Requirements: []SpecRequirement{},
	}

	delta := &DeltaSpec{
		Requirements: []DeltaRequirement{
			{
				Operation: DeltaAdded,
				Name:      "New Req",
				Content:   "The system SHALL do new things.",
				Scenarios: []Scenario{
					{Name: "Works", Content: "- **WHEN** triggered"},
				},
			},
		},
	}

	result, err := MergeDelta(main, []*DeltaSpec{delta})
	if err != nil {
		t.Fatalf("MergeDelta: %v", err)
	}
	if result.Capability != "auth" {
		t.Errorf("Capability = %q, want %q", result.Capability, "auth")
	}
	if len(result.Requirements) != 1 {
		t.Fatalf("Requirements count = %d, want 1", len(result.Requirements))
	}
}
