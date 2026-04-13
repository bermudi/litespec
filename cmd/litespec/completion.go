package main

import (
	"fmt"
	"os"

	"github.com/bermudi/litespec/internal"
)

const completionHelp = `Usage: litespec completion <shell>

Supported shells: bash, zsh, fish

Persist completions (loaded on every new shell):

  Bash:
    litespec completion bash > ~/.local/share/bash-completion/completions/litespec
    # Or add to ~/.bashrc:
    #   eval "$(litespec completion bash)"

  Zsh:
    litespec completion zsh > ~/.zfunc/_litespec
    # Ensure ~/.zfunc is in your fpath (add to ~/.zshrc):
    #   fpath+=~/.zfunc
    #   autoload -Uz compinit && compinit

  Fish:
    litespec completion fish > ~/.config/fish/completions/litespec.fish

Load temporarily (current session only):

  Bash:   eval "$(litespec completion bash)"
  Zsh:    eval "$(litespec completion zsh)"
  Fish:   source (litespec completion fish | psub)

  Note: Zsh may need 'autoload -Uz compinit && compinit' first
  if completions are not already initialized in the session.
`

func cmdCompletion(args []string) error {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		fmt.Print(completionHelp)
		return nil
	}

	shell := args[0]
	script, err := internal.CompletionScript(shell)
	if err != nil {
		return err
	}

	fmt.Print(script)
	return nil
}

func cmdComplete() {
	root, _ := internal.FindProjectRoot()

	words := os.Args[2:]
	completions := internal.Complete(root, words)
	for _, c := range completions {
		fmt.Printf("%s\t%s\n", c.Candidate, c.Description)
	}
}
