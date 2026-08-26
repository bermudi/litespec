package internal

import (
	"fmt"
	"strconv"
	"strings"
)

type queueUnitIdentity struct {
	Occurrence int    `json:"occurrence"`
	Heading    string `json:"heading"`
}

func formatRebuildRequest(identity queueUnitIdentity) string {
	return fmt.Sprintf(
		"Rebuild request:\nUnit occurrence: %d\nUnit heading: %s",
		identity.Occurrence,
		identity.Heading,
	)
}

func unresolvedRebuildRequests(units []queueUnit, comments []string) ([]queueUnitIdentity, []error) {
	identities := queueUnitIdentities(units)
	valid := make(map[queueUnitIdentity]int, len(identities))
	for i, identity := range identities {
		valid[identity] = i
	}

	unresolved := make(map[queueUnitIdentity]bool)
	var errs []error
	for commentIndex, comment := range comments {
		identity, kind, err := parseRebuildComment(comment, units)
		if err != nil {
			errs = append(errs, fmt.Errorf("comment %d: %w", commentIndex+1, err))
			continue
		}
		if kind == rebuildCommentOther {
			continue
		}
		if _, ok := valid[identity]; !ok {
			errs = append(errs, fmt.Errorf(
				"comment %d: unit occurrence %d with heading %q does not identify exactly one queue unit",
				commentIndex+1,
				identity.Occurrence,
				identity.Heading,
			))
			continue
		}
		if kind == rebuildCommentRequest {
			unresolved[identity] = true
			continue
		}
		delete(unresolved, identity)
	}

	result := make([]queueUnitIdentity, 0, len(unresolved))
	for _, identity := range identities {
		if unresolved[identity] {
			result = append(result, identity)
		}
	}
	return result, errs
}

func selectableUnitIdentities(units []queueUnit, comments []string) ([]queueUnitIdentity, []error) {
	unresolved, errs := unresolvedRebuildRequests(units, comments)
	unresolvedSet := make(map[queueUnitIdentity]bool, len(unresolved))
	for _, identity := range unresolved {
		unresolvedSet[identity] = true
	}

	identities := queueUnitIdentities(units)
	selectable := make([]queueUnitIdentity, 0)
	for i, unit := range units {
		if !isCheckedUnit(unit.Body) || unresolvedSet[identities[i]] {
			selectable = append(selectable, identities[i])
		}
	}
	return selectable, errs
}

type rebuildCommentKind int

const (
	rebuildCommentOther rebuildCommentKind = iota
	rebuildCommentRequest
	rebuildCommentEvidence
)

func parseRebuildComment(comment string, units []queueUnit) (queueUnitIdentity, rebuildCommentKind, error) {
	normalized := strings.TrimSpace(strings.ReplaceAll(comment, "\r\n", "\n"))
	lines := strings.Split(normalized, "\n")
	if len(lines) == 0 {
		return queueUnitIdentity{}, rebuildCommentOther, nil
	}

	if lines[0] == "Rebuild request:" {
		if len(lines) != 3 {
			return queueUnitIdentity{}, rebuildCommentOther, fmt.Errorf("malformed rebuild request")
		}
		identity, err := parseIdentityLines(lines[1], lines[2])
		if err != nil {
			return queueUnitIdentity{}, rebuildCommentOther, fmt.Errorf("malformed rebuild request: %w", err)
		}
		return identity, rebuildCommentRequest, nil
	}
	if strings.HasPrefix(lines[0], "Rebuild request") {
		return queueUnitIdentity{}, rebuildCommentOther, fmt.Errorf("malformed rebuild request")
	}

	hasOccurrence := strings.HasPrefix(lines[0], "Unit occurrence:")
	hasHeading := len(lines) > 1 && strings.HasPrefix(lines[1], "Unit heading:")
	if !hasOccurrence && !hasHeading {
		return queueUnitIdentity{}, rebuildCommentOther, nil
	}
	if !hasOccurrence || !hasHeading || len(lines) < 4 || strings.TrimSpace(lines[2]) != "Evidence:" {
		return queueUnitIdentity{}, rebuildCommentOther, fmt.Errorf("malformed identity-bearing evidence receipt")
	}
	identity, err := parseIdentityLines(lines[0], lines[1])
	if err != nil {
		return queueUnitIdentity{}, rebuildCommentOther, fmt.Errorf("malformed identity-bearing evidence receipt: %w", err)
	}
	unit, ok := findQueueUnit(units, identity)
	if !ok {
		return identity, rebuildCommentEvidence, nil
	}
	evidence := strings.Join(lines[2:], "\n")
	if issues := evidenceReceiptIssues(evidence, unitVerifyCommand(unit.Body), unitContractDigest(unit), "comment", identity.Heading); len(issues) != 0 {
		return queueUnitIdentity{}, rebuildCommentOther, fmt.Errorf("incomplete evidence receipt for occurrence %d heading %q", identity.Occurrence, identity.Heading)
	}
	return identity, rebuildCommentEvidence, nil
}

