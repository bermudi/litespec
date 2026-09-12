package internal

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
)

const assemblyTestVerify = "echo hi"

func assemblyTestRequest() ReceiptAssemblyRequest {
	return ReceiptAssemblyRequest{
		Occurrence: 1,
		Heading:    "My outcome",
		Verify:     assemblyTestVerify,
		UnitDigest: fixtureUnitDigest(assemblyTestVerify),
		Pre: ReceiptRunEvidence{
			SHA:    strings.Repeat("1", 40),
			Status: 1,
			Output: "missing outcome\n",
		},
		Post: ReceiptRunEvidence{
			SHA:    strings.Repeat("2", 40),
			Status: 0,
			Output: "outcome present\n",
		},
	}
}

func assemblyEmbeddedReceiptID(comments []string) string {
	match := regexp.MustCompile(`receipt-sha256-v1:[0-9a-f]{64}`).FindString(strings.Join(comments, "\n"))
	return match
}

func assemblySelfParse(t *testing.T, comments []string) parsedEvidenceReceipt {
	t.Helper()
	merged := mergeContinuedCommentRecords(comments)
	if len(merged) == 0 {
		t.Fatal("assembly produced no comments")
	}
	document := newEvidenceDocumentFromComment(merged[0])
	identity := queueUnitIdentity{Occurrence: 1, Heading: "My outcome"}
	receipt, issues := parseEvidenceReceiptDocument(
		evidencePayloadDocument(document),
		assemblyTestVerify,
		fixtureUnitDigest(assemblyTestVerify),
		"test",
		"My outcome",
		&identity,
	)
	if len(issues) > 0 {
		t.Fatalf("assembled receipt failed the validator grammar: %v", issues)
	}
	return receipt
}

func assemblyAssertCommentSizes(t *testing.T, comments []string, limit int) {
	t.Helper()
	for i, comment := range comments {
		if len(comment) > limit {
			t.Fatalf("comment %d is %d characters, over the %d limit", i+1, len(comment), limit)
		}
	}
}

func assemblyAssertContinuationChain(t *testing.T, comments []string) {
	t.Helper()
	for i, comment := range comments {
		lastLine := ""
		for _, line := range strings.Split(comment, "\n") {
			if strings.TrimSpace(line) != "" {
				lastLine = strings.TrimSpace(line)
			}
		}
		final := i == len(comments)-1
		if final && lastLine == receiptContinuationMarker {
			t.Fatalf("final comment %d ends with the continuation marker", i+1)
		}
		if !final && lastLine != receiptContinuationMarker {
			t.Fatalf("non-final comment %d does not end with the exact continuation marker", i+1)
		}
	}
}

func assemblyAssertFenceDelimiters(t *testing.T, comments []string) {
	t.Helper()
	for i, comment := range comments {
		openFence := ""
		var payloadLines []string
		for _, line := range strings.Split(comment, "\n") {
			delimiter := fenceDelimiter(strings.TrimSpace(line))
			if openFence == "" {
				if delimiter == "" {
					continue
				}
				openFence = delimiter
				payloadLines = nil
				continue
			}
			if strings.TrimSpace(line) == openFence {
				for _, payloadLine := range payloadLines {
					if run := len(fenceDelimiter(payloadLine)); run >= len(openFence) {
						t.Fatalf(
							"comment %d uses fence delimiter %q which does not exceed payload delimiter line %q",
							i+1, openFence, payloadLine,
						)
					}
				}
				openFence = ""
				continue
			}
			payloadLines = append(payloadLines, line)
		}
	}
}

