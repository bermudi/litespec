package importer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDetectOpenSpecProject(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(string)
		expected bool
	}{
		{
			name: "detects openspec/specs directory",
			setup: func(dir string) {
				os.MkdirAll(filepath.Join(dir, "openspec", "specs"), 0755)
			},
			expected: true,
		},
		{
			name: "detects openspec/changes directory",
			setup: func(dir string) {
				os.MkdirAll(filepath.Join(dir, "openspec", "changes"), 0755)
			},
			expected: true,
		},
		{
			name: "returns false for non-openspec project",
			setup: func(dir string) {
				os.MkdirAll(filepath.Join(dir, "specs", "canon"), 0755)
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			tt.setup(dir)
			result := DetectOpenSpecProject(dir)
			if result != tt.expected {
				t.Errorf("DetectOpenSpecProject() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestNormalizeH1(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "strips Specification suffix from H1",
			input:    "# cli-init Specification\n\n## Purpose\n...",
			expected: "# cli-init\n\n## Purpose\n...",
		},
		{
			name:     "leaves normal H1 unchanged",
			input:    "# cli-init\n\n## Purpose\n...",
			expected: "# cli-init\n\n## Purpose\n...",
		},
		{
			name:     "handles multi-word capability",
			input:    "# change-creation Specification\n\n## Purpose\n...",
			expected: "# change-creation\n\n## Purpose\n...",
		},
		{
			name:     "no H1 specification suffix",
			input:    "## Purpose\n\nSome content",
			expected: "## Purpose\n\nSome content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeH1(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeH1() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestNormalizeTasksPhases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "converts phase label format",
			input:    "## 1. Metadata Model\n\n- [ ] 1.1 Task one\n- [ ] 1.2 Task two\n\n## 2. Next Phase\n",
			expected: "## Phase 1: Metadata Model\n\n- [ ] 1.1 Task one\n- [ ] 1.2 Task two\n\n## Phase 2: Next Phase\n",
		},
		{
			name:     "leaves already normalized phases unchanged",
			input:    "## Phase 1: Metadata Model\n\n- [ ] 1.1 Task one\n",
			expected: "## Phase 1: Metadata Model\n\n- [ ] 1.1 Task one\n",
		},
		{
			name:     "handles multiple phase labels",
			input:    "## 1. First\n## 2. Second\n## 3. Third\n",
			expected: "## Phase 1: First\n## Phase 2: Second\n## Phase 3: Third\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeTasksPhases(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeTasksPhases() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestParseCreatedTime(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Time
		hasError bool
	}{
		{
			name:     "parses date-only format",
			input:    "2026-02-21",
			expected: time.Date(2026, 2, 21, 0, 0, 0, 0, time.UTC),
			hasError: false,
		},
		{
			name:     "parses quoted date",
			input:    "\"2026-02-21\"",
			expected: time.Date(2026, 2, 21, 0, 0, 0, 0, time.UTC),
			hasError: false,
		},
		{
			name:     "parses single-quoted date",
			input:    "'2026-02-21'",
			expected: time.Date(2026, 2, 21, 0, 0, 0, 0, time.UTC),
			hasError: false,
		},
		{
			name:     "parses RFC3339 format",
			input:    "2026-04-04T00:42:31Z",
			expected: time.Date(2026, 4, 4, 0, 42, 31, 0, time.UTC),
			hasError: false,
		},
		{
			name:     "empty string returns error",
			input:    "",
			expected: time.Time{},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseCreatedTime(tt.input)
			if tt.hasError {
				if err == nil {
					t.Error("parseCreatedTime() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("parseCreatedTime() unexpected error: %v", err)
				}
				if !result.Equal(tt.expected) {
					t.Errorf("parseCreatedTime() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

func TestImportOpenSpecProject(t *testing.T) {
	t.Run("copies canon specs with H1 normalization", func(t *testing.T) {
		src := t.TempDir()
		dst := t.TempDir()

		specsDir := filepath.Join(src, "openspec", "specs", "cli-init")
		os.MkdirAll(specsDir, 0755)
		specContent := "# cli-init Specification\n\n## Purpose\nTest spec\n"
		os.WriteFile(filepath.Join(specsDir, "spec.md"), []byte(specContent), 0644)

		stats, err := ImportOpenSpecProject(src, dst)
		if err != nil {
			t.Fatalf("ImportOpenSpecProject() error: %v", err)
		}

		if stats.CanonSpecs != 1 {
			t.Errorf("CanonSpecs = %d, want 1", stats.CanonSpecs)
		}

		copied, err := os.ReadFile(filepath.Join(dst, "specs", "canon", "cli-init", "spec.md"))
		if err != nil {
			t.Fatalf("read copied spec: %v", err)
		}

		expected := "# cli-init\n\n## Purpose\nTest spec\n"
		if string(copied) != expected {
			t.Errorf("spec content = %q, want %q", string(copied), expected)
		}
	})

	t.Run("migrates active changes with metadata conversion", func(t *testing.T) {
		src := t.TempDir()
		dst := t.TempDir()

		changeDir := filepath.Join(src, "openspec", "changes", "my-change")
		os.MkdirAll(changeDir, 0755)

		metaContent := "schema: spec-driven\ncreated: 2026-02-21\n"
		os.WriteFile(filepath.Join(changeDir, ".openspec.yaml"), []byte(metaContent), 0644)

		proposalContent := "## Motivation\nTest proposal\n"
		os.WriteFile(filepath.Join(changeDir, "proposal.md"), []byte(proposalContent), 0644)

		stats, err := ImportOpenSpecProject(src, dst)
		if err != nil {
			t.Fatalf("ImportOpenSpecProject() error: %v", err)
		}

		if stats.ActiveChanges != 1 {
			t.Errorf("ActiveChanges = %d, want 1", stats.ActiveChanges)
		}

		liteMeta, err := os.ReadFile(filepath.Join(dst, "specs", "changes", "my-change", ".litespec.yaml"))
		if err != nil {
			t.Fatalf("read litespec metadata: %v", err)
		}

		metaStr := string(liteMeta)
		if !strings.Contains(metaStr, "2026-02-21T00:00:00Z") {
			t.Errorf("metadata should contain RFC3339 date, got: %q", metaStr)
		}
		if !strings.Contains(metaStr, "spec-driven") {
			t.Errorf("metadata should preserve schema, got: %q", metaStr)
		}
	})

	t.Run("migrates archives with synthesized metadata", func(t *testing.T) {
		src := t.TempDir()
		dst := t.TempDir()

		archiveDir := filepath.Join(src, "openspec", "changes", "archive", "2025-01-11-my-feature")
		os.MkdirAll(archiveDir, 0755)

		proposalContent := "## Motivation\nArchived change\n"
		os.WriteFile(filepath.Join(archiveDir, "proposal.md"), []byte(proposalContent), 0644)

		stats, err := ImportOpenSpecProject(src, dst)
		if err != nil {
			t.Fatalf("ImportOpenSpecProject() error: %v", err)
		}

		if stats.Archives != 1 {
			t.Errorf("Archives = %d, want 1", stats.Archives)
		}

		liteMeta, err := os.ReadFile(filepath.Join(dst, "specs", "changes", "archive", "2025-01-11-my-feature", ".litespec.yaml"))
		if err != nil {
			t.Fatalf("read synthesized metadata: %v", err)
		}

		if string(liteMeta) == "" {
			t.Error("synthesized metadata is empty")
		}
	})

	t.Run("strips specs/ from archives", func(t *testing.T) {
		src := t.TempDir()
		dst := t.TempDir()

		archiveDir := filepath.Join(src, "openspec", "changes", "archive", "2025-01-11-old-change")
		os.MkdirAll(filepath.Join(archiveDir, "specs", "some-capability"), 0755)
		os.WriteFile(filepath.Join(archiveDir, "proposal.md"), []byte("## Motivation\n"), 0644)

		_, err := ImportOpenSpecProject(src, dst)
		if err != nil {
			t.Fatalf("ImportOpenSpecProject() error: %v", err)
		}

		strippedDir := filepath.Join(dst, "specs", "changes", "archive", "2025-01-11-old-change", "specs")
		if _, err := os.Stat(strippedDir); !os.IsNotExist(err) {
			t.Error("specs/ directory should have been stripped from archive")
		}
	})

	t.Run("normalizes task phase labels", func(t *testing.T) {
		src := t.TempDir()
		dst := t.TempDir()

		changeDir := filepath.Join(src, "openspec", "changes", "my-change")
		os.MkdirAll(changeDir, 0755)
		os.WriteFile(filepath.Join(changeDir, ".openspec.yaml"), []byte("schema: spec-driven\ncreated: 2026-03-01\n"), 0644)

		tasksContent := "## 1. First Phase\n\n- [ ] 1.1 Task one\n\n## 2. Second Phase\n"
		os.WriteFile(filepath.Join(changeDir, "tasks.md"), []byte(tasksContent), 0644)

		_, err := ImportOpenSpecProject(src, dst)
		if err != nil {
			t.Fatalf("ImportOpenSpecProject() error: %v", err)
		}

		copied, err := os.ReadFile(filepath.Join(dst, "specs", "changes", "my-change", "tasks.md"))
		if err != nil {
			t.Fatalf("read tasks: %v", err)
		}

		expected := "## Phase 1: First Phase\n\n- [ ] 1.1 Task one\n\n## Phase 2: Second Phase\n"
		if string(copied) != expected {
			t.Errorf("tasks content = %q, want %q", string(copied), expected)
		}
	})

	t.Run("converts archive metadata when .openspec.yaml exists", func(t *testing.T) {
		src := t.TempDir()
		dst := t.TempDir()

		archiveDir := filepath.Join(src, "openspec", "changes", "archive", "2025-03-15-convertible-feature")
		os.MkdirAll(archiveDir, 0755)
		os.WriteFile(filepath.Join(archiveDir, ".openspec.yaml"),
			[]byte("schema: spec-driven\ncreated: 2025-03-15\n"), 0644)
		os.WriteFile(filepath.Join(archiveDir, "proposal.md"), []byte("## Motivation\n"), 0644)

		stats, err := ImportOpenSpecProject(src, dst)
		if err != nil {
			t.Fatalf("ImportOpenSpecProject() error: %v", err)
		}

		if stats.Archives != 1 {
			t.Errorf("Archives = %d, want 1", stats.Archives)
		}

		liteMeta, err := os.ReadFile(filepath.Join(dst, "specs", "changes", "archive", "2025-03-15-convertible-feature", ".litespec.yaml"))
		if err != nil {
			t.Fatalf("read litespec metadata: %v", err)
		}

		content := string(liteMeta)
		if !strings.Contains(content, "created:") {
			t.Errorf("metadata missing created field: %q", content)
		}
		if !strings.Contains(content, "2025-03-15") {
			t.Errorf("metadata should contain date from original: %q", content)
		}
	})

	t.Run("synthesizes metadata for archive without date prefix", func(t *testing.T) {
		src := t.TempDir()
		dst := t.TempDir()

		archiveDir := filepath.Join(src, "openspec", "changes", "archive", "no-date-prefix-change")
		os.MkdirAll(archiveDir, 0755)
		os.WriteFile(filepath.Join(archiveDir, "proposal.md"), []byte("## Motivation\n"), 0644)

		_, err := ImportOpenSpecProject(src, dst)
		if err != nil {
			t.Fatalf("ImportOpenSpecProject() error: %v", err)
		}

		liteMeta, err := os.ReadFile(filepath.Join(dst, "specs", "changes", "archive", "no-date-prefix-change", ".litespec.yaml"))
		if err != nil {
			t.Fatalf("read synthesized metadata: %v", err)
		}

		content := string(liteMeta)
		if !strings.Contains(content, "spec-driven") {
			t.Errorf("synthesized metadata missing schema: %q", content)
		}
	})

	t.Run("preserves dependsOn through metadata conversion", func(t *testing.T) {
		src := t.TempDir()
		dst := t.TempDir()

		changeDir := filepath.Join(src, "openspec", "changes", "stacked-change")
		os.MkdirAll(changeDir, 0755)
		os.WriteFile(filepath.Join(changeDir, ".openspec.yaml"),
			[]byte("schema: spec-driven\ncreated: 2026-03-01\ndependsOn:\n  - base-change\n"), 0644)
		os.WriteFile(filepath.Join(changeDir, "proposal.md"), []byte("## Motivation\n"), 0644)

		_, err := ImportOpenSpecProject(src, dst)
		if err != nil {
			t.Fatalf("ImportOpenSpecProject() error: %v", err)
		}

		liteMeta, err := os.ReadFile(filepath.Join(dst, "specs", "changes", "stacked-change", ".litespec.yaml"))
		if err != nil {
			t.Fatalf("read litespec metadata: %v", err)
		}

		content := string(liteMeta)
		if !strings.Contains(content, "dependsOn") {
			t.Errorf("metadata missing dependsOn: %q", content)
		}
		if !strings.Contains(content, "base-change") {
			t.Errorf("metadata missing dependsOn value: %q", content)
		}
	})

	t.Run("metadata date is RFC3339 format", func(t *testing.T) {
		src := t.TempDir()
		dst := t.TempDir()

		changeDir := filepath.Join(src, "openspec", "changes", "date-check")
		os.MkdirAll(changeDir, 0755)
		os.WriteFile(filepath.Join(changeDir, ".openspec.yaml"),
			[]byte("schema: spec-driven\ncreated: 2026-02-21\n"), 0644)
		os.WriteFile(filepath.Join(changeDir, "proposal.md"), []byte("## Motivation\n"), 0644)

		_, err := ImportOpenSpecProject(src, dst)
		if err != nil {
			t.Fatalf("ImportOpenSpecProject() error: %v", err)
		}

		liteMeta, err := os.ReadFile(filepath.Join(dst, "specs", "changes", "date-check", ".litespec.yaml"))
		if err != nil {
			t.Fatalf("read litespec metadata: %v", err)
		}

		content := string(liteMeta)
		if !strings.Contains(content, "2026-02-21T00:00:00Z") {
			t.Errorf("metadata should have RFC3339 date, got: %q", content)
		}
	})

	t.Run("normalizes task phases in archived changes", func(t *testing.T) {
		src := t.TempDir()
		dst := t.TempDir()

		archiveDir := filepath.Join(src, "openspec", "changes", "archive", "2025-06-01-old-tasks")
		os.MkdirAll(archiveDir, 0755)
		os.WriteFile(filepath.Join(archiveDir, "tasks.md"),
			[]byte("## 1. Build\n\n- [ ] 1.1 Task\n\n## 2. Test\n"), 0644)

		_, err := ImportOpenSpecProject(src, dst)
		if err != nil {
			t.Fatalf("ImportOpenSpecProject() error: %v", err)
		}

		copied, err := os.ReadFile(filepath.Join(dst, "specs", "changes", "archive", "2025-06-01-old-tasks", "tasks.md"))
		if err != nil {
			t.Fatalf("read tasks: %v", err)
		}

		expected := "## Phase 1: Build\n\n- [ ] 1.1 Task\n\n## Phase 2: Test\n"
		if string(copied) != expected {
			t.Errorf("archive tasks = %q, want %q", string(copied), expected)
		}
	})

	t.Run("warns about IMPLEMENTATION_ORDER.md specifically", func(t *testing.T) {
		src := t.TempDir()
		dst := t.TempDir()

		changeDir := filepath.Join(src, "openspec", "changes")
		os.MkdirAll(changeDir, 0755)
		os.WriteFile(filepath.Join(changeDir, "IMPLEMENTATION_ORDER.md"), []byte("# Order\n"), 0644)

		stats, err := ImportOpenSpecProject(src, dst)
		if err != nil {
			t.Fatalf("ImportOpenSpecProject() error: %v", err)
		}

		found := false
		for _, w := range stats.Warnings {
			if strings.Contains(w, "IMPLEMENTATION_ORDER.md") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected IMPLEMENTATION_ORDER.md warning, got warnings: %v", stats.Warnings)
		}
	})

	t.Run("warns about skipped root-level items", func(t *testing.T) {
		src := t.TempDir()
		dst := t.TempDir()

		changesDir := filepath.Join(src, "openspec", "changes")
		os.MkdirAll(changesDir, 0755)

		os.WriteFile(filepath.Join(src, "openspec", "config.yaml"), []byte("schema: spec-driven\n"), 0644)
		os.WriteFile(filepath.Join(src, "openspec", "project.md"), []byte("# Project\n"), 0644)
		os.WriteFile(filepath.Join(src, "openspec", "AGENTS.md"), []byte("# Agents\n"), 0644)
		os.MkdirAll(filepath.Join(src, "openspec", "explorations"), 0755)

		stats, err := ImportOpenSpecProject(src, dst)
		if err != nil {
			t.Fatalf("ImportOpenSpecProject() error: %v", err)
		}

		expectedWarnings := []string{"config.yaml", "project.md", "AGENTS.md", "explorations"}
		for _, expected := range expectedWarnings {
			found := false
			for _, w := range stats.Warnings {
				if strings.Contains(w, expected) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected warning about %q, got warnings: %v", expected, stats.Warnings)
			}
		}
	})

	t.Run("warns on malformed date in metadata", func(t *testing.T) {
		src := t.TempDir()
		dst := t.TempDir()

		changeDir := filepath.Join(src, "openspec", "changes", "bad-date")
		os.MkdirAll(changeDir, 0755)
		os.WriteFile(filepath.Join(changeDir, ".openspec.yaml"),
			[]byte("schema: spec-driven\ncreated: not-a-date\n"), 0644)
		os.WriteFile(filepath.Join(changeDir, "proposal.md"), []byte("## Motivation\n"), 0644)

		stats, err := ImportOpenSpecProject(src, dst)
		if err != nil {
			t.Fatalf("ImportOpenSpecProject() error: %v", err)
		}

		found := false
		for _, w := range stats.Warnings {
			if strings.Contains(w, "not-a-date") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected warning about malformed date, got warnings: %v", stats.Warnings)
		}
	})

	t.Run("warns about dropped unsupported metadata fields", func(t *testing.T) {
		src := t.TempDir()
		dst := t.TempDir()

		changeDir := filepath.Join(src, "openspec", "changes", "dropped-fields")
		os.MkdirAll(changeDir, 0755)
		os.WriteFile(filepath.Join(changeDir, ".openspec.yaml"),
			[]byte("schema: spec-driven\ncreated: 2026-03-01\nprovides:\n  - foo\nrequires:\n  - bar\ntouches:\n  - baz\nparent: some-parent\n"), 0644)
		os.WriteFile(filepath.Join(changeDir, "proposal.md"), []byte("## Motivation\n"), 0644)

		stats, err := ImportOpenSpecProject(src, dst)
		if err != nil {
			t.Fatalf("ImportOpenSpecProject() error: %v", err)
		}

		found := false
		for _, w := range stats.Warnings {
			if strings.Contains(w, "skipped unsupported fields") &&
				strings.Contains(w, "provides") &&
				strings.Contains(w, "requires") &&
				strings.Contains(w, "touches") &&
				strings.Contains(w, "parent") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected warning about dropped fields, got warnings: %v", stats.Warnings)
		}

		liteMeta, err := os.ReadFile(filepath.Join(dst, "specs", "changes", "dropped-fields", ".litespec.yaml"))
		if err != nil {
			t.Fatalf("read litespec metadata: %v", err)
		}
		content := string(liteMeta)
		if strings.Contains(content, "provides") || strings.Contains(content, "requires") ||
			strings.Contains(content, "touches") || strings.Contains(content, "parent") {
			t.Errorf("dropped fields should not appear in output metadata: %q", content)
		}
	})
}

func TestPreviewImport(t *testing.T) {
	t.Run("counts without writing files", func(t *testing.T) {
		src := t.TempDir()

		specsDir := filepath.Join(src, "openspec", "specs", "capability1")
		os.MkdirAll(specsDir, 0755)
		os.WriteFile(filepath.Join(specsDir, "spec.md"), []byte("# capability1\n"), 0644)

		specsDir2 := filepath.Join(src, "openspec", "specs", "capability2")
		os.MkdirAll(specsDir2, 0755)
		os.WriteFile(filepath.Join(specsDir2, "spec.md"), []byte("# capability2\n"), 0644)

		stats, err := PreviewImport(src)
		if err != nil {
			t.Fatalf("PreviewImport() error: %v", err)
		}

		if stats.CanonSpecs != 2 {
			t.Errorf("CanonSpecs = %d, want 2", stats.CanonSpecs)
		}
	})

	t.Run("counts active changes and archives", func(t *testing.T) {
		src := t.TempDir()

		changeDir := filepath.Join(src, "openspec", "changes", "my-change")
		os.MkdirAll(changeDir, 0755)
		os.WriteFile(filepath.Join(changeDir, ".openspec.yaml"), []byte("schema: spec-driven\n"), 0644)

		archiveDir := filepath.Join(src, "openspec", "changes", "archive", "2025-01-11-old")
		os.MkdirAll(archiveDir, 0755)

		stats, err := PreviewImport(src)
		if err != nil {
			t.Fatalf("PreviewImport() error: %v", err)
		}

		if stats.ActiveChanges != 1 {
			t.Errorf("ActiveChanges = %d, want 1", stats.ActiveChanges)
		}
		if stats.Archives != 1 {
			t.Errorf("Archives = %d, want 1", stats.Archives)
		}
	})
	t.Run("collects entry names during preview", func(t *testing.T) {
		src := t.TempDir()

		for _, name := range []string{"alpha", "beta"} {
			dir := filepath.Join(src, "openspec", "specs", name)
			os.MkdirAll(dir, 0755)
			os.WriteFile(filepath.Join(dir, "spec.md"), []byte("# "+name+"\n"), 0644)
		}

		changeDir := filepath.Join(src, "openspec", "changes", "active-change")
		os.MkdirAll(changeDir, 0755)
		os.WriteFile(filepath.Join(changeDir, ".openspec.yaml"), []byte("schema: spec-driven\n"), 0644)

		archiveDir := filepath.Join(src, "openspec", "changes", "archive", "2025-01-11-old")
		os.MkdirAll(archiveDir, 0755)

		stats, err := PreviewImport(src)
		if err != nil {
			t.Fatalf("PreviewImport() error: %v", err)
		}

		if len(stats.CanonSpecNames) != 2 {
			t.Errorf("CanonSpecNames = %v, want 2 entries", stats.CanonSpecNames)
		}
		foundAlpha, foundBeta := false, false
		for _, n := range stats.CanonSpecNames {
			if n == "alpha" {
				foundAlpha = true
			}
			if n == "beta" {
				foundBeta = true
			}
		}
		if !foundAlpha || !foundBeta {
			t.Errorf("CanonSpecNames missing entries, got: %v", stats.CanonSpecNames)
		}
		if len(stats.ActiveChangeNames) != 1 || stats.ActiveChangeNames[0] != "active-change" {
			t.Errorf("ActiveChangeNames = %v, want [active-change]", stats.ActiveChangeNames)
		}
		if len(stats.ArchiveNames) != 1 || stats.ArchiveNames[0] != "2025-01-11-old" {
			t.Errorf("ArchiveNames = %v, want [2025-01-11-old]", stats.ArchiveNames)
		}
	})
}

func TestCheckConflicts(t *testing.T) {
	t.Run("detects existing canon directory", func(t *testing.T) {
		src := t.TempDir()
		dst := t.TempDir()

		os.MkdirAll(filepath.Join(dst, "specs", "canon"), 0755)

		conflicts, err := CheckConflicts(src, dst)
		if err != nil {
			t.Fatalf("CheckConflicts() error: %v", err)
		}

		if len(conflicts) == 0 {
			t.Error("expected conflicts for existing canon directory")
		}
	})

	t.Run("detects existing change directories", func(t *testing.T) {
		src := t.TempDir()
		dst := t.TempDir()

		os.MkdirAll(filepath.Join(dst, "specs", "changes", "existing-change"), 0755)

		conflicts, err := CheckConflicts(src, dst)
		if err != nil {
			t.Fatalf("CheckConflicts() error: %v", err)
		}

		if len(conflicts) == 0 {
			t.Error("expected conflicts for existing change directory")
		}
	})

	t.Run("ignores archive directory", func(t *testing.T) {
		src := t.TempDir()
		dst := t.TempDir()

		os.MkdirAll(filepath.Join(dst, "specs", "changes", "archive"), 0755)

		conflicts, err := CheckConflicts(src, dst)
		if err != nil {
			t.Fatalf("CheckConflicts() error: %v", err)
		}

		for _, c := range conflicts {
			if filepath.Base(c) == "archive" {
				t.Error("archive directory should not be a conflict")
			}
		}
	})

	t.Run("returns empty for clean target", func(t *testing.T) {
		src := t.TempDir()
		dst := t.TempDir()

		conflicts, err := CheckConflicts(src, dst)
		if err != nil {
			t.Fatalf("CheckConflicts() error: %v", err)
		}

		if len(conflicts) != 0 {
			t.Errorf("expected no conflicts, got %d", len(conflicts))
		}
	})
}
