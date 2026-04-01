package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runArchivePipeline(t *testing.T, root, changeName string) {
	t.Helper()

	changeSpecsDir := ChangeSpecsPath(root, changeName)
	entries, err := os.ReadDir(changeSpecsDir)
	if err != nil {
		t.Fatalf("read change specs: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		capability := entry.Name()
		capDir := filepath.Join(changeSpecsDir, capability)
		files, readErr := os.ReadDir(capDir)
		if readErr != nil {
			continue
		}

		var deltas []*DeltaSpec
		for _, f := range files {
			if filepath.Ext(f.Name()) != ".md" {
				continue
			}
			data, readErr := os.ReadFile(filepath.Join(capDir, f.Name()))
			if readErr != nil {
				t.Fatalf("read delta: %v", readErr)
			}
			delta, parseErr := ParseDeltaSpec(string(data))
			if parseErr != nil {
				t.Fatalf("parse delta: %v", parseErr)
			}
			deltas = append(deltas, delta)
		}

		if len(deltas) == 0 {
			continue
		}

		mainSpecDir := filepath.Join(SpecsPath(root), capability)
		mainSpecPath := filepath.Join(mainSpecDir, "spec.md")
		mainData, readErr := os.ReadFile(mainSpecPath)

		var mainSpec *Spec
		if readErr != nil {
			cap := deltas[0].Capability
			if cap == "" {
				cap = capability
			}
			mainSpec = &Spec{Capability: cap}
		} else {
			mainSpec, err = ParseMainSpec(string(mainData))
			if err != nil {
				t.Fatalf("parse main spec: %v", err)
			}
		}

		merged, err := MergeDelta(mainSpec, deltas)
		if err != nil {
			t.Fatalf("merge delta: %v", err)
		}

		if err := os.MkdirAll(mainSpecDir, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(mainSpecPath, []byte(SerializeSpec(merged)), 0o644); err != nil {
			t.Fatalf("write spec: %v", err)
		}
	}

	if err := ArchiveChange(root, changeName); err != nil {
		t.Fatalf("ArchiveChange: %v", err)
	}
}

func TestArchiveNewCapability(t *testing.T) {
	root := setupTestProject(t)
	writeChangeFile(t, root, "add-rate-limit", "proposal.md", "# Proposal\nMotivation.")
	writeChangeFile(t, root, "add-rate-limit", "design.md", "# Design\nApproach.")
	writeChangeFile(t, root, "add-rate-limit", "tasks.md", "## Phase 1\n- [x] Done")

	deltaContent := `# rate-limit

## ADDED Requirements

### Requirement: Rate Limiting
The system SHALL limit API requests.

#### Scenario: Exceeds limit
- **WHEN** user exceeds limit
- **THEN** return 429
`
	writeDeltaSpecFile(t, root, "add-rate-limit", "rate-limit", "spec.md", deltaContent)

	runArchivePipeline(t, root, "add-rate-limit")

	mainSpecPath := filepath.Join(SpecsPath(root), "rate-limit", "spec.md")
	data, err := os.ReadFile(mainSpecPath)
	if err != nil {
		t.Fatalf("read main spec: %v", err)
	}
	spec, err := ParseMainSpec(string(data))
	if err != nil {
		t.Fatalf("ParseMainSpec: %v", err)
	}
	if spec.Capability != "rate-limit" {
		t.Errorf("Capability = %q, want %q", spec.Capability, "rate-limit")
	}
	if len(spec.Requirements) != 1 {
		t.Fatalf("Requirements count = %d, want 1", len(spec.Requirements))
	}
	if spec.Requirements[0].Name != "Rate Limiting" {
		t.Errorf("Name = %q, want %q", spec.Requirements[0].Name, "Rate Limiting")
	}
	if len(spec.Requirements[0].Scenarios) != 1 {
		t.Errorf("Scenarios count = %d, want 1", len(spec.Requirements[0].Scenarios))
	}

	if _, err := os.Stat(ChangePath(root, "add-rate-limit")); !os.IsNotExist(err) {
		t.Error("change directory should be gone after archive")
	}
	archiveEntries, _ := os.ReadDir(ArchivePath(root))
	found := false
	for _, e := range archiveEntries {
		if strings.HasSuffix(e.Name(), "-add-rate-limit") {
			found = true
		}
	}
	if !found {
		t.Error("archived directory not found in archive")
	}
}

func TestArchiveModifyExistingCapability(t *testing.T) {
	root := setupTestProject(t)

	writeMainSpecFile(t, root, "auth", `# auth

### Requirement: Login
The system SHALL authenticate.

#### Scenario: Valid
- **WHEN** valid creds
`)

	writeChangeFile(t, root, "mod-auth", "proposal.md", "# Proposal")
	writeChangeFile(t, root, "mod-auth", "design.md", "# Design")
	writeChangeFile(t, root, "mod-auth", "tasks.md", "## Phase 1\n- [x] Done")

	writeDeltaSpecFile(t, root, "mod-auth", "auth", "spec.md", `## MODIFIED Requirements

### Requirement: Login
The system SHALL authenticate via OAuth.

#### Scenario: OAuth
- **WHEN** OAuth token valid
`)

	runArchivePipeline(t, root, "mod-auth")

	data, err := os.ReadFile(filepath.Join(SpecsPath(root), "auth", "spec.md"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	spec, err := ParseMainSpec(string(data))
	if err != nil {
		t.Fatalf("ParseMainSpec: %v", err)
	}
	if spec.Capability != "auth" {
		t.Errorf("Capability = %q", spec.Capability)
	}
	if len(spec.Requirements) != 1 {
		t.Fatalf("Requirements count = %d, want 1", len(spec.Requirements))
	}
	r := spec.Requirements[0]
	if r.Content != "The system SHALL authenticate via OAuth." {
		t.Errorf("Content = %q, want updated", r.Content)
	}
	if len(r.Scenarios) != 1 || r.Scenarios[0].Name != "OAuth" {
		t.Errorf("Scenarios not replaced: %+v", r.Scenarios)
	}
}

func TestArchiveRenameThenModify(t *testing.T) {
	root := setupTestProject(t)

	writeMainSpecFile(t, root, "auth", `# auth

### Requirement: Login
The system SHALL authenticate.

#### Scenario: Valid
- **WHEN** valid creds
`)

	writeChangeFile(t, root, "rename-mod", "proposal.md", "# Proposal")
	writeChangeFile(t, root, "rename-mod", "design.md", "# Design")
	writeChangeFile(t, root, "rename-mod", "tasks.md", "## Phase 1\n- [x] Done")

	writeDeltaSpecFile(t, root, "rename-mod", "auth", "spec.md", `## RENAMED Requirements

### Requirement: Login → Authenticate

## MODIFIED Requirements

### Requirement: Authenticate
The system SHALL authenticate via SSO.

#### Scenario: SSO
- **WHEN** SSO token valid
`)

	runArchivePipeline(t, root, "rename-mod")

	data, err := os.ReadFile(filepath.Join(SpecsPath(root), "auth", "spec.md"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	spec, err := ParseMainSpec(string(data))
	if err != nil {
		t.Fatalf("ParseMainSpec: %v", err)
	}
	if len(spec.Requirements) != 1 {
		t.Fatalf("Requirements count = %d, want 1", len(spec.Requirements))
	}
	r := spec.Requirements[0]
	if r.Name != "Authenticate" {
		t.Errorf("Name = %q, want %q (renamed)", r.Name, "Authenticate")
	}
	if r.Content != "The system SHALL authenticate via SSO." {
		t.Errorf("Content = %q, want modified content", r.Content)
	}
	if len(r.Scenarios) != 1 || r.Scenarios[0].Name != "SSO" {
		t.Errorf("Scenarios not replaced correctly: %+v", r.Scenarios)
	}
}

func TestArchiveDanglingDeltaRejected(t *testing.T) {
	root := setupTestProject(t)
	writeChangeFile(t, root, "bad-change", "proposal.md", "# Proposal")
	writeChangeFile(t, root, "bad-change", "design.md", "# Design")
	writeChangeFile(t, root, "bad-change", "tasks.md", "## Phase 1\n- [ ] Task")

	writeDeltaSpecFile(t, root, "bad-change", "auth", "spec.md", `## MODIFIED Requirements

### Requirement: Nonexistent
The system SHALL do something.

#### Scenario: S1
- **WHEN** triggered
`)

	result, err := ValidateChange(root, "bad-change")
	if err != nil {
		t.Fatalf("ValidateChange: %v", err)
	}
	if result.Valid {
		t.Fatal("validation should reject dangling delta")
	}
	if _, statErr := os.Stat(ChangePath(root, "bad-change")); os.IsNotExist(statErr) {
		t.Fatal("change directory should still exist (archive should not have run)")
	}
}

func TestArchiveNewCapabilityUsesDirNameAsFallback(t *testing.T) {
	root := setupTestProject(t)
	writeChangeFile(t, root, "add-cap", "proposal.md", "# Proposal")
	writeChangeFile(t, root, "add-cap", "design.md", "# Design")
	writeChangeFile(t, root, "add-cap", "tasks.md", "## Phase 1\n- [x] Done")

	writeDeltaSpecFile(t, root, "add-cap", "my-feature", "spec.md", `## ADDED Requirements

### Requirement: Core
The system SHALL provide the feature.

#### Scenario: Basic
- **WHEN** feature is used
`)

	runArchivePipeline(t, root, "add-cap")

	data, err := os.ReadFile(filepath.Join(SpecsPath(root), "my-feature", "spec.md"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	spec, err := ParseMainSpec(string(data))
	if err != nil {
		t.Fatalf("ParseMainSpec: %v", err)
	}
	if spec.Capability != "my-feature" {
		t.Errorf("Capability = %q, want %q (directory name fallback)", spec.Capability, "my-feature")
	}
}
