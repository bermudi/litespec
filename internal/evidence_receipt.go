package internal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	evidenceProtocolLegacy = "evidence/legacy-v0"
	evidenceProtocolV1     = "evidence/v1"
	receiptIDPrefix        = "receipt-sha256-v1:"
	legacyReceiptIDPrefix  = "legacy-receipt-sha256-v1:"
)

var receiptIDPattern = regexp.MustCompile(`^(?:receipt-sha256-v1|legacy-receipt-sha256-v1):[0-9a-f]{64}$`)

type evidenceReceiptHeader struct {
	protocol        string
	digestAlgorithm string
	receiptID       string
	recoveredFrom   string
	versioned       bool
}

type parsedEvidenceReceipt struct {
	header     evidenceReceiptHeader
	identity   *queueUnitIdentity
	heading    string
	verify     string
	digest     string
	preSHA     string
	preStatus  string
	preOutput  string
	preScope   string
	postSHA    string
	postStatus string
	postOutput string
	postScope  string
}

func receiptVersionField(line string) bool {
	for _, label := range []string{"Protocol:", "Digest algorithm:", "Receipt ID:", "Recovered from:"} {
		if strings.HasPrefix(line, label) {
			return true
		}
	}
	return false
}

func receiptVersionFieldsPresent(document evidenceDocument) bool {
	openFence := ""
	for _, line := range document.lines {
		if consumeMarkdownFenceLine(&openFence, line) {
			continue
		}
		if receiptVersionField(line) {
			return true
		}
	}
	return false
}

func consumeReceiptHeaderField(cursor *evidenceCursor, label string) (string, error) {
	cursor.skipBlanks()
	prefix := label + ": "
	if cursor.at >= len(cursor.lines) || !strings.HasPrefix(cursor.lines[cursor.at], prefix) {
		return "", fmt.Errorf("version metadata must contain %s: in the required order", label)
	}
	value := strings.TrimSpace(strings.TrimPrefix(cursor.lines[cursor.at], prefix))
	cursor.at++
	if value == "" {
		return "", fmt.Errorf("%s: must be nonempty", label)
	}
	return value, nil
}

func consumeVersionedReceiptHeader(cursor *evidenceCursor) (evidenceReceiptHeader, error) {
	protocol, err := consumeReceiptHeaderField(cursor, "Protocol")
	if err != nil {
		return evidenceReceiptHeader{}, err
	}
	if protocol != evidenceProtocolV1 {
		return evidenceReceiptHeader{}, fmt.Errorf("unknown evidence protocol %q", protocol)
	}

	algorithm, err := consumeReceiptHeaderField(cursor, "Digest algorithm")
	if err != nil {
		return evidenceReceiptHeader{}, err
	}
	if _, ok := unitContractDigestForAlgorithm(queueUnit{}, algorithm); !ok {
		return evidenceReceiptHeader{}, fmt.Errorf("unknown digest algorithm %q", algorithm)
	}

	receiptID, err := consumeReceiptHeaderField(cursor, "Receipt ID")
	if err != nil {
		return evidenceReceiptHeader{}, err
	}
	if !strings.HasPrefix(receiptID, receiptIDPrefix) || !unitDigestPattern.MatchString(strings.TrimPrefix(receiptID, receiptIDPrefix)) {
		return evidenceReceiptHeader{}, fmt.Errorf("Receipt ID must be %s followed by 64 lowercase hexadecimal characters", receiptIDPrefix)
	}

	header := evidenceReceiptHeader{
		protocol:        protocol,
		digestAlgorithm: algorithm,
		receiptID:       receiptID,
		versioned:       true,
	}
	cursor.skipBlanks()
	if cursor.at < len(cursor.lines) && strings.HasPrefix(cursor.lines[cursor.at], "Recovered from:") {
		recoveredFrom, err := consumeReceiptHeaderField(cursor, "Recovered from")
		if err != nil {
			return evidenceReceiptHeader{}, err
		}
		if !receiptIDPattern.MatchString(recoveredFrom) {
			return evidenceReceiptHeader{}, fmt.Errorf("Recovered from must be a complete receipt ID")
		}
		header.recoveredFrom = recoveredFrom
	}
	cursor.skipBlanks()
	if cursor.at < len(cursor.lines) && receiptVersionField(cursor.lines[cursor.at]) {
		return evidenceReceiptHeader{}, fmt.Errorf("version metadata fields must appear once and in the required order")
	}
	return header, nil
}

