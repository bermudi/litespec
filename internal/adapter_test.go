package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetAdapter_KnownID(t *testing.T) {
	adapter := GetAdapter("claude")
	if adapter == nil {
		t.Fatal("expected adapter for claude, got nil")
	}
	if adapter.ID != "claude" {
		t.Errorf("expected ID=claude, got %s", adapter.ID)
	}
}

func TestGetAdapter_UnknownID(t *testing.T) {
	adapter := GetAdapter("nonexistent")
	if adapter != nil {
		t.Error("expected nil for unknown tool ID")
	}
}

func TestGetAdapter_EmptyID(t *testing.T) {
	adapter := GetAdapter("")
	if adapter != nil {
		t.Error("expected nil for empty tool ID")
	}
}

func TestGenerateAdapterCommands_UnknownToolID(t *testing.T) {
	err := GenerateAdapterCommands(t.TempDir(), []string{"bogus-tool"})
	if err == nil {
		t.Fatal("expected error for unknown tool ID")
	}
	if !strings.Contains(err.Error(), "unknown tool") {
		t.Errorf("expected 'unknown tool' in error, got: %v", err)
	}
}

func TestGenerateAdapterCommands_ClaudeCreatesSymlinks(t *testing.T) {
	root := t.TempDir()

	skillsDir := filepath.Join(root, SkillsDir)
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, si := range Skills {
		if err := os.WriteFile(filepath.Join(skillsDir, si.Name), []byte("template content"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	err := GenerateAdapterCommands(root, []string{"claude"})
	if err != nil {
		t.Fatalf("GenerateAdapterCommands: %v", err)
	}

	claudeDir := filepath.Join(root, ".claude", "skills")
	entries, err := os.ReadDir(claudeDir)
	if err != nil {
		t.Fatalf("reading claude skills dir: %v", err)
	}

	if len(entries) != len(Skills) {
		t.Errorf("expected %d symlinks, got %d", len(Skills), len(entries))
	}

	for _, entry := range entries {
		linkPath := filepath.Join(claudeDir, entry.Name())
		target, err := os.Readlink(linkPath)
		if err != nil {
			t.Errorf("expected %s to be a symlink: %v", entry.Name(), err)
		}
		if !strings.Contains(target, ".agents/skills") {
			t.Errorf("expected symlink target to contain .agents/skills, got %s", target)
		}
	}
}

func TestValidToolIDs(t *testing.T) {
	ids := ValidToolIDs()
	found := false
	for _, id := range ids {
		if id == "claude" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected claude in valid tool IDs, got %v", ids)
	}
}
