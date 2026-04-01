package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bermudi/litespec/internal/skill"
)

func setupChangeWithDeps(t *testing.T, name string, artifacts map[string]string) string {
	t.Helper()
	root := t.TempDir()
	specsDir := filepath.Join(root, "specs", "specs")
	changesDir := filepath.Join(root, "specs", "changes")
	if err := os.MkdirAll(specsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(changesDir, 0o755); err != nil {
		t.Fatal(err)
	}

	changeDir := filepath.Join(changesDir, name)
	if err := os.MkdirAll(changeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	metaPath := filepath.Join(changeDir, ".litespec.yaml")
	if err := os.WriteFile(metaPath, []byte("schema: spec-driven\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	for id, content := range artifacts {
		switch id {
		case "proposal":
			if err := os.WriteFile(filepath.Join(changeDir, "proposal.md"), []byte(content), 0o644); err != nil {
				t.Fatal(err)
			}
		case "specs":
			specSubdir := filepath.Join(changeDir, "specs", "test-cap")
			if err := os.MkdirAll(specSubdir, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(specSubdir, "spec.md"), []byte(content), 0o644); err != nil {
				t.Fatal(err)
			}
		case "design":
			if err := os.WriteFile(filepath.Join(changeDir, "design.md"), []byte(content), 0o644); err != nil {
				t.Fatal(err)
			}
		case "tasks":
			if err := os.WriteFile(filepath.Join(changeDir, "tasks.md"), []byte(content), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}

	return root
}

func TestArtifactInstructionID(t *testing.T) {
	tests := []struct {
		artifactID string
		want       string
	}{
		{"proposal", "artifact-proposal"},
		{"specs", "artifact-specs"},
		{"design", "artifact-design"},
		{"tasks", "artifact-tasks"},
		{"unknown", "artifact-proposal"},
		{"", "artifact-proposal"},
	}

	for _, tc := range tests {
		t.Run(tc.artifactID, func(t *testing.T) {
			got := ArtifactInstructionID(tc.artifactID)
			if got != tc.want {
				t.Errorf("ArtifactInstructionID(%q) = %q, want %q", tc.artifactID, got, tc.want)
			}
		})
	}
}

func TestArtifactInstructionID_RegistersSkillTemplates(t *testing.T) {
	expected := map[string]string{
		"artifact-proposal": "Motivation",
		"artifact-specs":    "ADDED Requirements",
		"artifact-design":   "Architecture",
		"artifact-tasks":    "Phase 1",
	}

	for id, marker := range expected {
		t.Run(id, func(t *testing.T) {
			tmpl := skill.Get(id)
			if tmpl == "" {
				t.Fatalf("skill.Get(%q) returned empty string", id)
			}
			if !strings.Contains(tmpl, marker) {
				t.Errorf("template %q does not contain marker %q", id, marker)
			}
		})
	}
}

func TestArtifactInstructionID_AllTemplatesDistinct(t *testing.T) {
	ids := []string{"artifact-proposal", "artifact-specs", "artifact-design", "artifact-tasks"}
	seen := map[string]string{}
	for _, id := range ids {
		tmpl := skill.Get(id)
		if prev, ok := seen[tmpl]; ok {
			t.Errorf("templates %q and %q are identical", prev, id)
		}
		seen[tmpl] = id
	}
}

func TestBuildArtifactInstructionsJSON_InstructionDiffersFromTemplate(t *testing.T) {
	root := setupChangeWithDeps(t, "test-change", nil)

	instr, err := BuildArtifactInstructionsJSON(root, "test-change", "proposal")
	if err != nil {
		t.Fatalf("BuildArtifactInstructionsJSON: %v", err)
	}

	if instr.Instruction == "" {
		t.Error("Instruction is empty")
	}
	if instr.Template == "" {
		t.Error("Template is empty")
	}
	if instr.Instruction == instr.Template {
		t.Error("Instruction and Template are identical — artifact-specific instructions not working")
	}
}

func TestBuildArtifactInstructionsJSON_InstructionPerArtifact(t *testing.T) {
	root := setupChangeWithDeps(t, "test-change", nil)

	instructions := map[string]string{}
	for _, id := range []string{"proposal", "specs", "design", "tasks"} {
		instr, err := BuildArtifactInstructionsJSON(root, "test-change", id)
		if err != nil {
			t.Fatalf("BuildArtifactInstructionsJSON(%q): %v", id, err)
		}
		instructions[id] = instr.Instruction
	}

	for _, id := range []string{"proposal", "specs", "design", "tasks"} {
		if instructions[id] == "" {
			t.Errorf("artifact %q has empty instruction", id)
		}
	}

	pairs := [][2]string{
		{"proposal", "specs"},
		{"proposal", "design"},
		{"proposal", "tasks"},
		{"specs", "design"},
		{"specs", "tasks"},
		{"design", "tasks"},
	}
	for _, p := range pairs {
		if instructions[p[0]] == instructions[p[1]] {
			t.Errorf("artifacts %q and %q have identical instructions", p[0], p[1])
		}
	}
}

func TestBuildArtifactInstructionsJSON_TemplateIsAlwaysPropose(t *testing.T) {
	root := setupChangeWithDeps(t, "test-change", nil)

	for _, id := range []string{"proposal", "specs", "design", "tasks"} {
		instr, err := BuildArtifactInstructionsJSON(root, "test-change", id)
		if err != nil {
			t.Fatalf("BuildArtifactInstructionsJSON(%q): %v", id, err)
		}
		if instr.Template != skill.Get("propose") {
			t.Errorf("artifact %q: Template does not match propose skill", id)
		}
	}
}

func TestBuildArtifactInstructionsJSON_InstructionContainsArtifactMarkers(t *testing.T) {
	root := setupChangeWithDeps(t, "test-change", nil)

	expectedMarkers := map[string]string{
		"proposal": "Motivation",
		"specs":    "ADDED Requirements",
		"design":   "Architecture",
		"tasks":    "Phase 1",
	}

	for id, marker := range expectedMarkers {
		instr, err := BuildArtifactInstructionsJSON(root, "test-change", id)
		if err != nil {
			t.Fatalf("BuildArtifactInstructionsJSON(%q): %v", id, err)
		}
		if !strings.Contains(instr.Instruction, marker) {
			t.Errorf("artifact %q instruction missing marker %q", id, marker)
		}
	}
}

func TestBuildArtifactInstructionsJSON_Dependencies(t *testing.T) {
	root := setupChangeWithDeps(t, "test-change", map[string]string{
		"proposal": "# Test Proposal",
	})

	instr, err := BuildArtifactInstructionsJSON(root, "test-change", "design")
	if err != nil {
		t.Fatalf("BuildArtifactInstructionsJSON: %v", err)
	}

	if len(instr.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(instr.Dependencies))
	}
	if instr.Dependencies[0].ID != "proposal" {
		t.Errorf("dependency ID = %q, want %q", instr.Dependencies[0].ID, "proposal")
	}
	if !instr.Dependencies[0].Done {
		t.Error("proposal dependency should be done")
	}
}

func TestBuildArtifactInstructionsJSON_Unlocks(t *testing.T) {
	root := setupChangeWithDeps(t, "test-change", nil)

	instr, err := BuildArtifactInstructionsJSON(root, "test-change", "proposal")
	if err != nil {
		t.Fatalf("BuildArtifactInstructionsJSON: %v", err)
	}

	expectedUnlocks := map[string]bool{"specs": false, "design": false, "tasks": false}
	for _, u := range instr.Unlocks {
		expectedUnlocks[u] = true
	}
	for u, found := range expectedUnlocks {
		if !found {
			t.Errorf("proposal should unlock %q", u)
		}
	}
}

func TestBuildArtifactInstructionsJSON_UnknownArtifact(t *testing.T) {
	root := setupChangeWithDeps(t, "test-change", nil)

	_, err := BuildArtifactInstructionsJSON(root, "test-change", "nonexistent")
	if err == nil {
		t.Error("expected error for unknown artifact")
	}
}
