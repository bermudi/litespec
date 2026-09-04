package internal

import (
	"strings"
	"testing"
)

func bodyReceiptTestQueueBody(verify, evidence string) string {
	return ownedQueue(strings.Join([]string{
		"## My outcome",
		"Done means: something",
		"Verify:",
		"```",
		verify,
		"```",
		evidence,
		"- [x] done",
	}, "\n"))
}

func bodyReceiptTestVersionedReceipt(identity queueUnitIdentity, algorithm, digest, verify string) string {
	receipt := parsedEvidenceReceipt{
		header: evidenceReceiptHeader{
			protocol:        evidenceProtocolV1,
			digestAlgorithm: algorithm,
			versioned:       true,
		},
		identity:   &identity,
		heading:    identity.Heading,
		verify:     verify,
		digest:     digest,
		preSHA:     recoveryTestPreSHA,
		preStatus:  "1",
		preOutput:  "fallback failure",
		preScope:   "Pre-evidence scope: this command exited 1 at " + recoveryTestPreSHA + "; nothing else is inferred.",
		postSHA:    recoveryTestPostSHA,
		postStatus: "0",
		postOutput: "fallback success",
		postScope:  "Post-evidence scope: this command exited 0 at " + recoveryTestPostSHA + "; nothing else is inferred.",
	}
	receipt.header.receiptID = receiptIDForCanonicalReceipt(receipt)
	return recoveryTestIdentityEvidence(identity, strings.Join([]string{
		"Evidence:",
		"Protocol: " + evidenceProtocolV1,
		"Digest algorithm: " + algorithm,
		"Receipt ID: " + receipt.header.receiptID,
		verify,
		"unit digest: " + digest,
		"pre sha: " + recoveryTestPreSHA,
		"pre exit status: 1",
		"```",
		receipt.preOutput,
		"```",
		receipt.preScope,
		"post sha: " + recoveryTestPostSHA,
		"post exit status: 0",
		"```",
		receipt.postOutput,
		"```",
		receipt.postScope,
	}, "\n")) + "\n"
}

func bodyReceiptTestSupersededContract(unit queueUnit) (string, string) {
	historical := unit
	historical.Body = append([]string(nil), unit.Body...)
	for i, line := range historical.Body {
		switch line {
		case "- [outcome] something":
			historical.Body[i] = "- [outcome] historical something"
		case "echo original":
			historical.Body[i] = "echo historical"
		}
	}
	return unitContractDigest(historical), "echo historical"
}

