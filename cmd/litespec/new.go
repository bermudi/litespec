package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func cmdNew(args []string) error {
	fs := newFlagSet("new", printNewHelp)
	var asJSON, asMinimal bool
	var issue int
	fs.BoolVar(&asJSON, "json", false, "output as JSON")
	fs.BoolVar(&asMinimal, "minimal", false, "minimal output")
	fs.IntVar(&issue, "issue", 0, "link to GH issue number")

	ok, err := parseFlagSet(fs, args)
	if !ok {
		return err
	}

	positional := fs.Args()
	if len(positional) == 0 {
		return fmt.Errorf("usage: litespec new <change-name> --issue N")
	}
	if len(positional) > 1 {
		return fmt.Errorf("unexpected arguments. Usage: litespec new <name> --issue N")
	}
	name := positional[0]

	if err := validateChangeName(name); err != nil {
		return err
	}

	if issue == 0 {
		return fmt.Errorf("--issue is required (links the change name to the GH issue)")
	}
	if issue < 1 {
		return fmt.Errorf("--issue must be a positive integer, got %d", issue)
	}

	issueURL := githubIssueURL(issue)

	type newLinkedJSON struct {
		ChangeName string `json:"changeName"`
		Issue      int    `json:"issue"`
		IssueURL   string `json:"issueUrl,omitempty"`
		Linked     bool   `json:"linked"`
	}
	type newLinkedMinimalJSON struct {
		ChangeName string `json:"changeName"`
		Issue      int    `json:"issue"`
	}

	full := newLinkedJSON{
		ChangeName: name,
		Issue:      issue,
		IssueURL:   issueURL,
		Linked:     true,
	}
	minimal := newLinkedMinimalJSON{
		ChangeName: name,
		Issue:      issue,
	}

	var textSB strings.Builder
	textSB.WriteString(fmt.Sprintf("Linked: %s -> GH issue #%d\n", name, issue))
	if issueURL != "" {
		textSB.WriteString(fmt.Sprintf("Issue: %s\n", issueURL))
	}
	textSB.WriteString("\nGH issue body is proposal + design + queue (64k limit).\n")
	textSB.WriteString("No folder created — GH issue is the queue.\n")
	textSB.WriteString("Add the `litespec` label to the issue so `validate` discovers it.\n")
	textSB.WriteString("\nTemplate for GH issue body:\n")
	textSB.WriteString(fmt.Sprintf("## Proposal for %s\n...\n\n## Design\n...\n\n## Queue\n\n## <outcome>\nDone means: ...\nVerify: ```bash\n...\n```\n- [ ] pending\n", name))

	return Render(Response{
		Full:        full,
		Minimal:     minimal,
		Text:        textSB.String(),
		MinimalText: fmt.Sprintf("%s -> #%d", name, issue),
	}, asJSON, asMinimal)
}

func githubIssueURL(issue int) string {
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Sprintf("GH issue #%d", issue)
	}
	remote := strings.TrimSpace(string(out))
	if remote == "" {
		return fmt.Sprintf("GH issue #%d", issue)
	}
	if strings.HasPrefix(remote, "git@") {
		parts := strings.SplitN(remote, ":", 2)
		if len(parts) == 2 {
			hostPath := strings.TrimSuffix(parts[1], ".git")
			host := strings.TrimPrefix(parts[0], "git@")
			if strings.Contains(host, "github.com") {
				return fmt.Sprintf("https://github.com/%s/issues/%d", hostPath, issue)
			}
		}
	} else if strings.HasPrefix(remote, "https://") || strings.HasPrefix(remote, "http://") {
		remote = strings.TrimSuffix(remote, ".git")
		if strings.Contains(remote, "github.com") {
			return fmt.Sprintf("%s/issues/%d", remote, issue)
		}
	}
	return fmt.Sprintf("GH issue #%d", issue)
}