func parseEvidenceReceiptHeader(document evidenceDocument) (evidenceReceiptHeader, int, error) {
	document = document.trimSpace()
	if !receiptVersionFieldsPresent(document) {
		return evidenceReceiptHeader{
			protocol:        evidenceProtocolLegacy,
			digestAlgorithm: digestAlgorithmV1,
		}, 0, nil
	}

	cursor := newEvidenceCursorFromDocument(document)
	cursor.skipBlanks()
	if cursor.at >= len(cursor.lines) || !strings.HasPrefix(cursor.lines[cursor.at], "Protocol:") {
		return evidenceReceiptHeader{}, 0, fmt.Errorf("version metadata is partial or out of order; Protocol: must be first")
	}
	if !strings.HasPrefix(cursor.lines[cursor.at], "Protocol: ") {
		return evidenceReceiptHeader{}, 0, fmt.Errorf("Protocol: must use the exact versioned header grammar")
	}
	header, err := consumeVersionedReceiptHeader(cursor)
	if err != nil {
		return evidenceReceiptHeader{}, 0, err
	}
	return header, cursor.at, nil
}

func evidenceReceiptDigestAlgorithm(document evidenceDocument) (string, error) {
	header, _, err := parseEvidenceReceiptHeader(document)
	if err != nil {
		return "", err
	}
	return header.digestAlgorithm, nil
}

func evidenceReceiptExpectedDigest(unit queueUnit, document evidenceDocument) string {
	algorithm, err := evidenceReceiptDigestAlgorithm(document)
	if err != nil {
		return unitContractDigest(unit)
	}
	if digest, ok := unitContractDigestForAlgorithm(unit, algorithm); ok {
		return digest
	}
	return unitContractDigest(unit)
}

func (c *evidenceCursor) consumeContinuationReceiptHeader(expectedIdentity *queueUnitIdentity) error {
	if c.receiptHeader == nil || !c.trackCommentBoundaries || c.at >= len(c.lines) {
		return nil
	}
	part := c.partAt(c.at)
	if part < 1 || part-1 >= len(c.partContinued) || !c.partContinued[part-1] {
		return nil
	}
	if !strings.HasPrefix(c.lines[c.at], "Protocol:") {
		return nil
	}
	repeated, err := consumeVersionedReceiptHeader(c)
	if err != nil {
		return err
	}
	if repeated != *c.receiptHeader {
		return fmt.Errorf("continuation receipt header must match the top-level receipt")
	}
	if expectedIdentity == nil {
		return nil
	}
	c.skipBlanks()
	occurrenceText, ok := c.consumeField("Unit occurrence")
	if !ok {
		return fmt.Errorf("continuation receipt must repeat Unit occurrence:")
	}
	occurrence, err := strconv.Atoi(occurrenceText)
	if err != nil || occurrence != expectedIdentity.Occurrence {
		return fmt.Errorf("continuation receipt Unit occurrence must match the receipt identity")
	}
	c.skipBlanks()
	if !c.consumeExactLines("Unit heading: " + expectedIdentity.Heading) {
		return fmt.Errorf("continuation receipt Unit heading must match the receipt identity exactly")
	}
	return nil
}

func receiptCanonicalFields(receipt parsedEvidenceReceipt) []string {
	occurrence := ""
	heading := receipt.heading
	if receipt.identity != nil {
		occurrence = strconv.Itoa(receipt.identity.Occurrence)
		heading = receipt.identity.Heading
	}
	return []string{
		receipt.header.protocol,
		receipt.header.digestAlgorithm,
		receipt.header.recoveredFrom,
		occurrence,
		heading,
		receipt.verify,
		receipt.digest,
		receipt.preSHA,
		receipt.preStatus,
		receipt.preOutput,
		receipt.preScope,
		receipt.postSHA,
		receipt.postStatus,
		receipt.postOutput,
		receipt.postScope,
	}
}

