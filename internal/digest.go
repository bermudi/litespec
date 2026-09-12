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
	body, source, err := queueBody(root, issueNumber, queuePath)
	if err != nil {
		return nil, err
	}
	units, err := queueUnitsFromBody(body, source)
	if err != nil {
		return nil, err
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

// ResolvedQueueUnit is one queue unit located by exact heading and positive
// same-heading occurrence, carrying its current contract digest and Verify
// command exactly as validate reads them.
type ResolvedQueueUnit struct {
	Occurrence int
	Heading    string
	Digest     string
	Verify     string
}

// resolveQueueUnitIndex locates one unit with the same identity semantics
// as the digest command: exact heading match, disambiguated by a positive
// same-heading occurrence when the heading repeats. Ambiguous, unknown, or
// out-of-range resolution is a visible refusal.
func resolveQueueUnitIndex(units []queueUnit, identities []queueUnitIdentity, source, heading string, occurrence int) (int, error) {
	matches := make([]int, 0, 1)
	for i, unit := range units {
		if unit.Heading == heading {
			matches = append(matches, i)
		}
	}
	if len(matches) == 0 {
		return 0, fmt.Errorf("no unit with heading %q in %s", heading, source)
	}
	if occurrence >= 1 {
		for _, i := range matches {
			if identities[i].Occurrence == occurrence {
				return i, nil
			}
		}
		return 0, fmt.Errorf("heading %q has no occurrence %d in %s (valid: 1..%d)", heading, occurrence, source, len(matches))
	}
	if len(matches) > 1 {
		return 0, fmt.Errorf("heading %q matches %d occurrences in %s; pass --occurrence between 1 and %d", heading, len(matches), source, len(matches))
	}
	return matches[0], nil
}

// ResolveQueueUnit locates one queue unit with the same identity semantics
// as the digest command: exact heading match, disambiguated by a positive
// same-heading occurrence when the heading repeats. Ambiguous, unknown, or
// out-of-range resolution is a visible refusal.
func ResolveQueueUnit(root string, issueNumber int, queuePath, heading string, occurrence int) (ResolvedQueueUnit, error) {
	body, source, err := queueBody(root, issueNumber, queuePath)
	if err != nil {
		return ResolvedQueueUnit{}, err
	}
	units, err := queueUnitsFromBody(body, source)
	if err != nil {
		return ResolvedQueueUnit{}, err
	}

	identities := queueUnitIdentities(units)
	index, err := resolveQueueUnitIndex(units, identities, source, heading, occurrence)
	if err != nil {
		return ResolvedQueueUnit{}, err
	}
	command := unitVerifyCommand(units[index].Body)
	if strings.TrimSpace(command) == "" {
		return ResolvedQueueUnit{}, fmt.Errorf("unit %q in %s has no Verify command", heading, source)
	}
	return ResolvedQueueUnit{
		Occurrence: identities[index].Occurrence,
		Heading:    units[index].Heading,
		Digest:     unitContractDigest(units[index]),
		Verify:     command,
	}, nil
}

func queueBody(root string, issueNumber int, queuePath string) (string, string, error) {
	if queuePath != "" {
		raw, err := os.ReadFile(queuePath)
		if err != nil {
			return "", "", err
		}
		return string(raw), fmt.Sprintf("queue file %s", queuePath), nil
	}
	if _, err := lookPathGh("gh"); err != nil {
		return "", "", fmt.Errorf("gh not available")
	}
	out, err := ghIssueView(root, issueNumber)
	if err != nil {
		return "", "", fmt.Errorf("gh issue view %d failed: %w", issueNumber, err)
	}
	var issue ghIssue
	if err := json.Unmarshal(out, &issue); err != nil {
		return "", "", fmt.Errorf("parse gh issue: %w", err)
	}
	return issue.Body, fmt.Sprintf("GH issue #%d", issue.Number), nil
}

func queueUnitsFromBody(body, source string) ([]queueUnit, error) {
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
	return units, nil
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
