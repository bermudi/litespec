package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const littleGoblinPreSHA = "5555555555555555555555555555555555555555"
const littleGoblinPostSHA = "6666666666666666666666666666666666666666"

type littleGoblinUnitSpec struct {
	heading  string
	depends  string
	done     string
	scenario string
	verify   string
	checked  bool
}

func littleGoblinQueueBody(specs ...littleGoblinUnitSpec) string {
	lines := []string{
		"Base: 7777777777777777777777777777777777777777",
		"Branch: litespec/little-goblin-regression",
		"",
	}
	for i, spec := range specs {
		if i > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, "## "+spec.heading)
		if spec.depends != "" {
			lines = append(lines, "Depends: "+spec.depends)
		}
		lines = append(lines,
			"Done means:",
			"- [outcome] "+spec.done,
			"Scenarios:",
			"- [outcome] "+spec.scenario,
			"Verify:",
			"```bash",
			spec.verify,
			"```",
		)
		status := "- [ ] pending"
		if spec.checked {
			status = "- [x] done"
		}
		lines = append(lines, status)
	}
	return strings.Join(lines, "\n") + "\n"
}

func littleGoblinHistoricalUnit(spec littleGoblinUnitSpec) queueUnit {
	lines := make([]string, 0, 8)
	if spec.depends != "" {
		lines = append(lines, "Depends: "+spec.depends)
	}
	lines = append(lines,
		"Done means:",
		"- [outcome] "+spec.done,
		"Scenarios:",
		"- [outcome] "+spec.scenario,
		"Verify:",
		"```bash",
		spec.verify,
		"```",
	)
	return queueUnit{Heading: spec.heading, Body: lines}
}

func littleGoblinIdentityEvidence(identity queueUnitIdentity, evidence string) string {
	return strings.Join([]string{
		fmt.Sprintf("Unit occurrence: %d", identity.Occurrence),
		"Unit heading: " + identity.Heading,
		evidence,
	}, "\n")
}

func littleGoblinReceipt(identity queueUnitIdentity, digest, verify, recoveredFrom, preOutput, postOutput string, versioned bool) (string, string) {
	protocol := evidenceProtocolLegacy
	algorithm := digestAlgorithmV1
	if versioned {
		protocol = evidenceProtocolV1
	}
	receipt := parsedEvidenceReceipt{
		header: evidenceReceiptHeader{
			protocol:        protocol,
			digestAlgorithm: algorithm,
			recoveredFrom:   recoveredFrom,
			versioned:       versioned,
		},
		identity:   &identity,
		heading:    identity.Heading,
		verify:     verify,
		digest:     digest,
		preSHA:     littleGoblinPreSHA,
		preStatus:  "1",
		preOutput:  preOutput,
		preScope:   "Pre-evidence scope: this command exited 1 at " + littleGoblinPreSHA + "; nothing else is inferred.",
		postSHA:    littleGoblinPostSHA,
		postStatus: "0",
		postOutput: postOutput,
		postScope:  "Post-evidence scope: this command exited 0 at " + littleGoblinPostSHA + "; nothing else is inferred.",
	}
	receipt.header.receiptID = receiptIDForCanonicalReceipt(receipt)

	lines := []string{
		"Evidence:",
	}
	if versioned {
		lines = append(lines,
			"Protocol: "+protocol,
			"Digest algorithm: "+algorithm,
			"Receipt ID: "+receipt.header.receiptID,
		)
		if recoveredFrom != "" {
			lines = append(lines, "Recovered from: "+recoveredFrom)
		}
	}
	lines = append(lines,
		verify,
		"unit digest: "+digest,
		"pre sha: "+littleGoblinPreSHA,
		"pre exit status: 1",
		"```",
		preOutput,
		"```",
		receipt.preScope,
		"post sha: "+littleGoblinPostSHA,
		"post exit status: 0",
		"```",
		postOutput,
		"```",
		receipt.postScope,
	)
	return littleGoblinIdentityEvidence(identity, strings.Join(lines, "\n")), receipt.header.receiptID
}

