package internal

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

var (
	v2TestPreSHA  = strings.Repeat("a", 40)
	v2TestPostSHA = strings.Repeat("b", 40)
)

func v2TestRequest(preOutput, postOutput string) ReceiptAssemblyRequest {
	request := assemblyTestRequest()
	request.Pre.SHA = v2TestPreSHA
	request.Post.SHA = v2TestPostSHA
	request.Pre.Output = preOutput
	request.Post.Output = postOutput
	return request
}

func v2TestLargeOutput(prefix string) string {
	lines := make([]string, 0, 150)
	for i := 0; i < 150; i++ {
		lines = append(lines, fmt.Sprintf("%s line %03d of routine bounded-receipt output for the elision path", prefix, i))
	}
	return strings.Join(lines, "\n")
}

func v2TestIssues(t *testing.T, comments []string) []ValidationIssue {
	t.Helper()
	merged := mergeContinuedCommentRecords(comments)
	if len(merged) != 1 {
		t.Fatalf("expected one merged receipt comment, got %d", len(merged))
	}
	document := evidencePayloadDocument(newEvidenceDocumentFromComment(merged[0]))
	identity := queueUnitIdentity{Occurrence: 1, Heading: "My outcome"}
	issues, _ := evidenceReceiptIssuesForDocument(
		document,
		assemblyTestVerify,
		fixtureUnitDigest(assemblyTestVerify),
		"test",
		"My outcome",
		&identity,
	)
	return issues
}

func v2TestRequireNoIssues(t *testing.T, comments []string) {
	t.Helper()
	if issues := v2TestIssues(t, comments); len(issues) > 0 {
		t.Fatalf("expected the receipt to validate, got issues: %v", issues)
	}
}

func v2TestRequireIssue(t *testing.T, comments []string, substr string) string {
	t.Helper()
	issues := v2TestIssues(t, comments)
	if len(issues) == 0 {
		t.Fatalf("expected a validation issue naming %q, got none", substr)
	}
	for _, issue := range issues {
		if strings.Contains(issue.Message, substr) {
			if !strings.Contains(issue.Message, "My outcome") {
				t.Fatalf("validation issue does not name the unit: %v", issue.Message)
			}
			return issue.Message
		}
	}
	t.Fatalf("expected a validation issue containing %q, got: %v", substr, issues)
	return ""
}

// v2TestMutateFencePayload rewrites one run's fenced payload in an assembled
// receipt comment, leaving every other byte untouched.
func v2TestMutateFencePayload(t *testing.T, comment, phase string, transform func(string) string) string {
	t.Helper()
	lines := strings.Split(comment, "\n")
	anchor := phase + " exit status:"
	openIndex := -1
	delimiter := ""
	for i, line := range lines {
		if strings.HasPrefix(line, anchor) {
			if i+1 >= len(lines) {
				t.Fatalf("no fence follows the %s line", anchor)
			}
			delimiter = lines[i+1]
			if fenceDelimiter(strings.TrimSpace(delimiter)) == "" {
				t.Fatalf("line after %s is not a fence: %q", anchor, delimiter)
			}
			openIndex = i + 2
			break
		}
	}
	if openIndex < 0 {
		t.Fatalf("comment has no %s line", anchor)
	}
	closeIndex := openIndex
	for closeIndex < len(lines) && lines[closeIndex] != delimiter {
		closeIndex++
	}
	if closeIndex >= len(lines) {
		t.Fatal("fence is never closed")
	}
	payload := strings.Join(lines[openIndex:closeIndex], "\n")
	replacement := strings.Split(transform(payload), "\n")
	updated := append([]string{}, lines[:openIndex]...)
	updated = append(updated, replacement...)
	updated = append(updated, lines[closeIndex:]...)
	return strings.Join(updated, "\n")
}

