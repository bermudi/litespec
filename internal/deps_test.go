package internal

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestChangeMetaRoundTripWithDependsOn(t *testing.T) {
	meta := ChangeMeta{
		Schema:    "spec-driven",
		DependsOn: []string{"add-user-auth", "add-logging"},
	}

	data, err := yaml.Marshal(&meta)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var parsed ChangeMeta
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(parsed.DependsOn) != 2 {
		t.Fatalf("DependsOn length = %d, want 2", len(parsed.DependsOn))
	}
	if parsed.DependsOn[0] != "add-user-auth" {
		t.Errorf("DependsOn[0] = %q, want %q", parsed.DependsOn[0], "add-user-auth")
	}
	if parsed.DependsOn[1] != "add-logging" {
		t.Errorf("DependsOn[1] = %q, want %q", parsed.DependsOn[1], "add-logging")
	}
}

func TestChangeMetaRoundTripWithoutDependsOn(t *testing.T) {
	meta := ChangeMeta{
		Schema: "spec-driven",
	}

	data, err := yaml.Marshal(&meta)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var parsed ChangeMeta
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if parsed.DependsOn != nil {
		t.Errorf("DependsOn = %v, want nil", parsed.DependsOn)
	}
}

func TestChangeMetaReadFromDisk(t *testing.T) {
	root := setupTestProject(t)

	if err := CreateChange(root, "test-dep"); err != nil {
		t.Fatalf("CreateChange: %v", err)
	}

	meta, err := ReadChangeMeta(root, "test-dep")
	if err != nil {
		t.Fatalf("ReadChangeMeta: %v", err)
	}
	if meta.Schema != "spec-driven" {
		t.Errorf("Schema = %q, want %q", meta.Schema, "spec-driven")
	}
	if meta.DependsOn != nil {
		t.Errorf("DependsOn = %v, want nil for new change", meta.DependsOn)
	}
}

func TestChangeMetaReadWithDependsOn(t *testing.T) {
	root := setupTestProject(t)

	if err := CreateChange(root, "child-change"); err != nil {
		t.Fatalf("CreateChange: %v", err)
	}

	if err := UpdateChangeDeps(root, "child-change", []string{"parent-change"}); err != nil {
		t.Fatalf("UpdateChangeDeps: %v", err)
	}

	meta, err := ReadChangeMeta(root, "child-change")
	if err != nil {
		t.Fatalf("ReadChangeMeta: %v", err)
	}
	if len(meta.DependsOn) != 1 || meta.DependsOn[0] != "parent-change" {
		t.Errorf("DependsOn = %v, want [parent-change]", meta.DependsOn)
	}
}

