package internal

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestArtifactExists_FileArtifactExists(t *testing.T) {
	root := setupTestProject(t)
	writeChangeFile(t, root, "mychange", "proposal.md", "# Proposal")

	art := ArtifactInfo{ID: "proposal", Filename: "proposal.md"}
	if !artifactExists(root, "mychange", art) {
		t.Error("expected file artifact to exist")
	}
}

func TestArtifactExists_FileArtifactMissing(t *testing.T) {
	root := setupTestProject(t)
	writeChangeFile(t, root, "mychange", "design.md", "# Design")

	art := ArtifactInfo{ID: "proposal", Filename: "proposal.md"}
	if artifactExists(root, "mychange", art) {
		t.Error("expected missing file artifact to not exist")
	}
}

func TestArtifactExists_SpecsWithMarkdownFiles(t *testing.T) {
	root := setupTestProject(t)
	deltaContent := `## ADDED Requirements

### Requirement: R1
The system SHALL work.

#### Scenario: S1
- **WHEN** triggered
- **THEN** expected result
`
	writeDeltaSpecFile(t, root, "mychange", "auth", "spec.md", deltaContent)

	art := ArtifactInfo{ID: "specs", Filename: "specs"}
	if !artifactExists(root, "mychange", art) {
		t.Error("expected specs artifact to exist when subdirs have .md files")
	}
}

