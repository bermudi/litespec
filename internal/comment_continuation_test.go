package internal

import (
	"fmt"
	"slices"
	"strings"
	"testing"
)

const continuationPreSHA = "1111111111111111111111111111111111111111"
const continuationPostSHA = "2222222222222222222222222222222222222222"

func continuedReceiptHead(verifyCmd, digest string) string {
	return strings.Join([]string{
		"Unit occurrence: 1",
		"Unit heading: My outcome",
		"Evidence:",
		verifyCmd,
		"unit digest: " + digest,
		"pre sha: " + continuationPreSHA,
		"pre exit status: 1",
		"```",
		verifyCmd + " missing outcome",
		"```",
		"Pre-evidence scope: this command exited 1 at " + continuationPreSHA + "; nothing else is inferred.",
		receiptContinuationMarker,
	}, "\n")
}

func continuedReceiptTail(verifyCmd string) string {
	return strings.Join([]string{
		"post sha: " + continuationPostSHA,
		"post exit status: 0",
		"```",
		verifyCmd + " output",
		"```",
		"Post-evidence scope: this command exited 0 at " + continuationPostSHA + "; nothing else is inferred.",
	}, "\n")
}

func rawOutputChunk(phase string, number, total int, digest, payload string, marker bool) string {
	return rawOutputChunkFor(phase, number, total, 1, "My outcome", digest, payload, marker)
}

func rawOutputChunkFor(phase string, number, total, occurrence int, heading, digest, payload string, marker bool) string {
	lines := []string{
		"Raw output chunk:",
		"Output: " + phase,
		fmt.Sprintf("Chunk: %d/%d", number, total),
		fmt.Sprintf("Unit occurrence: %d", occurrence),
		"Unit heading: " + heading,
		"unit digest: " + digest,
		"```",
		payload,
		"```",
	}
	if marker {
		lines = append(lines, receiptContinuationMarker)
	}
	return strings.Join(lines, "\n")
}

func chunkedReceiptComments(verifyCmd, digest string) []string {
	first := strings.Join([]string{
		"Unit occurrence: 1",
		"Unit heading: My outcome",
		"Evidence:",
		verifyCmd,
		"unit digest: " + digest,
		"pre sha: " + continuationPreSHA,
		"pre exit status: 1",
		rawOutputChunk("pre", 1, 2, digest, "pre one\n", true),
	}, "\n")
	second := rawOutputChunk("pre", 2, 2, digest, "pre two", true)
	third := strings.Join([]string{
		"Pre-evidence scope: this command exited 1 at " + continuationPreSHA + "; nothing else is inferred.",
		"post sha: " + continuationPostSHA,
		"post exit status: 0",
		rawOutputChunk("post", 1, 2, digest, "post one\n", true),
	}, "\n")
	fourth := strings.Join([]string{
		rawOutputChunk("post", 2, 2, digest, "post two", false),
		"Post-evidence scope: this command exited 0 at " + continuationPostSHA + "; nothing else is inferred.",
	}, "\n")
	return []string{first, second, third, fourth}
}

func headingFormReceipt(verifyCmd, digest string) string {
	receipt := strings.Replace(evidenceReceipt(verifyCmd), fixtureUnitDigest(verifyCmd), digest, 1)
	return "## My outcome\n" + strings.TrimPrefix(receipt, "Evidence:\n")
}

func headingFormChunkedReceiptComments(verifyCmd, digest string) []string {
	comments := chunkedReceiptComments(verifyCmd, digest)
	lines := strings.SplitN(comments[0], "\n", 3)
	comments[0] = "## My outcome\n" + lines[2]
	return comments
}

func TestMergeContinuedCommentsJoinsMarkerWithNextComment(t *testing.T) {
	head := continuedReceiptHead("echo hi", "aaaa")
	tail := continuedReceiptTail("echo hi")
	merged := mergeContinuedComments([]string{head, tail, "unrelated"})

	want := strings.TrimSuffix(head, "\n"+receiptContinuationMarker) + "\n" + tail
	if merged[0] != want {
		t.Errorf("merged head = %q, want %q", merged[0], want)
	}
	if merged[1] != "" {
		t.Errorf("consumed continuation = %q, want empty", merged[1])
	}
	if merged[2] != "unrelated" {
		t.Errorf("unrelated comment = %q, want untouched", merged[2])
	}
}

