package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type DecisionStatus string

const (
	StatusProposed   DecisionStatus = "proposed"
	StatusAccepted   DecisionStatus = "accepted"
	StatusSuperseded DecisionStatus = "superseded"
)

var validStatuses = map[DecisionStatus]bool{
	StatusProposed:   true,
	StatusAccepted:   true,
	StatusSuperseded: true,
}

var decisionFileRe = regexp.MustCompile(`^(\d{4})-([a-z0-9][a-z0-9-]*[a-z0-9])\.md$`)

type Decision struct {
	Number       int
	Slug         string
	Title        string
	Status       DecisionStatus
	Context      string
	Decision     string
	Consequences string
	Supersedes   []string
	SupersededBy []string
	FilePath     string
	LastModified time.Time
}

func DecisionsPath(root string) string {
	return filepath.Join(root, ProjectDirName, "decisions")
}

func ParseDecision(path string) (*Decision, error) {
	base := filepath.Base(path)
	m := decisionFileRe.FindStringSubmatch(base)
	if m == nil {
		return nil, fmt.Errorf("invalid decision filename %q: must match NNNN-<slug>.md", base)
	}

	var number int
	fmt.Sscanf(m[1], "%d", &number)
	slug := m[2]

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read decision: %w", err)
	}

	content := string(data)
	title := extractH1(content)
	if title == "" {
		return nil, fmt.Errorf("decision file %q has no H1 title", base)
	}

	sections := splitH2Sections(content)

	statusStr, err := requireSection(sections, "Status", base)
	if err != nil {
		return nil, err
	}
	status := DecisionStatus(strings.TrimSpace(statusStr))
	if !validStatuses[status] {
		return nil, fmt.Errorf("decision %q has invalid status %q (valid: proposed, accepted, superseded)", base, statusStr)
	}

	context, err := requireSection(sections, "Context", base)
	if err != nil {
		return nil, err
	}

	decision, err := requireSection(sections, "Decision", base)
	if err != nil {
		return nil, err
	}

	consequences, err := requireSection(sections, "Consequences", base)
	if err != nil {
		return nil, err
	}

	supersedes := parseSlugList(sections["Supersedes"])
	supersededBy := parseSlugList(sections["Superseded-By"])

	fi, _ := os.Stat(path)
	var lastMod time.Time
	if fi != nil {
		lastMod = fi.ModTime()
	}

	return &Decision{
		Number:       number,
		Slug:         slug,
		Title:        title,
		Status:       status,
		Context:      context,
		Decision:     decision,
		Consequences: consequences,
		Supersedes:   supersedes,
		SupersededBy: supersededBy,
		FilePath:     path,
		LastModified: lastMod,
	}, nil
}

func ListDecisions(root string) ([]*Decision, error) {
	dir := DecisionsPath(root)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read decisions directory: %w", err)
	}

	var result []*Decision
	for _, entry := range entries {
		if entry.IsDir() || !decisionFileRe.MatchString(entry.Name()) {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		d, err := ParseDecision(path)
		if err != nil {
			continue
		}
		result = append(result, d)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Number < result[j].Number
	})

	return result, nil
}

func FindDecisionBySlug(root, slug string) (*Decision, error) {
	decisions, err := ListDecisions(root)
	if err != nil {
		return nil, err
	}

	// Accept either "NNNN-slug" or "slug"
	trimmed := regexp.MustCompile(`^\d{4}-`).ReplaceAllString(slug, "")

	for _, d := range decisions {
		if d.Slug == trimmed || fmt.Sprintf("%04d-%s", d.Number, d.Slug) == slug {
			return d, nil
		}
	}
	return nil, nil
}

func extractH1(content string) string {
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "# ") && !strings.HasPrefix(line, "## ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return ""
}

func splitH2Sections(content string) map[string]string {
	sections := make(map[string]string)
	var current string
	var buf strings.Builder

	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "## ") && !strings.HasPrefix(line, "### ") {
			if current != "" {
				sections[current] = strings.TrimSpace(buf.String())
			}
			current = strings.TrimSpace(strings.TrimPrefix(line, "## "))
			buf.Reset()
		} else if current != "" {
			if buf.Len() > 0 {
				buf.WriteByte('\n')
			}
			buf.WriteString(line)
		}
	}
	if current != "" {
		sections[current] = strings.TrimSpace(buf.String())
	}

	return sections
}

func requireSection(sections map[string]string, name, file string) (string, error) {
	val, ok := sections[name]
	if !ok {
		return "", fmt.Errorf("decision %q missing required section %q", file, name)
	}
	return val, nil
}

var slugListItemRe = regexp.MustCompile(`^\s*[-*]\s+\S`)

func parseSlugList(content string) []string {
	if content == "" {
		return nil
	}
	var slugs []string
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if slugListItemRe.MatchString(line) {
			slug := strings.TrimSpace(strings.TrimPrefix(line, "-"))
			slug = strings.TrimSpace(strings.TrimPrefix(slug, "*"))
			if slug != "" {
				slugs = append(slugs, slug)
			}
		} else if line != "" && !strings.HasPrefix(line, "<!--") {
			slugs = append(slugs, line)
		}
	}
	return slugs
}
