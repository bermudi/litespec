package internal

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestGetLastModifiedNestedFiles(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	oldFile := filepath.Join(root, "old.txt")
	newFile := filepath.Join(sub, "new.txt")

	writeFile(t, oldFile, "old")
	writeFile(t, newFile, "new")

	oldTime := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(oldFile, oldTime, oldTime); err != nil {
		t.Fatalf("chtimes: %v", err)
	}

	mtime, err := GetLastModified(root)
	if err != nil {
		t.Fatalf("GetLastModified: %v", err)
	}
	if mtime.Before(oldTime) {
		t.Errorf("mtime = %v, want at least %v", mtime, oldTime)
	}
}

func TestGetLastModifiedEmptyDirectory(t *testing.T) {
	root := t.TempDir()

	dirInfo, err := os.Stat(root)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	dirMtime := dirInfo.ModTime()

	mtime, err := GetLastModified(root)
	if err != nil {
		t.Fatalf("GetLastModified: %v", err)
	}
	if !mtime.Equal(dirMtime) {
		t.Errorf("mtime = %v, want dir mtime %v", mtime, dirMtime)
	}
}

func TestGetLastModifiedSingleFile(t *testing.T) {
	root := t.TempDir()
	f := filepath.Join(root, "only.txt")
	writeFile(t, f, "content")

	fi, err := os.Stat(f)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}

	mtime, err := GetLastModified(root)
	if err != nil {
		t.Fatalf("GetLastModified: %v", err)
	}
	if !mtime.Equal(fi.ModTime()) {
		t.Errorf("mtime = %v, want %v", mtime, fi.ModTime())
	}
}

func TestFormatRelativeTimeJustNow(t *testing.T) {
	result := FormatRelativeTime(time.Now())
	if result != "just now" {
		t.Errorf("got %q, want %q", result, "just now")
	}
}

func TestFormatRelativeTimeMinutes(t *testing.T) {
	result := FormatRelativeTime(time.Now().Add(-5 * time.Minute))
	if result != "5m ago" {
		t.Errorf("got %q, want %q", result, "5m ago")
	}
}

func TestFormatRelativeTimeHours(t *testing.T) {
	result := FormatRelativeTime(time.Now().Add(-3 * time.Hour))
	if result != "3h ago" {
		t.Errorf("got %q, want %q", result, "3h ago")
	}
}

func TestFormatRelativeTimeDays(t *testing.T) {
	result := FormatRelativeTime(time.Now().Add(-7 * 24 * time.Hour))
	if result != "7d ago" {
		t.Errorf("got %q, want %q", result, "7d ago")
	}
}

