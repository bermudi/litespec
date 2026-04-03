package main

import (
	"fmt"
	"os"

	"github.com/bermudi/litespec/internal"
)

func cmdCompletion(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: litespec completion <shell>\nSupported shells: bash, zsh, fish")
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