func TestMergeContinuedCommentsChainsAcrossComments(t *testing.T) {
	head := "## My outcome\nEvidence:\n" + receiptContinuationMarker
	middle := "echo hi\n" + receiptContinuationMarker
	tail := "unit digest: done"
	merged := mergeContinuedComments([]string{head, middle, tail})

	want := "## My outcome\nEvidence:\necho hi\nunit digest: done"
	if merged[0] != want {
		t.Errorf("merged head = %q, want %q", merged[0], want)
	}
	if merged[1] != "" || merged[2] != "" {
		t.Errorf("consumed slots = %q, %q, want empty", merged[1], merged[2])
	}
}

func TestMergeContinuedCommentsIgnoresMarkerInsideFence(t *testing.T) {
	comment := "## My outcome\nEvidence:\n```\nraw output\n" + receiptContinuationMarker + "\n```\n"
	merged := mergeContinuedComments([]string{comment, "next"})

	if merged[0] != comment {
		t.Errorf("fenced marker must stay raw content, got %q", merged[0])
	}
	if merged[1] != "next" {
		t.Errorf("next comment must be untouched, got %q", merged[1])
	}
}

func TestMergeContinuedCommentsLeavesDanglingMarkerIntact(t *testing.T) {
	head := continuedReceiptHead("echo hi", "aaaa")
	merged := mergeContinuedComments([]string{head})

	if !slices.Equal(merged, []string{head}) {
		t.Errorf("dangling marker must be left intact, got %v", merged)
	}
}

func TestMergeContinuedCommentsDoesNotSkipBlankInterveningComment(t *testing.T) {
	head := continuedReceiptHead("echo hi", "aaaa")
	tail := continuedReceiptTail("echo hi")
	merged := mergeContinuedComments([]string{head, "", tail})

	if strings.Contains(merged[0], tail) || !strings.Contains(merged[0], receiptContinuationMarker) {
		t.Fatalf("blank comment must interrupt continuation, got %q", merged[0])
	}
	if merged[2] != tail {
		t.Fatalf("tail after blank comment must remain separate, got %q", merged[2])
	}
}

func TestBlankInterveningCommentFailsReceiptValidation(t *testing.T) {
	source := "GH issue #1"
	units, issues := ValidateQueueBody(ownedQueue(checkedUnit("echo hi", "")), source)
	digest := unitContractDigest(units[0])
	comments := []string{continuedReceiptHead("echo hi", digest), "", continuedReceiptTail("echo hi")}
	result := &ValidationResult{Valid: true}
	applyQueueIssues(result, "GitHub comments", units, issues, comments)
	if len(result.Errors) == 0 {
		t.Fatal("expected blank intervening comment to fail validation")
	}
}

func TestMergeContinuedCommentsIsIdempotent(t *testing.T) {
	comments := []string{
		continuedReceiptHead("echo hi", "aaaa"),
		continuedReceiptTail("echo hi"),
		"unrelated",
	}
	once := mergeContinuedComments(comments)
	twice := mergeContinuedComments(once)
	if !slices.Equal(once, twice) {
		t.Errorf("merge is not idempotent: %v vs %v", once, twice)
	}
}

func TestSplitReceiptSatisfiesCheckedUnit(t *testing.T) {
	source := "GH issue #1"
	units, issues := ValidateQueueBody(ownedQueue(checkedUnit("echo hi", "")), source)
	digest := unitContractDigest(units[0])
	comments := []string{
		continuedReceiptHead("echo hi", digest),
		continuedReceiptTail("echo hi"),
	}
	result := &ValidationResult{Valid: true}
	applyQueueIssues(result, "GitHub comments", units, issues, comments)
	if len(result.Errors) > 0 {
		t.Fatalf("expected split receipt to satisfy checked unit, got %v", result.Errors)
	}
}