func TestParseArchivedName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"2026-04-01-add-auth", "add-auth"},
		{"2026-03-15-fix-validation-bug", "fix-validation-bug"},
		{"add-auth", "add-auth"},
		{"not-a-date-change", "not-a-date-change"},
		{"2026-13-45-bad-date", "bad-date"},
	}
	for _, tt := range tests {
		got := ParseArchivedName(tt.input)
		if got != tt.want {
			t.Errorf("ParseArchivedName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestListArchivedChanges(t *testing.T) {
	root := setupTestProject(t)

	os.MkdirAll(filepath.Join(ArchivePath(root), "2026-04-01-add-auth"), 0o755)
	os.MkdirAll(filepath.Join(ArchivePath(root), "2026-04-02-add-logging"), 0o755)

	names, err := ListArchivedChanges(root)
	if err != nil {
		t.Fatalf("ListArchivedChanges: %v", err)
	}

	found := map[string]bool{}
	for _, n := range names {
		found[n] = true
	}
	if !found["add-auth"] {
		t.Error("expected add-auth in archived names")
	}
	if !found["add-logging"] {
		t.Error("expected add-logging in archived names")
	}
}

func TestListArchivedChangesEmpty(t *testing.T) {
	root := setupTestProject(t)

	names, err := ListArchivedChanges(root)
	if err != nil {
		t.Fatalf("ListArchivedChanges: %v", err)
	}
	if len(names) != 0 {
		t.Errorf("expected 0 archived changes, got %d", len(names))
	}
}

func TestResolveDepActiveChange(t *testing.T) {
	root := setupTestProject(t)
	CreateChange(root, "add-auth")

	dep, found := ResolveDep(root, "add-auth")
	if !found {
		t.Fatal("expected to find active change")
	}
	if !dep.IsActive {
		t.Error("expected IsActive = true for active change")
	}
}

func TestResolveDepArchivedChange(t *testing.T) {
	root := setupTestProject(t)
	os.MkdirAll(filepath.Join(ArchivePath(root), "2026-04-01-add-auth"), 0o755)

	dep, found := ResolveDep(root, "add-auth")
	if !found {
		t.Fatal("expected to find archived change")
	}
	if dep.IsActive {
		t.Error("expected IsActive = false for archived change")
	}
}

func TestResolveDepActiveTakesPriority(t *testing.T) {
	root := setupTestProject(t)
	CreateChange(root, "add-auth")
	os.MkdirAll(filepath.Join(ArchivePath(root), "2026-04-01-add-auth"), 0o755)

	dep, found := ResolveDep(root, "add-auth")
	if !found {
		t.Fatal("expected to find change")
	}
	if !dep.IsActive {
		t.Error("expected active change to take priority")
	}
}

func TestResolveDepNotFound(t *testing.T) {
	root := setupTestProject(t)

	_, found := ResolveDep(root, "nonexistent")
	if found {
		t.Error("expected not to find nonexistent change")
	}
}

func TestResolveDepsAllValid(t *testing.T) {
	root := setupTestProject(t)
	CreateChange(root, "add-auth")
	os.MkdirAll(filepath.Join(ArchivePath(root), "2026-04-01-add-logging"), 0o755)

	resolved, err := ResolveDeps(root, []string{"add-auth", "add-logging"})
	if err != nil {
		t.Fatalf("ResolveDeps: %v", err)
	}
	if len(resolved) != 2 {
		t.Fatalf("expected 2 resolved deps, got %d", len(resolved))
	}
	if !resolved[0].IsActive {
		t.Error("first dep should be active")
	}
	if resolved[1].IsActive {
		t.Error("second dep should be archived")
	}
}

func TestResolveDepsMissingDep(t *testing.T) {
	root := setupTestProject(t)

	_, err := ResolveDeps(root, []string{"nonexistent"})
	if err == nil {
		t.Fatal("expected error for missing dependency")
	}
}

func TestResolveDepsEmpty(t *testing.T) {
	root := setupTestProject(t)

	resolved, err := ResolveDeps(root, nil)
	if err != nil {
		t.Fatalf("ResolveDeps: %v", err)
	}
	if resolved != nil {
		t.Errorf("expected nil for empty deps, got %v", resolved)
	}
}

func TestDetectCyclesNoCycles(t *testing.T) {
	depMap := map[string][]string{
		"a": {"b"},
		"b": {"c"},
	}
	cycles := DetectCycles(depMap)
	if len(cycles) != 0 {
		t.Errorf("expected 0 cycles, got %d", len(cycles))
	}
}

func TestDetectCyclesSimple(t *testing.T) {
	depMap := map[string][]string{
		"a": {"b"},
		"b": {"a"},
	}
	cycles := DetectCycles(depMap)
	if len(cycles) == 0 {
		t.Fatal("expected cycle to be detected")
	}
	found := false
	for _, c := range cycles {
		if len(c) >= 3 {
			names := map[string]bool{}
			for _, n := range c {
				names[n] = true
			}
			if names["a"] && names["b"] {
				found = true
			}
		}
	}
	if !found {
		t.Errorf("expected a<->b cycle, got %v", cycles)
	}
}

func TestDetectCyclesLonger(t *testing.T) {
	depMap := map[string][]string{
		"a": {"b"},
		"b": {"c"},
		"c": {"a"},
	}
	cycles := DetectCycles(depMap)
	if len(cycles) == 0 {
		t.Fatal("expected cycle to be detected")
	}
}

func TestDetectCyclesEmpty(t *testing.T) {
	cycles := DetectCycles(nil)
	if len(cycles) != 0 {
		t.Errorf("expected 0 cycles for empty dep map, got %d", len(cycles))
	}
}

func TestTopologicalSortSimple(t *testing.T) {
	changes := []ChangeInfo{
		{Name: "b"},
		{Name: "a"},
	}
	depMap := map[string][]string{
		"b": {"a"},
	}

	result := TopologicalSort(changes, depMap)
	if len(result) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result))
	}
	if result[0].Name != "a" {
		t.Errorf("first should be a, got %s", result[0].Name)
	}
	if result[1].Name != "b" {
		t.Errorf("second should be b, got %s", result[1].Name)
	}
}

