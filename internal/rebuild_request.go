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
	scan := scanQueueComments(units, comments)
	return scan.unresolved, scan.errors
}

func selectableUnitIdentities(units []queueUnit, comments []string) ([]queueUnitIdentity, []error) {
	scan := scanQueueComments(units, comments)
	unresolvedSet := make(map[queueUnitIdentity]bool, len(scan.unresolved))
	for _, identity := range scan.unresolved {
		unresolvedSet[identity] = true
	}
	markedSet := make(map[queueUnitIdentity]bool, len(scan.replanRequired))
	for _, identity := range scan.replanRequired {
		markedSet[identity] = true
	}

	identities := queueUnitIdentities(units)
	selectable := make([]queueUnitIdentity, 0)
	for i, unit := range units {
		if markedSet[identities[i]] {
			continue
		}
		if !isCheckedUnit(unit.Body) || unresolvedSet[identities[i]] {
			selectable = append(selectable, identities[i])
		}
	}
	return selectable, scan.errors
}

type rebuildCommentKind int

const (
	rebuildCommentOther rebuildCommentKind = iota
	rebuildCommentRequest
	rebuildCommentEvidence
	rebuildCommentStaleEvidence
)

type evidenceReceiptDeclaration struct {
	digest  string
	receipt parsedEvidenceReceipt
}

func receiptVerifyCommand(lines []string) (string, bool) {
	document := newEvidenceDocument(strings.Join(lines, "\n")).trimSpace()
	_, headerEnd, err := parseEvidenceReceiptHeader(document)
	if err != nil {
		return "", false
	}
	candidate := document.lines[headerEnd:]
	for len(candidate) > 0 && strings.TrimSpace(candidate[len(candidate)-1]) == "" {
		candidate = candidate[:len(candidate)-1]
	}
	if len(candidate) == 1 {
		if command, ok := verifyCommandFromLabel(candidate[0]); ok {
			return command, command != ""
		}
	}
	if len(candidate) == 0 {
		return "", false
	}
	return strings.Join(candidate, "\n"), true
}

func validEvidenceReceiptDeclarations(
	document evidenceDocument,
	source string,
	heading string,
	expectedIdentity *queueUnitIdentity,
) []evidenceReceiptDeclaration {
	document = document.trimSpace()
	var declarations []evidenceReceiptDeclaration
	openFence := ""
	for i, line := range document.lines {
		if consumeMarkdownFenceLine(&openFence, line) {
			continue
		}
		if !strings.HasPrefix(line, "unit digest:") {
			continue
		}
		digest := strings.TrimSpace(strings.TrimPrefix(line, "unit digest:"))
		if !unitDigestPattern.MatchString(digest) {
			continue
		}
		verifyCmd, ok := receiptVerifyCommand(document.lines[:i])
		if !ok {
			continue
		}
		receipt, issues := parseEvidenceReceiptDocument(
			document,
			verifyCmd,
			digest,
			source,
			heading,
			expectedIdentity,
		)
		if len(issues) == 0 {
			declarations = append(declarations, evidenceReceiptDeclaration{digest: digest, receipt: receipt})
		}
	}
	return declarations
}

func digestMatchesAnyUnit(digest string, units []queueUnit) bool {
	for _, unit := range units {
		if unitContractDigest(unit) == digest {
			return true
		}
	}
	return false
}

// parseRebuildComment classifies an identity-bearing comment. For evidence
// receipts it also returns the declared `unit digest:` value; a complete
// receipt for a superseded contract is reported as rebuildCommentStaleEvidence
// with the digest intact so the amendment chain can judge whether the contract
// edit was witnessed.
func parseRebuildComment(comment string, units []queueUnit) (queueUnitIdentity, rebuildCommentKind, string, error) {
	return parseRebuildCommentRecord(continuedComment{text: comment, parts: []continuedCommentPart{{text: comment}}}, units)
}

func parseHeadingFormEvidenceCommentRecord(comment continuedComment, units []queueUnit) (queueUnitIdentity, rebuildCommentKind, string, bool, error) {
	identities := queueUnitIdentities(units)
	for i, unit := range units {
		if !commentNamesUnit(comment.text, unit.Heading) || !commentHasEvidenceAfterHeading(comment, unit.Heading) {
			continue
		}
		document, ok := commentEvidenceDocument(comment, unit.Heading)
		if !ok {
			continue
		}
		identity := identities[i]
		currentDigest := unitContractDigest(unit)
		expectedDigest := evidenceReceiptExpectedDigest(unit, document)
		receipt, issues := parseEvidenceReceiptDocument(
			document,
			unitVerifyCommand(unit.Body),
			expectedDigest,
			"comment",
			identity.Heading,
			&identity,
		)
		if len(issues) == 0 {
			return identity, rebuildCommentEvidence, normalizedEvidenceReceiptDigest(unit, receipt), true, nil
		}

		declarations := validEvidenceReceiptDeclarations(document, "comment", identity.Heading, &identity)
		if len(declarations) == 1 && declarations[0].digest != currentDigest {
			return identity, rebuildCommentStaleEvidence, declarations[0].digest, true, nil
		}
		return identity, rebuildCommentOther, "", true, fmt.Errorf("malformed heading-form evidence: %s", issues[0].Message)
	}
	return queueUnitIdentity{}, rebuildCommentOther, "", false, nil
}

