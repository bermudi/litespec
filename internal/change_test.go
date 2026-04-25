package internal

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestIsPatchMode_True(t *testing.T) {
	root := setupTestProject(t)
	writeChangeFile(t, root, "pchange", ".gitkeep", "")
	meta := ChangeMeta{
		Schema:  "spec-driven",
		Created: time.Now().UTC().Truncate(time.Second),
		Mode:    "patch",
	}
	writeChangeMeta(t, root, "pchange", meta)

	if !IsPatchMode(root, "pchange") {
		t.Error("expected IsPatchMode=true for change with mode: patch")
	}
}

func TestIsPatchMode_FalseWithoutMode(t *testing.T) {
	root := setupTestProject(t)
	writeChangeFile(t, root, "fchange", ".gitkeep", "")
	meta := ChangeMeta{
		Schema:  "spec-driven",
		Created: time.Now().UTC().Truncate(time.Second),
	}
	writeChangeMeta(t, root, "fchange", meta)

	if IsPatchMode(root, "fchange") {
		t.Error("expected IsPatchMode=false for change without mode field")
	}
}

func TestIsPatchMode_FalseNoMetaFile(t *testing.T) {
	root := setupTestProject(t)
	changeDir := ChangePath(root, "nometa")
	if err := os.MkdirAll(changeDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	if IsPatchMode(root, "nometa") {
		t.Error("expected IsPatchMode=false when no .litespec.yaml exists")
	}
}

func TestIsPatchMode_FalseEmptyMode(t *testing.T) {
	root := setupTestProject(t)
	changeDir := ChangePath(root, "emptymode")
	if err := os.MkdirAll(changeDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	metaPath := filepath.Join(changeDir, MetaFileName)
	content := "schema: spec-driven\ncreated: 2026-01-01T00:00:00Z\nmode: \"\"\n"
	if err := os.WriteFile(metaPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	if IsPatchMode(root, "emptymode") {
		t.Error("expected IsPatchMode=false for empty mode string")
	}
}
