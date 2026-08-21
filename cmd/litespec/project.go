package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bermudi/litespec/v2/internal"
)

func requireProjectRoot() (string, error) {
	root, err := internal.FindProjectRoot()
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(filepath.Join(root, internal.ProjectDirName)); err != nil {
		return "", fmt.Errorf("not a litespec project. Run 'litespec init' first")
	}
	return root, nil
}

func requireProjectRootWithStaleCheck() (string, error) {
	root, err := requireProjectRoot()
	if err != nil {
		return "", err
	}
	if warn := internal.CheckStaleSkills(root); warn != "" {
		fmt.Fprintln(os.Stderr, "WARN", warn)
	}
	return root, nil
}
