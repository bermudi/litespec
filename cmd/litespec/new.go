package main

import (
	"fmt"

	"github.com/bermudi/litespec/internal"
)

func cmdNew(args []string) error {
	if hasHelpFlag(args) {
		printNewHelp()
		return nil
	}
	if err := checkUnknownFlags(args, map[string]bool{}); err != nil {
		return err
	}
	if len(args) > 1 {
		return fmt.Errorf("unexpected arguments. Usage: litespec new <name>")
	}

	if len(args) == 0 {
		return fmt.Errorf("usage: litespec new <change-name>")
	}

	name := args[0]
	if err := validateChangeName(name); err != nil {
		return err
	}

	root, err := internal.FindProjectRoot()
	if err != nil {
		return err
	}

	if err := internal.CreateChange(root, name); err != nil {
		return err
	}

	fmt.Println(internal.ChangePath(root, name))
	return nil
}
