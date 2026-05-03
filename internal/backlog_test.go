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
	if summary.Other != 0 {
		t.Errorf("Other = %d, want 0 (Nice-to-Have is unrecognized, not other)", summary.Other)
	}
	if len(summary.Unrecognized) != 1 || summary.Unrecognized[0] != "Nice-to-Have" {
		t.Errorf("Unrecognized = %v, want [Nice-to-Have]", summary.Unrecognized)
	}
}

func TestParseBacklog_ExplicitOtherSection(t *testing.T) {
	content := `## Deferred

- Item one

## Other

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
	if len(summary.Unrecognized) != 0 {
		t.Errorf("Unrecognized = %v, want empty", summary.Unrecognized)
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

func TestValidateBacklog_UnrecognizedSections(t *testing.T) {
	content := `## Deferred

- Real item

## Deferred Items

- Misplaced item

## Nice-to-Have

- Wish
`
	dir := t.TempDir()
	root := dir
	backlogDir := filepath.Join(root, "specs", "canon")
	os.MkdirAll(backlogDir, 0o755)
	backlogPath := filepath.Join(root, "specs", "backlog.md")
	if err := os.WriteFile(backlogPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	result := ValidateBacklog(root)
	if len(result.Warnings) != 2 {
		t.Fatalf("expected 2 warnings, got %d: %v", len(result.Warnings), result.Warnings)
	}
	if result.Warnings[0].Message != `"Deferred Items" is not a recognized section — use ## Deferred, ## Open Questions, ## Future Versions, or ## Other` {
		t.Errorf("unexpected warning: %s", result.Warnings[0].Message)
	}
	if result.Warnings[1].Message != `"Nice-to-Have" is not a recognized section — use ## Deferred, ## Open Questions, ## Future Versions, or ## Other` {
		t.Errorf("unexpected warning: %s", result.Warnings[1].Message)
	}
}

