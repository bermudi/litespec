package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bermudi/litespec/internal"
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

	changes, err := internal.ListChanges(root)
	if err != nil {
		return err
	}

	var draft, active, completed, patch []internal.ChangeInfo
	for _, c := range changes {
		if internal.IsPatchMode(root, c.Name) {
			patch = append(patch, c)
		} else if c.TotalTasks == 0 {
			draft = append(draft, c)
		} else if c.CompletedTasks == c.TotalTasks {
			completed = append(completed, c)
		} else {
			active = append(active, c)
		}
	}

	sort.Slice(active, func(i, j int) bool {
		pctI := float64(active[i].CompletedTasks) / float64(active[i].TotalTasks)
		pctJ := float64(active[j].CompletedTasks) / float64(active[j].TotalTasks)
		if pctI != pctJ {
			return pctI < pctJ
		}
		return active[i].Name < active[j].Name
	})

	totalReqs := 0
	for _, s := range specs {
		totalReqs += s.RequirementCount
	}

	totalCompletedTasks := 0
	totalTasks := 0
	for _, c := range active {
		totalCompletedTasks += c.CompletedTasks
		totalTasks += c.TotalTasks
	}

	decisions, decErr := internal.ListDecisions(root)
	depMap, _ := internal.LoadDepMap(root)

	productContent, productErr := os.ReadFile(internal.ProductPath(root))
	ghIssues := fetchViewGHIssues(root)

	return Render(Response{
		Full:        buildViewFullJSON(root, specs, draft, active, completed, patch, totalReqs, totalCompletedTasks, totalTasks, decisions, decErr, depMap, productContent, productErr, ghIssues),
		Minimal:     buildViewMinimalJSON(root, specs, draft, active, completed, patch, totalReqs, totalCompletedTasks, totalTasks, decisions, decErr),
		Text:        buildViewText(root, specs, draft, active, completed, patch, totalReqs, totalCompletedTasks, totalTasks, decisions, decErr, changes, depMap, productContent, productErr, ghIssues),
		MinimalText: fmt.Sprintf("%d specs\t%d reqs\t%d draft\t%d active\t%d ready\t%d/%d tasks", len(specs), totalReqs, len(draft), len(active), len(completed), totalCompletedTasks, totalTasks),
	}, asJSON, asMinimal)
}

func formatCount(n int, label string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, label)
	}
	return fmt.Sprintf("%d %s", n, label)
}

func formatTimestamps(c internal.ChangeInfo) string {
	var parts []string
	if !c.Created.IsZero() {
		parts = append(parts, "born "+c.Created.Format("2006-01-02"))
	}
	if !c.LastModified.IsZero() {
		parts = append(parts, "touched "+internal.FormatRelativeTime(c.LastModified))
	}
	if len(parts) == 0 {
		return ""
	}
	return "  (" + strings.Join(parts, ", ") + ")"
}

func createProgressBar(completed, total, width int) string {
	if total == 0 {
		return strings.Repeat("─", width)
	}
	pct := float64(completed) / float64(total)
	filled := int(math.Round(pct * float64(width)))
	empty := width - filled
	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", empty) + "]"
}

func renderDependencyGraph(w io.Writer, depMap map[string][]string, changes []internal.ChangeInfo) {
	changeMap := make(map[string]internal.ChangeInfo)
	for _, c := range changes {
		changeMap[c.Name] = c
	}

	activeNames := make(map[string]bool)
	for _, c := range changes {
		activeNames[c.Name] = true
	}

	var sb strings.Builder

	reverseMap := make(map[string][]string)
	for name, deps := range depMap {
		for _, dep := range deps {
			reverseMap[dep] = append(reverseMap[dep], name)
		}
	}

	related := make(map[string]bool)
	for name := range activeNames {
		if len(depMap[name]) > 0 || len(reverseMap[name]) > 0 {
			related[name] = true
		}
		for _, dep := range depMap[name] {
			if activeNames[dep] {
				related[dep] = true
			}
		}
	}

	var unrelated []string
	for name := range activeNames {
		if !related[name] {
			unrelated = append(unrelated, name)
		}
	}

	if len(related) == 0 {
		if len(unrelated) > 0 {
			sort.Strings(unrelated)
			sb.WriteString("\nUnrelated:\n")
			for _, name := range unrelated {
				fmt.Fprintf(&sb, "  - %s%s\n", name, formatTimestamps(changeMap[name]))
			}
		}
		fmt.Fprint(w, sb.String())
		return
	}

	var roots []string
	for name := range related {
		if len(depMap[name]) == 0 {
			roots = append(roots, name)
		}
	}

	sort.Strings(roots)

	seen := make(map[string]bool)

	var printNode func(name string, prefix string, isLast bool)
	printNode = func(name string, prefix string, isLast bool) {
		if seen[name] {
			return
		}
		seen[name] = true

		connector := "├── "
		if isLast {
			connector = "└── "
		}
		fmt.Fprintf(&sb, "%s%s%s%s\n", prefix, connector, name, formatTimestamps(changeMap[name]))

		children := reverseMap[name]
		sort.Strings(children)
		for i, child := range children {
			newPrefix := prefix
			if isLast {
				newPrefix += "    "
			} else {
				newPrefix += "│   "
			}
			printNode(child, newPrefix, i == len(children)-1)
		}
	}

	for i, root := range roots {
		printNode(root, "", i == len(roots)-1)
	}

	if len(unrelated) > 0 {
		sort.Strings(unrelated)
		sb.WriteString("\nUnrelated:\n")
		for _, name := range unrelated {
			fmt.Fprintf(&sb, "  - %s%s\n", name, formatTimestamps(changeMap[name]))
		}
	}

	fmt.Fprint(w, sb.String())
}

