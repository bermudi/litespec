package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bermudi/litespec/internal"
)

func cmdPatch(args []string) error {
	if hasHelpFlag(args) {
		printPatchHelp()
		return nil
	}
	if err := checkUnknownFlags(args, map[string]bool{"--json": true}); err != nil {
		return err
	}

	var asJSON bool
	var positional []string
	for _, arg := range args {
		switch arg {
		case jsonFlag:
			asJSON = true
		default:
			if !strings.HasPrefix(arg, "-") {
				positional = append(positional, arg)
			}
		}
	}

	if len(positional) != 2 {
		return fmt.Errorf("usage: litespec patch <name> <capability>")
	}

	name := positional[0]
	capability := positional[1]

	if err := validateChangeName(name); err != nil {
		return err
	}

	if err := validateChangeName(capability); err != nil {
		return fmt.Errorf("invalid capability name: %w", err)
	}

	root, err := internal.FindProjectRoot()
	if err != nil {
		return err
	}

	if _, err := os.Stat(filepath.Join(root, internal.ProjectDirName)); err != nil {
		return fmt.Errorf("not a litespec project. Run 'litespec init' first")
	}

	changeDir := internal.ChangePath(root, name)
	if _, err := os.Stat(changeDir); err == nil {
		return fmt.Errorf("change %q already exists", name)
	}

	specDir := filepath.Join(changeDir, "specs", capability)
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		return fmt.Errorf("create change directory: %w", err)
	}

	stub := fmt.Sprintf("# %s\n\n## ADDED Requirements\n", capability)
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(stub), 0o644); err != nil {
		return fmt.Errorf("write spec stub: %w", err)
	}

	meta := internal.ChangeMeta{
		Schema:  "spec-driven",
		Created: time.Now().UTC().Truncate(time.Second),
		Mode:    "patch",
	}
	if err := internal.WriteChangeMeta(root, name, &meta); err != nil {
		return err
	}

	ctx, err := internal.LoadChangeContext(root, name)
	if err != nil {
		return err
	}

	if asJSON {
		status := internal.BuildChangeStatusJSON(ctx)
		data, err := internal.MarshalJSON(status)
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	fmt.Printf("Created: %s (patch mode)\n\n", changeDir)
	fmt.Println("Artifacts:")
	fmt.Printf("  %-12s DONE       specs/%s/spec.md\n", "specs", capability)
	fmt.Println("\nWrite your delta spec, implement, then archive.")
	fmt.Println("Use 'litespec validate' to check your delta, 'litespec archive' to commit to canon.")
	return nil
}
