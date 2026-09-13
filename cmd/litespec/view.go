package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bermudi/litespec/v2/internal"
)

func cmdView(args []string) error {
	fs := newFlagSet("view", printViewHelp)
	var asJSON, asMinimal bool
	fs.BoolVar(&asJSON, "json", false, "output as JSON")
	fs.BoolVar(&asMinimal, "minimal", false, "minimal output")

	ok, err := parseFlagSet(fs, args)
	if !ok {
		return err
	}

	root, err := requireProjectRootWithStaleCheck()
	if err != nil {
		return err
	}

	specs, err := internal.ListSpecs(root)
	if err != nil {
		return err
	}

	totalReqs := 0
	for _, s := range specs {
		totalReqs += s.RequirementCount
	}

	decisions, decErr := internal.ListDecisions(root)

	productContent, productErr := os.ReadFile(internal.ProductPath(root))
	ghIssues := fetchViewGHIssues(root)

	return Render(Response{
		Full:        buildViewFullJSON(root, specs, totalReqs, decisions, decErr, productContent, productErr, ghIssues),
		Minimal:     buildViewMinimalJSON(specs, totalReqs, decisions, decErr, ghIssues),
		Text:        buildViewText(root, specs, totalReqs, decisions, decErr, productContent, productErr, ghIssues),
		MinimalText: fmt.Sprintf("%d specs\t%d reqs\t%d decisions\t%d issues", len(specs), totalReqs, decisionCount(decisions, decErr), len(ghIssues)),
	}, asJSON, asMinimal)
}

func decisionCount(decisions []*internal.Decision, decErr error) int {
	if decErr != nil {
		return 0
	}
	count := 0
	for _, d := range decisions {
		if d.Status != internal.StatusSuperseded {
			count++
		}
	}
	return count
}

type viewGHIssueJSON struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	State  string `json:"state"`
	URL    string `json:"url"`
}

type viewJSON struct {
	Summary   viewSummaryJSON    `json:"summary"`
	Specs     []viewSpecJSON     `json:"specs,omitempty"`
	Decisions []viewDecisionJSON `json:"decisions,omitempty"`
	Product   *viewProductJSON   `json:"product,omitempty"`
	GHIssues  []viewGHIssueJSON  `json:"ghIssues,omitempty"`
}

type viewProductJSON struct {
	Path    string `json:"path"`
	Exists  bool   `json:"exists"`
	Preview string `json:"preview,omitempty"`
}

type viewSummaryJSON struct {
	Specs        int                    `json:"specs"`
	Requirements int                    `json:"requirements"`
	Decisions    *viewDecisionCountJSON `json:"decisions,omitempty"`
	GHIssues     int                    `json:"ghIssues,omitempty"`
}

type viewDecisionCountJSON struct {
	Active int `json:"active"`
	Total  int `json:"total"`
}

type viewSpecJSON struct {
	Name             string `json:"name"`
	RequirementCount int    `json:"requirementCount"`
}

type viewDecisionJSON struct {
	Number int    `json:"number"`
	Slug   string `json:"slug"`
	Status string `json:"status"`
	Spine  bool   `json:"spine,omitempty"`
}

func fetchViewGHIssues(root string) []viewGHIssueJSON {
	if _, err := exec.LookPath("gh"); err != nil {
		return nil
	}
	if _, err := os.Stat(filepath.Join(root, ".git")); err != nil {
		if out, err := exec.Command("git", "rev-parse", "--is-inside-work-tree").Output(); err != nil || strings.TrimSpace(string(out)) != "true" {
			return nil
		}
	}
	cmd := exec.Command("gh", "issue", "list", "--label", "litespec", "--json", "number,title,state,url", "--state", "open", "--limit", "10000")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	var issues []viewGHIssueJSON
	if err := json.Unmarshal(out, &issues); err != nil {
		return nil
	}
	return issues
}

