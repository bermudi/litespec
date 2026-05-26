package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bermudi/litespec/internal"
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