func TestChunkedRawOutputReceiptReconstructsAndSatisfiesCheckedUnit(t *testing.T) {
	source := "GH issue #1"
	units, issues := ValidateQueueBody(ownedQueue(checkedUnit("echo hi", "")), source)
	digest := unitContractDigest(units[0])
	comments := chunkedReceiptComments("echo hi", digest)
	merged := mergeContinuedComments(comments)
	evidence := evidencePayload(merged[0], "echo hi")
	identity := queueUnitIdentity{Occurrence: 1, Heading: "My outcome"}
	cursor := newEvidenceCursor(evidence)
	if !cursor.consumeVerifyCommand("echo hi") {
		t.Fatal("expected Verify command")
	}
	cursor.skipBlanks()
	if _, ok := cursor.consumeField("unit digest"); !ok {
		t.Fatal("expected unit digest")
	}
	cursor.skipBlanks()
	if _, ok := cursor.consumeField("pre sha"); !ok {
		t.Fatal("expected pre sha")
	}
	cursor.skipBlanks()
	if _, ok := cursor.consumeField("pre exit status"); !ok {
		t.Fatal("expected pre exit status")
	}
	cursor.skipBlanks()
	pre, ok, reason := cursor.consumeRawOutput("pre", digest, identity.Heading, &identity)
	if !ok || pre != "pre one\npre two" {
		t.Fatalf("pre raw output = %q, ok=%t, reason=%s", pre, ok, reason)
	}

	result := &ValidationResult{Valid: true}
	applyQueueIssues(result, "GitHub comments", units, issues, comments)
	if len(result.Errors) > 0 {
		t.Fatalf("expected chunked receipt to satisfy checked unit, got %v", result.Errors)
	}
}

func multilineVerifyWithMisleadingDigest() string {
	return strings.Join([]string{
		"echo hi",
		"unit digest: " + strings.Repeat("0", 64),
		"echo done",
	}, "\n")
}

func assertMultilineVerifyReceipt(t *testing.T, comments func(string, string) []string) {
	t.Helper()
	source := "GH issue #1"
	verifyCmd := multilineVerifyWithMisleadingDigest()
	units, issues := ValidateQueueBody(ownedQueue(checkedUnit(verifyCmd, "")), source)
	if len(units) != 1 {
		t.Fatalf("expected one queue unit, got %d", len(units))
	}
	digest := unitContractDigest(units[0])
	if digest == strings.Repeat("0", 64) {
		t.Fatal("fixture digest must differ from the misleading Verify digest")
	}

	records := mergeContinuedCommentRecords(comments(verifyCmd, digest))
	identity := queueUnitIdentity{Occurrence: 1, Heading: "My outcome"}
	gotIdentity, kind, gotDigest, err := parseRebuildCommentRecord(records[0], units)
	if err != nil {
		t.Fatalf("multiline Verify receipt returned an error: %v", err)
	}
	if gotIdentity != identity || kind != rebuildCommentEvidence || gotDigest != digest {
		t.Fatalf("receipt = identity %v, kind %d, digest %q; want %v, evidence, %q", gotIdentity, kind, gotDigest, identity, digest)
	}

	result := &ValidationResult{Valid: true}
	applyQueueIssues(result, source, units, issues, comments(verifyCmd, digest))
	if len(result.Errors) > 0 {
		t.Fatalf("expected multiline Verify receipt to satisfy checked unit, got %v", result.Errors)
	}
}

func TestFieldBoundaryReceiptUsesStructuredDigestAfterMultilineVerify(t *testing.T) {
	assertMultilineVerifyReceipt(t, func(verifyCmd, digest string) []string {
		return []string{
			continuedReceiptHead(verifyCmd, digest),
			continuedReceiptTail(verifyCmd),
		}
	})
}

func TestChunkedReceiptUsesStructuredDigestAfterMultilineVerify(t *testing.T) {
	assertMultilineVerifyReceipt(t, chunkedReceiptComments)
}

