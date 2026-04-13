package main

import (
	"fmt"
	"os"

	"github.com/bermudi/litespec/internal"
)

func cmdCompletion(args []string) error {
	if len(args) == 0 {
		fmt.Print(`Usage: litespec completion <shell>

Supported shells: bash, zsh, fish

Loading completions:

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

`)
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
