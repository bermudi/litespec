package internal

import (
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

func issueMessages(issues []ValidationIssue) []string {
	messages := make([]string, 0, len(issues))
	for _, iss := range issues {
		messages = append(messages, iss.Message)
	}
	return messages
}