func renamedMultilineReceiptFixture(t *testing.T) (string, []queueUnit, string, string) {
	t.Helper()
	verifyCmd := multilineVerifyWithMisleadingDigest()
	oldBody := ownedQueue("## Old outcome\nDone means: target\nVerify:\n```\n" + verifyCmd + "\n```\n- [ ] pending\n")
	newBody := ownedQueue("## Prefix outcome\nDone means: prefix\nVerify:\n```\necho hi\n```\n- [ ] pending\n\n" +
		"## Renamed outcome\nDone means: target\nVerify:\n```\n" + verifyCmd + "\n```\n- [ ] pending\n")
	oldUnits := parseQueueUnits(oldBody)
	newUnits := parseQueueUnits(newBody)
	if len(oldUnits) != 1 || len(newUnits) != 2 {
		t.Fatalf("fixture units = %d old, %d new; want 1 old and 2 new", len(oldUnits), len(newUnits))
	}
	oldDigest := unitContractDigest(oldUnits[0])
	newDigest := unitContractDigest(newUnits[1])
	if oldDigest == newDigest {
		t.Fatal("fixture digests must differ across the renamed contract")
	}
	return verifyCmd, newUnits, oldDigest, newDigest
}

func receiptPartsForIdentity(
	makeReceipt func(string, string) []string,
	verifyCmd, heading, digest string,
) []string {
	parts := makeReceipt(verifyCmd, digest)
	for i := range parts {
		parts[i] = strings.ReplaceAll(parts[i], "Unit heading: My outcome", "Unit heading: "+heading)
	}
	return parts
}

func assertRenamedReceiptFallback(t *testing.T, makeReceipt func(string, string) []string) {
	t.Helper()
	verifyCmd, newUnits, oldDigest, newDigest := renamedMultilineReceiptFixture(t)
	oldIdentity := queueUnitIdentity{Occurrence: 1, Heading: "Old outcome"}
	newIdentity := queueUnitIdentity{Occurrence: 1, Heading: "Renamed outcome"}
	oldReceipt := receiptPartsForIdentity(makeReceipt, verifyCmd, oldIdentity.Heading, oldDigest)
	records := mergeContinuedCommentRecords(oldReceipt)
	gotIdentity, kind, gotDigest, err := parseRebuildCommentRecord(records[0], newUnits)
	if err != nil {
		t.Fatalf("renamed receipt returned an error: %v", err)
	}
	if gotIdentity != oldIdentity || kind != rebuildCommentEvidence || gotDigest != oldDigest {
		t.Fatalf("renamed receipt = identity %v, kind %d, digest %q; want %v, evidence, %q", gotIdentity, kind, gotDigest, oldIdentity, oldDigest)
	}

	comments := append([]string{}, oldReceipt...)
	comments = append(comments, amendmentRecord(1, newIdentity.Heading, oldDigest, newDigest, "rename the outcome"))
	comments = append(comments, receiptPartsForIdentity(makeReceipt, verifyCmd, newIdentity.Heading, newDigest)...)
	unresolved, errs := unresolvedRebuildRequests(newUnits, comments)
	if len(errs) > 0 {
		t.Fatalf("renamed receipt broke amendment routing: %v", errs)
	}
	if len(unresolved) > 0 {
		t.Fatalf("renamed receipt left amendment unresolved: %v", unresolved)
	}
}

func TestFieldBoundaryRenamedReceiptUsesCompleteCandidateGrammar(t *testing.T) {
	assertRenamedReceiptFallback(t, func(verifyCmd, digest string) []string {
		return []string{
			continuedReceiptHead(verifyCmd, digest),
			continuedReceiptTail(verifyCmd),
		}
	})
}

func TestChunkedRenamedReceiptUsesCompleteCandidateGrammar(t *testing.T) {
	assertRenamedReceiptFallback(t, chunkedReceiptComments)
}

func TestChunkedRawOutputDuplicateChunkFailsValidation(t *testing.T) {
	source := "GH issue #1"
	units, issues := ValidateQueueBody(ownedQueue(checkedUnit("echo hi", "")), source)
	digest := unitContractDigest(units[0])
	comments := chunkedReceiptComments("echo hi", digest)
	duplicate := rawOutputChunk("pre", 1, 2, digest, "duplicate", true)
	comments[1] = strings.Replace(comments[1], rawOutputChunk("pre", 2, 2, digest, "pre two", true), duplicate, 1)
	result := &ValidationResult{Valid: true}
	applyQueueIssues(result, "GitHub comments", units, issues, comments)
	if len(result.Errors) == 0 {
		t.Fatal("expected duplicate raw output chunk to fail validation")
	}
}

