package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const recoveryTestPreSHA = "3333333333333333333333333333333333333333"
const recoveryTestPostSHA = "4444444444444444444444444444444444444444"

func recoveryTestIdentity(heading string, occurrence int) queueUnitIdentity {
	return queueUnitIdentity{Occurrence: occurrence, Heading: heading}
}

func recoveryTestLegacyReceipt(digest, verify string) string {
	return strings.Join([]string{
		"Evidence:",
		verify,
		"unit digest: " + digest,
		"pre sha: " + recoveryTestPreSHA,
		"pre exit status: 1",
		"```",
		"historical failure",
		"```",
		"Pre-evidence scope: this command exited 1 at " + recoveryTestPreSHA + "; nothing else is inferred.",
		"post sha: " + recoveryTestPostSHA,
		"post exit status: 0",
		"```",
		"historical success",
		"```",
		"Post-evidence scope: this command exited 0 at " + recoveryTestPostSHA + "; nothing else is inferred.",
	}, "\n") + "\n"
}

func recoveryTestReceiptID(t *testing.T, evidence string, verify string, identity queueUnitIdentity, digest string) string {
	t.Helper()
	document := evidencePayloadDocument(newEvidenceDocument(evidence))
	receipt, issues := parseEvidenceReceiptDocument(
		document,
		verify,
		digest,
		"test",
		identity.Heading,
		&identity,
	)
	if len(issues) > 0 {
		t.Fatalf("receipt fixture is invalid: %v", issues)
	}
	return receipt.header.receiptID
}

func recoveryTestIdentityEvidence(identity queueUnitIdentity, evidence string) string {
	return strings.Join([]string{
		fmt.Sprintf("Unit occurrence: %d", identity.Occurrence),
		"Unit heading: " + identity.Heading,
		evidence,
	}, "\n")
}

func recoveryTestVersionedReceipt(identity queueUnitIdentity, digest, verify, recoveredFrom string) string {
	receipt := parsedEvidenceReceipt{
		header: evidenceReceiptHeader{
			protocol:        evidenceProtocolV1,
			digestAlgorithm: digestAlgorithmV1,
			recoveredFrom:   recoveredFrom,
			versioned:       true,
		},
		identity:   &identity,
		heading:    identity.Heading,
		verify:     verify,
		digest:     digest,
		preSHA:     recoveryTestPreSHA,
		preStatus:  "1",
		preOutput:  "recovered failure",
		preScope:   "Pre-evidence scope: this command exited 1 at " + recoveryTestPreSHA + "; nothing else is inferred.",
		postSHA:    recoveryTestPostSHA,
		postStatus: "0",
		postOutput: "recovered success",
		postScope:  "Post-evidence scope: this command exited 0 at " + recoveryTestPostSHA + "; nothing else is inferred.",
	}
	receipt.header.receiptID = receiptIDForCanonicalReceipt(receipt)
	return recoveryTestIdentityEvidence(identity, strings.Join([]string{
		"Evidence:",
		"Protocol: " + evidenceProtocolV1,
		"Digest algorithm: " + digestAlgorithmV1,
		"Receipt ID: " + receipt.header.receiptID,
		func() string {
			if recoveredFrom == "" {
				return ""
			}
			return "Recovered from: " + recoveredFrom
		}(),
		verify,
		"unit digest: " + digest,
		"pre sha: " + recoveryTestPreSHA,
		"pre exit status: 1",
		"```",
		"recovered failure",
		"```",
		receipt.preScope,
		"post sha: " + recoveryTestPostSHA,
		"post exit status: 0",
		"```",
		"recovered success",
		"```",
		receipt.postScope,
	}, "\n"))
}

func recoveryTestQueueBody(evidence string, checked bool) string {
	status := "- [ ] pending"
	if checked {
		status = "- [x] done"
	}
	return ownedQueue(strings.Join([]string{
		"## My outcome",
		"Done means: something",
		"Verify:",
		"```",
		"echo hi",
		"```",
		evidence,
		status,
	}, "\n"))
}

func recoveryTestLegacyIdentityReceipt(t *testing.T, identity queueUnitIdentity, digest, verify string) (string, string) {
	t.Helper()
	evidence := recoveryTestLegacyReceipt(digest, verify)
	identityEvidence := recoveryTestIdentityEvidence(identity, evidence)
	return identityEvidence, recoveryTestReceiptID(t, identityEvidence, verify, identity, digest)
}

func recoveryTestResult(t *testing.T, body string, comments []string) *ValidationResult {
	t.Helper()
	units, unitIssues := ValidateQueueBody(body, "GH issue #15")
	result := &ValidationResult{Valid: true}
	applyQueueIssues(result, "GitHub comments", units, unitIssues, comments)
	return result
}

