package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runArchivePipeline(t *testing.T, root, changeName string) {
	t.Helper()

	writes, err := PrepareArchiveWrites(root, changeName)
	if err != nil {
		t.Fatalf("PrepareArchiveWrites: %v", err)
	}

	if err := WritePendingSpecs(writes); err != nil {
		t.Fatalf("WritePendingSpecs: %v", err)
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

	mainSpecPath := filepath.Join(CanonPath(root), "rate-limit", "spec.md")
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

func TestArchiveStripsSpecsSubtree(t *testing.T) {
	root := setupTestProject(t)
	writeChangeFile(t, root, "strip-test", "proposal.md", "# Proposal")
	writeChangeFile(t, root, "strip-test", "design.md", "# Design")
	writeChangeFile(t, root, "strip-test", "tasks.md", "## Phase 1\n- [x] Done")

	writeDeltaSpecFile(t, root, "strip-test", "auth", "spec.md", `## ADDED Requirements

### Requirement: Login
The system SHALL authenticate.

#### Scenario: Valid
- **WHEN** valid creds
`)

	runArchivePipeline(t, root, "strip-test")

	archiveEntries, _ := os.ReadDir(ArchivePath(root))
	var archivedName string
	for _, e := range archiveEntries {
		if strings.HasSuffix(e.Name(), "-strip-test") {
			archivedName = e.Name()
		}
	}
	if archivedName == "" {
		t.Fatal("archived directory not found")
	}

	specsSubtree := filepath.Join(ArchivePath(root), archivedName, "specs")
	if _, err := os.Stat(specsSubtree); !os.IsNotExist(err) {
		t.Errorf("archived directory MUST NOT contain specs/ subtree, but %s exists", specsSubtree)
	}

	proposalPath := filepath.Join(ArchivePath(root), archivedName, "proposal.md")
	if _, err := os.Stat(proposalPath); os.IsNotExist(err) {
		t.Error("archived directory should still contain proposal.md")
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

	data, err := os.ReadFile(filepath.Join(CanonPath(root), "auth", "spec.md"))
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

	data, err := os.ReadFile(filepath.Join(CanonPath(root), "auth", "spec.md"))
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

	data, err := os.ReadFile(filepath.Join(CanonPath(root), "my-feature", "spec.md"))
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

func TestArchiveRENAMEDDanglingRejected(t *testing.T) {
	root := setupTestProject(t)
	writeMainSpecFile(t, root, "auth", `# auth

### Requirement: Login
The system SHALL authenticate.

#### Scenario: Valid
- **WHEN** valid creds
`)
	writeChangeFile(t, root, "bad-rename", "proposal.md", "# Proposal")
	writeChangeFile(t, root, "bad-rename", "design.md", "# Design")
	writeChangeFile(t, root, "bad-rename", "tasks.md", "## Phase 1\n- [ ] Task")

	writeDeltaSpecFile(t, root, "bad-rename", "auth", "spec.md", `## RENAMED Requirements

### Requirement: Nonexistent → Phantom
`)

	result, err := ValidateChange(root, "bad-rename")
	if err != nil {
		t.Fatalf("ValidateChange: %v", err)
	}
	if result.Valid {
		t.Fatal("validation should reject RENAMED dangling delta")
	}

	mainData, _ := os.ReadFile(filepath.Join(CanonPath(root), "auth", "spec.md"))
	mainSpec, _ := ParseMainSpec(string(mainData))
	if len(mainSpec.Requirements) != 1 || mainSpec.Requirements[0].Name != "Login" {
		t.Fatal("main spec should be unchanged after rejected archive")
	}
	if _, statErr := os.Stat(ChangePath(root, "bad-rename")); os.IsNotExist(statErr) {
		t.Fatal("change directory should still exist")
	}
}

func TestArchiveCrossDeltaConflictRejected(t *testing.T) {
	root := setupTestProject(t)
	writeMainSpecFile(t, root, "auth", `# auth

### Requirement: Login
The system SHALL authenticate.

#### Scenario: Valid
- **WHEN** valid creds
`)
	writeChangeFile(t, root, "conflict", "proposal.md", "# Proposal")
	writeChangeFile(t, root, "conflict", "design.md", "# Design")
	writeChangeFile(t, root, "conflict", "tasks.md", "## Phase 1\n- [ ] Task")

	writeDeltaSpecFile(t, root, "conflict", "auth", "part1.md", `## MODIFIED Requirements

### Requirement: Login
The system SHALL authenticate via OAuth.

#### Scenario: OAuth
- **WHEN** OAuth token valid
`)
	writeDeltaSpecFile(t, root, "conflict", "auth", "part2.md", `## MODIFIED Requirements

### Requirement: Login
The system SHALL authenticate via SAML.

#### Scenario: SAML
- **WHEN** SAML assertion valid
`)

	result, err := ValidateChange(root, "conflict")
	if err != nil {
		t.Fatalf("ValidateChange: %v", err)
	}
	if !result.Valid {
		t.Fatalf("validate should pass (cross-delta conflicts are a merge concern): %v", result.Errors)
	}

	changeSpecsDir := ChangeSpecsPath(root, "conflict")
	entries, _ := os.ReadDir(changeSpecsDir)
	var deltas []*DeltaSpec
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		capDir := filepath.Join(changeSpecsDir, entry.Name())
		files, _ := os.ReadDir(capDir)
		for _, f := range files {
			if filepath.Ext(f.Name()) != ".md" {
				continue
			}
			data, _ := os.ReadFile(filepath.Join(capDir, f.Name()))
			delta, _ := ParseDeltaSpec(string(data))
			deltas = append(deltas, delta)
		}
	}

	mainData, _ := os.ReadFile(filepath.Join(CanonPath(root), "auth", "spec.md"))
	mainSpec, _ := ParseMainSpec(string(mainData))

	_, mergeErr := MergeDelta(mainSpec, deltas)
	if mergeErr == nil {
		t.Fatal("merge should reject cross-delta MODIFIED conflict")
	}
	if !strings.Contains(mergeErr.Error(), "multiple deltas modify") {
		t.Errorf("error = %q, want mention of multiple deltas modify", mergeErr.Error())
	}

	mainData2, _ := os.ReadFile(filepath.Join(CanonPath(root), "auth", "spec.md"))
	mainSpec2, _ := ParseMainSpec(string(mainData2))
	if mainSpec2.Requirements[0].Content != "The system SHALL authenticate." {
		t.Fatal("main spec should be unchanged after rejected merge")
	}
}

func TestArchiveADDEDDuplicateRejected(t *testing.T) {
	root := setupTestProject(t)
	writeMainSpecFile(t, root, "auth", `# auth

### Requirement: Login
The system SHALL authenticate.

#### Scenario: Valid
- **WHEN** valid creds
`)
	writeChangeFile(t, root, "dup-add", "proposal.md", "# Proposal")
	writeChangeFile(t, root, "dup-add", "design.md", "# Design")
	writeChangeFile(t, root, "dup-add", "tasks.md", "## Phase 1\n- [ ] Task")

	writeDeltaSpecFile(t, root, "dup-add", "auth", "spec.md", `## ADDED Requirements

### Requirement: Login
The system SHALL authenticate differently.

#### Scenario: S1
- **WHEN** triggered
`)

	result, err := ValidateChange(root, "dup-add")
	if err != nil {
		t.Fatalf("ValidateChange: %v", err)
	}
	if result.Valid {
		t.Fatal("validation should reject ADDED duplicate of existing requirement")
	}

	mainData, _ := os.ReadFile(filepath.Join(CanonPath(root), "auth", "spec.md"))
	mainSpec, _ := ParseMainSpec(string(mainData))
	if len(mainSpec.Requirements) != 1 || mainSpec.Requirements[0].Content != "The system SHALL authenticate." {
		t.Fatal("main spec should be unchanged after rejected archive")
	}
	if _, statErr := os.Stat(ChangePath(root, "dup-add")); os.IsNotExist(statErr) {
		t.Fatal("change directory should still exist")
	}
}

func TestArchiveMultiCapabilityAtomicity(t *testing.T) {
	root := setupTestProject(t)

	writeMainSpecFile(t, root, "auth", `# auth

### Requirement: Login
The system SHALL authenticate.

#### Scenario: Valid
- **WHEN** valid creds
`)

	writeMainSpecFile(t, root, "api", `# api

### Requirement: Rate Limit
The system SHALL limit requests.

#### Scenario: Exceeds
- **WHEN** limit exceeded
`)

	writeChangeFile(t, root, "multi", "proposal.md", "# Proposal")
	writeChangeFile(t, root, "multi", "design.md", "# Design")
	writeChangeFile(t, root, "multi", "tasks.md", "## Phase 1\n- [x] Done")

	writeDeltaSpecFile(t, root, "multi", "auth", "spec.md", `## MODIFIED Requirements

### Requirement: Login
The system SHALL authenticate via OAuth.

#### Scenario: OAuth
- **WHEN** OAuth token valid
`)

	writeDeltaSpecFile(t, root, "multi", "api", "part1.md", `## MODIFIED Requirements

### Requirement: Rate Limit
The system SHALL limit to 100.

#### Scenario: Over 100
- **WHEN** 101 requests
`)

	writeDeltaSpecFile(t, root, "multi", "api", "part2.md", `## MODIFIED Requirements

### Requirement: Rate Limit
The system SHALL limit to 200.

#### Scenario: Over 200
- **WHEN** 201 requests
`)

	result, err := ValidateChange(root, "multi")
	if err != nil {
		t.Fatalf("ValidateChange: %v", err)
	}
	if !result.Valid {
		t.Fatalf("validate should pass (cross-delta conflicts are a merge concern): %v", result.Errors)
	}

	_, prepareErr := PrepareArchiveWrites(root, "multi")
	if prepareErr == nil {
		t.Fatal("PrepareArchiveWrites should reject cross-delta conflict in api capability")
	}
	if !strings.Contains(prepareErr.Error(), "multiple deltas modify") {
		t.Errorf("error = %q, want mention of multiple deltas modify", prepareErr.Error())
	}

	authData, _ := os.ReadFile(filepath.Join(CanonPath(root), "auth", "spec.md"))
	authSpec, _ := ParseMainSpec(string(authData))
	if len(authSpec.Requirements) != 1 || authSpec.Requirements[0].Content != "The system SHALL authenticate." {
		t.Fatal("auth spec should be unchanged — no writes should have occurred")
	}

	apiData, _ := os.ReadFile(filepath.Join(CanonPath(root), "api", "spec.md"))
	apiSpec, _ := ParseMainSpec(string(apiData))
	if len(apiSpec.Requirements) != 1 || apiSpec.Requirements[0].Content != "The system SHALL limit requests." {
		t.Fatal("api spec should be unchanged — no writes should have occurred")
	}

	if _, statErr := os.Stat(ChangePath(root, "multi")); os.IsNotExist(statErr) {
		t.Fatal("change directory should still exist")
	}
}

func TestArchiveMultiCapabilityHappyPath(t *testing.T) {
	root := setupTestProject(t)

	writeMainSpecFile(t, root, "auth", `# auth

### Requirement: Login
The system SHALL authenticate.

#### Scenario: Valid
- **WHEN** valid creds
`)

	writeChangeFile(t, root, "multi-ok", "proposal.md", "# Proposal")
	writeChangeFile(t, root, "multi-ok", "design.md", "# Design")
	writeChangeFile(t, root, "multi-ok", "tasks.md", "## Phase 1\n- [x] Done")

	writeDeltaSpecFile(t, root, "multi-ok", "auth", "spec.md", `## MODIFIED Requirements

### Requirement: Login
The system SHALL authenticate via SSO.

#### Scenario: SSO
- **WHEN** SSO token valid
`)

	writeDeltaSpecFile(t, root, "multi-ok", "rate-limit", "spec.md", `# rate-limit

## ADDED Requirements

### Requirement: Throttle
The system SHALL throttle requests.

#### Scenario: Over limit
- **WHEN** limit exceeded
`)

	runArchivePipeline(t, root, "multi-ok")

	authData, err := os.ReadFile(filepath.Join(CanonPath(root), "auth", "spec.md"))
	if err != nil {
		t.Fatalf("read auth spec: %v", err)
	}
	authSpec, err := ParseMainSpec(string(authData))
	if err != nil {
		t.Fatalf("parse auth spec: %v", err)
	}
	if len(authSpec.Requirements) != 1 {
		t.Fatalf("auth requirements = %d, want 1", len(authSpec.Requirements))
	}
	if authSpec.Requirements[0].Content != "The system SHALL authenticate via SSO." {
		t.Errorf("auth content = %q, want SSO updated", authSpec.Requirements[0].Content)
	}

	rlData, err := os.ReadFile(filepath.Join(CanonPath(root), "rate-limit", "spec.md"))
	if err != nil {
		t.Fatalf("read rate-limit spec: %v", err)
	}
	rlSpec, err := ParseMainSpec(string(rlData))
	if err != nil {
		t.Fatalf("parse rate-limit spec: %v", err)
	}
	if rlSpec.Capability != "rate-limit" {
		t.Errorf("rate-limit capability = %q, want %q", rlSpec.Capability, "rate-limit")
	}
	if len(rlSpec.Requirements) != 1 || rlSpec.Requirements[0].Name != "Throttle" {
		t.Errorf("rate-limit requirements unexpected: %+v", rlSpec.Requirements)
	}

	if _, err := os.Stat(ChangePath(root, "multi-ok")); !os.IsNotExist(err) {
		t.Fatal("change directory should be gone after archive")
	}
	archiveEntries, _ := os.ReadDir(ArchivePath(root))
	found := false
	for _, e := range archiveEntries {
		if strings.HasSuffix(e.Name(), "-multi-ok") {
			found = true
		}
	}
	if !found {
		t.Error("archived directory not found in archive")
	}
}

func TestPrepareArchiveWritesNoDeltas(t *testing.T) {
	root := setupTestProject(t)

	writeChangeFile(t, root, "no-deltas", "proposal.md", "# Proposal")
	writeChangeFile(t, root, "no-deltas", "design.md", "# Design")
	writeChangeFile(t, root, "no-deltas", "tasks.md", "## Phase 1\n- [x] Done")

	if err := os.MkdirAll(ChangeSpecsPath(root, "no-deltas"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	writes, err := PrepareArchiveWrites(root, "no-deltas")
	if err != nil {
		t.Fatalf("PrepareArchiveWrites: %v", err)
	}
	if len(writes) != 0 {
		t.Fatalf("writes = %d, want 0", len(writes))
	}

	if err := ArchiveChange(root, "no-deltas"); err != nil {
		t.Fatalf("ArchiveChange: %v", err)
	}
	if _, err := os.Stat(ChangePath(root, "no-deltas")); !os.IsNotExist(err) {
		t.Fatal("change directory should be gone after archive")
	}
}

func TestWritePendingSpecsCreatesDirectories(t *testing.T) {
	root := setupTestProject(t)

	writes := []PendingWrite{
		{
			Capability: "brand-new",
			Path:       filepath.Join(CanonPath(root), "brand-new", "spec.md"),
			Dir:        filepath.Join(CanonPath(root), "brand-new"),
			Content:    "# brand-new\n\n### Requirement: Core\nThe system SHALL work.\n",
		},
	}

	if err := WritePendingSpecs(writes); err != nil {
		t.Fatalf("WritePendingSpecs: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(CanonPath(root), "brand-new", "spec.md"))
	if err != nil {
		t.Fatalf("read written spec: %v", err)
	}
	spec, err := ParseMainSpec(string(data))
	if err != nil {
		t.Fatalf("ParseMainSpec: %v", err)
	}
	if spec.Capability != "brand-new" {
		t.Errorf("Capability = %q, want %q", spec.Capability, "brand-new")
	}
	if len(spec.Requirements) != 1 || spec.Requirements[0].Name != "Core" {
		t.Errorf("requirements unexpected: %+v", spec.Requirements)
	}
}
