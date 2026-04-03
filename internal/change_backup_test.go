package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRestoreBackups_WithBackup(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file.txt")

	original := []byte("original content")
	if err := os.WriteFile(path, original, 0o644); err != nil {
		t.Fatal(err)
	}

	newContent := []byte("new content")
	if err := os.WriteFile(path, newContent, 0o644); err != nil {
		t.Fatal(err)
	}

	writes := []PendingWrite{
		{Path: path, Backup: original},
	}
	restoreBackups(writes, 1)

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(original) {
		t.Fatalf("expected %q, got %q", original, got)
	}
}

func TestRestoreBackups_NoBackup(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file.txt")

	if err := os.WriteFile(path, []byte("new content"), 0o644); err != nil {
		t.Fatal(err)
	}

	writes := []PendingWrite{
		{Path: path, Backup: nil},
	}
	restoreBackups(writes, 1)

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("expected file to be removed")
	}
}

func TestRestoreBackups_PartialCount(t *testing.T) {
	dir := t.TempDir()

	paths := make([]string, 3)
	originals := make([][]byte, 3)
	for i := range paths {
		paths[i] = filepath.Join(dir, string(rune('a'+i))+".txt")
		originals[i] = []byte(string(rune('a'+i)) + " original")
		if err := os.WriteFile(paths[i], originals[i], 0o644); err != nil {
			t.Fatal(err)
		}
	}

	for i := range paths {
		if err := os.WriteFile(paths[i], []byte("new"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	writes := []PendingWrite{
		{Path: paths[0], Backup: originals[0]},
		{Path: paths[1], Backup: originals[1]},
		{Path: paths[2], Backup: originals[2]},
	}
	restoreBackups(writes, 2)

	for i := 0; i < 2; i++ {
		got, err := os.ReadFile(paths[i])
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(originals[i]) {
			t.Fatalf("file %d: expected %q, got %q", i, originals[i], got)
		}
	}

	got2, err := os.ReadFile(paths[2])
	if err != nil {
		t.Fatal(err)
	}
	if string(got2) != "new" {
		t.Fatalf("file 2: expected %q, got %q", "new", got2)
	}
}

func TestRestoreBackups_EmptyWrites(t *testing.T) {
	restoreBackups([]PendingWrite{}, 0)
}

func TestCleanupTmps_RemovesFiles(t *testing.T) {
	dir := t.TempDir()
	paths := make([]string, 3)
	for i := range paths {
		paths[i] = filepath.Join(dir, string(rune('a'+i))+".tmp")
		if err := os.WriteFile(paths[i], []byte("tmp"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	cleanupTmps(paths)

	for i, p := range paths {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Fatalf("file %d: expected to be removed", i)
		}
	}
}

func TestCleanupTmps_NonexistentFiles(t *testing.T) {
	cleanupTmps([]string{
		"/nonexistent/path/aaa.tmp",
		"/nonexistent/path/bbb.tmp",
	})
}

func TestCleanupTmps_EmptySlice(t *testing.T) {
	cleanupTmps([]string{})
}