func TestFormatRelativeTimeOldDate(t *testing.T) {
	old := time.Now().Add(-60 * 24 * time.Hour)
	result := FormatRelativeTime(old)
	expected := old.Format("2006-01-02")
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestListChangesEnriched(t *testing.T) {
	root := setupTestProject(t)

	writeChangeFile(t, root, "my-change", "tasks.md", "## Phase 1\n- [x] Done\n- [ ] Not done\n")
	writeChangeFile(t, root, "my-change", "proposal.md", "# Proposal")

	changes, err := ListChanges(root)
	if err != nil {
		t.Fatalf("ListChanges: %v", err)
	}
	if len(changes) != 1 {
		t.Fatalf("changes count = %d, want 1", len(changes))
	}
	c := changes[0]
	if c.Name != "my-change" {
		t.Errorf("Name = %q, want %q", c.Name, "my-change")
	}
	if c.CompletedTasks != 1 {
		t.Errorf("CompletedTasks = %d, want 1", c.CompletedTasks)
	}
	if c.TotalTasks != 2 {
		t.Errorf("TotalTasks = %d, want 2", c.TotalTasks)
	}
	if c.LastModified.IsZero() {
		t.Error("LastModified is zero, should be populated")
	}
}

func TestListChangesNoTasksMD(t *testing.T) {
	root := setupTestProject(t)
	writeChangeFile(t, root, "bare-change", "proposal.md", "# Proposal")

	changes, err := ListChanges(root)
	if err != nil {
		t.Fatalf("ListChanges: %v", err)
	}
	if len(changes) != 1 {
		t.Fatalf("changes count = %d, want 1", len(changes))
	}
	c := changes[0]
	if c.CompletedTasks != 0 || c.TotalTasks != 0 {
		t.Errorf("got %d/%d, want 0/0", c.CompletedTasks, c.TotalTasks)
	}
}

func TestListSpecsEnriched(t *testing.T) {
	root := setupTestProject(t)
	writeMainSpecFile(t, root, "auth", `# auth

## Requirements

### Requirement: Login
The system SHALL authenticate.

#### Scenario: Valid
- **WHEN** valid creds

### Requirement: Logout
The system SHALL logout.
`)

	specs, err := ListSpecs(root)
	if err != nil {
		t.Fatalf("ListSpecs: %v", err)
	}
	if len(specs) != 1 {
		t.Fatalf("specs count = %d, want 1", len(specs))
	}
	s := specs[0]
	if s.Name != "auth" {
		t.Errorf("Name = %q, want %q", s.Name, "auth")
	}
	if s.RequirementCount != 2 {
		t.Errorf("RequirementCount = %d, want 2", s.RequirementCount)
	}
}

func TestListSpecsParseFailureReturnsZero(t *testing.T) {
	root := setupTestProject(t)
	dir := filepath.Join(CanonPath(root), "broken")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte("not valid spec"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	specs, err := ListSpecs(root)
	if err != nil {
		t.Fatalf("ListSpecs: %v", err)
	}
	if len(specs) != 1 {
		t.Fatalf("specs count = %d, want 1", len(specs))
	}
	if specs[0].RequirementCount != 0 {
		t.Errorf("RequirementCount = %d, want 0 on parse failure", specs[0].RequirementCount)
	}
}

func TestListSpecsNoSpecMD(t *testing.T) {
	root := setupTestProject(t)
	dir := filepath.Join(CanonPath(root), "empty-cap")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	specs, err := ListSpecs(root)
	if err != nil {
		t.Fatalf("ListSpecs: %v", err)
	}
	if len(specs) != 1 {
		t.Fatalf("specs count = %d, want 1", len(specs))
	}
	if specs[0].RequirementCount != 0 {
		t.Errorf("RequirementCount = %d, want 0 when no spec.md", specs[0].RequirementCount)
	}
}

func TestChangeListStatus(t *testing.T) {
	tests := []struct {
		completed int
		total     int
		want      string
	}{
		{0, 0, "no-tasks"},
		{5, 5, "complete"},
		{3, 5, "in-progress"},
		{0, 3, "in-progress"},
	}
	for _, tt := range tests {
		got := ChangeListStatus(tt.completed, tt.total)
		if got != tt.want {
			t.Errorf("ChangeListStatus(%d, %d) = %q, want %q", tt.completed, tt.total, got, tt.want)
		}
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestFormatRelativeTimeBoundaries(t *testing.T) {
	tests := []struct {
		name string
		dt   time.Duration
		want string
	}{
		{"exactly 1 minute", 1 * time.Minute, "1m ago"},
		{"just under 1 hour", 59*time.Minute + 59*time.Second, "59m ago"},
		{"exactly 1 hour", 1 * time.Hour, "1h ago"},
		{"just under 24 hours", 23*time.Hour + 59*time.Minute + 59*time.Second, "23h ago"},
		{"exactly 24 hours", 24 * time.Hour, "1d ago"},
		{"just under 30 days", 30*24*time.Hour - time.Second, "29d ago"},
		{"just over 30 days", 30*24*time.Hour + time.Second, time.Now().Add(-30*24*time.Hour - time.Second).Format("2006-01-02")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatRelativeTime(time.Now().Add(-tt.dt))
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestListChangesCreatedFromMeta(t *testing.T) {
	root := setupTestProject(t)

	changeDir := ChangePath(root, "ts-change")
	if err := os.MkdirAll(changeDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	createdTime := time.Date(2026, 3, 15, 10, 30, 0, 0, time.UTC)
	meta := ChangeMeta{
		Schema:  "spec-driven",
		Created: createdTime,
	}
	data, err := yaml.Marshal(&meta)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(changeDir, MetaFileName), data, 0o644); err != nil {
		t.Fatalf("write meta: %v", err)
	}

	changes, err := ListChanges(root)
	if err != nil {
		t.Fatalf("ListChanges: %v", err)
	}
	if len(changes) != 1 {
		t.Fatalf("changes count = %d, want 1", len(changes))
	}
	c := changes[0]
	if c.Created.IsZero() {
		t.Error("Created is zero, should be populated from .litespec.yaml")
	}
	if !c.Created.Equal(createdTime) {
		t.Errorf("Created = %v, want %v", c.Created, createdTime)
	}
}

func TestListChangesCreatedMissingMeta(t *testing.T) {
	root := setupTestProject(t)

	changeDir := ChangePath(root, "no-meta-change")
	if err := os.MkdirAll(changeDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	changes, err := ListChanges(root)
	if err != nil {
		t.Fatalf("ListChanges: %v", err)
	}
	if len(changes) != 1 {
		t.Fatalf("changes count = %d, want 1", len(changes))
	}
	if !changes[0].Created.IsZero() {
		t.Error("Created should be zero when no .litespec.yaml exists")
	}
}

func TestGetLastModifiedNonexistentDir(t *testing.T) {
	_, err := GetLastModified("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}
