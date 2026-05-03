package internal

import (
	"os"
	"path/filepath"
	"strings"
)

type BacklogItem struct {
	Section string
	Title   string
}

type BacklogSummary struct {
	Deferred      int
	OpenQuestions int
	Future        int
	Other         int
	Unrecognized  []string
}

func BacklogPath(root string) string {
	return filepath.Join(root, ProjectDirName, BacklogFileName)
}

func normalizeBacklogSection(header string) (key string, ok bool) {
	switch strings.ToLower(strings.TrimSpace(header)) {
	case "deferred":
		return "deferred", true
	case "open questions":
		return "open-questions", true
	case "future versions", "future":
		return "future", true
	case "other":
		return "other", true
	default:
		return "", false
	}
}

func ParseBacklog(path string) (*BacklogSummary, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var currentSection string
	var deferred, openQuestions, future, other int
	var unrecognized []string

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSuffix(line, "\r")
		if strings.HasPrefix(line, "## ") && !strings.HasPrefix(line, "### ") {
			header := strings.TrimPrefix(line, "## ")
			key, ok := normalizeBacklogSection(header)
			if ok {
				currentSection = key
			} else {
				currentSection = "unrecognized"
				unrecognized = append(unrecognized, strings.TrimSpace(header))
			}
			continue
		}

		if currentSection == "" {
			continue
		}

		// Count top-level `- ` or `* ` items only (no leading whitespace)
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			switch currentSection {
			case "deferred":
				deferred++
			case "open-questions":
				openQuestions++
			case "future":
				future++
			case "other":
				other++
			}
		}
	}

	if deferred+openQuestions+future+other == 0 && len(unrecognized) == 0 {
		return nil, nil
	}

	return &BacklogSummary{
		Deferred:      deferred,
		OpenQuestions: openQuestions,
		Future:        future,
		Other:         other,
		Unrecognized:  unrecognized,
	}, nil
}

func ParseBacklogItems(path string) ([]BacklogItem, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var items []BacklogItem
	var currentSection string

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSuffix(line, "\r")
		if strings.HasPrefix(line, "## ") && !strings.HasPrefix(line, "### ") {
			header := strings.TrimPrefix(line, "## ")
			key, ok := normalizeBacklogSection(header)
			if ok {
				currentSection = key
			} else {
				currentSection = ""
			}
			continue
		}

		if currentSection == "" {
			continue
		}

		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			title := extractBacklogTitle(strings.TrimPrefix(strings.TrimPrefix(line, "- "), "* "))
			items = append(items, BacklogItem{
				Section: currentSection,
				Title:   title,
			})
		}
	}

	return items, nil
}

func extractBacklogTitle(line string) string {
	// Extract bold title: **Title** — rest
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "**") {
		return line
	}
	line = strings.TrimPrefix(line, "**")
	idx := strings.Index(line, "**")
	if idx < 0 {
		return line
	}
	return line[:idx]
}