func TestTopologicalSortUnrelated(t *testing.T) {
	changes := []ChangeInfo{
		{Name: "c"},
		{Name: "a"},
		{Name: "b"},
	}
	depMap := map[string][]string{
		"b": {"a"},
	}

	result := TopologicalSort(changes, depMap)
	if result[0].Name != "a" {
		t.Errorf("first should be a, got %s", result[0].Name)
	}
	if result[1].Name != "c" {
		t.Errorf("second should be c (level-0 alphabetical), got %s", result[1].Name)
	}
	if result[2].Name != "b" {
		t.Errorf("third should be b (level-1), got %s", result[2].Name)
	}
}

func TestTopologicalSortNoDeps(t *testing.T) {
	changes := []ChangeInfo{
		{Name: "c"},
		{Name: "a"},
		{Name: "b"},
	}

	result := TopologicalSort(changes, nil)
	if result[0].Name != "a" {
		t.Errorf("first should be a, got %s", result[0].Name)
	}
	if result[1].Name != "b" {
		t.Errorf("second should be b, got %s", result[1].Name)
	}
	if result[2].Name != "c" {
		t.Errorf("third should be c, got %s", result[2].Name)
	}
}

func TestGetDependents(t *testing.T) {
	root := setupTestProject(t)
	CreateChange(root, "add-auth")
	CreateChange(root, "add-rate-limiting")

	if err := UpdateChangeDeps(root, "add-rate-limiting", []string{"add-auth"}); err != nil {
		t.Fatalf("UpdateChangeDeps: %v", err)
	}

	dependents, err := GetDependents(root, "add-auth")
	if err != nil {
		t.Fatalf("GetDependents: %v", err)
	}
	if len(dependents) != 1 || dependents[0] != "add-rate-limiting" {
		t.Errorf("dependents = %v, want [add-rate-limiting]", dependents)
	}
}

func TestGetDependentsNone(t *testing.T) {
	root := setupTestProject(t)
	CreateChange(root, "add-auth")

	dependents, err := GetDependents(root, "add-auth")
	if err != nil {
		t.Fatalf("GetDependents: %v", err)
	}
	if len(dependents) != 0 {
		t.Errorf("expected 0 dependents, got %d", len(dependents))
	}
}

func TestLoadDepMap(t *testing.T) {
	root := setupTestProject(t)
	CreateChange(root, "add-auth")
	CreateChange(root, "add-rate-limiting")

	if err := UpdateChangeDeps(root, "add-rate-limiting", []string{"add-auth"}); err != nil {
		t.Fatalf("UpdateChangeDeps: %v", err)
	}

	depMap, err := LoadDepMap(root)
	if err != nil {
		t.Fatalf("LoadDepMap: %v", err)
	}
	if len(depMap) != 1 {
		t.Fatalf("expected 1 entry in depMap, got %d", len(depMap))
	}
	if depMap["add-rate-limiting"][0] != "add-auth" {
		t.Errorf("expected add-rate-limiting -> [add-auth], got %v", depMap)
	}
}