func parseIdentityLines(occurrenceLine, headingLine string) (queueUnitIdentity, error) {
	const occurrencePrefix = "Unit occurrence: "
	const headingPrefix = "Unit heading: "
	if !strings.HasPrefix(occurrenceLine, occurrencePrefix) || !strings.HasPrefix(headingLine, headingPrefix) {
		return queueUnitIdentity{}, fmt.Errorf("identity fields must be exact and ordered")
	}
	occurrence, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(occurrenceLine, occurrencePrefix)))
	if err != nil || occurrence < 1 {
		return queueUnitIdentity{}, fmt.Errorf("unit occurrence must be a positive integer")
	}
	heading := strings.TrimPrefix(headingLine, headingPrefix)
	if strings.TrimSpace(heading) == "" || heading != strings.TrimSpace(heading) {
		return queueUnitIdentity{}, fmt.Errorf("unit heading must be nonempty and exact")
	}
	return queueUnitIdentity{Occurrence: occurrence, Heading: heading}, nil
}

func queueUnitIdentities(units []queueUnit) []queueUnitIdentity {
	occurrences := make(map[string]int)
	identities := make([]queueUnitIdentity, 0, len(units))
	for _, unit := range units {
		occurrences[unit.Heading]++
		identities = append(identities, queueUnitIdentity{
			Occurrence: occurrences[unit.Heading],
			Heading:    unit.Heading,
		})
	}
	return identities
}

func findQueueUnit(units []queueUnit, wanted queueUnitIdentity) (queueUnit, bool) {
	identities := queueUnitIdentities(units)
	for i, identity := range identities {
		if identity == wanted {
			return units[i], true
		}
	}
	return queueUnit{}, false
}

func persistLocalRebuildRouting(
	body string,
	affected []queueUnitIdentity,
	persist func(string) error,
) (string, error) {
	allSections := parseQueueUnits(body)
	units := make([]queueUnit, 0, len(allSections))
	sectionIdentities := make(map[int]queueUnitIdentity)
	occurrences := make(map[string]int)
	for i, section := range allSections {
		if !isUnit(section) {
			continue
		}
		occurrences[section.Heading]++
		identity := queueUnitIdentity{Occurrence: occurrences[section.Heading], Heading: section.Heading}
		sectionIdentities[i] = identity
		units = append(units, section)
	}
	valid := make(map[queueUnitIdentity]bool)
	for i, identity := range queueUnitIdentities(units) {
		if isCheckedUnit(units[i].Body) {
			valid[identity] = true
		}
	}
	wanted := make(map[queueUnitIdentity]bool, len(affected))
	for _, identity := range affected {
		if !valid[identity] {
			return body, fmt.Errorf(
				"affected unit occurrence %d with heading %q is missing, ambiguous, or unchecked",
				identity.Occurrence,
				identity.Heading,
			)
		}
		wanted[identity] = true
	}

	lines := strings.Split(body, "\n")
	current := queueUnitIdentity{}
	inUnit := false
	openFence := ""
	changed := make(map[queueUnitIdentity]bool)
	sectionIndex := -1
	for i, line := range lines {
		if consumeMarkdownFenceLine(&openFence, line) {
			continue
		}
		if strings.HasPrefix(line, "## ") {
			sectionIndex++
			current, inUnit = sectionIdentities[sectionIndex]
			continue
		}
		if !inUnit || !wanted[current] || !isCheckedLine(strings.TrimSpace(line)) {
			continue
		}
		statusAt := strings.Index(line, "[x]")
		if statusAt < 0 {
			statusAt = strings.Index(line, "[X]")
		}
		if statusAt < 0 {
			continue
		}
		lines[i] = line[:statusAt] + "[ ]" + line[statusAt+3:]
		changed[current] = true
	}
	if len(changed) != len(wanted) {
		return body, fmt.Errorf("could not change every affected local unit status")
	}

	updated := strings.Join(lines, "\n")
	if err := persist(updated); err != nil {
		return body, err
	}
	return updated, nil
}
