package internal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

const protocolTestPreSHA = "1111111111111111111111111111111111111111"
const protocolTestPostSHA = "2222222222222222222222222222222222222222"

func protocolTestScope(phase, status, sha string) string {
	return fmt.Sprintf("%s-evidence scope: this command exited %s at %s; nothing else is inferred.", phase, status, sha)
}

func protocolTestReceiptID(protocol, algorithm string, occurrence int, heading, verify, digest, preOutput, postOutput string) string {
	preScope := protocolTestScope("Pre", "1", protocolTestPreSHA)
	postScope := protocolTestScope("Post", "0", protocolTestPostSHA)
	fields := []string{
		protocol,
		algorithm,
		"",
		strconv.Itoa(occurrence),
		heading,
		verify,
		digest,
		protocolTestPreSHA,
		"1",
		preOutput,
		preScope,
		protocolTestPostSHA,
		"0",
		postOutput,
		postScope,
	}
	var canonical strings.Builder
	for _, field := range fields {
		canonical.WriteString(strconv.Itoa(len([]byte(field))))
		canonical.WriteByte(':')
		canonical.WriteString(field)
	}
	sum := sha256.Sum256([]byte(canonical.String()))
	return "receipt-sha256-v1:" + hex.EncodeToString(sum[:])
}

func protocolTestVersionedReceipt(algorithm, digest string, occurrence int, heading, preOutput, postOutput string) string {
	const protocol = "evidence/v1"
	const verify = "echo hi"
	receiptID := protocolTestReceiptID(protocol, algorithm, occurrence, heading, verify, digest, preOutput, postOutput)
	return strings.Join([]string{
		"Evidence:",
		"Protocol: " + protocol,
		"Digest algorithm: " + algorithm,
		"Receipt ID: " + receiptID,
		verify,
		"unit digest: " + digest,
		"pre sha: " + protocolTestPreSHA,
		"pre exit status: 1",
		"```",
		preOutput,
		"```",
		protocolTestScope("Pre", "1", protocolTestPreSHA),
		"post sha: " + protocolTestPostSHA,
		"post exit status: 0",
		"```",
		postOutput,
		"```",
		protocolTestScope("Post", "0", protocolTestPostSHA),
	}, "\n") + "\n"
}

func protocolTestIdentityReceipt(algorithm, digest string, occurrence int, heading, preOutput, postOutput string) string {
	return strings.Join([]string{
		fmt.Sprintf("Unit occurrence: %d", occurrence),
		"Unit heading: " + heading,
		protocolTestVersionedReceipt(algorithm, digest, occurrence, heading, preOutput, postOutput),
	}, "\n")
}

func protocolTestDigestV0(unit queueUnit) string {
	fields := []string{unit.Heading}
	for _, prefix := range []string{"Read first:", "Constraints:", "Depends:", "Boundary:"} {
		if value, ok := queueUnitFieldValue(unit.Body, prefix); ok {
			fields = append(fields, value)
		}
	}
	doneMeans, _ := queueUnitFieldLines(unit.Body, "Done means:")
	fields = append(fields, strings.Join(doneMeans, "\n"))
	if scenarios, ok := queueUnitFieldLines(unit.Body, "Scenarios:"); ok {
		fields = append(fields, strings.Join(scenarios, "\n"))
	}
	if risks, ok := queueUnitFieldLines(unit.Body, "Risk cases:"); ok {
		fields = append(fields, strings.Join(risks, "\n"))
	}
	fields = append(fields, unitVerifyCommand(unit.Body))
	sum := sha256.Sum256([]byte(strings.Join(fields, "\x00")))
	return hex.EncodeToString(sum[:])
}

func protocolTestChunk(phase string, number, total int, algorithm, receiptID, digest, payload string, continued bool) string {
	lines := []string{
		"Raw output chunk:",
		"Protocol: evidence/v1",
		"Digest algorithm: " + algorithm,
		"Receipt ID: " + receiptID,
		"Output: " + phase,
		fmt.Sprintf("Chunk: %d/%d", number, total),
		"Unit occurrence: 1",
		"Unit heading: My outcome",
		"unit digest: " + digest,
		"```",
		payload,
		"```",
	}
	if continued {
		lines = append(lines, receiptContinuationMarker)
	}
	return strings.Join(lines, "\n")
}

