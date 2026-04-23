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

	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSuffix(line, "\r")
		if strings.HasPrefix(line, "## ") && !strings.HasPrefix(line, "### ") {
			section := strings.TrimSpace(strings.TrimPrefix(line, "## "))
			switch section {
			case "Deferred":
				currentSection = "deferred"
			case "Open Questions":
				currentSection = "open-questions"
			case "Future Versions":
				currentSection = "future"
			default:
				currentSection = "other"
			}
			continue
		}

		if currentSection == "" {
			continue
		}

		// Count top-level `- ` items only (no leading whitespace)
		if strings.HasPrefix(line, "- ") {
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

	if deferred+openQuestions+future+other == 0 {
		return nil, nil
	}

	return &BacklogSummary{
		Deferred:      deferred,
		OpenQuestions: openQuestions,
		Future:        future,
		Other:         other,
	}, nil
}
