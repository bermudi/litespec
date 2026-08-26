package internal

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strconv"
	"strings"
)

var unitDigestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

// unitContractDigest binds a unit's contract text: heading, optional Read
// first/Constraints/Depends values, Done means value, and the Verify command
// content, each length-prefixed in that fixed order. Status checkbox,
// Evidence content, and every other unit line are excluded by design.
func unitContractDigest(unit queueUnit) string {
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
	doneMeans, _ := queueUnitFieldValue(unit.Body, "Done means:")
	fields = append(fields, contractFieldBytes(doneMeans))
	fields = append(fields, contractFieldBytes(unitVerifyCommand(unit.Body)))

	var b strings.Builder
	for _, f := range fields {
		b.WriteString(strconv.Itoa(len(f)))
		b.WriteByte(':')
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
