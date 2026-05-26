package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func setupBenchProject(b *testing.B, numChanges, capsPerChange, reqsPerCap int) string {
	b.Helper()
	root := b.TempDir()
	for _, dir := range []string{CanonPath(root), ChangesPath(root), ArchivePath(root)} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			b.Fatalf("setup: %v", err)
		}
	}

	for i := 0; i < numChanges; i++ {
		name := fmt.Sprintf("change-%03d", i)
		changeDir := ChangePath(root, name)
		if err := os.MkdirAll(changeDir, 0o755); err != nil {
			b.Fatalf("mkdir change: %v", err)
		}
		meta := ChangeMeta{Schema: "spec-driven", Created: time.Now().UTC().Truncate(time.Second)}
		metaData, _ := yaml.Marshal(&meta)
		metaData = append(metaData, []byte("\n")...)
		if err := os.WriteFile(filepath.Join(changeDir, MetaFileName), metaData, 0o644); err != nil {
			b.Fatalf("write meta: %v", err)
		}

		for _, fname := range []string{"proposal.md", "design.md"} {
			if err := os.WriteFile(filepath.Join(changeDir, fname), []byte("# "+fname), 0o644); err != nil {
				b.Fatalf("write %s: %v", fname, err)
			}
		}

		tasksContent := "## Phase 1\n"
		for j := 0; j < 3; j++ {
			tasksContent += "- [x] Task " + fmt.Sprintf("%d", j) + "\n"
		}
		if err := os.WriteFile(filepath.Join(changeDir, "tasks.md"), []byte(tasksContent), 0o644); err != nil {
			b.Fatalf("write tasks: %v", err)
		}

		for c := 0; c < capsPerChange; c++ {
			capName := fmt.Sprintf("cap-%03d", c)
			capDir := filepath.Join(ChangeSpecsPath(root, name), capName)
			if err := os.MkdirAll(capDir, 0o755); err != nil {
				b.Fatalf("mkdir cap: %v", err)
			}

			var delta strings.Builder
			delta.WriteString("# " + capName + "\n\n## ADDED Requirements\n")
			for r := 0; r < reqsPerCap; r++ {
				delta.WriteString(fmt.Sprintf("### Requirement: Req-%d\nThe system SHALL do thing %d.\n\n#### Scenario: S-%d\n- **WHEN** X\n- **THEN** Y\n\n", r, r, r))
			}
			if err := os.WriteFile(filepath.Join(capDir, "spec.md"), []byte(delta.String()), 0o644); err != nil {
				b.Fatalf("write delta: %v", err)
			}
		}
	}

	return root
}

func BenchmarkValidateAll(b *testing.B) {
	b.Run("1change_2caps_5reqs", func(b *testing.B) {
		root := setupBenchProject(b, 1, 2, 5)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := ValidateAll(root, false); err != nil {
				b.Fatalf("ValidateAll: %v", err)
			}
		}
	})

	b.Run("5changes_3caps_10reqs", func(b *testing.B) {
		root := setupBenchProject(b, 5, 3, 10)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := ValidateAll(root, false); err != nil {
				b.Fatalf("ValidateAll: %v", err)
			}
		}
	})

	b.Run("20changes_2caps_5reqs", func(b *testing.B) {
		root := setupBenchProject(b, 20, 2, 5)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := ValidateAll(root, false); err != nil {
				b.Fatalf("ValidateAll: %v", err)
			}
		}
	})
}

func BenchmarkMergeDelta(b *testing.B) {
	b.Run("add_10reqs", func(b *testing.B) {
		main := &Spec{Capability: "cap", Requirements: []SpecRequirement{}}
		delta := &DeltaSpec{
			Requirements: make([]DeltaRequirement, 10),
		}
		for i := 0; i < 10; i++ {
			delta.Requirements[i] = DeltaRequirement{
				Operation: DeltaAdded,
				Name:      fmt.Sprintf("Req-%d", i),
				Content:   fmt.Sprintf("The system SHALL do thing %d.", i),
				Scenarios: []Scenario{{Name: fmt.Sprintf("S-%d", i), Content: "- **WHEN** X\n- **THEN** Y"}},
			}
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := MergeDelta(main, []*DeltaSpec{delta}); err != nil {
				b.Fatalf("MergeDelta: %v", err)
			}
		}
	})

	b.Run("modify_50reqs", func(b *testing.B) {
		main := &Spec{Capability: "cap", Requirements: make([]SpecRequirement, 50)}
		for i := 0; i < 50; i++ {
			main.Requirements[i] = SpecRequirement{
				Name:    fmt.Sprintf("Req-%d", i),
				Content: fmt.Sprintf("Original content %d SHALL work.", i),
			}
		}
		delta := &DeltaSpec{
			Requirements: make([]DeltaRequirement, 50),
		}
		for i := 0; i < 50; i++ {
			delta.Requirements[i] = DeltaRequirement{
				Operation: DeltaModified,
				Name:      fmt.Sprintf("Req-%d", i),
				Content:   fmt.Sprintf("Updated content %d SHALL work better.", i),
			}
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := MergeDelta(main, []*DeltaSpec{delta}); err != nil {
				b.Fatalf("MergeDelta: %v", err)
			}
		}
	})

	b.Run("mixed_ops_20reqs", func(b *testing.B) {
		main := &Spec{Capability: "cap", Requirements: make([]SpecRequirement, 30)}
		for i := 0; i < 30; i++ {
			main.Requirements[i] = SpecRequirement{
				Name:    fmt.Sprintf("Req-%d", i),
				Content: fmt.Sprintf("Content %d SHALL exist.", i),
			}
		}
		delta := &DeltaSpec{
			Requirements: make([]DeltaRequirement, 20),
		}
		for i := 0; i < 10; i++ {
			delta.Requirements[i] = DeltaRequirement{
				Operation: DeltaRenamed,
				OldName:   fmt.Sprintf("Req-%d", i),
				Name:      fmt.Sprintf("Renamed-%d", i),
			}
		}
		for i := 10; i < 15; i++ {
			delta.Requirements[i] = DeltaRequirement{
				Operation: DeltaRemoved,
				Name:      fmt.Sprintf("Req-%d", i),
			}
		}
		for i := 15; i < 20; i++ {
			delta.Requirements[i] = DeltaRequirement{
				Operation: DeltaModified,
				Name:      fmt.Sprintf("Req-%d", i),
				Content:   fmt.Sprintf("Modified %d SHALL be different.", i),
			}
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := MergeDelta(main, []*DeltaSpec{delta}); err != nil {
				b.Fatalf("MergeDelta: %v", err)
			}
		}
	})
}
