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
	fs := newFlagSet("patch", printPatchHelp)
	var asJSON, asMinimal bool
	fs.BoolVar(&asJSON, "json", false, "output as JSON")
	fs.BoolVar(&asMinimal, "minimal", false, "minimal output")

	ok, err := parseFlagSet(fs, args)
	if !ok {
		return err
	}

	positional := fs.Args()
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

	root, err := requireProjectRoot()
	if err != nil {
		return err
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

	status := internal.BuildChangeStatusJSON(ctx)

	type patchMinimalJSON struct {
		ChangeName string `json:"changeName"`
		IsComplete  bool   `json:"isComplete"`
	}

	var textSB strings.Builder
	textSB.WriteString(fmt.Sprintf("Created: %s (patch mode)\n\n", changeDir))
	textSB.WriteString("Artifacts:\n")
	textSB.WriteString(fmt.Sprintf("  %-12s DONE       specs/%s/spec.md\n", "specs", capability))
	textSB.WriteString("\nWrite your delta spec, implement, then archive.\n")
	textSB.WriteString("Use 'litespec validate' to check your delta, 'litespec archive' to commit to canon.\n")

	return Render(Response{
		Full:        status,
		Minimal:     patchMinimalJSON{ChangeName: status.ChangeName, IsComplete: status.IsComplete},
		Text:        textSB.String(),
		MinimalText: changeDir,
	}, asJSON, asMinimal)
}
