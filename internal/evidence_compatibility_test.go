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
		t.Run("observed early heading-form history and witnessed digest transitions", func(t *testing.T) {
			t.Fatal("heading-form history fixture not implemented")
		})
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
		comments := []string{oldReceipt, middleReceipt, amendOne, amendTwo, request, recovery}
		result := &ValidationResult{Valid: true}
		applyQueueIssues(result, "synthetic #52 comments", units, unitIssues, comments)
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
		if !strings.Contains(comments[0], oldReceipt) || !strings.Contains(comments[1], middleReceipt) {
			t.Fatal("#52 comment history was not append-only")
		}

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
		body := littleGoblinQueueBody(
			littleGoblinUnitSpec{
				heading:  "Pin source",
				done:     "the source pin is structurally valid",
				scenario: "source pin is structurally valid",
				verify:   "echo source pin",
			},
			littleGoblinUnitSpec{
				heading:  "Pin target",
				depends:  "Pin source",
				done:     "the target pin is structurally valid",
				scenario: "little-goblin-shaped regression pins execute as named validation scenarios",
				verify:   "echo target pin",
			},
		)
		units, issues := ValidateQueueBody(body, "synthetic little-goblin regression pins")
		if len(issues) > 0 {
			t.Fatalf("named regression queue was rejected: %v", issues)
		}
		if len(units) != 2 || len(units[1].Depends) != 1 || units[1].Depends[0] != "Pin source" {
			t.Fatalf("named regression queue lost its dependency shape: %+v", units)
		}
	})
}