func receiptIDForCanonicalReceipt(receipt parsedEvidenceReceipt) string {
	var canonical strings.Builder
	for _, field := range receiptCanonicalFields(receipt) {
		canonical.WriteString(strconv.Itoa(len([]byte(field))))
		canonical.WriteByte(':')
		canonical.WriteString(field)
	}
	sum := sha256.Sum256([]byte(canonical.String()))
	prefix := receiptIDPrefix
	if !receipt.header.versioned {
		prefix = legacyReceiptIDPrefix
	}
	return prefix + hex.EncodeToString(sum[:])
}

func parseEvidenceReceiptDocument(
	document evidenceDocument,
	verifyCmd string,
	expectedDigest string,
	source string,
	heading string,
	expectedIdentity *queueUnitIdentity,
) (parsedEvidenceReceipt, []ValidationIssue) {
	document = document.trimSpace()
	receipt := parsedEvidenceReceipt{
		heading:  heading,
		identity: expectedIdentity,
	}
	if len(document.lines) == 0 || strings.TrimSpace(strings.Join(document.lines, "\n")) == "" {
		return receipt, []ValidationIssue{{
			Severity: SeverityError,
			Message:  fmt.Sprintf("%s: checked unit %q missing Evidence receipt", source, heading),
			File:     source,
		}}
	}

	var issues []ValidationIssue
	fail := func(reason string) {
		issues = append(issues, ValidationIssue{
			Severity: SeverityError,
			Message:  fmt.Sprintf("%s: checked unit %q Evidence receipt %s", source, heading, reason),
			File:     source,
		})
	}

	header, headerEnd, err := parseEvidenceReceiptHeader(document)
	if err != nil {
		fail(fmt.Sprintf("has invalid version metadata: %v", err))
		return receipt, issues
	}
	receipt.header = header

	cursor := newEvidenceCursorFromDocument(document)
	cursor.at = headerEnd
	if header.versioned {
		cursor.receiptHeader = &receipt.header
	}
	if verifyCmd == "" || !cursor.consumeVerifyCommand(verifyCmd) {
		fail("must quote the Verify command verbatim")
		return receipt, issues
	}
	receipt.verify = verifyCmd

	cursor.skipBlanks()
	digestText, ok := cursor.consumeField("unit digest")
	if !ok {
		fail("must include a `unit digest:` field between the Verify command and pre sha")
		return receipt, issues
	}
	receipt.digest = digestText
	if !unitDigestPattern.MatchString(digestText) {
		fail("unit digest must be 64 lowercase hexadecimal characters")
	} else if expectedDigest != "" && digestText != expectedDigest {
		fail(fmt.Sprintf("unit digest mismatch: expected %s, actual %s — recompute with litespec or amend the contract", expectedDigest, digestText))
	}

	cursor.skipBlanks()
	preSHA, ok := cursor.consumeField("pre sha")
	if !ok {
		fail("fields must appear in order beginning with pre sha")
		return receipt, issues
	}
	receipt.preSHA = preSHA
	if !evidenceCommitPattern.MatchString(preSHA) {
		fail("pre sha must be a full 40- or 64-character hexadecimal commit ID")
	}

	cursor.skipBlanks()
	preStatusText, ok := cursor.consumeField("pre exit status")
	if !ok {
		fail("fields must appear in order: pre exit status must follow pre sha")
		return receipt, issues
	}
	receipt.preStatus = preStatusText
	preStatus, err := strconv.Atoi(preStatusText)
	if err != nil {
		fail("pre exit status must be an integer")
	} else if preStatus == 0 {
		fail("pre exit status must be non-zero")
	}

	cursor.skipBlanks()
	preOutput, ok, reason := cursor.consumeRawOutput("pre", digestText, heading, expectedIdentity)
	if !ok {
		fail(reason)
		return receipt, issues
	}
	receipt.preOutput = preOutput

	cursor.skipBlanks()
	if cursor.at >= len(cursor.lines) {
		fail("must include a matching Pre-evidence scope line")
		return receipt, issues
	}
	preScope := preEvidenceScopePattern.FindStringSubmatch(cursor.lines[cursor.at])
	if preScope == nil {
		fail("must include a matching Pre-evidence scope line")
	} else {
		receipt.preScope = cursor.lines[cursor.at]
		if preScope[1] != preStatusText {
			fail("pre scope line status must match pre exit status")
		}
		if !strings.EqualFold(preScope[2], preSHA) {
			fail("pre scope line sha must match pre sha")
		}
	}
	cursor.at++

	if header.versioned {
		if err := cursor.consumeContinuationReceiptHeader(expectedIdentity); err != nil {
			fail(fmt.Sprintf("has invalid continuation metadata: %v", err))
			return receipt, issues
		}
	}
	cursor.skipBlanks()
	postSHA, ok := cursor.consumeField("post sha")
	if !ok {
		fail("fields must appear in order: post sha must follow the pre scope line")
		return receipt, issues
	}
	receipt.postSHA = postSHA
	if !evidenceCommitPattern.MatchString(postSHA) {
		fail("post sha must be a full 40- or 64-character hexadecimal commit ID")
	}
	if evidenceCommitPattern.MatchString(preSHA) && strings.EqualFold(preSHA, postSHA) {
		fail("pre and post sha must differ")
	}

	cursor.skipBlanks()
	postStatusText, ok := cursor.consumeField("post exit status")
	if !ok {
		fail("fields must appear in order: post exit status must follow post sha")
		return receipt, issues
	}
	receipt.postStatus = postStatusText
	if postStatusText != "0" {
		fail("post exit status must be 0")
	}

	cursor.skipBlanks()
	postOutput, ok, reason := cursor.consumeRawOutput("post", digestText, heading, expectedIdentity)
	if !ok {
		fail(reason)
		return receipt, issues
	}
	receipt.postOutput = postOutput

	cursor.skipBlanks()
	if cursor.at >= len(cursor.lines) {
		fail("must include a matching Post-evidence scope line")
		return receipt, issues
	}
	postScope := postEvidenceScopePattern.FindStringSubmatch(cursor.lines[cursor.at])
	if postScope == nil {
		fail("must include a matching Post-evidence scope line")
	} else {
		receipt.postScope = cursor.lines[cursor.at]
		if postScope[1] != postStatusText {
			fail("post scope line status must match post exit status")
		}
		if !strings.EqualFold(postScope[2], postSHA) {
			fail("post scope line sha must match post sha")
		}
	}
	cursor.at++

	for cursor.at < len(cursor.lines) && strings.TrimSpace(cursor.lines[cursor.at]) == "" {
		cursor.at++
	}
	if cursor.at != len(cursor.lines) {
		fail("fields must appear in order with no unexpected trailing content")
	}

	if header.versioned && receiptIDPattern.MatchString(header.receiptID) {
		expectedID := receiptIDForCanonicalReceipt(receipt)
		if header.receiptID != expectedID {
			fail(fmt.Sprintf("Receipt ID mismatch: expected %s, actual %s", expectedID, header.receiptID))
		}
	} else if !header.versioned {
		receipt.header.receiptID = receiptIDForCanonicalReceipt(receipt)
	}
	return receipt, issues
}

func evidenceReceiptIssuesForDocument(
	document evidenceDocument,
	verifyCmd string,
	expectedDigest string,
	source string,
	heading string,
	expectedIdentity *queueUnitIdentity,
) ([]ValidationIssue, string) {
	receipt, issues := parseEvidenceReceiptDocument(document, verifyCmd, expectedDigest, source, heading, expectedIdentity)
	return issues, receipt.digest
}