func TestReceiptAssemblyEngine(t *testing.T) {
	t.Run("assembled receipt ID matches the validator derivation", func(t *testing.T) {
		comments, err := AssembleEvidenceReceiptComments(assemblyTestRequest())
		if err != nil {
			t.Fatalf("assembly refused valid run evidence: %v", err)
		}
		if len(comments) != 1 {
			t.Fatalf("expected one comment for small outputs, got %d", len(comments))
		}
		receipt := assemblySelfParse(t, comments)
		derived := receiptIDForCanonicalReceipt(receipt)
		embedded := assemblyEmbeddedReceiptID(comments)
		if embedded == "" {
			t.Fatal("assembled comment carries no Receipt ID")
		}
		if embedded != derived {
			t.Fatalf("Receipt ID mismatch: engine emitted %s, validator derives %s", embedded, derived)
		}
		if !strings.Contains(comments[0], assemblyTestVerify) {
			t.Fatal("assembled receipt must quote the exact Verify command")
		}
		if !strings.Contains(comments[0], "unit digest: "+fixtureUnitDigest(assemblyTestVerify)) {
			t.Fatal("assembled receipt must quote the current unit digest")
		}

		body := ownedQueue(checkedUnit(assemblyTestVerify, ""))
		units, unitIssues := ValidateQueueBody(body, "queue")
		if len(units) != 1 {
			t.Fatalf("expected one unit, got %d", len(units))
		}
		result := &ValidationResult{Valid: true}
		applyQueueIssues(result, "comments", units, unitIssues, comments)
		if containsIssue(result.Errors, "Evidence receipt") {
			t.Fatalf("validator did not accept the assembled receipt as unit evidence: %v", result.Errors)
		}

		mutated := assemblyTestRequest()
		mutated.Post.Output = "different outcome\n"
		mutatedComments, err := AssembleEvidenceReceiptComments(mutated)
		if err != nil {
			t.Fatalf("assembly refused mutated run evidence: %v", err)
		}
		if assemblyEmbeddedReceiptID(mutatedComments) == embedded {
			t.Fatal("Receipt ID did not change when post output changed")
		}
	})

	t.Run("oversized outputs split at legal boundaries with markers and chunk identity", func(t *testing.T) {
		t.Run("field boundary split", func(t *testing.T) {
			request := assemblyTestRequest()
			request.Pre.Output = strings.Repeat("pre line\n", 30)
			request.Post.Output = strings.Repeat("post line\n", 30)
			const limit = 820
			request.CommentLimit = limit
			comments, err := AssembleEvidenceReceiptComments(request)
			if err != nil {
				t.Fatalf("assembly refused splittable run evidence: %v", err)
			}
			if len(comments) < 2 {
				t.Fatalf("expected a split across comments, got %d", len(comments))
			}
			assemblyAssertCommentSizes(t, comments, limit)
			assemblyAssertContinuationChain(t, comments)
			assemblyAssertFenceDelimiters(t, comments)
			receipt := assemblySelfParse(t, comments)
			if receipt.preOutput != request.Pre.Output {
				t.Fatal("pre output was not reconstructed verbatim across the split")
			}
			if receipt.postOutput != request.Post.Output {
				t.Fatal("post output was not reconstructed verbatim across the split")
			}
			if receiptIDForCanonicalReceipt(receipt) != assemblyEmbeddedReceiptID(comments) {
				t.Fatal("split receipt changed the canonical Receipt ID")
			}
			if !strings.Contains(comments[1], "Protocol: evidence/v1") ||
				!strings.Contains(comments[1], "Unit occurrence: 1") ||
				!strings.Contains(comments[1], "Unit heading: My outcome") {
				t.Fatal("continuation comment does not repeat the receipt identity")
			}
		})

		t.Run("explicit chunk form", func(t *testing.T) {
			request := assemblyTestRequest()
			request.Pre.Output = strings.Repeat("0123456789", 200) + "\n"
			request.Post.Output = strings.Repeat("abcdef\n", 120)
			const limit = 800
			request.CommentLimit = limit
			comments, err := AssembleEvidenceReceiptComments(request)
			if err != nil {
				t.Fatalf("assembly refused chunkable run evidence: %v", err)
			}
			if len(comments) < 3 {
				t.Fatalf("expected chunked comments, got %d", len(comments))
			}
			assemblyAssertCommentSizes(t, comments, limit)
			assemblyAssertContinuationChain(t, comments)
			assemblyAssertFenceDelimiters(t, comments)

			joined := strings.Join(comments, "\n")
			type chunkRecord struct {
				phase  string
				number int
				total  int
			}
			var records []chunkRecord
			lastPhase := ""
			for _, line := range strings.Split(joined, "\n") {
				trimmed := strings.TrimSpace(line)
				if phase, ok := strings.CutPrefix(trimmed, "Output: "); ok {
					lastPhase = phase
					continue
				}
				number, total, ok := strings.Cut(trimmed, "/")
				if !ok || !strings.HasPrefix(trimmed, "Chunk: ") {
					continue
				}
				chunkNumber, err := strconv.Atoi(strings.TrimPrefix(number, "Chunk: "))
				if err != nil {
					t.Fatalf("chunk record has a non-numeric number: %s", trimmed)
				}
				chunkTotal, err := strconv.Atoi(total)
				if err != nil {
					t.Fatalf("chunk record has a non-numeric total: %s", trimmed)
				}
				records = append(records, chunkRecord{phase: lastPhase, number: chunkNumber, total: chunkTotal})
			}
			if len(records) < 2 {
				t.Fatalf("expected explicit chunk records, got %d", len(records))
			}
			seenPhase := map[string]int{}
			for _, record := range records {
				seenPhase[record.phase]++
				if record.number != seenPhase[record.phase] {
					t.Fatalf("chunk numbers are not consecutive from 1 for %s: got %d", record.phase, record.number)
				}
			}
			for phase, count := range seenPhase {
				for _, record := range records {
					if record.phase == phase && record.total != count {
						t.Fatalf("%s chunks declare total %d but %d exist", phase, record.total, count)
					}
				}
			}
			for _, comment := range comments {
				count := strings.Count(comment, "Raw output chunk:")
				if count > 1 {
					t.Fatal("chunk records must not share one comment")
				}
			}
			if strings.Count(joined, "Output: pre") < 2 {
				t.Fatal("pre output was not chunked with repeated identity")
			}

			receipt := assemblySelfParse(t, comments)
			if receipt.preOutput != request.Pre.Output {
				t.Fatal("chunked pre output was not reconstructed byte-for-byte")
			}
			if receipt.postOutput != request.Post.Output {
				t.Fatal("chunked post output was not reconstructed byte-for-byte")
			}
			if receiptIDForCanonicalReceipt(receipt) != assemblyEmbeddedReceiptID(comments) {
				t.Fatal("chunked receipt changed the canonical Receipt ID")
			}
		})
	})

	t.Run("invalid run evidence refused without partial output", func(t *testing.T) {
		cases := []struct {
			name    string
			mutate  func(*ReceiptAssemblyRequest)
			message string
		}{
			{
				name: "equal pre and post shas",
				mutate: func(r *ReceiptAssemblyRequest) {
					r.Post.SHA = r.Pre.SHA
				},
				message: "pre and post",
			},
			{
				name: "zero pre exit status",
				mutate: func(r *ReceiptAssemblyRequest) {
					r.Pre.Status = 0
				},
				message: "pre exit status",
			},
			{
				name: "nonzero post exit status",
				mutate: func(r *ReceiptAssemblyRequest) {
					r.Post.Status = 3
				},
				message: "post exit status",
			},
			{
				name: "malformed recovered-from receipt id",
				mutate: func(r *ReceiptAssemblyRequest) {
					r.RecoveredFrom = "receipt-sha256-v1:nothex"
				},
				message: "Recovered from",
			},
		}
		for _, testCase := range cases {
			request := assemblyTestRequest()
			testCase.mutate(&request)
			comments, err := AssembleEvidenceReceiptComments(request)
			if err == nil {
				t.Fatalf("%s: expected a visible refusal", testCase.name)
			}
			if !strings.Contains(err.Error(), testCase.message) {
				t.Fatalf("%s: refusal does not name the problem: %v", testCase.name, err)
			}
			if comments != nil {
				t.Fatalf("%s: refusal produced partial output: %d comments", testCase.name, len(comments))
			}
		}
	})

	t.Run("rebuild identity and recovery provenance assemble canonically", func(t *testing.T) {
		request := assemblyTestRequest()
		request.Rebuild = true
		request.RecoveredFrom = "receipt-sha256-v1:" + strings.Repeat("cd", 32)
		comments, err := AssembleEvidenceReceiptComments(request)
		if err != nil {
			t.Fatalf("assembly refused rebuild evidence: %v", err)
		}
		head := comments[0]
		occurrenceIndex := strings.Index(head, "Unit occurrence: 1")
		headingIndex := strings.Index(head, "Unit heading: My outcome")
		evidenceIndex := strings.Index(head, "Evidence:")
		if occurrenceIndex < 0 || headingIndex < 0 || evidenceIndex < 0 ||
			occurrenceIndex > evidenceIndex || headingIndex > evidenceIndex {
			t.Fatal("rebuild receipt must open with its routing identity before Evidence:")
		}
		if !strings.Contains(head, "Recovered from: "+request.RecoveredFrom) {
			t.Fatal("recovery provenance is missing from the assembled receipt")
		}

		receipt := assemblySelfParse(t, comments)
		if receiptIDForCanonicalReceipt(receipt) != assemblyEmbeddedReceiptID(comments) {
			t.Fatal("rebuild receipt ID does not match the validator derivation")
		}

		other := request
		other.RecoveredFrom = "receipt-sha256-v1:" + strings.Repeat("ef", 32)
		otherComments, err := AssembleEvidenceReceiptComments(other)
		if err != nil {
			t.Fatalf("assembly refused alternate recovery provenance: %v", err)
		}
		if assemblyEmbeddedReceiptID(otherComments) == assemblyEmbeddedReceiptID(comments) {
			t.Fatal("Recovered from does not participate in the canonical Receipt ID")
		}

		body := ownedQueue(checkedUnit(assemblyTestVerify, ""))
		units, _ := ValidateQueueBody(body, "queue")
		identity, kind, _, err := parseRebuildComment(comments[0], units)
		if err != nil {
			t.Fatalf("assembled rebuild receipt does not parse: %v", err)
		}
		if kind != rebuildCommentEvidence {
			t.Fatalf("assembled rebuild receipt is not evidence, got kind %d", kind)
		}
		if identity != (queueUnitIdentity{Occurrence: 1, Heading: "My outcome"}) {
			t.Fatalf("assembled rebuild receipt resolves to the wrong identity: %+v", identity)
		}
	})
}