func v2TestAssembleOne(t *testing.T, request ReceiptAssemblyRequest) string {
	t.Helper()
	comments, err := AssembleEvidenceV2ReceiptComments(request)
	if err != nil {
		t.Fatalf("assembly refused: %v", err)
	}
	if len(comments) != 1 {
		t.Fatalf("expected one comment, got %d", len(comments))
	}
	return comments[0]
}

func TestValidateEvidenceV2Receipts(t *testing.T) {
	t.Run("v2 receipt with consistent excerpt fields validates", func(t *testing.T) {
		small := v2TestAssembleOne(t, v2TestRequest("missing outcome\n", "outcome present\n"))
		v2TestRequireNoIssues(t, []string{small})

		elided := v2TestAssembleOne(t, v2TestRequest(v2TestLargeOutput("pre"), v2TestLargeOutput("post")))
		if !strings.Contains(elided, "bytes elided") {
			t.Fatalf("expected the large-output receipt to carry an elision marker:\n%s", elided)
		}
		v2TestRequireNoIssues(t, []string{elided})

		empty := v2TestAssembleOne(t, v2TestRequest("", ""))
		if !strings.Contains(empty, "<no output>") {
			t.Fatalf("expected the empty-output receipt to carry `<no output>`:\n%s", empty)
		}
		v2TestRequireNoIssues(t, []string{empty})

		asymmetric := v2TestAssembleOne(t, v2TestRequest("", "outcome present\n"))
		v2TestRequireNoIssues(t, []string{asymmetric})

		rebuild := v2TestAssembleOne(t, func() ReceiptAssemblyRequest {
			request := v2TestRequest("missing outcome\n", "outcome present\n")
			request.Rebuild = true
			return request
		}())
		if !strings.Contains(rebuild, "Unit occurrence: 1") {
			t.Fatalf("expected the rebuild receipt to be identity-bearing:\n%s", rebuild)
		}
		v2TestRequireNoIssues(t, []string{rebuild})

		markerShaped := v2TestAssembleOne(t, v2TestRequest(
			"audit start\n... 123 bytes elided ...\naudit end\n",
			"outcome present\n",
		))
		v2TestRequireNoIssues(t, []string{markerShaped})

		inconsistent := v2TestMutateFencePayload(t, elided, "pre", func(payload string) string {
			return regexp.MustCompile(`\.\.\. (\d+) bytes elided \.\.\.`).ReplaceAllStringFunc(
				payload,
				func(marker string) string {
					var n int
					fmt.Sscanf(marker, "... %d bytes elided ...", &n)
					return fmt.Sprintf("... %d bytes elided ...", n+1)
				},
			)
		})
		v2TestRequireIssue(t, []string{inconsistent}, "reconstruction arithmetic does not hold")

		zeroMarker := v2TestMutateFencePayload(t, elided, "pre", func(payload string) string {
			return regexp.MustCompile(`\.\.\. \d+ bytes elided \.\.\.`).ReplaceAllLiteralString(payload, "... 0 bytes elided ...")
		})
		v2TestRequireIssue(t, []string{zeroMarker}, "zero elided")

		missingMarker := v2TestMutateFencePayload(t, elided, "pre", func(payload string) string {
			return regexp.MustCompile(`(?m)^\.\.\. \d+ bytes elided \.\.\.\n`).ReplaceAllLiteralString(payload, "")
		})
		v2TestRequireIssue(t, []string{missingMarker}, "neither holds the full output nor carries an elision marker")
	})

	t.Run("v2 unelided output must match declared bytes and sha256", func(t *testing.T) {
		small := v2TestAssembleOne(t, v2TestRequest("missing outcome\n", "outcome present\n"))

		hashMismatch := v2TestMutateFencePayload(t, small, "pre", func(payload string) string {
			return strings.Replace(payload, "missing outcome", "missing 0utcome", 1)
		})
		v2TestRequireIssue(t, []string{hashMismatch}, "does not hash to the declared")

		byteMismatch := v2TestMutateFencePayload(t, small, "pre", func(payload string) string {
			return payload
		})
		byteMismatch = strings.Replace(byteMismatch, "pre bytes: 16\n", "pre bytes: 15\n", 1)
		v2TestRequireIssue(t, []string{byteMismatch}, "neither holds the full output nor carries an elision marker")

		zeroDeclared := strings.Replace(small, "pre bytes: 16\n", "pre bytes: 0\n", 1)
		v2TestRequireIssue(t, []string{zeroDeclared}, "does not hold exactly `<no output>`")
	})

	t.Run("v2 receipt must not continue or chunk across comments", func(t *testing.T) {
		large := v2TestRequest(v2TestLargeOutput("pre"), v2TestLargeOutput("post"))
		large.CommentLimit = 4096
		v1Chain, err := AssembleEvidenceReceiptComments(large)
		if err != nil {
			t.Fatalf("v1 assembly refused: %v", err)
		}

		v2Chunked := make([]string, len(v1Chain))
		for i, comment := range v1Chain {
			v2Chunked[i] = strings.ReplaceAll(comment, "Protocol: "+evidenceProtocolV1, "Protocol: "+evidenceProtocolV2)
			v2Chunked[i] = strings.ReplaceAll(v2Chunked[i], receiptIDPrefix, receiptIDV2Prefix)
		}
		v2TestRequireIssue(t, v2Chunked, "must not use the raw output chunk form")

		small := v2TestAssembleOne(t, v2TestRequest("missing outcome\n", "outcome present\n"))
		dangling := small + "\n" + receiptContinuationMarker
		v2TestRequireIssue(t, []string{dangling}, "must not end with the continuation marker")

		lines := strings.Split(small, "\n")
		splitAt := -1
		for i, line := range lines {
			if strings.HasPrefix(line, "Pre-evidence scope:") {
				splitAt = i + 1
				break
			}
		}
		if splitAt < 0 {
			t.Fatal("receipt has no Pre-evidence scope line")
		}
		first := strings.Join(lines[:splitAt], "\n") + "\n" + receiptContinuationMarker
		second := strings.Join(lines[splitAt:], "\n")
		v2TestRequireIssue(t, []string{first, second}, "must not continue across comments")

		padded := v2TestMutateFencePayload(t, small, "pre", func(payload string) string {
			return payload + "\n" + strings.Repeat("x", receiptV2TotalBudget)
		})
		v2TestRequireIssue(t, []string{padded}, "byte evidence/v2 receipt budget")
	})

	t.Run("v1 and legacy receipts still validate verbatim", func(t *testing.T) {
		smallV1, err := AssembleEvidenceReceiptComments(v2TestRequest("missing outcome\n", "outcome present\n"))
		if err != nil {
			t.Fatalf("v1 assembly refused: %v", err)
		}
		v2TestRequireNoIssues(t, smallV1)

		largeRequest := v2TestRequest(v2TestLargeOutput("pre"), v2TestLargeOutput("post"))
		largeRequest.CommentLimit = 4096
		largeV1, err := AssembleEvidenceReceiptComments(largeRequest)
		if err != nil {
			t.Fatalf("v1 assembly refused: %v", err)
		}
		if len(largeV1) < 2 {
			t.Fatalf("expected the oversized v1 receipt to continue across comments, got %d comments", len(largeV1))
		}
		v2TestRequireNoIssues(t, largeV1)

		legacy := strings.Join([]string{
			"## My outcome",
			"Evidence:",
			assemblyTestVerify,
			"unit digest: " + fixtureUnitDigest(assemblyTestVerify),
			"pre sha: " + v2TestPreSHA,
			"pre exit status: 1",
			"```",
			"missing outcome",
			"```",
			receiptScopeLine("Pre", 1, v2TestPreSHA),
			"post sha: " + v2TestPostSHA,
			"post exit status: 0",
			"```",
			"outcome present",
			"```",
			receiptScopeLine("Post", 0, v2TestPostSHA),
		}, "\n")
		v2TestRequireNoIssues(t, []string{legacy})
	})
}
