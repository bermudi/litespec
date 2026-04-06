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

func TestCleanStaleSymlinks_RemovesStale(t *testing.T) {
	dir := t.TempDir()
	os.Symlink("/nonexistent/stale-link", filepath.Join(dir, "litespec-old-skill"))
	os.WriteFile(filepath.Join(dir, "regular-file.txt"), []byte("data"), 0o644)

	err := cleanStaleSymlinks(dir)
	if err != nil {
		t.Fatalf("cleanStaleSymlinks: %v", err)
	}

	if _, err := os.Lstat(filepath.Join(dir, "litespec-old-skill")); !os.IsNotExist(err) {
		t.Error("expected stale symlink to be removed")
	}
	if _, err := os.Lstat(filepath.Join(dir, "regular-file.txt")); err != nil {
		t.Error("expected regular file to be preserved")
	}
}

func TestCleanStaleSymlinks_PreservesValid(t *testing.T) {
	dir := t.TempDir()
	for _, si := range Skills {
		os.Symlink("/nonexistent", filepath.Join(dir, si.Name))
	}

	err := cleanStaleSymlinks(dir)
	if err != nil {
		t.Fatalf("cleanStaleSymlinks: %v", err)
	}

	entries, _ := os.ReadDir(dir)
	if len(entries) != len(Skills) {
		t.Errorf("expected %d entries, got %d", len(Skills), len(entries))
	}
}

func TestCleanStaleSymlinks_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	if err := cleanStaleSymlinks(dir); err != nil {
		t.Fatalf("cleanStaleSymlinks on empty dir: %v", err)
	}
}

func TestCleanStaleSymlinks_SkipsNonSymlinks(t *testing.T) {
	dir := t.TempDir()
	os.Mkdir(filepath.Join(dir, "subdir"), 0o755)
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("data"), 0o644)
	os.Symlink("/nonexistent/stale", filepath.Join(dir, "stale-link"))

	if err := cleanStaleSymlinks(dir); err != nil {
		t.Fatalf("cleanStaleSymlinks: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "subdir")); err != nil {
		t.Error("expected subdir to be preserved")
	}
	if _, err := os.Stat(filepath.Join(dir, "file.txt")); err != nil {
		t.Error("expected regular file to be preserved")
	}
	if _, err := os.Lstat(filepath.Join(dir, "stale-link")); !os.IsNotExist(err) {
		t.Error("expected stale symlink to be removed")
	}
}

func TestGenerateAdapterCommands_CleansStaleSymlinks(t *testing.T) {
	root := t.TempDir()

	skillsDir := filepath.Join(root, SkillsDir)
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, si := range Skills {
		if err := os.WriteFile(filepath.Join(skillsDir, si.Name), []byte("template"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	claudeDir := filepath.Join(root, ".claude", "skills")
	os.MkdirAll(claudeDir, 0o755)
	os.Symlink("/nonexistent/stale", filepath.Join(claudeDir, "litespec-archive"))
	os.Symlink("/nonexistent/stale", filepath.Join(claudeDir, "litespec-continue"))
	os.WriteFile(filepath.Join(claudeDir, "user-notes.txt"), []byte("keep me"), 0o644)

	err := GenerateAdapterCommands(root, []string{"claude"})
	if err != nil {
		t.Fatalf("GenerateAdapterCommands: %v", err)
	}

	if _, err := os.Lstat(filepath.Join(claudeDir, "litespec-archive")); !os.IsNotExist(err) {
		t.Error("expected stale litespec-archive symlink to be removed")
	}
	if _, err := os.Lstat(filepath.Join(claudeDir, "litespec-continue")); !os.IsNotExist(err) {
		t.Error("expected stale litespec-continue symlink to be removed")
	}
	if _, err := os.Stat(filepath.Join(claudeDir, "user-notes.txt")); err != nil {
		t.Error("expected regular file to be preserved")
	}

	entries, _ := os.ReadDir(claudeDir)
	names := make(map[string]bool)
	for _, e := range entries {
		names[e.Name()] = true
	}
	for _, si := range Skills {
		if !names[si.Name] {
			t.Errorf("expected symlink %s to exist", si.Name)
		}
	}
	if len(entries) != len(Skills)+1 {
		t.Errorf("expected %d entries (%d skills + user-notes.txt), got %d", len(Skills)+1, len(Skills), len(entries))
	}
}
