package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bermudi/litespec/internal"
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
		return fmt.Errorf("usage: litespec new <change-name> [--issue N]")
	}
	if len(positional) > 1 {
		return fmt.Errorf("unexpected arguments. Usage: litespec new <name> [--issue N]")
	}
	name := positional[0]

	if err := validateChangeName(name); err != nil {
		return err
	}

	root, err := requireProjectRootWithStaleCheck()
	if err != nil {
		return err
	}

	if issue != 0 {
		if issue < 1 {
			return fmt.Errorf("--issue must be a positive integer, got %d", issue)
		}

		if _, lookErr := exec.LookPath("gh"); lookErr == nil {
			cmd := exec.Command("gh", "issue", "view", fmt.Sprintf("%d", issue), "--json", "number,title,url")
			cmd.Dir = root
			if out, viewErr := cmd.CombinedOutput(); viewErr != nil {
				fmt.Fprintf(os.Stderr, "WARN could not verify GH issue #%d: %v\n%s\n", issue, viewErr, strings.TrimSpace(string(out)))
			}
		}

		issueURL := githubIssueURL(root, issue)

		type newLinkedJSON struct {
			ChangeName string `json:"changeName"`
			Issue      int    `json:"issue"`
			IssueURL   string `json:"issueUrl,omitempty"`
			Linked     bool   `json:"linked"`
			IsComplete bool   `json:"isComplete"`
		}
		type newLinkedMinimalJSON struct {
			ChangeName string `json:"changeName"`
			Issue      int    `json:"issue"`
			IsComplete bool   `json:"isComplete"`
		}

		full := newLinkedJSON{
			ChangeName: name,
			Issue:      issue,
			IssueURL:   issueURL,
			Linked:     true,
			IsComplete: false,
		}
		minimal := newLinkedMinimalJSON{
			ChangeName: name,
			Issue:      issue,
			IsComplete: false,
		}

		var textSB strings.Builder
		textSB.WriteString(fmt.Sprintf("Linked: %s -> GH issue #%d\n", name, issue))
		if issueURL != "" {
			textSB.WriteString(fmt.Sprintf("Issue: %s\n", issueURL))
		}
		textSB.WriteString("\nGH issue body is proposal + design + queue (v2 lean, 64k limit).\n")
		textSB.WriteString("No folder created in lean — GH issue is the queue.\n")
		textSB.WriteString("\nTemplate for GH issue body:\n")
		textSB.WriteString(fmt.Sprintf("## Proposal for %s\n...\n\n## Design\n...\n\n## Queue\n\n## <outcome>\nDone means: ...\nVerify: ```bash\n...\n```\n- [ ] pending\n", name))
		if _, lookErr := exec.LookPath("gh"); lookErr != nil {
			textSB.WriteString("\nOffline fallback when gh unavailable: specs/changes/<name>/QUEUE.md\n")
		}

		return Render(Response{
			Full:        full,
			Minimal:     minimal,
			Text:        textSB.String(),
			MinimalText: fmt.Sprintf("%s -> #%d", name, issue),
		}, asJSON, asMinimal)
	}

	changeDir := internal.ChangePath(root, name)
	if _, statErr := os.Stat(changeDir); statErr == nil {
		return fmt.Errorf("change %q already exists", name)
	}

	if err := internal.CreateChange(root, name); err != nil {
		return err
	}

	queuePath := filepath.Join(changeDir, "QUEUE.md")
	if _, statErr := os.Stat(queuePath); os.IsNotExist(statErr) {
		queueTemplate := fmt.Sprintf("# %s\n\nOffline fallback — GH unavailable. Queue lives here. Sync to GH issue when online.\n\n## Proposal\n\n...\n\n## Design\n\n...\n\n## Queue\n\n## <outcome>\nDone means: ...\nVerify: ```bash\n...\n```\n- [ ] pending\n", name)
		if writeErr := os.WriteFile(queuePath, []byte(queueTemplate), 0o644); writeErr != nil {
			fmt.Fprintf(os.Stderr, "WARN could not write QUEUE.md: %v\n", writeErr)
		}
	}

	ctx, err := internal.LoadChangeContext(root, name)
	if err != nil {
		return err
	}

	status := internal.BuildChangeStatusJSON(ctx)

	type newMinimalJSON struct {
		ChangeName string `json:"changeName"`
		IsComplete bool   `json:"isComplete"`
	}

	var textSB strings.Builder
	textSB.WriteString(fmt.Sprintf("Created: %s\n\n", internal.ChangePath(root, name)))
	textSB.WriteString("Artifacts:\n")
	for _, art := range internal.Artifacts {
		state := ctx.Artifacts[art.ID]
		var deps string
		if len(art.Requires) > 0 {
			deps = fmt.Sprintf(" (needs: %s)", strings.Join(art.Requires, ", "))
		}
		textSB.WriteString(fmt.Sprintf("  %-12s %-10s %s%s\n", art.ID, state, art.Filename, deps))
	}
	textSB.WriteString("\nOffline fallback: QUEUE.md created. When online, move queue to GH issue body.\n")
	textSB.WriteString("Use 'litespec instructions <artifact>' for per-artifact guidance.\n")

	return Render(Response{
		Full:        status,
		Minimal:     newMinimalJSON{ChangeName: status.ChangeName, IsComplete: status.IsComplete},
		Text:        textSB.String(),
		MinimalText: internal.ChangePath(root, name),
	}, asJSON, asMinimal)
}

func githubIssueURL(root string, issue int) string {
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return fmt.Sprintf("GH issue #%d", issue)
	}
	remote := strings.TrimSpace(string(out))
	if remote == "" {
		return fmt.Sprintf("GH issue #%d", issue)
	}
	var hostPath string
	if strings.HasPrefix(remote, "git@") {
		parts := strings.SplitN(remote, ":", 2)
		if len(parts) == 2 {
			hostPath = parts[1]
			hostPath = strings.TrimSuffix(hostPath, ".git")
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