func TestChunkedRawOutputInterruptedByCommentFailsValidation(t *testing.T) {
	source := "GH issue #1"
	units, issues := ValidateQueueBody(ownedQueue(checkedUnit("echo hi", "")), source)
	digest := unitContractDigest(units[0])
	comments := chunkedReceiptComments("echo hi", digest)
	comments = []string{comments[0], "unrelated comment", comments[1], comments[2]}
	result := &ValidationResult{Valid: true}
	applyQueueIssues(result, "GitHub comments", units, issues, comments)
	if len(result.Errors) == 0 {
		t.Fatal("expected interrupted raw output continuation to fail validation")
	}
}

func TestSplitReceiptResolvesAmendmentRequest(t *testing.T) {
	body := func(doneMeans string) string {
		return ownedQueue("## My outcome\nDone means: " + doneMeans + "\nVerify:\n```\necho hi\n```\n- [x] done\n")
	}
	oldUnits, _ := ValidateQueueBody(body("something"), "fixture")
	newUnits, _ := ValidateQueueBody(body("something tighter"), "fixture")
	oldDigest := unitContractDigest(oldUnits[0])
	newDigest := unitContractDigest(newUnits[0])
	comments := []string{
		amendmentRecord(1, "My outcome", oldDigest, newDigest, "tighten the outcome"),
		continuedReceiptHead("echo hi", newDigest),
		continuedReceiptTail("echo hi"),
	}
	unresolved, errs := unresolvedRebuildRequests(newUnits, comments)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(unresolved) != 0 {
		t.Fatalf("expected split receipt to resolve the amendment, got unresolved %v", unresolved)
	}
}

func TestChunkedPreAmendmentReceiptClassifiesAsStaleEvidence(t *testing.T) {
	body := func(doneMeans string) string {
		return ownedQueue("## My outcome\nDone means: " + doneMeans + "\nVerify:\n```\necho hi\n```\n- [x] done\n")
	}
	oldUnits, _ := ValidateQueueBody(body("something"), "fixture")
	newUnits, _ := ValidateQueueBody(body("something tighter"), "fixture")
	oldDigest := unitContractDigest(oldUnits[0])
	newDigest := unitContractDigest(newUnits[0])
	identity := queueUnitIdentity{Occurrence: 1, Heading: "My outcome"}

	records := mergeContinuedCommentRecords(chunkedReceiptComments("echo hi", oldDigest))
	gotIdentity, kind, digest, err := parseRebuildCommentRecord(records[0], newUnits)
	if err != nil {
		t.Fatalf("chunked pre-amendment receipt returned an error: %v", err)
	}
	if gotIdentity != identity || kind != rebuildCommentStaleEvidence || digest != oldDigest {
		t.Fatalf("chunked pre-amendment receipt = identity %v, kind %d, digest %q; want %v, stale evidence, %q", gotIdentity, kind, digest, identity, oldDigest)
	}
	if newDigest == oldDigest {
		t.Fatal("fixture digests must differ across the amendment")
	}
}

func TestDanglingSplitReceiptStillFailsValidation(t *testing.T) {
	source := "GH issue #1"
	units, issues := ValidateQueueBody(ownedQueue(checkedUnit("echo hi", "")), source)
	digest := unitContractDigest(units[0])
	result := &ValidationResult{Valid: true}
	applyQueueIssues(result, "GitHub comments", units, issues, []string{continuedReceiptHead("echo hi", digest)})
	if len(result.Errors) == 0 {
		t.Fatal("expected dangling continuation to fail validation, got none")
	}
	joined := strings.ToLower(strings.Join(issueMessages(result.Errors), "\n"))
	if !strings.Contains(joined, "evidence receipt") {
		t.Fatalf("expected an evidence receipt error, got %v", result.Errors)
	}
}