func protocolTestChunkedReceiptComments(digest string) []string {
	preOutput := "pre one\npre two"
	postOutput := "post one\npost two"
	receiptID := protocolTestReceiptID("evidence/v1", "unit-contract-sha256-v1", 1, "My outcome", "echo hi", digest, preOutput, postOutput)
	first := strings.Join([]string{
		"Unit occurrence: 1",
		"Unit heading: My outcome",
		"Evidence:",
		"Protocol: evidence/v1",
		"Digest algorithm: unit-contract-sha256-v1",
		"Receipt ID: " + receiptID,
		"echo hi",
		"unit digest: " + digest,
		"pre sha: " + protocolTestPreSHA,
		"pre exit status: 1",
		protocolTestChunk("pre", 1, 2, "unit-contract-sha256-v1", receiptID, digest, "pre one\n", true),
	}, "\n")
	second := protocolTestChunk("pre", 2, 2, "unit-contract-sha256-v1", receiptID, digest, "pre two", true)
	third := strings.Join([]string{
		protocolTestScope("Pre", "1", protocolTestPreSHA),
		"post sha: " + protocolTestPostSHA,
		"post exit status: 0",
		protocolTestChunk("post", 1, 2, "unit-contract-sha256-v1", receiptID, digest, "post one\n", true),
	}, "\n")
	fourth := strings.Join([]string{
		protocolTestChunk("post", 2, 2, "unit-contract-sha256-v1", receiptID, digest, "post two", false),
		protocolTestScope("Post", "0", protocolTestPostSHA),
	}, "\n")
	return []string{first, second, third, fourth}
}

func TestVersionedEvidenceProtocol(t *testing.T) {
	source := "GH issue #1"
	body := ownedQueue("## My outcome\nDone means: something\nVerify:\n```\necho hi\n```\n- [x] done\n")
	units, unitIssues := ValidateQueueBody(body, source)
	if len(units) != 1 {
		t.Fatalf("expected one unit, got %d", len(units))
	}
	digest := unitContractDigest(units[0])

	t.Run("versioned and legacy receipts dispatch through fixed protocols", func(t *testing.T) {
		versioned := checkedUnit("echo hi", protocolTestVersionedReceipt("unit-contract-sha256-v1", digest, 1, "My outcome", "missing outcome", "output"))
		_, issues := ValidateQueueBody(ownedQueue(versioned), source)
		if len(issues) > 0 {
			t.Fatalf("versioned receipt was rejected: %v", issues)
		}

		legacy := checkedUnit("echo hi", evidenceReceipt("echo hi"))
		_, issues = ValidateQueueBody(ownedQueue(legacy), source)
		if len(issues) > 0 {
			t.Fatalf("legacy receipt was rejected: %v", issues)
		}
	})

	t.Run("unknown or partial receipt versions fail without legacy fallback", func(t *testing.T) {
		versioned := protocolTestVersionedReceipt("unit-contract-sha256-v1", digest, 1, "My outcome", "missing outcome", "output")
		cases := map[string]string{
			"unknown protocol":     strings.Replace(versioned, "Protocol: evidence/v1", "Protocol: evidence/v99", 1),
			"partial metadata":     strings.Replace(versioned, "Digest algorithm: unit-contract-sha256-v1\n", "", 1),
			"legacy field mixture": strings.Replace(versioned, "Protocol: evidence/v1\n", "", 1),
		}
		for name, evidence := range cases {
			t.Run(name, func(t *testing.T) {
				_, issues := ValidateQueueBody(ownedQueue(checkedUnit("echo hi", evidence)), source)
				if len(issues) == 0 {
					t.Fatal("expected compatibility error")
				}
			})
		}
	})

	t.Run("digest algorithm changes do not fabricate contract edits", func(t *testing.T) {
		algorithm := "unit-contract-sha256-v0"
		legacyDigest := protocolTestDigestV0(units[0])
		if legacyDigest == digest {
			t.Fatal("algorithm fixture must produce a distinct digest")
		}
		comments := []string{
			protocolTestIdentityReceipt("unit-contract-sha256-v1", digest, 1, "My outcome", "missing outcome", "output"),
			protocolTestIdentityReceipt(algorithm, legacyDigest, 1, "My outcome", "missing outcome", "output"),
		}
		unresolved, errors := unresolvedRebuildRequests(units, comments)
		if len(errors) > 0 {
			t.Fatalf("algorithm-only receipt change produced errors: %v", errors)
		}
		if len(unresolved) > 0 {
			t.Fatalf("algorithm-only receipt change produced unresolved requests: %v", unresolved)
		}

		changedUnits, _ := ValidateQueueBody(ownedQueue("## My outcome\nDone means: changed\nVerify:\n```\necho hi\n```\n- [x] done\n"), source)
		if len(scanQueueComments(changedUnits, []string{comments[1]}).errors) == 0 {
			t.Fatal("real contract change must still require an amendment edge")
		}
	})

	t.Run("stable receipt identity survives continuation and chunking", func(t *testing.T) {
		comments := protocolTestChunkedReceiptComments(digest)
		result := &ValidationResult{Valid: true}
		applyQueueIssues(result, "GitHub comments", units, unitIssues, comments)
		if len(result.Errors) > 0 {
			t.Fatalf("versioned chunked receipt was rejected: %v", result.Errors)
		}

		changed := append([]string(nil), comments...)
		changed[1] = strings.Replace(changed[1], "Receipt ID: receipt-sha256-v1:", "Receipt ID: receipt-sha256-v1:0", 1)
		result = &ValidationResult{Valid: true}
		applyQueueIssues(result, "GitHub comments", units, unitIssues, changed)
		if len(result.Errors) == 0 {
			t.Fatal("changed chunk identity must fail validation")
		}
	})
}