func TestAppendOnlyEvidenceRecovery(t *testing.T) {
	t.Run("append-only recovery resolves prior complete evidence in both queue lanes", func(t *testing.T) {
		body := recoveryTestQueueBody("", true)
		units, _ := ValidateQueueBody(body, "GH issue #15")
		identity := recoveryTestIdentity("My outcome", 1)
		digest := unitContractDigest(units[0])
		oldReceipt, oldID := recoveryTestLegacyIdentityReceipt(t, identity, digest, "echo hi")
		body = recoveryTestQueueBody(oldReceipt, true)
		units, issues := ValidateQueueBody(body, "GH issue #15")
		if len(issues) > 0 {
			t.Fatalf("historical body receipt was rejected: %v", issues)
		}
		recovery := recoveryTestVersionedReceipt(identity, digest, "echo hi", oldID)

		result := &ValidationResult{Valid: true}
		applyQueueIssues(result, "GitHub comments", units, issues, []string{
			formatRebuildRequest(identity),
			recovery,
		})
		if len(result.Errors) > 0 {
			t.Fatalf("GitHub recovery was rejected: %v", result.Errors)
		}

		root := t.TempDir()
		queuePath := filepath.Join(root, "specs", "queues", "recovery.md")
		if err := os.MkdirAll(filepath.Dir(queuePath), 0o755); err != nil {
			t.Fatalf("mkdir queue: %v", err)
		}
		localBody := body + "\n" + recovery
		if err := os.WriteFile(queuePath, []byte(localBody), 0o644); err != nil {
			t.Fatalf("write queue: %v", err)
		}
		localResult, err := ValidateQueueFile(queuePath)
		if err != nil {
			t.Fatalf("ValidateQueueFile: %v", err)
		}
		if !localResult.Valid || len(localResult.Errors) > 0 {
			t.Fatalf("local recovery was rejected: %+v", localResult.Errors)
		}
		stripped, metadata := splitLocalQueueMetadataBlocks(localBody)
		if !strings.Contains(stripped, oldReceipt) || !strings.Contains(localBody, recovery) {
			t.Fatal("local recovery did not preserve the prior evidence and appended receipt")
		}
		if len(metadata) != 1 || metadata[0] != recovery {
			t.Fatalf("local metadata = %q, want the appended recovery receipt", metadata)
		}
	})

	t.Run("recovery rejects broken references and receipt identity drift", func(t *testing.T) {
		body := recoveryTestQueueBody("", true)
		units, _ := ValidateQueueBody(body, "GH issue #15")
		identity := recoveryTestIdentity("My outcome", 1)
		digest := unitContractDigest(units[0])
		_, oldID := recoveryTestLegacyIdentityReceipt(t, identity, digest, "echo hi")

		dangling := recoveryTestVersionedReceipt(identity, digest, "echo hi", "receipt-sha256-v1:"+strings.Repeat("0", 64))
		if result := recoveryTestResult(t, body, []string{dangling}); len(result.Errors) == 0 {
			t.Fatal("dangling recovery reference was accepted")
		}

		otherIdentity := recoveryTestIdentity("Other outcome", 1)
		myIdentity := recoveryTestIdentity("My outcome", 1)
		otherBody := ownedQueue(strings.Join([]string{
			"## Other outcome",
			"Done means: other",
			"Verify:",
			"```",
			"echo hi",
			"```",
			"- [x] done",
			"",
			"## My outcome",
			"Done means: something",
			"Verify:",
			"```",
			"echo hi",
			"```",
			"- [x] done",
		}, "\n"))
		otherUnits, _ := ValidateQueueBody(otherBody, "GH issue #15")
		otherDigest := unitContractDigest(otherUnits[0])
		myDigest := unitContractDigest(otherUnits[1])
		otherReceipt, otherID := recoveryTestLegacyIdentityReceipt(t, otherIdentity, otherDigest, "echo hi")
		myReceipt, _ := recoveryTestLegacyIdentityReceipt(t, myIdentity, myDigest, "echo hi")
		otherBody = ownedQueue(strings.Join([]string{
			"## Other outcome",
			"Done means: other",
			"Verify:",
			"```",
			"echo hi",
			"```",
			otherReceipt,
			"- [x] done",
			"",
			"## My outcome",
			"Done means: something",
			"Verify:",
			"```",
			"echo hi",
			"```",
			myReceipt,
			"- [x] done",
		}, "\n"))
		otherUnits, otherIssues := ValidateQueueBody(otherBody, "GH issue #15")
		if len(otherIssues) > 0 {
			t.Fatalf("historical two-unit body was rejected: %v", otherIssues)
		}
		mismatched := recoveryTestVersionedReceipt(myIdentity, myDigest, "echo hi", otherID)
		result := &ValidationResult{Valid: true}
		applyQueueIssues(result, "GitHub comments", otherUnits, otherIssues, []string{mismatched})
		if len(result.Errors) == 0 {
			t.Fatal("recovery reference to another unit was accepted")
		}

		divergent := strings.Replace(
			recoveryTestVersionedReceipt(identity, digest, "echo hi", oldID),
			"recovered success",
			"divergent success",
			1,
		)
		if result := recoveryTestResult(t, body, []string{divergent}); len(result.Errors) == 0 {
			t.Fatal("divergent duplicate receipt identity was accepted")
		}

		chunked := recoveryTestChunkedReceipt(identity, digest, oldID)
		chunked[1] = strings.Replace(chunked[1], "Unit heading: My outcome", "Unit heading: Other outcome", 1)
		if result := recoveryTestResult(t, body, chunked); len(result.Errors) == 0 {
			t.Fatal("changed continuation identity was accepted")
		}
	})

	t.Run("unauthenticated supersession and quarantine do not become controls", func(t *testing.T) {
		body := recoveryTestQueueBody("", true)
		units, issues := ValidateQueueBody(body, "GH issue #15")
		identity := recoveryTestIdentity("My outcome", 1)
		digest := unitContractDigest(units[0])
		oldReceipt, oldID := recoveryTestLegacyIdentityReceipt(t, identity, digest, "echo hi")
		recovery := recoveryTestVersionedReceipt(identity, digest, "echo hi", oldID)
		malformed := strings.TrimSuffix(oldReceipt, "Post-evidence scope: this command exited 0 at "+recoveryTestPostSHA+"; nothing else is inferred.\n")
		for _, control := range []string{
			"Supersedes: " + oldID,
			"quarantine: " + oldID,
			"Authorized by: repository-controller",
		} {
			result := &ValidationResult{Valid: true}
			applyQueueIssues(result, "GitHub comments", units, issues, []string{
				oldReceipt,
				malformed,
				recovery,
				control,
			})
			if len(result.Errors) == 0 {
				t.Fatalf("control %q hid the malformed historical record", control)
			}
		}
	})
}