func TestInterruptedContinuationFailsValidation(t *testing.T) {
	source := "GH issue #1"
	units, issues := ValidateQueueBody(ownedQueue(checkedUnit("echo hi", "")), source)
	digest := unitContractDigest(units[0])
	comments := []string{
		continuedReceiptHead("echo hi", digest),
		"unrelated comment posted in between",
		continuedReceiptTail("echo hi"),
	}
	result := &ValidationResult{Valid: true}
	applyQueueIssues(result, "GitHub comments", units, issues, comments)
	if len(result.Errors) == 0 {
		t.Fatal("expected interrupted continuation to fail validation, got none")
	}
	joined := strings.ToLower(strings.Join(issueMessages(result.Errors), "\n"))
	if !strings.Contains(joined, "evidence receipt") {
		t.Fatalf("expected an evidence receipt error, got %v", result.Errors)
	}
}

func TestRawOutputChunkMissingMarkerFailsValidation(t *testing.T) {
	units, issues := ValidateQueueBody(ownedQueue(checkedUnit("echo hi", "")), "GH issue #1")
	digest := unitContractDigest(units[0])
	comments := chunkedReceiptComments("echo hi", digest)
	comments[0] = strings.TrimSuffix(comments[0], "\n"+receiptContinuationMarker)
	result := &ValidationResult{Valid: true}
	applyQueueIssues(result, "GitHub comments", units, issues, comments[:1])
	if len(result.Errors) == 0 {
		t.Fatal("expected missing raw output continuation marker to fail validation")
	}
}

func TestRawOutputChunksInOneCommentFailValidation(t *testing.T) {
	units, issues := ValidateQueueBody(ownedQueue(checkedUnit("echo hi", "")), "GH issue #1")
	digest := unitContractDigest(units[0])
	comments := chunkedReceiptComments("echo hi", digest)
	first := strings.TrimSuffix(comments[0], "\n"+receiptContinuationMarker)
	comments[0] = first + "\n" + rawOutputChunk("pre", 2, 2, digest, "pre two", false)
	result := &ValidationResult{Valid: true}
	applyQueueIssues(result, "GitHub comments", units, issues, comments[:1])
	if len(result.Errors) == 0 {
		t.Fatal("expected multiple raw output chunks in one comment to fail validation")
	}
}

func TestRawOutputChunksBeforeHeadingFormReceiptFailValidation(t *testing.T) {
	source := "GH issue #1"
	units, issues := ValidateQueueBody(ownedQueue(checkedUnit("echo hi", evidenceReceipt("echo hi"))), source)
	digest := unitContractDigest(units[0])
	headingReceipt := headingFormReceipt("echo hi", digest)

	tests := []struct {
		name     string
		comments []string
	}{
		{
			name: "same physical comment",
			comments: []string{strings.Join([]string{
				rawOutputChunk("pre", 1, 2, digest, "orphan one", false),
				rawOutputChunk("pre", 2, 2, digest, "orphan two", false),
				headingReceipt,
			}, "\n")},
		},
		{
			name: "earlier marker-joined comments",
			comments: []string{
				rawOutputChunk("pre", 1, 2, digest, "orphan one", true),
				rawOutputChunk("pre", 2, 2, digest, "orphan two", true),
				headingReceipt,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := &ValidationResult{Valid: true}
			applyQueueIssues(result, "GitHub comments", units, issues, test.comments)
			if !containsIssue(result.Errors, "orphan raw output chunk") {
				t.Fatalf("expected orphan raw output chunk error, got %v", result.Errors)
			}
		})
	}
}

func TestRawOutputChunkUnitHeadingMustMatchByteForByte(t *testing.T) {
	source := "GH issue #1"
	units, issues := ValidateQueueBody(ownedQueue(checkedUnit("echo hi", evidenceReceipt("echo hi"))), source)
	digest := unitContractDigest(units[0])
	comments := headingFormChunkedReceiptComments("echo hi", digest)
	comments[0] = strings.Replace(comments[0], "Unit heading: My outcome", "Unit heading: My outcome ", 1)

	records := mergeContinuedCommentRecords(comments)
	document, ok := commentEvidenceDocument(records[0], "My outcome")
	if !ok {
		t.Fatal("expected heading-form receipt to be located")
	}
	identity := queueUnitIdentity{Occurrence: 1, Heading: "My outcome"}
	receiptIssues, _ := evidenceReceiptIssuesForDocument(document, "echo hi", digest, "comment", identity.Heading, &identity)
	if !containsIssue(receiptIssues, "Unit heading must match the receipt identity exactly") {
		t.Fatalf("expected byte-for-byte heading error, got %v", receiptIssues)
	}

	result := &ValidationResult{Valid: true}
	applyQueueIssues(result, "GitHub comments", units, issues, comments)
	if !containsIssue(result.Errors, "orphan raw output chunk") {
		t.Fatalf("expected malformed chunk receipt to be rejected, got %v", result.Errors)
	}
}