func TestBodyReceiptDeclarationFallback(t *testing.T) {
	source := "GH issue #15"
	identity := recoveryTestIdentity("My outcome", 1)
	const currentVerify = "echo original"
	const editedVerify = "echo edited"

	cleanUnits, _ := ValidateQueueBody(bodyReceiptTestQueueBody(currentVerify, ""), source)
	if len(cleanUnits) != 1 {
		t.Fatalf("expected one unit, got %d", len(cleanUnits))
	}
	currentDigest := unitContractDigest(cleanUnits[0])

	t.Run("current-digest body receipt with edited Verify is rejected", func(t *testing.T) {
		tampered := bodyReceiptTestVersionedReceipt(identity, digestAlgorithmV1, currentDigest, editedVerify)
		units, unitIssues := ValidateQueueBody(bodyReceiptTestQueueBody(currentVerify, tampered), source)
		result := &ValidationResult{Valid: true}
		applyQueueIssues(result, source, units, unitIssues, nil)
		if !recoveryTestHasErrorContaining(result, "must quote the Verify command verbatim") {
			t.Fatalf("current-digest body receipt with edited Verify was accepted: %v", result.Errors)
		}
	})

	t.Run("body and comment lanes reject the tampered receipt consistently", func(t *testing.T) {
		tampered := bodyReceiptTestVersionedReceipt(identity, digestAlgorithmV1, currentDigest, editedVerify)

		bodyUnits, bodyIssues := ValidateQueueBody(bodyReceiptTestQueueBody(currentVerify, tampered), source)
		bodyResult := &ValidationResult{Valid: true}
		applyQueueIssues(bodyResult, source, bodyUnits, bodyIssues, nil)
		if len(bodyResult.Errors) == 0 {
			t.Fatal("body lane accepted the tampered receipt")
		}

		cleanUnits, cleanIssues := ValidateQueueBody(bodyReceiptTestQueueBody(currentVerify, ""), source)
		commentResult := &ValidationResult{Valid: true}
		applyQueueIssues(commentResult, source, cleanUnits, cleanIssues, []string{tampered})
		if len(commentResult.Errors) == 0 {
			t.Fatal("comment lane accepted the tampered receipt")
		}

		document := evidencePayloadDocument(newEvidenceDocument(tampered))
		for _, lane := range []string{"queue", "comment"} {
			if _, ok := completeEvidenceReceiptObservation(document, identity, cleanUnits, lane); ok {
				t.Fatalf("declaration fallback accepted a current-digest receipt in the %s lane", lane)
			}
		}
	})

	t.Run("historical body receipt recovery keeps working", func(t *testing.T) {
		oldDigest, oldVerify := bodyReceiptTestSupersededContract(cleanUnits[0])
		oldEvidence := recoveryTestLegacyReceipt(oldDigest, oldVerify)
		oldID := recoveryTestReceiptID(t, oldEvidence, oldVerify, identity, oldDigest)
		body := bodyReceiptTestQueueBody(currentVerify, oldEvidence)
		units, unitIssues := ValidateQueueBody(body, source)
		amendment := recoveryTestAmendment(identity, oldDigest, currentDigest)
		recovery := recoveryTestVersionedReceipt(identity, currentDigest, currentVerify, oldID)

		unamended := &ValidationResult{Valid: true}
		applyQueueIssues(unamended, source, units, unitIssues, []string{recovery})
		if !recoveryTestHasErrorContaining(unamended, "observed receipt digests") {
			t.Fatalf("unwitnessed superseded body receipt was accepted: %v", unamended.Errors)
		}

		amended := &ValidationResult{Valid: true}
		applyQueueIssues(amended, source, units, unitIssues, []string{amendment, recovery})
		if len(amended.Errors) > 0 {
			t.Fatalf("witnessed historical body receipt recovery was rejected: %v", amended.Errors)
		}
	})

	t.Run("algorithm-equivalent normalization is not regressed", func(t *testing.T) {
		v0Digest, ok := unitContractDigestForAlgorithm(cleanUnits[0], digestAlgorithmV0)
		if !ok || v0Digest == currentDigest {
			t.Fatal("algorithm fixture must produce a distinct digest")
		}

		equivalent := bodyReceiptTestVersionedReceipt(identity, digestAlgorithmV0, v0Digest, currentVerify)
		units, unitIssues := ValidateQueueBody(bodyReceiptTestQueueBody(currentVerify, equivalent), source)
		if len(unitIssues) > 0 {
			t.Fatalf("algorithm-equivalent body receipt was rejected: %v", unitIssues)
		}
		result := &ValidationResult{Valid: true}
		applyQueueIssues(result, source, units, unitIssues, nil)
		if len(result.Errors) > 0 {
			t.Fatalf("algorithm-equivalent body receipt fabricated a contract transition: %v", result.Errors)
		}

		tampered := bodyReceiptTestVersionedReceipt(identity, digestAlgorithmV0, v0Digest, editedVerify)
		tamperedUnits, tamperedIssues := ValidateQueueBody(bodyReceiptTestQueueBody(currentVerify, tampered), source)
		tamperedResult := &ValidationResult{Valid: true}
		applyQueueIssues(tamperedResult, source, tamperedUnits, tamperedIssues, nil)
		if !recoveryTestHasErrorContaining(tamperedResult, "must quote the Verify command verbatim") {
			t.Fatalf("algorithm-equivalent current-digest receipt with edited Verify was accepted: %v", tamperedResult.Errors)
		}
	})
}