func TestArtifactExists_SpecsEmptySubdirs(t *testing.T) {
	root := setupTestProject(t)
	specsDir := filepath.Join(ChangePath(root, "mychange"), "specs", "auth")
	if err := os.MkdirAll(specsDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	art := ArtifactInfo{ID: "specs", Filename: "specs"}
	if artifactExists(root, "mychange", art) {
		t.Error("expected specs artifact to not exist with empty subdirs")
	}
}

func TestHasMarkdownFiles_DirWithMarkdown(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte("content"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	if !hasMarkdownFiles(dir) {
		t.Error("expected true for dir with .md file")
	}
}

func TestHasMarkdownFiles_DirWithNoMarkdown(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "data.txt"), []byte("content"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	if hasMarkdownFiles(dir) {
		t.Error("expected false for dir without .md files")
	}
}

func TestHasMarkdownFiles_NonexistentDir(t *testing.T) {
	if hasMarkdownFiles("/nonexistent/path/xyz") {
		t.Error("expected false for nonexistent dir")
	}
}

func TestLoadArtifactStates_AllBlockedExceptProposal(t *testing.T) {
	root := setupTestProject(t)
	writeChangeFile(t, root, "mychange", "placeholder.txt", "")

	states, err := LoadArtifactStates(root, "mychange")
	if err != nil {
		t.Fatalf("LoadArtifactStates: %v", err)
	}

	if states["proposal"] != ArtifactReady {
		t.Errorf("expected proposal=READY (no prereqs), got %s", states["proposal"])
	}
	for _, art := range Artifacts {
		if art.ID == "proposal" {
			continue
		}
		if states[art.ID] != ArtifactBlocked {
			t.Errorf("expected %s to be BLOCKED, got %s", art.ID, states[art.ID])
		}
	}
}

func TestLoadArtifactStates_OnlyProposalExists(t *testing.T) {
	root := setupTestProject(t)
	writeChangeFile(t, root, "mychange", "proposal.md", "# Proposal")

	states, err := LoadArtifactStates(root, "mychange")
	if err != nil {
		t.Fatalf("LoadArtifactStates: %v", err)
	}

	if states["proposal"] != ArtifactDone {
		t.Errorf("expected proposal=DONE, got %s", states["proposal"])
	}
	if states["specs"] != ArtifactReady {
		t.Errorf("expected specs=READY, got %s", states["specs"])
	}
	if states["design"] != ArtifactReady {
		t.Errorf("expected design=READY, got %s", states["design"])
	}
	if states["tasks"] != ArtifactBlocked {
		t.Errorf("expected tasks=BLOCKED, got %s", states["tasks"])
	}
}

func TestLoadArtifactStates_ProposalSpecsDesignExist(t *testing.T) {
	root := setupTestProject(t)
	writeChangeFile(t, root, "mychange", "proposal.md", "# Proposal")
	writeChangeFile(t, root, "mychange", "design.md", "# Design")
	deltaContent := `## ADDED Requirements

### Requirement: R1
The system SHALL work.

#### Scenario: S1
- **WHEN** triggered
- **THEN** expected result
`
	writeDeltaSpecFile(t, root, "mychange", "cap", "spec.md", deltaContent)

	states, err := LoadArtifactStates(root, "mychange")
	if err != nil {
		t.Fatalf("LoadArtifactStates: %v", err)
	}

	if states["proposal"] != ArtifactDone {
		t.Errorf("expected proposal=DONE, got %s", states["proposal"])
	}
	if states["specs"] != ArtifactDone {
		t.Errorf("expected specs=DONE, got %s", states["specs"])
	}
	if states["design"] != ArtifactDone {
		t.Errorf("expected design=DONE, got %s", states["design"])
	}
	if states["tasks"] != ArtifactReady {
		t.Errorf("expected tasks=READY, got %s", states["tasks"])
	}
}

func TestLoadArtifactStates_AllDone(t *testing.T) {
	root := setupTestProject(t)
	writeChangeFile(t, root, "mychange", "proposal.md", "# Proposal")
	writeChangeFile(t, root, "mychange", "design.md", "# Design")
	writeChangeFile(t, root, "mychange", "tasks.md", "## Phase 1\n- [ ] Task")
	deltaContent := `## ADDED Requirements

### Requirement: R1
The system SHALL work.

#### Scenario: S1
- **WHEN** triggered
- **THEN** expected result
`
	writeDeltaSpecFile(t, root, "mychange", "cap", "spec.md", deltaContent)

	states, err := LoadArtifactStates(root, "mychange")
	if err != nil {
		t.Fatalf("LoadArtifactStates: %v", err)
	}

	for _, art := range Artifacts {
		if states[art.ID] != ArtifactDone {
			t.Errorf("expected %s=DONE, got %s", art.ID, states[art.ID])
		}
	}
}

func TestGetReadyArtifacts_MixedStates(t *testing.T) {
	states := map[string]ArtifactState{
		"proposal": ArtifactDone,
		"specs":    ArtifactReady,
		"design":   ArtifactReady,
		"tasks":    ArtifactBlocked,
	}

	ready := GetReadyArtifacts(states)
	if len(ready) != 2 {
		t.Fatalf("expected 2 ready artifacts, got %d", len(ready))
	}

	found := map[string]bool{}
	for _, id := range ready {
		found[id] = true
	}
	if !found["specs"] || !found["design"] {
		t.Errorf("expected specs and design to be ready, got %v", ready)
	}
}

func TestGetReadyArtifacts_AllBlocked(t *testing.T) {
	states := map[string]ArtifactState{
		"proposal": ArtifactBlocked,
		"specs":    ArtifactBlocked,
		"design":   ArtifactBlocked,
		"tasks":    ArtifactBlocked,
	}

	ready := GetReadyArtifacts(states)
	if len(ready) != 0 {
		t.Errorf("expected 0 ready artifacts, got %d", len(ready))
	}
}

func TestGetReadyArtifacts_AllDone(t *testing.T) {
	states := map[string]ArtifactState{
		"proposal": ArtifactDone,
		"specs":    ArtifactDone,
		"design":   ArtifactDone,
		"tasks":    ArtifactDone,
	}

	ready := GetReadyArtifacts(states)
	if len(ready) != 0 {
		t.Errorf("expected 0 ready artifacts, got %d", len(ready))
	}
}

func TestGetNextArtifact_FirstReadyIsProposal(t *testing.T) {
	states := map[string]ArtifactState{
		"proposal": ArtifactReady,
		"specs":    ArtifactBlocked,
		"design":   ArtifactBlocked,
		"tasks":    ArtifactBlocked,
	}

	next := GetNextArtifact(states)
	if next != "proposal" {
		t.Errorf("expected proposal, got %q", next)
	}
}

func TestGetNextArtifact_AllBlocked(t *testing.T) {
	states := map[string]ArtifactState{
		"proposal": ArtifactBlocked,
		"specs":    ArtifactBlocked,
		"design":   ArtifactBlocked,
		"tasks":    ArtifactBlocked,
	}

	next := GetNextArtifact(states)
	if next != "" {
		t.Errorf("expected empty string, got %q", next)
	}
}

func TestGetNextArtifact_AllDone(t *testing.T) {
	states := map[string]ArtifactState{
		"proposal": ArtifactDone,
		"specs":    ArtifactDone,
		"design":   ArtifactDone,
		"tasks":    ArtifactDone,
	}

	next := GetNextArtifact(states)
	if next != "" {
		t.Errorf("expected empty string, got %q", next)
	}
}

func TestLoadChangeContext_ValidWithMetadata(t *testing.T) {
	root := setupTestProject(t)
	writeChangeFile(t, root, "mychange", "proposal.md", "# Proposal")

	ts := time.Date(2026, 4, 3, 12, 0, 0, 0, time.UTC)
	meta := ChangeMeta{
		Schema:  "spec-driven",
		Created: ts,
	}
	writeChangeMeta(t, root, "mychange", meta)

	ch, err := LoadChangeContext(root, "mychange")
	if err != nil {
		t.Fatalf("LoadChangeContext: %v", err)
	}
	if ch.Name != "mychange" {
		t.Errorf("expected name=mychange, got %q", ch.Name)
	}
	if ch.Schema != "spec-driven" {
		t.Errorf("expected schema=spec-driven, got %q", ch.Schema)
	}
	if !ch.Created.Equal(ts) {
		t.Errorf("expected created=%v, got %v", ts, ch.Created)
	}
	if ch.Artifacts["proposal"] != ArtifactDone {
		t.Errorf("expected proposal=DONE, got %s", ch.Artifacts["proposal"])
	}
}

func TestLoadChangeContext_NonexistentChange(t *testing.T) {
	root := setupTestProject(t)

	_, err := LoadChangeContext(root, "ghost")
	if err == nil {
		t.Fatal("expected error for nonexistent change")
	}
}

func TestLoadChangeContext_CorruptedMetadata(t *testing.T) {
	root := setupTestProject(t)
	changeDir := ChangePath(root, "badchange")
	if err := os.MkdirAll(changeDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	metaPath := filepath.Join(changeDir, MetaFileName)
	if err := os.WriteFile(metaPath, []byte("{{invalid yaml: ["), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	_, err := LoadChangeContext(root, "badchange")
	if err == nil {
		t.Fatal("expected error for corrupted metadata")
	}
}
