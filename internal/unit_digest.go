package internal

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strconv"
	"strings"
)

var unitDigestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

// Digest algorithm identifiers. Unversioned receipts always dispatch to v1;
// v0 is retained and applies only when a receipt declares it explicitly.
const (
	digestAlgorithmV0 = "unit-contract-sha256-v0"
	digestAlgorithmV1 = "unit-contract-sha256-v1"
)

// unitContractDigest binds a unit's contract text: heading, optional Read
// first/Constraints/Depends/Boundary values, Done means clauses, scenario
// mappings, risk cases, and Verify content, each length-prefixed in order.
// Status checkbox, Evidence content, and every other unit line are excluded.
func unitContractDigest(unit queueUnit) string {
	return digestLengthPrefixed(unitContractFields(unit))
}

func unitContractDigestForAlgorithm(unit queueUnit, algorithm string) (string, bool) {
	switch algorithm {
	case digestAlgorithmV1:
		return unitContractDigest(unit), true
	case digestAlgorithmV0:
		return digestNulSeparated(unitContractFields(unit)), true
	default:
		return "", false
	}
}

func unitContractFields(unit queueUnit) [][]byte {
	fields := [][]byte{contractFieldBytes(unit.Heading)}
	if v, ok := queueUnitFieldValue(unit.Body, "Read first:"); ok {
		fields = append(fields, contractFieldBytes(v))
	}
	if v, ok := queueUnitFieldValue(unit.Body, "Constraints:"); ok {
		fields = append(fields, contractFieldBytes(v))
	}
	if v, ok := queueUnitFieldValue(unit.Body, "Depends:"); ok {
		fields = append(fields, contractFieldBytes(v))
	}
	if v, ok := queueUnitFieldValue(unit.Body, "Boundary:"); ok {
		fields = append(fields, contractFieldBytes(v))
	}
	doneMeans, _ := queueUnitFieldLines(unit.Body, "Done means:")
	fields = append(fields, contractFieldBytes(strings.Join(doneMeans, "\n")))
	if scenarios, ok := queueUnitFieldLines(unit.Body, "Scenarios:"); ok {
		fields = append(fields, contractFieldBytes(strings.Join(scenarios, "\n")))
	}
	if risks, ok := queueUnitFieldLines(unit.Body, "Risk cases:"); ok {
		fields = append(fields, contractFieldBytes(strings.Join(risks, "\n")))
	}
	fields = append(fields, contractFieldBytes(unitVerifyCommand(unit.Body)))
	return fields
}

func digestLengthPrefixed(fields [][]byte) string {
	var b strings.Builder
	for _, f := range fields {
		b.WriteString(strconv.Itoa(len(f)))
		b.WriteByte(':')
		b.Write(f)
	}
	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}

func digestNulSeparated(fields [][]byte) string {
	var b strings.Builder
	for i, f := range fields {
		if i > 0 {
			b.WriteByte(0)
		}
		b.Write(f)
	}
	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}

func contractFieldBytes(v string) []byte {
	v = strings.ReplaceAll(v, "\r\n", "\n")
	return []byte(strings.TrimRight(v, " \t\n\r"))
}

func queueUnitFieldValue(body []string, prefix string) (string, bool) {
	openFence := ""
	for _, line := range body {
		if consumeMarkdownFenceLine(&openFence, line) {
			continue
		}
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, prefix)), true
		}
	}
	return "", false
}
