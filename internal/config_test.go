package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadProjectConfig_MissingFile(t *testing.T) {
	cfg, err := ReadProjectConfig(t.TempDir())
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if len(cfg.Tools) != 0 {
		t.Errorf("expected empty tools, got %v", cfg.Tools)
	}
}

func TestReadProjectConfig_CorruptYAML(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, ProjectDirName), 0o755)
	os.WriteFile(ConfigPath(root), []byte("tools: [unclosed\n"), 0o644)

	_, err := ReadProjectConfig(root)
	if err == nil {
		t.Fatal("expected error for corrupt YAML")
	}
}

func TestReadProjectConfig_EmptyFile(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, ProjectDirName), 0o755)
	os.WriteFile(ConfigPath(root), []byte(""), 0o644)

	cfg, err := ReadProjectConfig(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Tools) != 0 {
		t.Errorf("expected empty tools, got %v", cfg.Tools)
	}
}

func TestWriteThenReadProjectConfig(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, ProjectDirName), 0o755)

	original := &ProjectConfig{Tools: []string{"claude"}}
	if err := WriteProjectConfig(root, original); err != nil {
		t.Fatalf("WriteProjectConfig: %v", err)
	}

	cfg, err := ReadProjectConfig(root)
	if err != nil {
		t.Fatalf("ReadProjectConfig: %v", err)
	}
	if len(cfg.Tools) != 1 || cfg.Tools[0] != "claude" {
		t.Errorf("expected tools=[claude], got %v", cfg.Tools)
	}
}

func TestWriteProjectConfig_CreatesFile(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, ProjectDirName), 0o755)

	if err := WriteProjectConfig(root, &ProjectConfig{Tools: []string{"claude"}}); err != nil {
		t.Fatalf("WriteProjectConfig: %v", err)
	}
	if _, err := os.Stat(ConfigPath(root)); os.IsNotExist(err) {
		t.Error("expected config file to exist")
	}
}
