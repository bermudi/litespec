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

func TestOrphanRawOutputChunkFailsValidation(t *testing.T) {
	units, issues := ValidateQueueBody(ownedQueue(checkedUnit("echo hi", "")), "GH issue #1")
	digest := unitContractDigest(units[0])
	result := &ValidationResult{Valid: true}
	applyQueueIssues(result, "GitHub comments", units, issues, []string{rawOutputChunk("pre", 1, 2, digest, "orphan", false)})
	if len(result.Errors) == 0 {
		t.Fatal("expected orphan raw output chunk to fail validation")
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
	comments := chunkedReceiptComments("echo hi", firstDigest)
	lines := strings.SplitN(comments[0], "\n", 3)
	comments[0] = "## My outcome\n" + lines[2]
	comments[0] = strings.Replace(comments[0], "Unit occurrence: 1", "Unit occurrence: 999", 1)
	result := &ValidationResult{Valid: true}
	applyQueueIssues(result, "GitHub comments", units, issues, comments)
	if len(result.Errors) == 0 {
		t.Fatal("expected heading-form receipt with wrong chunk occurrence to fail validation")
	}
}

func issueMessages(issues []ValidationIssue) []string {
	messages := make([]string, 0, len(issues))
	for _, iss := range issues {
		messages = append(messages, iss.Message)
	}
	return messages
}
