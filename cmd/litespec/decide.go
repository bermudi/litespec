package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bermudi/litespec/internal"
)

func cmdDecide(args []string) error {
	if hasHelpFlag(args) {
		printDecideHelp()
		return nil
	}
	if err := checkUnknownFlags(args, map[string]bool{"--json": true, "--minimal": true}); err != nil {
		return err
	}

	asJSON, asMinimal := parseOutputFlags(args)

	if len(args) == 0 {
		return fmt.Errorf("usage: litespec decide <slug>")
	}

	slug := args[0]
	if err := validateDecisionSlug(slug); err != nil {
		return err
	}

	root, err := internal.FindProjectRoot()
	if err != nil {
		return err
	}

	if _, err := os.Stat(filepath.Join(root, internal.ProjectDirName)); err != nil {
		return fmt.Errorf("not a litespec project. Run 'litespec init' first")
	}

	decisions, err := internal.ListDecisions(root)
	if err != nil {
		return err
	}

	nextNum := 1
	for _, d := range decisions {
		if d.Number >= nextNum {
			nextNum = d.Number + 1
		}
		if d.Slug == slug {
			return fmt.Errorf("decision with slug %q already exists (%04d-%s)", slug, d.Number, d.Slug)
		}
	}

	filename := fmt.Sprintf("%04d-%s.md", nextNum, slug)
	decisionsDir := internal.DecisionsPath(root)
	if err := os.MkdirAll(decisionsDir, 0o755); err != nil {
		return fmt.Errorf("create decisions directory: %w", err)
	}

	content := fmt.Sprintf(`# %s

## Status

proposed

## Context

<!-- What forces are at play? What constraints apply? -->

## Decision

<!-- What we decided and why. Use SHALL/MUST where intent is normative. -->

## Consequences

<!-- What becomes easier? What becomes harder? What must change elsewhere? -->
`, slugToTitle(slug))

	filePath := filepath.Join(decisionsDir, filename)
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write decision file: %w", err)
	}

	if asJSON {
		type decideResultJSON struct {
			Number   int    `json:"number"`
			Slug     string `json:"slug"`
			Title    string `json:"title"`
			FilePath string `json:"filePath"`
		}
		out := decideResultJSON{
			Number:   nextNum,
			Slug:     slug,
			Title:    slugToTitle(slug),
			FilePath: filePath,
		}
		if asMinimal {
			type decideMinimalJSON struct {
				Number   int    `json:"number"`
				Slug     string `json:"slug"`
				FilePath string `json:"filePath"`
			}
			data, err := internal.MarshalJSON(decideMinimalJSON{Number: nextNum, Slug: slug, FilePath: filePath})
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}
		data, err := internal.MarshalJSON(out)
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	if asMinimal {
		fmt.Println(filePath)
		return nil
	}

	fmt.Printf("Created: %s\n", filePath)
	return nil
}

func validateDecisionSlug(slug string) error {
	if slug == "" {
		return fmt.Errorf("slug cannot be empty")
	}
	if len(slug) < 2 {
		return fmt.Errorf("slug must be at least 2 characters")
	}
	if strings.HasPrefix(slug, "-") || strings.HasSuffix(slug, "-") {
		return fmt.Errorf("slug must not start or end with a hyphen")
	}
	for _, c := range slug {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return fmt.Errorf("slug must contain only lowercase letters, digits, and hyphens (got %q)", slug)
		}
	}
	if strings.Contains(slug, "--") {
		return fmt.Errorf("slug must not contain consecutive hyphens")
	}
	return nil
}

func slugToTitle(slug string) string {
	parts := strings.Split(slug, "-")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}

func printDecideHelp() {
	fmt.Print(`Usage: litespec decide <slug> [--json] [--minimal]

Create a new architectural decision record.

The decision is created in specs/decisions/ with a scaffolded structure.
Status starts as "proposed". Edit the file to change status and fill in sections.

Arguments:
  <slug>            Decision slug (lowercase, hyphens, no spaces)

Flags:
  --json            Output as JSON
  --minimal         Minimal output

Examples:
  litespec decide single-shared-workspace
  litespec decide beta-tool-binding --json
`)
}