type viewGHIssueJSON struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	State  string `json:"state"`
	URL    string `json:"url"`
}

type viewJSON struct {
	Summary   viewSummaryJSON     `json:"summary"`
	Changes   []viewChangeJSON    `json:"changes,omitempty"`
	Specs     []viewSpecJSON      `json:"specs,omitempty"`
	Decisions []viewDecisionJSON  `json:"decisions,omitempty"`
	Graph     *viewGraphJSON      `json:"graph,omitempty"`
	Product   *viewProductJSON    `json:"product,omitempty"`
	GHIssues  []viewGHIssueJSON   `json:"ghIssues,omitempty"`
}

type viewProductJSON struct {
	Path    string `json:"path"`
	Exists  bool   `json:"exists"`
	Preview string `json:"preview,omitempty"`
}

type viewSummaryJSON struct {
	Specs            int                    `json:"specs"`
	Requirements     int                    `json:"requirements"`
	DraftChanges     int                    `json:"draftChanges"`
	ActiveChanges    int                    `json:"activeChanges"`
	PatchChanges     int                    `json:"patchChanges,omitempty"`
	ReadyToArchive   int                    `json:"readyToArchive"`
	TaskProgress     *viewProgressJSON      `json:"taskProgress,omitempty"`
	Decisions        *viewDecisionCountJSON `json:"decisions,omitempty"`
	Backlog          *viewBacklogJSON       `json:"backlog,omitempty"`
}

type viewProgressJSON struct {
	Completed  int `json:"completed"`
	Total      int `json:"total"`
	Percentage int `json:"percentage"`
}

type viewDecisionCountJSON struct {
	Active int `json:"active"`
	Total  int `json:"total"`
}

type viewBacklogJSON struct {
	Deferred      int      `json:"deferred"`
	OpenQuestions int      `json:"openQuestions"`
	Future        int      `json:"future"`
	Other         int      `json:"other,omitempty"`
	Unrecognized  []string `json:"unrecognized,omitempty"`
}

