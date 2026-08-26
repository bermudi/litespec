package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type UnitDigestLine struct {
	Occurrence int    `json:"occurrence"`
	Heading    string `json:"heading"`
	Digest     string `json:"digest"`
}

func DigestQueueUnits(root string, issueNumber int, queuePath string) ([]UnitDigestLine, error) {
	var body, source string
	if queuePath != "" {
		raw, err := os.ReadFile(queuePath)
		if err != nil {
			return nil, err
		}
		body = string(raw)
		source = fmt.Sprintf("queue file %s", queuePath)
	} else {
		if _, err := lookPathGh("gh"); err != nil {
			return nil, fmt.Errorf("gh not available")
		}
		out, err := ghIssueView(root, issueNumber)
		if err != nil {
			return nil, fmt.Errorf("gh issue view %d failed: %w", issueNumber, err)
		}
		var issue ghIssue
		if err := json.Unmarshal(out, &issue); err != nil {
			return nil, fmt.Errorf("parse gh issue: %w", err)
		}
		body = issue.Body
		source = fmt.Sprintf("GH issue #%d", issue.Number)
	}

	allSections := parseQueueUnits(body)
	units := make([]queueUnit, 0, len(allSections))
	for _, section := range allSections {
		if isUnit(section) {
			units = append(units, section)
		}
	}
	if len(units) == 0 {
		return nil, fmt.Errorf("%s contains no queue units", source)
	}

	identities := queueUnitIdentities(units)
	lines := make([]UnitDigestLine, 0, len(units))
	for i, unit := range units {
		lines = append(lines, UnitDigestLine{
			Occurrence: identities[i].Occurrence,
			Heading:    unit.Heading,
			Digest:     unitContractDigest(unit),
		})
	}
	return lines, nil
}

func FormatUnitDigestLines(lines []UnitDigestLine) string {
	var b strings.Builder
	for _, line := range lines {
		b.WriteString(strconv.Itoa(line.Occurrence))
		b.WriteByte('\t')
		b.WriteString(line.Heading)
		b.WriteByte('\t')
		b.WriteString(line.Digest)
		b.WriteByte('\n')
	}
	return b.String()
}