func recoveryTestChunkedReceipt(identity queueUnitIdentity, digest, recoveredFrom string) []string {
	preOutput := "recovered pre one\nrecovered pre two"
	postOutput := "recovered post one\nrecovered post two"
	receipt := parsedEvidenceReceipt{
		header: evidenceReceiptHeader{
			protocol:        evidenceProtocolV1,
			digestAlgorithm: digestAlgorithmV1,
			recoveredFrom:   recoveredFrom,
			versioned:       true,
		},
		identity:   &identity,
		heading:    identity.Heading,
		verify:     "echo hi",
		digest:     digest,
		preSHA:     recoveryTestPreSHA,
		preStatus:  "1",
		preOutput:  preOutput,
		preScope:   "Pre-evidence scope: this command exited 1 at " + recoveryTestPreSHA + "; nothing else is inferred.",
		postSHA:    recoveryTestPostSHA,
		postStatus: "0",
		postOutput: postOutput,
		postScope:  "Post-evidence scope: this command exited 0 at " + recoveryTestPostSHA + "; nothing else is inferred.",
	}
	receipt.header.receiptID = receiptIDForCanonicalReceipt(receipt)
	header := func() []string {
		return []string{
			"Protocol: " + evidenceProtocolV1,
			"Digest algorithm: " + digestAlgorithmV1,
			"Receipt ID: " + receipt.header.receiptID,
			"Recovered from: " + recoveredFrom,
		}
	}
	chunk := func(phase string, number, total int, payload string, continued bool) string {
		lines := []string{
			"Raw output chunk:",
		}
		lines = append(lines, header()...)
		lines = append(lines,
			"Output: "+phase,
			fmt.Sprintf("Chunk: %d/%d", number, total),
			fmt.Sprintf("Unit occurrence: %d", identity.Occurrence),
			"Unit heading: "+identity.Heading,
			"unit digest: "+digest,
			"```",
			payload,
			"```",
		)
		if continued {
			lines = append(lines, receiptContinuationMarker)
		}
		return strings.Join(lines, "\n")
	}
	first := strings.Join(append([]string{
		fmt.Sprintf("Unit occurrence: %d", identity.Occurrence),
		"Unit heading: " + identity.Heading,
		"Evidence:",
		"Protocol: " + evidenceProtocolV1,
		"Digest algorithm: " + digestAlgorithmV1,
		"Receipt ID: " + receipt.header.receiptID,
		"Recovered from: " + recoveredFrom,
		"echo hi",
		"unit digest: " + digest,
		"pre sha: " + recoveryTestPreSHA,
		"pre exit status: 1",
	}, chunk("pre", 1, 2, "recovered pre one\n", true)), "\n")
	second := chunk("pre", 2, 2, "recovered pre two", true)
	third := strings.Join([]string{
		"Pre-evidence scope: this command exited 1 at " + recoveryTestPreSHA + "; nothing else is inferred.",
		"post sha: " + recoveryTestPostSHA,
		"post exit status: 0",
		chunk("post", 1, 2, "recovered post one\n", true),
	}, "\n")
	fourth := strings.Join([]string{
		chunk("post", 2, 2, "recovered post two", false),
		"Post-evidence scope: this command exited 0 at " + recoveryTestPostSHA + "; nothing else is inferred.",
	}, "\n")
	return []string{first, second, third, fourth}
}