func TestValidateBacklog_NoUnrecognizedSections(t *testing.T) {
	content := `## Deferred

- Item one

## Other

- Wish A
`
	dir := t.TempDir()
	root := dir
	backlogDir := filepath.Join(root, "specs", "canon")
	os.MkdirAll(backlogDir, 0o755)
	backlogPath := filepath.Join(root, "specs", "backlog.md")
	if err := os.WriteFile(backlogPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	result := ValidateBacklog(root)
	if len(result.Warnings) != 0 {
		t.Fatalf("expected 0 warnings, got %d: %v", len(result.Warnings), result.Warnings)
	}
}

func TestParseBacklogItems_AllSections(t *testing.T) {
	content := `## Deferred

- **Item one** — description
- **Item two** — more text

## Open Questions

- **Question A** — details

## Future Versions

- **Feature X** — desc
`
	dir := t.TempDir()
	path := filepath.Join(dir, "backlog.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	items, err := ParseBacklogItems(path)
	if err != nil {
		t.Fatalf("ParseBacklogItems: %v", err)
	}
	if len(items) != 4 {
		t.Fatalf("expected 4 items, got %d", len(items))
	}

	checks := []struct {
		section string
		title   string
	}{
		{"deferred", "Item one"},
		{"deferred", "Item two"},
		{"open-questions", "Question A"},
		{"future", "Feature X"},
	}
	for i, want := range checks {
		if items[i].Section != want.section {
			t.Errorf("items[%d].Section = %q, want %q", i, items[i].Section, want.section)
		}
		if items[i].Title != want.title {
			t.Errorf("items[%d].Title = %q, want %q", i, items[i].Title, want.title)
		}
	}
}

func TestParseBacklogItems_NoBoldTitle(t *testing.T) {
	content := `## Deferred

- Plain text item
`
	dir := t.TempDir()
	path := filepath.Join(dir, "backlog.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	items, err := ParseBacklogItems(path)
	if err != nil {
		t.Fatalf("ParseBacklogItems: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Title != "Plain text item" {
		t.Errorf("Title = %q, want %q", items[0].Title, "Plain text item")
	}
}

func TestParseBacklogItems_MissingFileReturnsNil(t *testing.T) {
	path := filepath.Join(t.TempDir(), "backlog.md")
	items, err := ParseBacklogItems(path)
	if err != nil {
		t.Fatalf("ParseBacklogItems: %v", err)
	}
	if items != nil {
		t.Fatalf("expected nil, got %v", items)
	}
}

func TestParseBacklogItems_SkipsUnrecognizedSections(t *testing.T) {
	content := `## Deferred

- **Real item** — desc

## Nice-to-Have

- **Wish** — desc
`
	dir := t.TempDir()
	path := filepath.Join(dir, "backlog.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	items, err := ParseBacklogItems(path)
	if err != nil {
		t.Fatalf("ParseBacklogItems: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Title != "Real item" {
		t.Errorf("Title = %q, want %q", items[0].Title, "Real item")
	}
}

func TestParseBacklogItems_IgnoresNestedBullets(t *testing.T) {
	content := `## Deferred

- **Top level** — desc
  - nested sub-item
  - another nested
- **Another top** — desc
`
	dir := t.TempDir()
	path := filepath.Join(dir, "backlog.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	items, err := ParseBacklogItems(path)
	if err != nil {
		t.Fatalf("ParseBacklogItems: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}

func TestParseBacklogItems_CRLFLineEndings(t *testing.T) {
	content := "## Deferred\r\n\r\n- **Item one** — desc\r\n- **Item two** — desc\r\n"
	dir := t.TempDir()
	path := filepath.Join(dir, "backlog.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	items, err := ParseBacklogItems(path)
	if err != nil {
		t.Fatalf("ParseBacklogItems: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}

func TestParseBacklogItems_OtherSection(t *testing.T) {
	content := `## Deferred

- **Item one** — desc

## Other

- **Wish A** — desc
- **Wish B** — desc
`
	dir := t.TempDir()
	path := filepath.Join(dir, "backlog.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	items, err := ParseBacklogItems(path)
	if err != nil {
		t.Fatalf("ParseBacklogItems: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
	if items[1].Section != "other" {
		t.Errorf("items[1].Section = %q, want %q", items[1].Section, "other")
	}
	if items[1].Title != "Wish A" {
		t.Errorf("items[1].Title = %q, want %q", items[1].Title, "Wish A")
	}
}

func TestParseBacklogItems_FutureShorthand(t *testing.T) {
	content := `## Future

- **Item one** — desc
- **Item two** — desc
`
	dir := t.TempDir()
	path := filepath.Join(dir, "backlog.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	items, err := ParseBacklogItems(path)
	if err != nil {
		t.Fatalf("ParseBacklogItems: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Section != "future" {
		t.Errorf("items[0].Section = %q, want %q", items[0].Section, "future")
	}
}

func TestParseBacklogItems_AsteriskBullets(t *testing.T) {
	content := `## Deferred

* **Item one** — desc
* **Item two** — desc
`
	dir := t.TempDir()
	path := filepath.Join(dir, "backlog.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	items, err := ParseBacklogItems(path)
	if err != nil {
		t.Fatalf("ParseBacklogItems: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}

func TestParseBacklogItems_MixedBoldAndPlain(t *testing.T) {
	content := `## Deferred

- **Bold item** — desc
- Plain text item
`
	dir := t.TempDir()
	path := filepath.Join(dir, "backlog.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	items, err := ParseBacklogItems(path)
	if err != nil {
		t.Fatalf("ParseBacklogItems: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Title != "Bold item" {
		t.Errorf("items[0].Title = %q, want %q", items[0].Title, "Bold item")
	}
	if items[1].Title != "Plain text item" {
		t.Errorf("items[1].Title = %q, want %q", items[1].Title, "Plain text item")
	}
}

func TestParseBacklogItems_UnclosedBoldMarker(t *testing.T) {
	content := `## Deferred

- **Title with no closing
`
	dir := t.TempDir()
	path := filepath.Join(dir, "backlog.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	items, err := ParseBacklogItems(path)
	if err != nil {
		t.Fatalf("ParseBacklogItems: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Title != "Title with no closing" {
		t.Errorf("Title = %q, want %q", items[0].Title, "Title with no closing")
	}
}
