package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCmdImportDryRun(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	specsDir := filepath.Join(srcDir, "openspec", "specs", "test-capability")
	os.MkdirAll(specsDir, 0755)
	os.WriteFile(filepath.Join(specsDir, "spec.md"), []byte("# test-capability Specification\n\n## Purpose\n"), 0644)

	changesDir := filepath.Join(srcDir, "openspec", "changes", "my-change")
	os.MkdirAll(changesDir, 0755)
	os.WriteFile(filepath.Join(changesDir, ".openspec.yaml"), []byte("schema: spec-driven\ncreated: 2026-03-01\n"), 0644)
	os.WriteFile(filepath.Join(changesDir, "proposal.md"), []byte("## Motivation\n"), 0644)

	origWd, _ := os.Getwd()
	os.Chdir(dstDir)
	defer os.Chdir(origWd)

	args := []string{"--dry-run", "--source", srcDir}
	err := cmdImport(args)
	if err != nil {
		t.Errorf("cmdImport dry-run failed: %v", err)
	}

	canonPath := filepath.Join(dstDir, "specs", "canon", "test-capability", "spec.md")
	if _, err := os.Stat(canonPath); err == nil {
		t.Error("dry-run should not create files")
	}
}

func TestCmdImportBasic(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	specsDir := filepath.Join(srcDir, "openspec", "specs", "cli-init")
	os.MkdirAll(specsDir, 0755)
	os.WriteFile(filepath.Join(specsDir, "spec.md"), []byte("# cli-init Specification\n\n## Purpose\n"), 0644)

	origWd, _ := os.Getwd()
	os.Chdir(dstDir)
	defer os.Chdir(origWd)

	args := []string{"--source", srcDir}
	err := cmdImport(args)
	if err != nil {
		t.Fatalf("cmdImport failed: %v", err)
	}

	canonPath := filepath.Join(dstDir, "specs", "canon", "cli-init", "spec.md")
	data, err := os.ReadFile(canonPath)
	if err != nil {
		t.Fatalf("read imported spec: %v", err)
	}

	if string(data) != "# cli-init\n\n## Purpose\n" {
		t.Errorf("H1 normalization failed: got %q", string(data))
	}
}

func TestCmdImportNoOpenSpec(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	origWd, _ := os.Getwd()
	os.Chdir(dstDir)
	defer os.Chdir(origWd)

	args := []string{"--source", srcDir}
	err := cmdImport(args)
	if err == nil {
		t.Error("expected error for non-OpenSpec project")
	}
}

func TestCmdImportForceFlag(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	specsDir := filepath.Join(srcDir, "openspec", "specs", "test-spec")
	os.MkdirAll(specsDir, 0755)
	os.WriteFile(filepath.Join(specsDir, "spec.md"), []byte("# test-spec\n\n## Purpose\n"), 0644)

	existingCanon := filepath.Join(dstDir, "specs", "canon", "existing-spec")
	os.MkdirAll(existingCanon, 0755)
	os.WriteFile(filepath.Join(existingCanon, "spec.md"), []byte("# existing-spec\n"), 0644)

	origWd, _ := os.Getwd()
	os.Chdir(dstDir)
	defer os.Chdir(origWd)

	args := []string{"--source", srcDir}
	err := cmdImport(args)
	if err == nil {
		t.Error("expected error for conflicts without --force")
	}

	args = []string{"--source", srcDir, "--force"}
	err = cmdImport(args)
	if err != nil {
		t.Errorf("import with --force failed: %v", err)
	}
}

func TestCmdImportWithArchive(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	changeDir := filepath.Join(srcDir, "openspec", "changes", "test-change")
	os.MkdirAll(changeDir, 0755)
	os.WriteFile(filepath.Join(changeDir, ".openspec.yaml"), []byte("schema: spec-driven\ncreated: 2026-03-01\n"), 0644)

	archiveDir := filepath.Join(srcDir, "openspec", "changes", "archive", "2025-01-11-old-change")
	os.MkdirAll(archiveDir, 0755)
	os.WriteFile(filepath.Join(archiveDir, "proposal.md"), []byte("## Motivation\n"), 0644)

	origWd, _ := os.Getwd()
	os.Chdir(dstDir)
	defer os.Chdir(origWd)

	args := []string{"--source", srcDir}
	err := cmdImport(args)
	if err != nil {
		t.Fatalf("cmdImport failed: %v", err)
	}

	archiveMeta := filepath.Join(dstDir, "specs", "changes", "archive", "2025-01-11-old-change", ".litespec.yaml")
	if _, err := os.Stat(archiveMeta); os.IsNotExist(err) {
		t.Error("synthesized metadata not created for archive")
	}
}
