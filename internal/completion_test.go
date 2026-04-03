package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCompleteCommandNames(t *testing.T) {
	result := Complete("", []string{})
	if len(result) == 0 {
		t.Fatal("expected completions for empty words")
	}

	names := make(map[string]bool)
	for _, c := range result {
		names[c.Candidate] = true
	}

	for _, cmd := range []string{"init", "new", "list", "status", "validate", "instructions", "archive", "update", "completion"} {
		if !names[cmd] {
			t.Errorf("missing command %q in completions", cmd)
		}
	}

	if names["__complete"] {
		t.Error("__complete should not appear in command completions")
	}
}

func TestCompleteSingleWord(t *testing.T) {
	result := Complete("", []string{"v"})
	for _, c := range result {
		if c.Candidate == "validate" {
			return
		}
	}
	t.Error("expected 'validate' in completions for 'v'")
}

func TestCompleteSingleWordDash(t *testing.T) {
	result := Complete("", []string{"--"})
	for _, c := range result {
		if c.Candidate == "--version" {
			return
		}
	}
	t.Error("expected '--version' in completions for '--'")
}

func TestCompleteInstructionsArtifactIDs(t *testing.T) {
	result := Complete("", []string{"instructions", ""})
	names := make(map[string]bool)
	for _, c := range result {
		names[c.Candidate] = true
	}
	for _, id := range []string{"proposal", "specs", "design", "tasks"} {
		if !names[id] {
			t.Errorf("missing artifact %q", id)
		}
	}
}

func TestCompleteCompletionShells(t *testing.T) {
	result := Complete("", []string{"completion", ""})
	if len(result) != 3 {
		t.Fatalf("expected 3 shells, got %d", len(result))
	}
	names := make(map[string]bool)
	for _, c := range result {
		names[c.Candidate] = true
	}
	for _, shell := range []string{"bash", "zsh", "fish"} {
		if !names[shell] {
			t.Errorf("missing shell %q", shell)
		}
	}
}

func TestCompleteStatusChangeNames(t *testing.T) {
	root := t.TempDir()
	changesDir := filepath.Join(root, "specs", "changes")
	if err := os.MkdirAll(changesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"foo", "bar"} {
		if err := os.MkdirAll(filepath.Join(changesDir, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	result := Complete(root, []string{"status", ""})
	names := make(map[string]bool)
	for _, c := range result {
		names[c.Candidate] = true
	}
	if !names["foo"] || !names["bar"] {
		t.Errorf("expected foo and bar, got %v", names)
	}
}

func TestCompleteArchiveChangeNames(t *testing.T) {
	root := t.TempDir()
	changesDir := filepath.Join(root, "specs", "changes")
	if err := os.MkdirAll(changesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(changesDir, "my-change"), 0o755); err != nil {
		t.Fatal(err)
	}

	result := Complete(root, []string{"archive", ""})
	found := false
	for _, c := range result {
		if c.Candidate == "my-change" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'my-change' in archive completions")
	}
}

func TestCompleteInitTools(t *testing.T) {
	result := Complete("", []string{"init", "--tools", ""})
	names := make(map[string]bool)
	for _, c := range result {
		names[c.Candidate] = true
	}
	if !names["claude"] {
		t.Error("expected 'claude' in tool completions")
	}
}

func TestCompleteValidateFlags(t *testing.T) {
	result := Complete("", []string{"validate", "--"})
	names := make(map[string]bool)
	for _, c := range result {
		names[c.Candidate] = true
	}
	for _, flag := range []string{"--all", "--changes", "--specs", "--strict", "--json", "--type"} {
		if !names[flag] {
			t.Errorf("missing flag %q in validate completions", flag)
		}
	}
}

func TestCompleteSortValues(t *testing.T) {
	result := Complete("", []string{"list", "--sort", ""})
	if len(result) != 3 {
		t.Fatalf("expected 3 sort values, got %d", len(result))
	}
	names := make(map[string]bool)
	for _, c := range result {
		names[c.Candidate] = true
	}
	if !names["recent"] || !names["name"] || !names["deps"] {
		t.Errorf("expected recent, name, and deps, got %v", names)
	}
}

func TestCompleteTypeValues(t *testing.T) {
	result := Complete("", []string{"validate", "--type", ""})
	if len(result) != 2 {
		t.Fatalf("expected 2 type values, got %d", len(result))
	}
	names := make(map[string]bool)
	for _, c := range result {
		names[c.Candidate] = true
	}
	if !names["change"] || !names["spec"] {
		t.Errorf("expected change and spec, got %v", names)
	}
}

func TestCompleteUnknownCommand(t *testing.T) {
	result := Complete("", []string{"nonexistent", ""})
	if len(result) != 0 {
		t.Errorf("expected no completions for unknown command, got %d", len(result))
	}
}

func TestCompleteHiddenCompleteCommand(t *testing.T) {
	result := Complete("", []string{"__complete", "something"})
	if len(result) != 0 {
		t.Errorf("expected no completions for __complete, got %d", len(result))
	}
}

func TestCompleteErrorSilentFallback(t *testing.T) {
	result := Complete("/nonexistent/path", []string{"status", ""})
	if len(result) != 0 {
		t.Errorf("expected empty completions on error, got %d", len(result))
	}
}

func TestCompleteUpdateTools(t *testing.T) {
	result := Complete("", []string{"update", "--tools", ""})
	found := false
	for _, c := range result {
		if c.Candidate == "claude" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'claude' in update --tools completions")
	}
}

func TestCompletionScriptValidShells(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish"} {
		script, err := CompletionScript(shell)
		if err != nil {
			t.Errorf("CompletionScript(%q) error: %v", shell, err)
		}
		if script == "" {
			t.Errorf("CompletionScript(%q) returned empty string", shell)
		}
	}
}

func TestCompletionScriptInvalidShell(t *testing.T) {
	_, err := CompletionScript("powershell")
	if err == nil {
		t.Error("expected error for invalid shell")
	}
}

func TestCompletionScriptContent(t *testing.T) {
	bash, _ := CompletionScript("bash")
	if !containsSubstring(bash, "_litespec") {
		t.Error("bash script missing _litespec function")
	}
	if !containsSubstring(bash, "complete -F") {
		t.Error("bash script missing complete -F")
	}

	zsh, _ := CompletionScript("zsh")
	if !containsSubstring(zsh, "#compdef litespec") {
		t.Error("zsh script missing #compdef")
	}

	fish, _ := CompletionScript("fish")
	if !containsSubstring(fish, "complete -c litespec") {
		t.Error("fish script missing complete -c litespec")
	}
}

func containsSubstring(s, sub string) bool {
	return len(s) >= len(sub) && searchString(s, sub)
}

func searchString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