type viewChangeJSON struct {
	Name           string   `json:"name"`
	Status         string   `json:"status"`
	Born           string   `json:"born,omitempty"`
	LastModified   string   `json:"lastModified,omitempty"`
	CompletedTasks int      `json:"completedTasks,omitempty"`
	TotalTasks     int      `json:"totalTasks,omitempty"`
	Percentage     int      `json:"percentage,omitempty"`
	DependsOn      []string `json:"dependsOn,omitempty"`
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

type viewGraphJSON struct {
	Roots     []string `json:"roots"`
	Unrelated []string `json:"unrelated,omitempty"`
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
	cmd := exec.Command("gh", "issue", "list", "--json", "number,title,state,url", "--state", "open", "--limit", "50")
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

func buildViewText(root string, specs []internal.SpecInfo, draft, active, completed, patch []internal.ChangeInfo, totalReqs, totalCompletedTasks, totalTasks int, decisions []*internal.Decision, decErr error, changes []internal.ChangeInfo, depMap map[string][]string, productContent []byte, productErr error, ghIssues []viewGHIssueJSON) string {
	var sb strings.Builder
	sep := strings.Repeat("═", 60)

	fmt.Fprintln(&sb)
	fmt.Fprintln(&sb, "Litespec Dashboard")
	fmt.Fprintln(&sb)
	fmt.Fprintln(&sb, sep)

	// Product section — v2 lean durable
	fmt.Fprintln(&sb, "Product:")
	if productErr == nil {
		preview := strings.TrimSpace(string(productContent))
		if len(preview) > 400 {
			preview = preview[:400] + "…"
		}
		firstLine := strings.Split(preview, "\n")[0]
		fmt.Fprintf(&sb, "  specs/product.md — %s\n", strings.TrimSpace(firstLine))
		if strings.Contains(strings.ToLower(preview), "product") || true {
			fmt.Fprintln(&sb, "  product: mental models + flows")
		}
	} else {
		fmt.Fprintln(&sb, "  specs/product.md — missing (run litespec init to scaffold)")
		fmt.Fprintln(&sb, "  product: not yet initialized")
	}

	fmt.Fprintln(&sb)
	fmt.Fprintln(&sb, "Summary:")
	fmt.Fprintf(&sb, "  ● Specifications: %d specs, %d requirements\n", len(specs), totalReqs)
	fmt.Fprintf(&sb, "  ● Draft Changes: %d\n", len(draft))
	fmt.Fprintf(&sb, "  ● Active Changes: %d in progress\n", len(active))
	if len(patch) > 0 {
		fmt.Fprintf(&sb, "  ● Patch Changes: %d\n", len(patch))
	}
	fmt.Fprintf(&sb, "  ● Ready to Archive: %d\n", len(completed))
	if totalTasks > 0 {
		pct := int(math.Round(float64(totalCompletedTasks) / float64(totalTasks) * 100))
		fmt.Fprintf(&sb, "  ● Task Progress: %d/%d (%d%% complete)\n", totalCompletedTasks, totalTasks, pct)
	}
	if decErr == nil && len(decisions) > 0 {
		activeDec := 0
		for _, d := range decisions {
			if d.Status != internal.StatusSuperseded {
				activeDec++
			}
		}
		fmt.Fprintf(&sb, "  ● Decisions: %d/%d\n", activeDec, len(decisions))
	}

	backlog, _ := internal.ParseBacklog(internal.BacklogPath(root))
	if backlog != nil {
		var parts []string
		if backlog.Deferred > 0 {
			parts = append(parts, formatCount(backlog.Deferred, "deferred"))
		}
		if backlog.OpenQuestions > 0 {
			parts = append(parts, formatCount(backlog.OpenQuestions, "open questions"))
		}
		if backlog.Future > 0 {
			parts = append(parts, formatCount(backlog.Future, "future"))
		}
		line := "  ● Backlog: " + strings.Join(parts, ", ")
		if backlog.Other > 0 {
			line += " — " + formatCount(backlog.Other, "other")
		}
		fmt.Fprintln(&sb, line)
	}

	if len(active) > 0 {
		fmt.Fprintln(&sb)
		fmt.Fprintln(&sb, "Active Changes")
		fmt.Fprintln(&sb, strings.Repeat("─", 60))
		for _, c := range active {
			bar := createProgressBar(c.CompletedTasks, c.TotalTasks, 20)
			pct := int(math.Round(float64(c.CompletedTasks) / float64(c.TotalTasks) * 100))
			fmt.Fprintf(&sb, "  ◉ %-30s %s %d%%%s\n", c.Name, bar, pct, formatTimestamps(c))
		}
	}

	if len(draft) > 0 {
		fmt.Fprintln(&sb)
		fmt.Fprintln(&sb, "Draft Changes")
		fmt.Fprintln(&sb, strings.Repeat("─", 60))
		for _, c := range draft {
			fmt.Fprintf(&sb, "  ○ %s%s\n", c.Name, formatTimestamps(c))
		}
	}

	if len(patch) > 0 {
		fmt.Fprintln(&sb)
		fmt.Fprintln(&sb, "Patch Changes")
		fmt.Fprintln(&sb, strings.Repeat("─", 60))
		for _, c := range patch {
			fmt.Fprintf(&sb, "  ◆ %s%s\n", c.Name, formatTimestamps(c))
		}
	}

	if len(completed) > 0 {
		fmt.Fprintln(&sb)
		fmt.Fprintln(&sb, "Ready to Archive")
		fmt.Fprintln(&sb, strings.Repeat("─", 60))
		for _, c := range completed {
			fmt.Fprintf(&sb, "  ✓ %s%s\n", c.Name, formatTimestamps(c))
		}
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

	if depMap == nil {
		fmt.Fprintln(&sb)
		fmt.Fprintln(&sb, sep)
	} else {
		hasDeps := false
		for _, deps := range depMap {
			if len(deps) > 0 {
				hasDeps = true
				break
			}
		}

		if hasDeps {
			fmt.Fprintln(&sb)
			fmt.Fprintln(&sb, "Dependency Graph")
			fmt.Fprintln(&sb, strings.Repeat("─", 60))
			renderDependencyGraph(&sb, depMap, changes)
		}

		fmt.Fprintln(&sb)
		fmt.Fprintln(&sb, sep)
	}

	fmt.Fprintf(&sb, "\nUse litespec list --changes or litespec list --specs for detailed views\n")
	return sb.String()
}

func buildViewMinimalJSON(root string, specs []internal.SpecInfo, draft, active, completed, patch []internal.ChangeInfo, totalReqs, totalCompletedTasks, totalTasks int, decisions []*internal.Decision, decErr error) viewJSON {
	summary := viewSummaryJSON{
		Specs:          len(specs),
		Requirements:   totalReqs,
		DraftChanges:   len(draft),
		ActiveChanges:  len(active),
		PatchChanges:   len(patch),
		ReadyToArchive: len(completed),
	}

	if totalTasks > 0 {
		pct := int(math.Round(float64(totalCompletedTasks) / float64(totalTasks) * 100))
		summary.TaskProgress = &viewProgressJSON{
			Completed:  totalCompletedTasks,
			Total:      totalTasks,
			Percentage: pct,
		}
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

	backlog, _ := internal.ParseBacklog(internal.BacklogPath(root))
	if backlog != nil {
		summary.Backlog = &viewBacklogJSON{
			Deferred:      backlog.Deferred,
			OpenQuestions: backlog.OpenQuestions,
			Future:        backlog.Future,
			Other:         backlog.Other,
			Unrecognized:  backlog.Unrecognized,
		}
	}

	return viewJSON{Summary: summary}
}

func buildViewFullJSON(root string, specs []internal.SpecInfo, draft, active, completed, patch []internal.ChangeInfo, totalReqs, totalCompletedTasks, totalTasks int, decisions []*internal.Decision, decErr error, depMap map[string][]string, productContent []byte, productErr error, ghIssues []viewGHIssueJSON) viewJSON {
	out := buildViewMinimalJSON(root, specs, draft, active, completed, patch, totalReqs, totalCompletedTasks, totalTasks, decisions, decErr)

	var changes []viewChangeJSON
	for _, c := range draft {
		changes = append(changes, viewChangeJSON{
			Name: c.Name, Status: "draft", Born: bornStr(c), LastModified: modifiedStr(c),
		})
	}
	for _, c := range patch {
		changes = append(changes, viewChangeJSON{
			Name: c.Name, Status: "patch", Born: bornStr(c), LastModified: modifiedStr(c),
		})
	}
	for _, c := range active {
		pct := 0
		if c.TotalTasks > 0 {
			pct = int(math.Round(float64(c.CompletedTasks) / float64(c.TotalTasks) * 100))
		}
		changes = append(changes, viewChangeJSON{
			Name: c.Name, Status: "active", Born: bornStr(c), LastModified: modifiedStr(c),
			CompletedTasks: c.CompletedTasks, TotalTasks: c.TotalTasks, Percentage: pct, DependsOn: c.DependsOn,
		})
	}
	for _, c := range completed {
		changes = append(changes, viewChangeJSON{
			Name: c.Name, Status: "completed", Born: bornStr(c), LastModified: modifiedStr(c),
			CompletedTasks: c.CompletedTasks, TotalTasks: c.TotalTasks,
		})
	}

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

	var graph *viewGraphJSON
	if depMap != nil {
		hasDeps := false
		for _, deps := range depMap {
			if len(deps) > 0 {
				hasDeps = true
				break
			}
		}
		if hasDeps {
			var roots []string
			for name, deps := range depMap {
				if len(deps) == 0 {
					roots = append(roots, name)
				}
			}
			sort.Strings(roots)

			related := make(map[string]bool)
			allChanges := append(draft, active...)
			allChanges = append(allChanges, completed...)
			for _, c := range allChanges {
				related[c.Name] = related[c.Name] || len(depMap[c.Name]) > 0
				for _, dep := range depMap[c.Name] {
					related[dep] = true
				}
			}

			var unrelated []string
			for _, c := range allChanges {
				if !related[c.Name] && len(depMap[c.Name]) == 0 {
					unrelated = append(unrelated, c.Name)
				}
			}
			sort.Strings(unrelated)

			graph = &viewGraphJSON{Roots: roots, Unrelated: unrelated}
		}
	}

	out.Changes = changes
	out.Specs = specEntries
	out.Decisions = decEntries
	out.Graph = graph
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

func bornStr(c internal.ChangeInfo) string {
	if c.Created.IsZero() {
		return ""
	}
	return c.Created.Format("2006-01-02")
}

func modifiedStr(c internal.ChangeInfo) string {
	if c.LastModified.IsZero() {
		return ""
	}
	return internal.FormatRelativeTime(c.LastModified)
}