func littleGoblinHeadingFormReceipt(identity queueUnitIdentity, receipt string) string {
	prefix := strings.Join([]string{
		fmt.Sprintf("Unit occurrence: %d", identity.Occurrence),
		"Unit heading: " + identity.Heading,
	}, "\n") + "\n"
	return "## " + identity.Heading + "\n" + strings.TrimPrefix(receipt, prefix)
}

func littleGoblinErrorContains(result *ValidationResult, text string) bool {
	for _, issue := range result.Errors {
		if strings.Contains(issue.Message, text) {
			return true
		}
	}
	return false
}

func littleGoblinAmendment(identity queueUnitIdentity, oldDigest, newDigest, reason string) string {
	return strings.Join([]string{
		"Amendment:",
		fmt.Sprintf("Unit occurrence: %d", identity.Occurrence),
		"Unit heading: " + identity.Heading,
		"Old digest: " + oldDigest,
		"New digest: " + newDigest,
		"Reason: " + reason,
	}, "\n")
}

func littleGoblinChunk(phase string, number, total int, identity queueUnitIdentity, digest, payload string, continued bool) string {
	lines := []string{
		"Raw output chunk:",
		"Output: " + phase,
		fmt.Sprintf("Chunk: %d/%d", number, total),
		fmt.Sprintf("Unit occurrence: %d", identity.Occurrence),
		"Unit heading: " + identity.Heading,
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

func littleGoblinSplitOutput(output string) (string, string) {
	split := len(output) / 2
	if newline := strings.LastIndex(output[:split], "\n"); newline >= 0 {
		split = newline + 1
	}
	return output[:split], output[split:]
}

func littleGoblinChunkedLegacyReceipt(identity queueUnitIdentity, digest, verify, preOutput, postOutput string) ([]string, string) {
	preFirst, preSecond := littleGoblinSplitOutput(preOutput)
	postFirst, postSecond := littleGoblinSplitOutput(postOutput)
	receipt := parsedEvidenceReceipt{
		header: evidenceReceiptHeader{
			protocol:        evidenceProtocolLegacy,
			digestAlgorithm: digestAlgorithmV1,
		},
		identity:   &identity,
		heading:    identity.Heading,
		verify:     verify,
		digest:     digest,
		preSHA:     littleGoblinPreSHA,
		preStatus:  "1",
		preOutput:  preOutput,
		preScope:   "Pre-evidence scope: this command exited 1 at " + littleGoblinPreSHA + "; nothing else is inferred.",
		postSHA:    littleGoblinPostSHA,
		postStatus: "0",
		postOutput: postOutput,
		postScope:  "Post-evidence scope: this command exited 0 at " + littleGoblinPostSHA + "; nothing else is inferred.",
	}
	receipt.header.receiptID = receiptIDForCanonicalReceipt(receipt)
	chunk := func(phase string, number int, payload string, continued bool) string {
		return littleGoblinChunk(phase, number, 2, identity, digest, payload, continued)
	}
	first := strings.Join([]string{
		fmt.Sprintf("Unit occurrence: %d", identity.Occurrence),
		"Unit heading: " + identity.Heading,
		"Evidence:",
		verify,
		"unit digest: " + digest,
		"pre sha: " + littleGoblinPreSHA,
		"pre exit status: 1",
		chunk("pre", 1, preFirst, true),
	}, "\n")
	second := chunk("pre", 2, preSecond, true)
	third := strings.Join([]string{
		receipt.preScope,
		"post sha: " + littleGoblinPostSHA,
		"post exit status: 0",
		chunk("post", 1, postFirst, true),
	}, "\n")
	fourth := strings.Join([]string{
		chunk("post", 2, postSecond, false),
		receipt.postScope,
	}, "\n")
	return []string{first, second, third, fourth}, receipt.header.receiptID
}

func littleGoblinValidationResult(t *testing.T, body string, comments []string) *ValidationResult {
	t.Helper()
	units, unitIssues := ValidateQueueBody(body, "synthetic little-goblin fixture")
	result := &ValidationResult{Valid: true}
	applyQueueIssues(result, "synthetic comments", units, unitIssues, comments)
	return result
}

func TestLittleGoblinEvidenceRegressionFixtures(t *testing.T) {
	t.Run("issue #52 historical receipts survive parser and digest evolution through append-only recovery", func(t *testing.T) {
		identity := queueUnitIdentity{Occurrence: 1, Heading: "Historical receipt routing"}
		current := littleGoblinUnitSpec{
			heading:  identity.Heading,
			done:     "the current receipt interpretation is routable",
			scenario: "issue #52 historical receipts survive parser and digest evolution through append-only recovery",
			verify:   "echo current historical receipt",
			checked:  true,
		}
		body := littleGoblinQueueBody(current)
		units, unitIssues := ValidateQueueBody(body, "synthetic little-goblin #52")
		if len(units) != 1 {
			t.Fatalf("expected one #52-shaped unit, got %d", len(units))
		}
		if len(unitIssues) != 1 || !strings.Contains(unitIssues[0].Message, "missing Evidence receipt") {
			t.Fatalf("expected only the checked-unit evidence gap before comments, got %v", unitIssues)
		}

		currentDigest := unitContractDigest(units[0])
		historicalOne := littleGoblinHistoricalUnit(littleGoblinUnitSpec{
			heading:  identity.Heading,
			done:     "the original receipt interpretation is retained",
			scenario: "historical parser interpretation one",
			verify:   "echo historical receipt one",
		})
		historicalTwo := littleGoblinHistoricalUnit(littleGoblinUnitSpec{
			heading:  identity.Heading,
			done:     "the revised receipt interpretation is retained",
			scenario: "historical parser interpretation two",
			verify:   "echo historical receipt two",
		})
		oldDigest := unitContractDigest(historicalOne)
		middleDigest := unitContractDigest(historicalTwo)
		if oldDigest == middleDigest || middleDigest == currentDigest || oldDigest == currentDigest {
			t.Fatal("#52 historical contract digests must represent three revisions")
		}
		oldReceipt, oldID := littleGoblinReceipt(
			identity,
			oldDigest,
			"echo historical receipt one",
			"",
			"historical parser failure one",
			"historical parser success one",
			false,
		)
		middleReceipt, _ := littleGoblinReceipt(
			identity,
			middleDigest,
			"echo historical receipt two",
			"",
			"historical parser failure two",
			"historical parser success two",
			false,
		)
		amendOne := littleGoblinAmendment(identity, oldDigest, middleDigest, "record the first synthetic contract interpretation change")
		amendTwo := littleGoblinAmendment(identity, middleDigest, currentDigest, "record the second synthetic contract interpretation change")
		request := formatRebuildRequest(identity)
		recovery, recoveryID := littleGoblinReceipt(
			identity,
			currentDigest,
			current.verify,
			oldID,
			"current pre-run after append-only recovery",
			"current post-run after append-only recovery",
			true,
		)
		oldHeadingReceipt := littleGoblinHeadingFormReceipt(identity, oldReceipt)
		middleHeadingReceipt := littleGoblinHeadingFormReceipt(identity, middleReceipt)

		t.Run("unamended digest transition fails", func(t *testing.T) {
			comments := []string{oldHeadingReceipt, middleHeadingReceipt, request, recovery}
			result := littleGoblinValidationResult(t, body, comments)
			if !littleGoblinErrorContains(result, "not bridged by an amendment") {
				t.Fatalf("#52 unamended digest transition was not rejected: %v", result.Errors)
			}
		})

		comments := []string{oldHeadingReceipt, middleHeadingReceipt, amendOne, amendTwo, request, recovery}
		result := littleGoblinValidationResult(t, body, comments)
		if len(result.Errors) > 0 {
			t.Fatalf("#52 comment-shaped history was rejected: %v", result.Errors)
		}
		unresolved, scanErrors := unresolvedRebuildRequests(units, comments)
		if len(scanErrors) > 0 || len(unresolved) > 0 {
			t.Fatalf("#52 rebuild routing remained unresolved: unresolved=%v errors=%v", unresolved, scanErrors)
		}
		if !strings.HasPrefix(oldID, legacyReceiptIDPrefix) || !strings.HasPrefix(recoveryID, receiptIDPrefix) {
			t.Fatalf("#52 receipts used unexpected IDs: legacy=%q recovery=%q", oldID, recoveryID)
		}
		if !strings.Contains(recovery, "Recovered from: "+oldID) {
			t.Fatal("#52 recovery did not retain provenance to the first historical receipt")
		}
		if comments[0] != oldHeadingReceipt || comments[1] != middleHeadingReceipt {
			t.Fatal("#52 heading-form comment history was not append-only")
		}

		t.Run("malformed heading history remains visible after recovery", func(t *testing.T) {
			malformed := strings.TrimSuffix(
				oldHeadingReceipt,
				"Post-evidence scope: this command exited 0 at "+littleGoblinPostSHA+"; nothing else is inferred.",
			)
			malformedComments := []string{oldHeadingReceipt, malformed, middleHeadingReceipt, amendOne, amendTwo, request, recovery}
			malformedResult := littleGoblinValidationResult(t, body, malformedComments)
			if !littleGoblinErrorContains(malformedResult, "comment 2: malformed heading-form evidence") {
				t.Fatalf("#52 malformed heading history was hidden by recovery: %v", malformedResult.Errors)
			}
		})

		root := t.TempDir()
		queuePath := filepath.Join(root, "specs", "queues", "little-goblin-52.md")
		if err := os.MkdirAll(filepath.Dir(queuePath), 0o755); err != nil {
			t.Fatalf("mkdir #52 queue: %v", err)
		}
		localHistory := strings.Join([]string{body, oldReceipt, middleReceipt, amendOne, amendTwo, recovery}, "\n\n")
		if err := os.WriteFile(queuePath, []byte(localHistory), 0o644); err != nil {
			t.Fatalf("write #52 queue: %v", err)
		}
		localResult, err := ValidateQueueFile(queuePath)
		if err != nil {
			t.Fatalf("ValidateQueueFile #52: %v", err)
		}
		if !localResult.Valid || len(localResult.Errors) > 0 {
			t.Fatalf("#52 local metadata history was rejected: %v", localResult.Errors)
		}
		stripped, metadata := splitLocalQueueMetadataBlocks(localHistory)
		if !strings.Contains(stripped, body) || !strings.Contains(localHistory, oldReceipt) || !strings.Contains(localHistory, recovery) {
			t.Fatal("#52 local recovery did not preserve the prior queue and receipts")
		}
		if len(metadata) != 5 || metadata[0] != oldReceipt || metadata[1] != middleReceipt || metadata[4] != recovery {
			t.Fatalf("#52 local metadata = %q, want the two historical receipts, two amendments, and recovery", metadata)
		}
	})

	t.Run("issue #53 historical chunked receipt recovers without editing prior evidence", func(t *testing.T) {
		sourceHeading := "Synthetic source archive"
		indexHeading := "Synthetic receipt index"
		targetHeading := "Historical chunk recovery"
		currentVerify := "echo current chunk recovery"
		body := littleGoblinQueueBody(
			littleGoblinUnitSpec{
				heading:  sourceHeading,
				done:     "the synthetic source is available",
				scenario: "synthetic source is available",
				verify:   "echo source",
			},
			littleGoblinUnitSpec{
				heading:  indexHeading,
				depends:  sourceHeading,
				done:     "the synthetic receipt index is assembled",
				scenario: "synthetic receipt index is assembled",
				verify:   "echo index",
			},
			littleGoblinUnitSpec{
				heading:  targetHeading,
				depends:  indexHeading,
				done:     "the historical chunked receipt is recoverable",
				scenario: "issue #53 historical chunked receipt recovers without editing prior evidence",
				verify:   currentVerify,
				checked:  true,
			},
		)
		units, unitIssues := ValidateQueueBody(body, "synthetic little-goblin #53")
		if len(units) != 3 {
			t.Fatalf("expected three #53-shaped units, got %d", len(units))
		}
		if len(units[2].Depends) != 1 || units[2].Depends[0] != indexHeading || len(units[1].Depends) != 1 || units[1].Depends[0] != sourceHeading {
			t.Fatalf("#53 dependency queue was not parsed, units=%+v", units)
		}
		if len(unitIssues) != 1 || !strings.Contains(unitIssues[0].Message, "missing Evidence receipt") {
			t.Fatalf("expected only the checked #53 unit evidence gap before comments, got %v", unitIssues)
		}

		identity := queueUnitIdentity{Occurrence: 1, Heading: targetHeading}
		currentDigest := unitContractDigest(units[2])
		historicalVerify := "echo historical chunk recovery"
		historical := littleGoblinHistoricalUnit(littleGoblinUnitSpec{
			heading:  targetHeading,
			depends:  indexHeading,
			done:     "the earlier chunked interpretation is retained",
			scenario: "historical chunked receipt is retained",
			verify:   historicalVerify,
		})
		oldDigest := unitContractDigest(historical)
		preOutput := strings.Repeat("synthetic historical pre output line\n", 1800)
		postOutput := strings.Repeat("synthetic historical post output line\n", 1500)
		if len(preOutput)+len(postOutput) <= 65536 {
			t.Fatal("#53 fixture must model an oversized logical receipt")
		}
		chunks, oldID := littleGoblinChunkedLegacyReceipt(identity, oldDigest, historicalVerify, preOutput, postOutput)
		amendment := littleGoblinAmendment(identity, oldDigest, currentDigest, "record the synthetic #53 contract interpretation change")
		recovery, recoveryID := littleGoblinReceipt(
			identity,
			currentDigest,
			currentVerify,
			oldID,
			"current chunk recovery pre-run",
			"current chunk recovery post-run",
			true,
		)
		comments := append(append([]string{}, chunks...), amendment, recovery)
		result := littleGoblinValidationResult(t, body, comments)
		if len(result.Errors) > 0 {
			t.Fatalf("#53 chunked history was rejected: %v", result.Errors)
		}
		if !strings.HasPrefix(oldID, legacyReceiptIDPrefix) || !strings.HasPrefix(recoveryID, receiptIDPrefix) {
			t.Fatalf("#53 receipts used unexpected IDs: legacy=%q recovery=%q", oldID, recoveryID)
		}
		if !strings.Contains(comments[0], "Raw output chunk:") || !strings.Contains(comments[1], receiptContinuationMarker) || !strings.Contains(comments[2], receiptContinuationMarker) {
			t.Fatal("#53 historical receipt was not represented as consecutive continuation chunks")
		}
		if !strings.Contains(comments[len(comments)-1], "Recovered from: "+oldID) {
			t.Fatal("#53 versioned recovery did not point to the legacy chunked receipt")
		}
		for i := range chunks {
			if !strings.Contains(comments[i], "Unit occurrence: 1") || !strings.Contains(comments[i], "Unit heading: "+targetHeading) {
				t.Fatalf("#53 chunk %d lost the historical receipt identity", i+1)
			}
		}

		mutated := append([]string{}, comments...)
		mutated[1] = strings.Replace(mutated[1], "Unit heading: "+targetHeading, "Unit heading: Wrong chunk identity", 1)
		mutatedResult := littleGoblinValidationResult(t, body, mutated)
		if len(mutatedResult.Errors) == 0 {
			t.Fatal("#53 changed continuation identity was accepted")
		}
	})

	t.Run("little-goblin-shaped regression pins execute as named validation scenarios", func(t *testing.T) {
		archiveHeading := "Pin archive"
		indexHeading := "Pin index"
		routingHeading := "Pin routing"
		chunkHeading := "Pin chunk target"
		routingVerify := "echo pin routing current"
		chunkVerify := "echo pin chunk target current"
		body := littleGoblinQueueBody(
			littleGoblinUnitSpec{
				heading:  archiveHeading,
				done:     "the pin archive is available",
				scenario: "pin archive is available",
				verify:   "echo pin archive",
			},
			littleGoblinUnitSpec{
				heading:  indexHeading,
				depends:  archiveHeading,
				done:     "the pin index is assembled",
				scenario: "pin index is assembled",
				verify:   "echo pin index",
			},
			littleGoblinUnitSpec{
				heading:  routingHeading,
				depends:  indexHeading,
				done:     "historical pin routing receipts stay recoverable",
				scenario: "little-goblin-shaped regression pins execute as named validation scenarios",
				verify:   routingVerify,
				checked:  true,
			},
			littleGoblinUnitSpec{
				heading:  chunkHeading,
				depends:  indexHeading,
				done:     "the oversized historical pin receipt stays recoverable",
				scenario: "the oversized historical pin receipt stays recoverable",
				verify:   chunkVerify,
				checked:  true,
			},
		)
		units, unitIssues := ValidateQueueBody(body, "synthetic little-goblin regression pins")
		if len(units) != 4 {
			t.Fatalf("expected four pinned units, got %d", len(units))
		}
		if len(units[1].Depends) != 1 || units[1].Depends[0] != archiveHeading || len(units[2].Depends) != 1 || units[2].Depends[0] != indexHeading || len(units[3].Depends) != 1 || units[3].Depends[0] != indexHeading {
			t.Fatalf("named regression queue lost its dependency shape: %+v", units)
		}
		if len(unitIssues) != 2 {
			t.Fatalf("expected only the two checked-unit evidence gaps before comments, got %v", unitIssues)
		}
		for _, issue := range unitIssues {
			if !strings.Contains(issue.Message, "missing Evidence receipt") {
				t.Fatalf("unexpected pre-comment issue: %v", unitIssues)
			}
		}

		routingIdentity := queueUnitIdentity{Occurrence: 1, Heading: routingHeading}
		chunkIdentity := queueUnitIdentity{Occurrence: 1, Heading: chunkHeading}
		routingDigest := unitContractDigest(units[2])
		chunkDigest := unitContractDigest(units[3])

		routingHistoricalOne := littleGoblinHistoricalUnit(littleGoblinUnitSpec{
			heading:  routingHeading,
			depends:  indexHeading,
			done:     "the first routing interpretation is retained",
			scenario: "routing interpretation one",
			verify:   "echo pin routing one",
		})
		routingHistoricalTwo := littleGoblinHistoricalUnit(littleGoblinUnitSpec{
			heading:  routingHeading,
			depends:  indexHeading,
			done:     "the second routing interpretation is retained",
			scenario: "routing interpretation two",
			verify:   "echo pin routing two",
		})
		routingOldDigest := unitContractDigest(routingHistoricalOne)
		routingMiddleDigest := unitContractDigest(routingHistoricalTwo)
		routingOldReceipt, routingOldID := littleGoblinReceipt(
			routingIdentity,
			routingOldDigest,
			"echo pin routing one",
			"",
			"routing interpretation one failure",
			"routing interpretation one success",
			false,
		)
		routingMiddleReceipt, _ := littleGoblinReceipt(
			routingIdentity,
			routingMiddleDigest,
			"echo pin routing two",
			"",
			"routing interpretation two failure",
			"routing interpretation two success",
			false,
		)
		routingAmendOne := littleGoblinAmendment(routingIdentity, routingOldDigest, routingMiddleDigest, "record the first pinned routing interpretation change")
		routingAmendTwo := littleGoblinAmendment(routingIdentity, routingMiddleDigest, routingDigest, "record the second pinned routing interpretation change")
		routingRequest := formatRebuildRequest(routingIdentity)
		routingRecovery, routingRecoveryID := littleGoblinReceipt(
			routingIdentity,
			routingDigest,
			routingVerify,
			routingOldID,
			"routing recovery failure",
			"routing recovery success",
			true,
		)

		chunkHistorical := littleGoblinHistoricalUnit(littleGoblinUnitSpec{
			heading:  chunkHeading,
			depends:  indexHeading,
			done:     "the chunked interpretation is retained",
			scenario: "chunk interpretation one",
			verify:   "echo pin chunk historical",
		})
		chunkOldDigest := unitContractDigest(chunkHistorical)
		preOutput := strings.Repeat("pinned pre output line\n", 1800)
		postOutput := strings.Repeat("pinned post output line\n", 1500)
		if len(preOutput)+len(postOutput) <= 65536 {
			t.Fatal("pinned #53 fixture must model an oversized logical receipt")
		}
		chunks, chunkOldID := littleGoblinChunkedLegacyReceipt(chunkIdentity, chunkOldDigest, "echo pin chunk historical", preOutput, postOutput)
		chunkAmendment := littleGoblinAmendment(chunkIdentity, chunkOldDigest, chunkDigest, "record the pinned chunk interpretation change")
		chunkRecovery, chunkRecoveryID := littleGoblinReceipt(
			chunkIdentity,
			chunkDigest,
			chunkVerify,
			chunkOldID,
			"chunk recovery failure",
			"chunk recovery success",
			true,
		)

		if !strings.HasPrefix(routingOldID, legacyReceiptIDPrefix) || !strings.HasPrefix(chunkOldID, legacyReceiptIDPrefix) || !strings.HasPrefix(routingRecoveryID, receiptIDPrefix) || !strings.HasPrefix(chunkRecoveryID, receiptIDPrefix) {
			t.Fatalf("pinned receipts used unexpected IDs: routing legacy=%q chunk legacy=%q routing recovery=%q chunk recovery=%q", routingOldID, chunkOldID, routingRecoveryID, chunkRecoveryID)
		}

		comments := []string{
			routingOldReceipt,
			routingMiddleReceipt,
			routingAmendOne,
			routingAmendTwo,
			routingRequest,
			chunks[0],
			chunks[1],
			chunks[2],
			chunks[3],
			chunkAmendment,
			routingRecovery,
			chunkRecovery,
		}
		result := littleGoblinValidationResult(t, body, comments)
		if len(result.Errors) > 0 {
			t.Fatalf("pinned GH comment history was rejected: %v", result.Errors)
		}
		unresolved, scanErrors := unresolvedRebuildRequests(units, comments)
		if len(scanErrors) > 0 || len(unresolved) > 0 {
			t.Fatalf("pinned rebuild routing remained unresolved: unresolved=%v errors=%v", unresolved, scanErrors)
		}

		localHistory := strings.Join(append([]string{body}, comments...), "\n\n")
		root := t.TempDir()
		queuePath := filepath.Join(root, "specs", "queues", "little-goblin-pins.md")
		if err := os.MkdirAll(filepath.Dir(queuePath), 0o755); err != nil {
			t.Fatalf("mkdir pinned queue: %v", err)
		}
		if err := os.WriteFile(queuePath, []byte(localHistory), 0o644); err != nil {
			t.Fatalf("write pinned queue: %v", err)
		}
		localResult, err := ValidateQueueFile(queuePath)
		if err != nil {
			t.Fatalf("ValidateQueueFile pinned queue: %v", err)
		}
		if !localResult.Valid || len(localResult.Errors) > 0 {
			t.Fatalf("pinned local queue history was rejected: %v", localResult.Errors)
		}
		stripped, metadata := splitLocalQueueMetadataBlocks(localHistory)
		if strings.Contains(stripped, "Raw output chunk:") || strings.Contains(stripped, "unit digest: ") {
			t.Fatal("pinned local queue leaked receipt records into the contract body")
		}
		if len(metadata) != len(comments) {
			t.Fatalf("pinned local metadata = %d blocks, want the %d appended records", len(metadata), len(comments))
		}
		for i := range comments {
			if metadata[i] != comments[i] {
				t.Fatalf("pinned local metadata block %d did not mirror its appended record", i+1)
			}
		}

		t.Run("unamended pinned digest transition fails", func(t *testing.T) {
			unamended := []string{routingOldReceipt, routingMiddleReceipt, routingRequest, routingRecovery, chunkAmendment, chunkRecovery}
			unamendedResult := littleGoblinValidationResult(t, body, unamended)
			if !littleGoblinErrorContains(unamendedResult, "not bridged by an amendment") {
				t.Fatalf("pinned unamended digest transition was not rejected: %v", unamendedResult.Errors)
			}
		})

		t.Run("dangling pinned recovery reference fails", func(t *testing.T) {
			dangling, _ := littleGoblinReceipt(
				chunkIdentity,
				chunkDigest,
				chunkVerify,
				legacyReceiptIDPrefix+strings.Repeat("0", 64),
				"chunk recovery failure",
				"chunk recovery success",
				true,
			)
			danglingComments := append(append([]string{}, comments[:len(comments)-1]...), dangling)
			danglingResult := littleGoblinValidationResult(t, body, danglingComments)
			if !littleGoblinErrorContains(danglingResult, "does not identify an earlier complete receipt") {
				t.Fatalf("pinned dangling recovery reference was not rejected: %v", danglingResult.Errors)
			}
		})

		t.Run("unauthenticated supersession text does not repair broken pins", func(t *testing.T) {
			supersession := strings.Join([]string{
				"Supersedes: receipt-sha256-v1:" + strings.Repeat("a", 64),
				"Quarantined: yes",
				"Authorized by: synthetic actor",
			}, "\n")
			repaired := append(append([]string{}, comments...), supersession)
			unbridged := make([]string, 0, len(repaired))
			for _, comment := range repaired {
				if comment != routingAmendOne && comment != routingAmendTwo {
					unbridged = append(unbridged, comment)
				}
			}
			repairedResult := littleGoblinValidationResult(t, body, unbridged)
			if !littleGoblinErrorContains(repairedResult, "not bridged by an amendment") {
				t.Fatalf("supersession text changed pinned validation state: %v", repairedResult.Errors)
			}
		})

		t.Run("local chunk identity drift fails", func(t *testing.T) {
			driftedChunks := append([]string{}, chunks...)
			driftedChunks[1] = strings.Replace(driftedChunks[1], "Unit heading: "+chunkHeading, "Unit heading: Drifted chunk identity", 1)
			driftedComments := append(append([]string{}, comments[:5]...), driftedChunks...)
			driftedComments = append(driftedComments, chunkAmendment, routingRecovery, chunkRecovery)
			driftedHistory := strings.Join(append([]string{body}, driftedComments...), "\n\n")
			driftedPath := filepath.Join(root, "specs", "queues", "little-goblin-pins-drift.md")
			if err := os.WriteFile(driftedPath, []byte(driftedHistory), 0o644); err != nil {
				t.Fatalf("write drifted queue: %v", err)
			}
			driftedResult, err := ValidateQueueFile(driftedPath)
			if err != nil {
				t.Fatalf("ValidateQueueFile drifted queue: %v", err)
			}
			if driftedResult.Valid || len(driftedResult.Errors) == 0 {
				t.Fatal("pinned local chunk identity drift was accepted")
			}
		})

		t.Run("interrupted local chunk chain fails", func(t *testing.T) {
			coverage := strings.Join([]string{
				"Review coverage:",
				"HEAD: " + littleGoblinPostSHA,
				"Unit occurrence: 1",
				"Unit heading: " + chunkHeading,
				"Exercised:",
				"interrupting coverage record",
			}, "\n")
			interruptedComments := append(append([]string{}, comments[:6]...), coverage)
			interruptedComments = append(interruptedComments, comments[6:]...)
			interruptedHistory := strings.Join(append([]string{body}, interruptedComments...), "\n\n")
			interruptedPath := filepath.Join(root, "specs", "queues", "little-goblin-pins-interrupted.md")
			if err := os.WriteFile(interruptedPath, []byte(interruptedHistory), 0o644); err != nil {
				t.Fatalf("write interrupted queue: %v", err)
			}
			interruptedResult, err := ValidateQueueFile(interruptedPath)
			if err != nil {
				t.Fatalf("ValidateQueueFile interrupted queue: %v", err)
			}
			if interruptedResult.Valid || len(interruptedResult.Errors) == 0 {
				t.Fatal("pinned interrupted local chunk chain was accepted")
			}
		})
	})
}
