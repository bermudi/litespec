package internal

import (
	"os"
	"path/filepath"
	"strings"
)

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

func ParseBacklog(path string) (*BacklogSummary, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	content := string(data)
	var currentSection string
	var deferred, openQuestions, future, other int
	var unrecognized []string

	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSuffix(line, "\r")
		if strings.HasPrefix(line, "## ") && !strings.HasPrefix(line, "### ") {
			section := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(line, "## ")))
			switch section {
			case "deferred":
				currentSection = "deferred"
			case "open questions":
				currentSection = "open-questions"
			case "future versions", "future":
				currentSection = "future"
			case "other":
				currentSection = "other"
			default:
				currentSection = "unrecognized"
				unrecognized = append(unrecognized, strings.TrimSpace(strings.TrimPrefix(line, "## ")))
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
