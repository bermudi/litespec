package internal

import (
	"embed"
	"strings"
)

type Completion struct {
	Candidate   string
	Description string
}

//go:embed completion/scripts/litespec.bash
//go:embed completion/scripts/litespec.zsh
//go:embed completion/scripts/litespec.fish
var completionScripts embed.FS

func Complete(root string, words []string) []Completion {
	words = stripProgramName(words)

	if len(words) == 0 {
		return completeCommands()
	}

	if len(words) == 1 {
		w := words[0]
		if strings.HasPrefix(w, "-") {
			return completeFlags(root, "")
		}
		if _, ok := commandFlagDefs[w]; ok {
			return completeCommandArgs(root, w, []string{""})
		}
		return filterCompletions(completeCommands(), w)
	}

	cmd := words[0]
	rest := words[1:]

	if cmd == "__complete" {
		return nil
	}

	return completeCommandArgs(root, cmd, rest)
}

func stripProgramName(words []string) []string {
	for _, w := range words {
		if w == "litespec" {
			return words[1:]
		}
		break
	}
	return words
}

func CompletionScript(shell string) (string, error) {
	var filename string
	switch shell {
	case "bash":
		filename = "completion/scripts/litespec.bash"
	case "zsh":
		filename = "completion/scripts/litespec.zsh"
	case "fish":
		filename = "completion/scripts/litespec.fish"
	default:
		return "", errInvalidShell
	}

	data, err := completionScripts.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

var errInvalidShell = invalidShellError{}

type invalidShellError struct{}

func (invalidShellError) Error() string {
	return "invalid shell (valid: bash, zsh, fish)"
}

func completeCommands() []Completion {
	return []Completion{
		{"init", "Initialize project structure"},
		{"new", "Create a new change"},
		{"list", "List specs or changes"},
		{"status", "Show artifact states"},
		{"validate", "Validate changes and specs"},
		{"instructions", "Get artifact instructions"},
		{"archive", "Apply deltas and archive change"},
		{"view", "Dashboard overview with dependency graph"},
		{"decide", "Create a new architectural decision record"},
		{"update", "Regenerate skills and adapters"},
		{"completion", "Generate shell completion script"},
	}
}

func completeArtifactIDs() []Completion {
	return []Completion{
		{"proposal", "Why and what — motivation, scope, approach"},
		{"specs", "Delta specifications — ADDED/MODIFIED/REMOVED/RENAMED"},
		{"design", "How — technical approach, architecture decisions"},
		{"tasks", "What to do — phased implementation checklist"},
	}
}

func completeShells() []Completion {
	return []Completion{
		{"bash", "Bash completion"},
		{"zsh", "Zsh completion"},
		{"fish", "Fish completion"},
	}
}

func completeToolIDs() []Completion {
	var result []Completion
	for _, a := range Adapters {
		result = append(result, Completion{Candidate: a.ID, Description: a.Name})
	}
	return result
}

func completeChangeNames(root string) []Completion {
	changes, err := ListChanges(root)
	if err != nil {
		return nil
	}
	var result []Completion
	for _, c := range changes {
		result = append(result, Completion{Candidate: c.Name, Description: "change"})
	}
	return result
}

func completeSpecNames(root string) []Completion {
	specs, err := ListSpecs(root)
	if err != nil {
		return nil
	}
	var result []Completion
	for _, s := range specs {
		result = append(result, Completion{Candidate: s.Name, Description: "spec"})
	}
	return result
}

type commandFlags struct {
	flags         map[string]string
	hasPositional bool
	posResolver   func(root string) []Completion
}

var commandFlagDefs = map[string]commandFlags{
	"init": {
		flags: map[string]string{
			"--tools": "Tool IDs (comma-separated)",
		},
	},
	"new": {
		hasPositional: true,
	},
	"list": {
		flags: map[string]string{
			"--specs":     "List specs instead of changes",
			"--changes":   "List changes (default)",
			"--decisions": "List architectural decision records",
			"--sort":      "Sort by 'recent', 'name', 'deps', or 'number'",
			"--status":    "Filter decisions by status (requires --decisions)",
			"--json":      "Output as JSON",
		},
	},
	"status": {
		flags: map[string]string{
			"--json": "Output as JSON",
		},
		hasPositional: true,
		posResolver:   completeChangeNames,
	},
	"validate": {
		flags: map[string]string{
			"--all":       "Validate all changes, specs, and decisions",
			"--changes":   "Validate all changes only",
			"--specs":     "Validate all specs only",
			"--decisions": "Validate all decisions only",
			"--strict":    "Treat warnings as errors",
			"--json":      "Output as JSON",
			"--type":      "Disambiguate name: change|spec|decision",
		},
	},
	"instructions": {
		flags: map[string]string{
			"--json": "Output as JSON",
		},
		hasPositional: true,
		posResolver:   func(root string) []Completion { return completeArtifactIDs() },
	},
	"archive": {
		flags: map[string]string{
			"--allow-incomplete": "Archive even with incomplete tasks or unarchived dependencies",
		},
		hasPositional: true,
		posResolver:   completeChangeNames,
	},
	"decide": {
		hasPositional: true,
	},
	"update": {
		flags: map[string]string{
			"--tools": "Tool IDs (comma-separated)",
		},
	},
	"view": {},
	"completion": {
		hasPositional: true,
		posResolver:   func(root string) []Completion { return completeShells() },
	},
}

func completeFlags(root string, cmd string) []Completion {
	if cmd == "" {
		return completeGlobalFlags()
	}

	def, ok := commandFlagDefs[cmd]
	if !ok {
		return completeGlobalFlags()
	}

	var result []Completion
	for f, desc := range def.flags {
		result = append(result, Completion{Candidate: f, Description: desc})
	}
	return result
}

func completeGlobalFlags() []Completion {
	return []Completion{
		{"--version", "Print version"},
		{"--help", "Print help message"},
	}
}

func completeCommandArgs(root string, cmd string, rest []string) []Completion {
	def, ok := commandFlagDefs[cmd]
	if !ok {
		return nil
	}

	lastIdx := len(rest) - 1
	last := rest[lastIdx]

	if strings.HasPrefix(last, "-") {
		if _, hasFlagArg := def.flags[last]; hasFlagArg && flagTakesValue(cmd, last) {
			return completeFlagValue(root, cmd, last)
		}
		var result []Completion
		for f, desc := range def.flags {
			if strings.HasPrefix(f, last) {
				result = append(result, Completion{Candidate: f, Description: desc})
			}
		}
		return result
	}

	prevWord := ""
	if lastIdx > 0 {
		prevWord = rest[lastIdx-1]
	}

	if prevWord != "" && flagTakesValue(cmd, prevWord) {
		return filterCompletions(completeFlagValue(root, cmd, prevWord), last)
	}

	if def.hasPositional && def.posResolver != nil {
		completedPositionals := countPositionalArgs(rest[:lastIdx], cmd)
		if completedPositionals == 0 {
			return filterCompletions(def.posResolver(root), last)
		}
	}

	return nil
}

func flagTakesValue(cmd string, flag string) bool {
	switch flag {
	case "--tools":
		return true
	case "--sort":
		return true
	case "--type":
		return true
	case "--status":
		return true
	}
	return false
}

func completeFlagValue(root string, cmd string, flag string) []Completion {
	switch flag {
	case "--tools":
		return completeToolIDs()
	case "--sort":
		return []Completion{
			{"recent", "Sort by last modified"},
			{"name", "Sort alphabetically"},
			{"deps", "Sort by dependency order"},
			{"number", "Sort by decision number"},
		}
	case "--type":
		return []Completion{
			{"change", "Disambiguate as change"},
			{"spec", "Disambiguate as spec"},
			{"decision", "Disambiguate as decision"},
		}
	case "--status":
		return []Completion{
			{"proposed", "Proposed decisions"},
			{"accepted", "Accepted decisions"},
			{"superseded", "Superseded decisions"},
		}
	}
	return nil
}

func countPositionalArgs(rest []string, cmd string) int {
	count := 0
	for i, w := range rest {
		if w == "" || strings.HasPrefix(w, "-") {
			continue
		}
		if i > 0 {
			prev := rest[i-1]
			if flagTakesValue(cmd, prev) {
				continue
			}
		}
		count++
	}
	return count
}

func filterCompletions(candidates []Completion, prefix string) []Completion {
	var result []Completion
	for _, c := range candidates {
		if strings.HasPrefix(c.Candidate, prefix) {
			result = append(result, c)
		}
	}
	return result
}
