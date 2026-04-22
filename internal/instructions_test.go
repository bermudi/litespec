package internal

import (
	"strings"
	"testing"

	"github.com/bermudi/litespec/internal/skill"
)

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

func TestBuildArtifactInstructionsStandaloneJSON_InstructionNotEmpty(t *testing.T) {
	instr, err := BuildArtifactInstructionsStandaloneJSON("proposal")
	if err != nil {
		t.Fatalf("BuildArtifactInstructionsStandaloneJSON: %v", err)
	}

	if instr.Instruction == "" {
		t.Error("Instruction is empty")
	}
}

func TestBuildArtifactInstructionsStandaloneJSON_InstructionPerArtifact(t *testing.T) {
	instructions := map[string]string{}
	for _, id := range []string{"proposal", "specs", "design", "tasks"} {
		instr, err := BuildArtifactInstructionsStandaloneJSON(id)
		if err != nil {
			t.Fatalf("BuildArtifactInstructionsStandaloneJSON(%q): %v", id, err)
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


func TestBuildArtifactInstructionsStandaloneJSON_InstructionContainsArtifactMarkers(t *testing.T) {
	expectedMarkers := map[string]string{
		"proposal": "Motivation",
		"specs":    "ADDED Requirements",
		"design":   "Architecture",
		"tasks":    "Phase 1",
	}

	for id, marker := range expectedMarkers {
		instr, err := BuildArtifactInstructionsStandaloneJSON(id)
		if err != nil {
			t.Fatalf("BuildArtifactInstructionsStandaloneJSON(%q): %v", id, err)
		}
		if !strings.Contains(instr.Instruction, marker) {
			t.Errorf("artifact %q instruction missing marker %q", id, marker)
		}
	}
}

func TestBuildArtifactInstructionsStandaloneJSON_Fields(t *testing.T) {
	instr, err := BuildArtifactInstructionsStandaloneJSON("design")
	if err != nil {
		t.Fatalf("BuildArtifactInstructionsStandaloneJSON: %v", err)
	}

	if instr.ArtifactID != "design" {
		t.Errorf("ArtifactID = %q, want %q", instr.ArtifactID, "design")
	}
	if instr.OutputPath != "design.md" {
		t.Errorf("OutputPath = %q, want %q", instr.OutputPath, "design.md")
	}
	if instr.Description == "" {
		t.Error("Description is empty")
	}
}

func TestBuildArtifactInstructionsStandaloneJSON_UnknownArtifact(t *testing.T) {
	_, err := BuildArtifactInstructionsStandaloneJSON("nonexistent")
	if err == nil {
		t.Error("expected error for unknown artifact")
	}
}
