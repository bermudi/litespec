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
			return completeFlags("")
		}
		if spec := findCommandSpec(w); spec != nil {
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

func findCommandSpec(name string) *CommandSpec {
	for i := range CommandSpecs {
		if CommandSpecs[i].Name == name {
			return &CommandSpecs[i]
		}
	}
	return nil
}

func completeCommands() []Completion {
	var result []Completion
	for _, c := range CommandSpecs {
		if c.Hidden {
			continue
		}
		result = append(result, Completion{c.Name, c.Description})
	}
	return result
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

func completeFlags(cmd string) []Completion {
	if cmd == "" {
		return completeGlobalFlags()
	}

	spec := findCommandSpec(cmd)
	if spec == nil {
		return completeGlobalFlags()
	}

	var result []Completion
	for _, f := range spec.Flags {
		result = append(result, Completion{f.Name, f.Description})
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
	spec := findCommandSpec(cmd)
	if spec == nil {
		return nil
	}

	lastIdx := len(rest) - 1
	last := rest[lastIdx]

	if strings.HasPrefix(last, "-") {
		if flag := findFlagSpec(spec, last); flag != nil && flag.TakesValue {
			return resolveFlagValues(root, flag)
		}
		var result []Completion
		for _, f := range spec.Flags {
			if strings.HasPrefix(f.Name, last) {
				result = append(result, Completion{f.Name, f.Description})
			}
		}
		return result
	}

	prevWord := ""
	if lastIdx > 0 {
		prevWord = rest[lastIdx-1]
	}

	if prevWord != "" {
		if flag := findFlagSpec(spec, prevWord); flag != nil && flag.TakesValue {
			return filterCompletions(resolveFlagValues(root, flag), last)
		}
	}

	if spec.Positional != nil && spec.Positional.Resolver != nil {
		completedPositionals := countPositionalArgs(rest[:lastIdx], spec)
		if completedPositionals == 0 {
			return filterCompletions(spec.Positional.Resolver(root), last)
		}
	}

	return nil
}

func findFlagSpec(spec *CommandSpec, name string) *FlagSpec {
	for i := range spec.Flags {
		if spec.Flags[i].Name == name {
			return &spec.Flags[i]
		}
	}
	return nil
}

func resolveFlagValues(root string, flag *FlagSpec) []Completion {
	if flag.ValuesFunc != nil {
		return flag.ValuesFunc(root)
	}
	return flag.Values
}

func countPositionalArgs(rest []string, spec *CommandSpec) int {
	count := 0
	for i, w := range rest {
		if w == "" || strings.HasPrefix(w, "-") {
			continue
		}
		if i > 0 {
			prev := rest[i-1]
			if flag := findFlagSpec(spec, prev); flag != nil && flag.TakesValue {
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