func buildViewText(root string, specs []internal.SpecInfo, totalReqs int, decisions []*internal.Decision, decErr error, productContent []byte, productErr error, ghIssues []viewGHIssueJSON) string {
	var sb strings.Builder
	sep := strings.Repeat("═", 60)

	fmt.Fprintln(&sb)
	fmt.Fprintln(&sb, "Litespec Dashboard")
	fmt.Fprintln(&sb)
	fmt.Fprintln(&sb, sep)

	fmt.Fprintln(&sb, "Product:")
	if productErr == nil {
		preview := strings.TrimSpace(string(productContent))
		if len(preview) > 400 {
			preview = preview[:400] + "…"
		}
		firstLine := strings.Split(preview, "\n")[0]
		fmt.Fprintf(&sb, "  specs/product.md — %s\n", strings.TrimSpace(firstLine))
		fmt.Fprintln(&sb, "  product: mental models + flows")
	} else {
		fmt.Fprintln(&sb, "  specs/product.md — missing (run litespec init to scaffold)")
		fmt.Fprintln(&sb, "  product: not yet initialized")
	}

	fmt.Fprintln(&sb)
	fmt.Fprintln(&sb, "Summary:")
	fmt.Fprintf(&sb, "  ● Specifications: %d specs, %d requirements\n", len(specs), totalReqs)
	if decErr == nil && len(decisions) > 0 {
		activeDec := 0
		for _, d := range decisions {
			if d.Status != internal.StatusSuperseded {
				activeDec++
			}
		}
		fmt.Fprintf(&sb, "  ● Decisions: %d/%d\n", activeDec, len(decisions))
	}
	if len(ghIssues) > 0 {
		fmt.Fprintf(&sb, "  ● GH Issues: %d open\n", len(ghIssues))
	}

	if len(specs) > 0 {
		fmt.Fprintln(&sb)
		fmt.Fprintln(&sb, "Specifications")
		fmt.Fprintln(&sb, strings.Repeat("─", 60))
		sort.Slice(specs, func(i, j int) bool {
			return specs[i].RequirementCount > specs[j].RequirementCount
		})
		for _, s := range specs {
			label := "requirement"
			if s.RequirementCount != 1 {
				label = "requirements"
			}
			fmt.Fprintf(&sb, "  ▪ %-30s %d %s  (specs/%s/spec.md)\n", s.Name, s.RequirementCount, label, s.Name)
		}
	} else {
		fmt.Fprintln(&sb)
		fmt.Fprintln(&sb, "Specifications")
		fmt.Fprintln(&sb, strings.Repeat("─", 60))
		fmt.Fprintln(&sb, "  (no feature specs yet — add specs/<feature>/spec.md)")
	}

	if decErr == nil && len(decisions) > 0 {
		var activeDecs []*internal.Decision
		supersededCount := 0
		for _, d := range decisions {
			if d.Status != internal.StatusSuperseded {
				activeDecs = append(activeDecs, d)
			} else {
				supersededCount++
			}
		}
		sort.Slice(activeDecs, func(i, j int) bool {
			return activeDecs[i].Number < activeDecs[j].Number
		})
		fmt.Fprintln(&sb)
		fmt.Fprintln(&sb, "Decisions")
		fmt.Fprintln(&sb, strings.Repeat("─", 60))
		for _, d := range activeDecs {
			marker := " "
			if d.Spine {
				marker = "*"
			}
			fmt.Fprintf(&sb, "  %s%04d  %-30s  %s\n", marker, d.Number, d.Slug, d.Status)
		}
		if supersededCount > 0 {
			fmt.Fprintf(&sb, "  superseded: %d\n", supersededCount)
		}
	}

	if len(ghIssues) > 0 {
		fmt.Fprintln(&sb)
		fmt.Fprintln(&sb, "GH Issues (open)")
		fmt.Fprintln(&sb, strings.Repeat("─", 60))
		for _, iss := range ghIssues {
			fmt.Fprintf(&sb, "  #%-6d %-40s %s\n", iss.Number, iss.Title, iss.URL)
		}
	} else {
		if _, err := exec.LookPath("gh"); err != nil {
			fmt.Fprintln(&sb)
			fmt.Fprintln(&sb, "GH Issues")
			fmt.Fprintln(&sb, strings.Repeat("─", 60))
			fmt.Fprintln(&sb, "  (gh not available — showing local specs only)")
		}
	}

	fmt.Fprintln(&sb)
	fmt.Fprintln(&sb, sep)
	return sb.String()
}

func buildViewMinimalJSON(specs []internal.SpecInfo, totalReqs int, decisions []*internal.Decision, decErr error, ghIssues []viewGHIssueJSON) viewJSON {
	summary := viewSummaryJSON{
		Specs:        len(specs),
		Requirements: totalReqs,
		GHIssues:     len(ghIssues),
	}

	if decErr == nil && len(decisions) > 0 {
		activeDec := 0
		for _, d := range decisions {
			if d.Status != internal.StatusSuperseded {
				activeDec++
			}
		}
		summary.Decisions = &viewDecisionCountJSON{Active: activeDec, Total: len(decisions)}
	}

	return viewJSON{Summary: summary}
}

func buildViewFullJSON(root string, specs []internal.SpecInfo, totalReqs int, decisions []*internal.Decision, decErr error, productContent []byte, productErr error, ghIssues []viewGHIssueJSON) viewJSON {
	out := buildViewMinimalJSON(specs, totalReqs, decisions, decErr, ghIssues)

	var specEntries []viewSpecJSON
	for _, s := range specs {
		specEntries = append(specEntries, viewSpecJSON{Name: s.Name, RequirementCount: s.RequirementCount})
	}

	var decEntries []viewDecisionJSON
	if decErr == nil {
		var activeDecs []*internal.Decision
		for _, d := range decisions {
			if d.Status != internal.StatusSuperseded {
				activeDecs = append(activeDecs, d)
			}
		}
		sort.Slice(activeDecs, func(i, j int) bool {
			return activeDecs[i].Number < activeDecs[j].Number
		})
		for _, d := range activeDecs {
			decEntries = append(decEntries, viewDecisionJSON{Number: d.Number, Slug: d.Slug, Status: string(d.Status), Spine: d.Spine})
		}
	}

	out.Specs = specEntries
	out.Decisions = decEntries
	if productErr == nil {
		preview := strings.TrimSpace(string(productContent))
		if len(preview) > 500 {
			preview = preview[:500]
		}
		out.Product = &viewProductJSON{Path: internal.ProductPath(root), Exists: true, Preview: preview}
	} else {
		out.Product = &viewProductJSON{Path: internal.ProductPath(root), Exists: false}
	}
	out.GHIssues = ghIssues
	return out
}
