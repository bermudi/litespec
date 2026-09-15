package internal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var evidenceV2ElisionMarkerPattern = regexp.MustCompile(`^\.\.\. ([1-9][0-9]*) bytes elided \.\.\.$`)
var evidenceV2OutputSHAPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

// receiptV2DocumentBoundsIssue reports the first document-level v2 bound the
// receipt violates: the fixed byte budget, measured over the same span
// assembly budgets — the wrapper lines dropped ahead of the payload
// (heading or identity lines plus the Evidence: label) together with the
// payload — or the single-comment rule: a v2 receipt never ends with the
// continuation marker.
func receiptV2DocumentBoundsIssue(document evidenceDocument, payloadBytes int) string {
	size := document.prefixBytes + payloadBytes
	if size > receiptV2TotalBudget {
		return fmt.Sprintf("exceeds the %d-byte evidence/v2 receipt budget (%d bytes including the heading wrapper)", receiptV2TotalBudget, size)
	}
	for i := len(document.lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(document.lines[i]) == "" {
			continue
		}
		if document.lines[i] == receiptContinuationMarker {
			return "must not end with the continuation marker: evidence/v2 is a single-comment protocol"
		}
		return ""
	}
	return ""
}

// receiptV2EnforceOutputBounds enforces one run's declared metadata against
// its fenced payload: zero bytes require the exact `<no output>` fence, an
// unelided fence must match the declared byte count and hash to the declared
// SHA-256, and an elided fence must satisfy the reconstruction arithmetic.
func receiptV2EnforceOutputBounds(phase, payload, declaredBytes, declaredSHA string, fail func(string)) {
	declared, err := strconv.Atoi(declaredBytes)
	if err != nil || declared < 0 {
		fail(fmt.Sprintf("declares %s bytes %q, which is not a byte count", phase, declaredBytes))
		return
	}
	if !evidenceV2OutputSHAPattern.MatchString(declaredSHA) {
		fail(fmt.Sprintf("declares %s output sha256 %q, which is not 64 lowercase hexadecimal characters", phase, declaredSHA))
		return
	}
	if declared == 0 {
		if payload != receiptOutputField("") {
			fail(fmt.Sprintf("declares 0 %s bytes but the fence does not hold exactly `<no output>`", phase))
		}
		return
	}
	if len(payload) == declared {
		sum := sha256.Sum256([]byte(payload))
		if hex.EncodeToString(sum[:]) != declaredSHA {
			fail(fmt.Sprintf("declares %d unelided %s bytes but the fence does not hash to the declared %s output sha256", declared, phase, phase))
		}
		return
	}
	receiptV2EnforceElisionArithmetic(phase, payload, declared, fail)
}

// receiptV2EnforceElisionArithmetic checks an elided fence: exactly one
// canonical elision marker — a positive count without leading zeros, the
// only form assembly emits — and fence bytes minus the marker plus the
// elided count must equal the declared byte count. Marker-shaped lines
// that deviate from the canonical form count as output content; a
// spurious canonical marker inside head or tail inflates the fence bytes,
// breaks the arithmetic, and surfaces here as a visible error.
func receiptV2EnforceElisionArithmetic(phase, payload string, declared int, fail func(string)) {
	lines := strings.Split(payload, "\n")
	var markers []int
	for i, line := range lines {
		if evidenceV2ElisionMarkerPattern.MatchString(line) {
			markers = append(markers, i)
		}
	}
	switch {
	case len(markers) == 0:
		fail(fmt.Sprintf("declares %d %s bytes but the fence neither holds the full output nor carries an elision marker", declared, phase))
	case len(markers) > 1:
		fail(fmt.Sprintf("carries multiple elision-marker lines, so the %s output reconstruction arithmetic is ambiguous", phase))
	default:
		index := markers[0]
		elided, _ := strconv.Atoi(evidenceV2ElisionMarkerPattern.FindStringSubmatch(lines[index])[1])
		separator := 0
		if index != len(lines)-1 {
			separator = 1
		}
		covered := len(payload) - len(lines[index]) - separator + elided
		if covered != declared {
			fail(fmt.Sprintf("reconstruction arithmetic does not hold: fence bytes plus %d elided cover %d %s bytes, not the declared %d", elided, covered, phase, declared))
		}
	}
}
