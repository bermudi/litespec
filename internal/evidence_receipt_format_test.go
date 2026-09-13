package internal

import (
	"strings"
	"testing"
)

func blankSeparatedReceipt(verifyCmd, digest, verifyLine string) string {
	return strings.Join([]string{
		"Unit occurrence: 1",
		"Unit heading: My outcome",
		"Evidence:",
		"",
		verifyLine,
		"",
		"unit digest: " + digest,
		"",
		"pre sha: " + evidenceTestSHA,
		"",
		"pre exit status: 1",
		"",
		"```",
		verifyCmd + " missing outcome",
		"```",
		"",
		"Pre-evidence scope: this command exited 1 at " + evidenceTestSHA + "; nothing else is inferred.",
		"",
		"post sha: " + evidencePostTestSHA,
		"",
		"post exit status: 0",
		"",
		"```",
		verifyCmd + " output",
		"```",
		"",
		"Post-evidence scope: this command exited 0 at " + evidencePostTestSHA + "; nothing else is inferred.",
		"",
	}, "\n")
}

func TestEvidenceReceiptToleratesBlankLinesBetweenFields(t *testing.T) {
	source := "GH issue #1"
	units, issues := ValidateQueueBody(ownedQueue(checkedUnit("echo hi", "")), source)
	digest := unitContractDigest(units[0])
	result := &ValidationResult{Valid: true}
	applyQueueIssues(result, "GitHub comments", units, issues, []string{blankSeparatedReceipt("echo hi", digest, "echo hi")})
	if len(result.Errors) > 0 {
		t.Fatalf("expected blank-separated receipt to satisfy checked unit, got %v", result.Errors)
	}
}

func TestEvidenceReceiptAcceptsVerifyLabelForm(t *testing.T) {
	source := "GH issue #1"
	units, issues := ValidateQueueBody(ownedQueue(checkedUnit("echo hi", "")), source)
	digest := unitContractDigest(units[0])
	result := &ValidationResult{Valid: true}
	applyQueueIssues(result, "GitHub comments", units, issues, []string{blankSeparatedReceipt("echo hi", digest, "Verify: `echo hi`")})
	if len(result.Errors) > 0 {
		t.Fatalf("expected Verify-label receipt to satisfy checked unit, got %v", result.Errors)
	}
}

func TestEvidenceReceiptStillRejectsProseBetweenFields(t *testing.T) {
	source := "GH issue #1"
	units, _ := ValidateQueueBody(ownedQueue(checkedUnit("echo hi", "")), source)
	digest := unitContractDigest(units[0])
	receipt := strings.Replace(
		blankSeparatedReceipt("echo hi", digest, "echo hi"),
		"pre sha: "+evidenceTestSHA,
		"pre sha: "+evidenceTestSHA+"\n\nthe run failed as expected",
		1,
	)
	issues := evidenceReceiptIssues(receipt, "echo hi", digest, "comment", "My outcome")
	if len(issues) == 0 {
		t.Fatal("expected prose between fields to fail parsing, got none")
	}
	if !strings.Contains(issues[0].Message, "fields must appear in order") {
		t.Fatalf("expected field-order error, got %v", issues)
	}
}

func TestEvidenceReceiptRejectsLabelFormWithWrongCommand(t *testing.T) {
	source := "GH issue #1"
	units, _ := ValidateQueueBody(ownedQueue(checkedUnit("echo hi", "")), source)
	digest := unitContractDigest(units[0])
	issues := evidenceReceiptIssues(blankSeparatedReceipt("echo hi", digest, "Verify: `echo goodbye`"), "echo hi", digest, "comment", "My outcome")
	if len(issues) == 0 {
		t.Fatal("expected wrong Verify command to fail parsing, got none")
	}
	if !strings.Contains(issues[0].Message, "must quote the Verify command verbatim") {
		t.Fatalf("expected verbatim-command error, got %v", issues)
	}
}
