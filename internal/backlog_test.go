package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseBacklog_AllCategories(t *testing.T) {
	content := `## Deferred

- Item one
- Item two

## Open Questions

- Question A
- Question B
- Question C

## Future Versions

- Feature X
`
	dir := t.TempDir()
	path := filepath.Join(dir, "backlog.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	summary, err := ParseBacklog(path)
	if err != nil {
		t.Fatalf("ParseBacklog: %v", err)
	}
	if summary == nil {
		t.Fatal("expected non-nil summary")
	}
	if summary.Deferred != 2 {
		t.Errorf("Deferred = %d, want 2", summary.Deferred)
	}
	if summary.OpenQuestions != 3 {
		t.Errorf("OpenQuestions = %d, want 3", summary.OpenQuestions)
	}
	if summary.Future != 1 {
		t.Errorf("Future = %d, want 1", summary.Future)
	}
	if summary.Other != 0 {
		t.Errorf("Other = %d, want 0", summary.Other)
	}
}

func TestParseBacklog_UnknownSectionsAsOther(t *testing.T) {
	content := `## Deferred

- Item one

## Nice-to-Have

- Wish A
- Wish B
`
	dir := t.TempDir()
	path := filepath.Join(dir, "backlog.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	summary, err := ParseBacklog(path)
	if err != nil {
		t.Fatalf("ParseBacklog: %v", err)
	}
	if summary == nil {
		t.Fatal("expected non-nil summary")
	}
	if summary.Deferred != 1 {
		t.Errorf("Deferred = %d, want 1", summary.Deferred)
	}
	if summary.Other != 2 {
		t.Errorf("Other = %d, want 2", summary.Other)
	}
}

func TestParseBacklog_NestedBulletsIgnored(t *testing.T) {
	content := `## Deferred

- Top level item
  - Nested one
  - Nested two
- Another top level
`
	dir := t.TempDir()
	path := filepath.Join(dir, "backlog.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	summary, err := ParseBacklog(path)
	if err != nil {
		t.Fatalf("ParseBacklog: %v", err)
	}
	if summary == nil {
		t.Fatal("expected non-nil summary")
	}
	if summary.Deferred != 2 {
		t.Errorf("Deferred = %d, want 2", summary.Deferred)
	}
}

func TestParseBacklog_MissingFileReturnsNil(t *testing.T) {
	path := filepath.Join(t.TempDir(), "backlog.md")

	summary, err := ParseBacklog(path)
	if err != nil {
		t.Fatalf("ParseBacklog: %v", err)
	}
	if summary != nil {
		t.Errorf("expected nil summary for missing file, got %+v", summary)
	}
}

func TestParseBacklog_EmptyFileReturnsNil(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "backlog.md")
	if err := os.WriteFile(path, []byte("## Deferred\n\n## Open Questions\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	summary, err := ParseBacklog(path)
	if err != nil {
		t.Fatalf("ParseBacklog: %v", err)
	}
	if summary != nil {
		t.Errorf("expected nil summary for empty file, got %+v", summary)
	}
}

func TestParseBacklog_CRLFLineEndings(t *testing.T) {
	content := "## Deferred\r\n\r\n- Item one\r\n- Item two\r\n\r\n## Open Questions\r\n\r\n- Question A\r\n"

	dir := t.TempDir()
	path := filepath.Join(dir, "backlog.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	summary, err := ParseBacklog(path)
	if err != nil {
		t.Fatalf("ParseBacklog: %v", err)
	}
	if summary == nil {
		t.Fatal("expected non-nil summary")
	}
	if summary.Deferred != 2 {
		t.Errorf("Deferred = %d, want 2", summary.Deferred)
	}
	if summary.OpenQuestions != 1 {
		t.Errorf("OpenQuestions = %d, want 1", summary.OpenQuestions)
	}
}

func TestParseBacklog_FutureShorthand(t *testing.T) {
	content := `## Future

- Item one
- Item two
`
	dir := t.TempDir()
	path := filepath.Join(dir, "backlog.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	summary, err := ParseBacklog(path)
	if err != nil {
		t.Fatalf("ParseBacklog: %v", err)
	}
	if summary == nil {
		t.Fatal("expected non-nil summary")
	}
	if summary.Future != 2 {
		t.Errorf("Future = %d, want 2", summary.Future)
	}
	if summary.Other != 0 {
		t.Errorf("Other = %d, want 0", summary.Other)
	}
}

func TestParseBacklog_AsteriskBullets(t *testing.T) {
	content := `## Deferred

* Item one
* Item two

## Open Questions

* Question A
`
	dir := t.TempDir()
	path := filepath.Join(dir, "backlog.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	summary, err := ParseBacklog(path)
	if err != nil {
		t.Fatalf("ParseBacklog: %v", err)
	}
	if summary == nil {
		t.Fatal("expected non-nil summary")
	}
	if summary.Deferred != 2 {
		t.Errorf("Deferred = %d, want 2", summary.Deferred)
	}
	if summary.OpenQuestions != 1 {
		t.Errorf("OpenQuestions = %d, want 1", summary.OpenQuestions)
	}
}

func TestParseBacklog_CaseInsensitiveHeaders(t *testing.T) {
	content := `## deferred

- Item one

## open questions

- Question A

## future versions

- Feature X
`
	dir := t.TempDir()
	path := filepath.Join(dir, "backlog.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	summary, err := ParseBacklog(path)
	if err != nil {
		t.Fatalf("ParseBacklog: %v", err)
	}
	if summary == nil {
		t.Fatal("expected non-nil summary")
	}
	if summary.Deferred != 1 {
		t.Errorf("Deferred = %d, want 1", summary.Deferred)
	}
	if summary.OpenQuestions != 1 {
		t.Errorf("OpenQuestions = %d, want 1", summary.OpenQuestions)
	}
	if summary.Future != 1 {
		t.Errorf("Future = %d, want 1", summary.Future)
	}
}

func TestBacklogPath(t *testing.T) {
	root := "/project"
	got := BacklogPath(root)
	want := filepath.Join(root, ProjectDirName, BacklogFileName)
	if got != want {
		t.Errorf("BacklogPath = %q, want %q", got, want)
	}
}