func TestRawOutputChunkDuplicatedAfterReceiptFailsValidation(t *testing.T) {
	units, issues := ValidateQueueBody(ownedQueue(checkedUnit("echo hi", "")), "GH issue #1")
	digest := unitContractDigest(units[0])
	comments := chunkedReceiptComments("echo hi", digest)
	comments[3] += "\n" + rawOutputChunk("post", 1, 2, digest, "duplicate", false)
	result := &ValidationResult{Valid: true}
	applyQueueIssues(result, "GitHub comments", units, issues, comments)
	if len(result.Errors) == 0 {
		t.Fatal("expected duplicated raw output chunk after a complete receipt to fail validation")
	}
}

func TestHeadingFormChunkReceiptResolves(t *testing.T) {
	units, issues := ValidateQueueBody(ownedQueue(checkedUnit("echo hi", "")), "GH issue #1")
	digest := unitContractDigest(units[0])
	comments := chunkedReceiptComments("echo hi", digest)
	lines := strings.SplitN(comments[0], "\n", 3)
	comments[0] = "## My outcome\n" + lines[2]
	result := &ValidationResult{Valid: true}
	applyQueueIssues(result, "GitHub comments", units, issues, comments)
	if len(result.Errors) > 0 {
		t.Fatalf("expected heading-form chunk receipt to satisfy checked unit, got %v", result.Errors)
	}
}

func TestHeadingFormChunkOccurrenceMustMatchUnit(t *testing.T) {
	body := ownedQueue(
		"## My outcome\nDone means: first thing\nVerify:\n```\necho hi\n```\n- [x] done\n\n" +
			"## My outcome\nDone means: second thing\nVerify:\n```\necho hi\n```\n- [x] done\n",
	)
	units, issues := ValidateQueueBody(body, "GH issue #1")
	firstDigest := unitContractDigest(units[0])
	secondDigest := unitContractDigest(units[1])
	firstComments := chunkedReceiptComments("echo hi", firstDigest)
	secondComments := headingFormChunkedReceiptComments("echo hi", secondDigest)
	comments := append(firstComments, secondComments...)
	records := mergeContinuedCommentRecords(comments)
	firstIdentity := queueUnitIdentity{Occurrence: 1, Heading: "My outcome"}
	secondIdentity := queueUnitIdentity{Occurrence: 2, Heading: "My outcome"}
	used := make(map[int]bool)
	if _, ok := matchingEvidenceCommentForUnit(firstIdentity, units[0], firstDigest, units, records, used); !ok {
		t.Fatal("expected complete receipt for occurrence 1 to match occurrence 1")
	}
	used[0] = true
	if _, ok := matchingEvidenceCommentForUnit(secondIdentity, units[1], secondDigest, units, records, used); ok {
		t.Fatal("expected complete heading-form receipt declaring occurrence 1 not to match occurrence 2")
	}

	document, ok := commentEvidenceDocument(records[len(firstComments)], "My outcome")
	if !ok {
		t.Fatal("expected heading-form receipt to be located")
	}
	receiptIssues, _ := evidenceReceiptIssuesForDocument(document, "echo hi", secondDigest, "comment", secondIdentity.Heading, &secondIdentity)
	if !containsIssue(receiptIssues, "Unit occurrence must match the receipt identity") {
		t.Fatalf("expected occurrence mismatch error, got %v", receiptIssues)
	}

	result := &ValidationResult{Valid: true}
	applyQueueIssues(result, "GH issue #1", units, issues, comments)
	if len(result.Errors) == 0 {
		t.Fatal("expected occurrence 2 to remain unsatisfied by occurrence 1's receipt")
	}
}

func issueMessages(issues []ValidationIssue) []string {
	messages := make([]string, 0, len(issues))
	for _, iss := range issues {
		messages = append(messages, iss.Message)
	}
	return messages
}