func parseRebuildCommentRecord(comment continuedComment, units []queueUnit) (queueUnitIdentity, rebuildCommentKind, string, error) {
	document := newEvidenceDocumentFromComment(comment).trimSpace()
	lines := document.lines
	if len(lines) == 0 {
		return queueUnitIdentity{}, rebuildCommentOther, "", nil
	}

	if lines[0] == "Rebuild request:" {
		if len(lines) != 3 {
			return queueUnitIdentity{}, rebuildCommentOther, "", fmt.Errorf("malformed rebuild request")
		}
		identity, err := parseIdentityLines(lines[1], lines[2])
		if err != nil {
			return queueUnitIdentity{}, rebuildCommentOther, "", fmt.Errorf("malformed rebuild request: %w", err)
		}
		return identity, rebuildCommentRequest, "", nil
	}
	if strings.HasPrefix(lines[0], "Rebuild request") {
		return queueUnitIdentity{}, rebuildCommentOther, "", fmt.Errorf("malformed rebuild request")
	}

	hasOccurrence := strings.HasPrefix(lines[0], "Unit occurrence:")
	hasHeading := len(lines) > 1 && strings.HasPrefix(lines[1], "Unit heading:")
	if !hasOccurrence && !hasHeading {
		identity, kind, digest, shaped, err := parseHeadingFormEvidenceCommentRecord(comment, units)
		if err != nil {
			return queueUnitIdentity{}, rebuildCommentOther, "", err
		}
		if shaped {
			return identity, kind, digest, nil
		}
		return queueUnitIdentity{}, rebuildCommentOther, "", nil
	}
	if !hasOccurrence || !hasHeading || len(lines) < 4 || strings.TrimSpace(lines[2]) != "Evidence:" {
		return queueUnitIdentity{}, rebuildCommentOther, "", fmt.Errorf("malformed identity-bearing evidence receipt")
	}
	identity, err := parseIdentityLines(lines[0], lines[1])
	if err != nil {
		return queueUnitIdentity{}, rebuildCommentOther, "", fmt.Errorf("malformed identity-bearing evidence receipt: %w", err)
	}
	evidenceDocument := evidencePayloadDocument(document.afterLine(2))
	unit, ok := findQueueUnit(units, identity)
	if !ok {
		declarations := validEvidenceReceiptDeclarations(evidenceDocument, "comment", identity.Heading, &identity)
		if len(declarations) != 1 {
			return queueUnitIdentity{}, rebuildCommentOther, "", fmt.Errorf("incomplete evidence receipt for occurrence %d heading %q", identity.Occurrence, identity.Heading)
		}
		declaration := declarations[0]
		if digestMatchesAnyUnit(declaration.digest, units) {
			return queueUnitIdentity{}, rebuildCommentOther, "", fmt.Errorf("incomplete evidence receipt for occurrence %d heading %q", identity.Occurrence, identity.Heading)
		}
		return identity, rebuildCommentEvidence, declaration.digest, nil
	}

	currentDigest := unitContractDigest(unit)
	expectedDigest := evidenceReceiptExpectedDigest(unit, evidenceDocument)
	receipt, issues := parseEvidenceReceiptDocument(
		evidenceDocument,
		unitVerifyCommand(unit.Body),
		expectedDigest,
		"comment",
		identity.Heading,
		&identity,
	)
	if len(issues) == 0 {
		return identity, rebuildCommentEvidence, normalizedEvidenceReceiptDigest(unit, receipt), nil
	}

	declarations := validEvidenceReceiptDeclarations(evidenceDocument, "comment", identity.Heading, &identity)
	if len(declarations) == 1 && declarations[0].digest != currentDigest {
		return identity, rebuildCommentStaleEvidence, declarations[0].digest, nil
	}
	return queueUnitIdentity{}, rebuildCommentOther, "", fmt.Errorf("incomplete evidence receipt for occurrence %d heading %q", identity.Occurrence, identity.Heading)
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
