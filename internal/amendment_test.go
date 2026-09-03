package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const amendmentTestOldDoneMeans = "it works as specified v1"
const amendmentTestNewDoneMeans = "it works as specified v2"

func amendmentTestBody(doneMeans, heading string) string {
	return strings.Join([]string{
		"Base: 0000000000000000000000000000000000000002",
		"Branch: litespec/demo",
		"",
		"## " + heading,
		"",
		"Depends:",
		"Done means:",
		"- [outcome] " + doneMeans,
		"Scenarios:",
		"- [outcome] TestOutcome",
		"",
		"Verify:",
		digestTestFence + "bash",
		"echo first",
		digestTestFence,
		"",
		"- [x] done",
	}, "\n")
}

func amendmentTestUnits(t *testing.T, body string) []queueUnit {
	t.Helper()
	var units []queueUnit
	for _, section := range parseQueueUnits(body) {
		if isUnit(section) {
			units = append(units, section)
		}
	}
	if len(units) != 1 {
		t.Fatalf("fixture should yield 1 unit, got %d", len(units))
	}
	return units
}

func amendmentReceipt(occ int, heading, digest string) string {
	const preSHA = "1111111111111111111111111111111111111111"
	const postSHA = "2222222222222222222222222222222222222222"
	return strings.Join([]string{
		fmt.Sprintf("Unit occurrence: %d", occ),
		fmt.Sprintf("Unit heading: %s", heading),
		"Evidence:",
		"echo first",
		"unit digest: " + digest,
		"pre sha: " + preSHA,
		"pre exit status: 1",
		digestTestFence,
		"boom",
		digestTestFence,
		fmt.Sprintf("Pre-evidence scope: this command exited 1 at %s; nothing else is inferred.", preSHA),
		"post sha: " + postSHA,
		"post exit status: 0",
		digestTestFence,
		"fine",
		digestTestFence,
		fmt.Sprintf("Post-evidence scope: this command exited 0 at %s; nothing else is inferred.", postSHA),
	}, "\n")
}

func amendmentRecord(occ int, heading, oldDigest, newDigest, reason string) string {
	return strings.Join([]string{
		"Amendment:",
		fmt.Sprintf("Unit occurrence: %d", occ),
		fmt.Sprintf("Unit heading: %s", heading),
		"Old digest: " + oldDigest,
		"New digest: " + newDigest,
		"Reason: " + reason,
	}, "\n")
}

const amendmentFakeDigest = "3333333333333333333333333333333333333333333333333333333333333333"

func amendmentTestBodyWithVerify(doneMeans, heading, verify string) string {
	return strings.Replace(amendmentTestBody(doneMeans, heading), "echo first", verify, 1)
}

func amendmentReceiptWithVerify(occurrence int, heading, digest, verify string) string {
	return strings.Replace(amendmentReceipt(occurrence, heading, digest), "echo first", verify, 1)
}

func TestAmendmentImpliesUnresolvedRebuildRequest(t *testing.T) {
	oldUnits := amendmentTestUnits(t, amendmentTestBody(amendmentTestOldDoneMeans, "First unit"))
	newBody := amendmentTestBody(amendmentTestNewDoneMeans, "First unit")
	newUnits := amendmentTestUnits(t, newBody)
	oldDigest := unitContractDigest(oldUnits[0])
	newDigest := unitContractDigest(newUnits[0])
	verifyCmd := unitVerifyCommand(newUnits[0].Body)
	if verifyCmd != unitVerifyCommand(oldUnits[0].Body) {
		t.Fatal("fixture verify command must not change between contract states")
	}
	if oldDigest == newDigest {
		t.Fatal("fixture digests must differ across the contract edit")
	}

	identity := queueUnitIdentity{Occurrence: 1, Heading: "First unit"}

	t.Run("amendment alone leaves the checked unit selectable and unresolved", func(t *testing.T) {
		comments := []string{
			amendmentReceipt(1, "First unit", oldDigest),
			amendmentRecord(1, "First unit", oldDigest, newDigest, "tighten the outcome"),
		}
		unresolved, errs := unresolvedRebuildRequests(newUnits, comments)
		if len(errs) != 0 {
			t.Fatalf("unexpected errors: %v", errs)
		}
		if len(unresolved) != 1 || unresolved[0] != identity {
			t.Errorf("unresolved = %v, want [occurrence 1 heading %q]", unresolved, identity.Heading)
		}
		selectable, errs := selectableUnitIdentities(newUnits, comments)
		if len(errs) != 0 {
			t.Fatalf("unexpected selection errors: %v", errs)
		}
		if len(selectable) != 1 || selectable[0] != identity {
			t.Errorf("selectable = %v, want the amended checked unit", selectable)
		}
	})

	t.Run("later complete receipt carrying the new digest resolves it", func(t *testing.T) {
		comments := []string{
			amendmentReceipt(1, "First unit", oldDigest),
			amendmentRecord(1, "First unit", oldDigest, newDigest, "tighten the outcome"),
			amendmentReceipt(1, "First unit", newDigest),
		}
		unresolved, errs := unresolvedRebuildRequests(newUnits, comments)
		if len(errs) != 0 {
			t.Fatalf("unexpected errors: %v", errs)
		}
		if len(unresolved) != 0 {
			t.Errorf("unresolved = %v, want none after resolution", unresolved)
		}
	})

	t.Run("heading rename accepted via post-amendment identity and Old digest provenance", func(t *testing.T) {
		renamedUnits := amendmentTestUnits(t, amendmentTestBody(amendmentTestNewDoneMeans, "Renamed unit"))
		renamedIdentity := queueUnitIdentity{Occurrence: 1, Heading: "Renamed unit"}
		renamedDigest := unitContractDigest(renamedUnits[0])
		comments := []string{
			amendmentReceipt(1, "First unit", oldDigest),
			amendmentRecord(1, "Renamed unit", oldDigest, renamedDigest, "clarify the heading"),
			amendmentReceipt(1, "Renamed unit", renamedDigest),
		}
		unresolved, errs := unresolvedRebuildRequests(renamedUnits, comments)
		if len(errs) != 0 {
			t.Fatalf("unexpected errors: %v", errs)
		}
		if len(unresolved) != 0 {
			t.Errorf("unresolved = %v, want none after rename resolution", unresolved)
		}
		selectable, errs := selectableUnitIdentities(renamedUnits, comments[:2])
		if len(errs) != 0 {
			t.Fatalf("unexpected selection errors: %v", errs)
		}
		if len(selectable) != 1 || selectable[0] != renamedIdentity {
			t.Errorf("selectable = %v, want the renamed identity still open", selectable)
		}
	})

	t.Run("silent edit plus fresh receipt without amendment is a broken chain error", func(t *testing.T) {
		comments := []string{
			amendmentReceipt(1, "First unit", oldDigest),
			amendmentReceipt(1, "First unit", newDigest),
		}
		unresolved, errs := unresolvedRebuildRequests(newUnits, comments)
		if len(unresolved) != 0 {
			t.Errorf("unresolved = %v, want none", unresolved)
		}
		found := false
		for _, err := range errs {
			if strings.Contains(err.Error(), "not bridged by an amendment") &&
				strings.Contains(err.Error(), `"First unit"`) {
				found = true
			}
		}
		if !found {
			t.Errorf("expected a broken-chain error naming the unit, got %v", errs)
		}
	})

	t.Run("two amendments between receipts bridge the chain", func(t *testing.T) {
		midDigest := amendmentFakeDigest
		comments := []string{
			amendmentReceipt(1, "First unit", oldDigest),
			amendmentRecord(1, "First unit", oldDigest, midDigest, "first edit"),
			amendmentRecord(1, "First unit", midDigest, newDigest, "second edit"),
			amendmentReceipt(1, "First unit", newDigest),
		}
		unresolved, errs := unresolvedRebuildRequests(newUnits, comments)
		if len(errs) != 0 {
			t.Fatalf("unexpected errors: %v", errs)
		}
		if len(unresolved) != 0 {
			t.Errorf("unresolved = %v, want none", unresolved)
		}
	})

	t.Run("fake bridging amendment from a foreign old digest does not connect the chain", func(t *testing.T) {
		comments := []string{
			amendmentReceipt(1, "First unit", oldDigest),
			amendmentRecord(1, "First unit", amendmentFakeDigest, newDigest, "fabricated provenance"),
			amendmentReceipt(1, "First unit", newDigest),
		}
		_, errs := unresolvedRebuildRequests(newUnits, comments)
		found := false
		for _, err := range errs {
			if strings.Contains(err.Error(), "not bridged by an amendment") {
				found = true
			}
		}
		if !found {
			t.Errorf("expected a broken-chain error, got %v", errs)
		}
	})

	t.Run("malformed amendment is a visible error", func(t *testing.T) {
		malformed := strings.Join([]string{
			"Amendment:",
			"Unit occurrence: 1",
			"Unit heading: First unit",
			"Old digest: " + oldDigest,
			"New digest: " + newDigest,
		}, "\n")
		unresolved, errs := unresolvedRebuildRequests(newUnits, []string{malformed})
		if len(unresolved) != 0 {
			t.Errorf("unresolved = %v, want none for malformed input", unresolved)
		}
		found := false
		for _, err := range errs {
			if strings.Contains(err.Error(), "malformed amendment") {
				found = true
			}
		}
		if !found {
			t.Errorf("expected malformed-amendment error, got %v", errs)
		}
	})

	t.Run("final amendment New digest must match the current contract", func(t *testing.T) {
		comments := []string{
			amendmentRecord(1, "First unit", oldDigest, amendmentFakeDigest, "digest invented out of thin air"),
		}
		_, errs := unresolvedRebuildRequests(newUnits, comments)
		found := false
		for _, err := range errs {
			if strings.Contains(err.Error(), "current contract digest") && strings.Contains(err.Error(), amendmentFakeDigest) {
				found = true
			}
		}
		if !found {
			t.Errorf("expected New-digest mismatch error, got %v", errs)
		}
	})

	t.Run("stale receipt that no amendment claims stays an error", func(t *testing.T) {
		otherDigest := "4444444444444444444444444444444444444444444444444444444444444444"
		comments := []string{
			amendmentReceipt(1, "Nonexistent unit", otherDigest),
		}
		_, errs := unresolvedRebuildRequests(newUnits, comments)
		found := false
		for _, err := range errs {
			if strings.Contains(err.Error(), "does not identify exactly one queue unit") {
				found = true
			}
		}
		if !found {
			t.Errorf("expected unresolvable-identity error, got %v", errs)
		}
	})
}

func TestSuccessiveRenamedAmendmentsWithChangedVerify(t *testing.T) {
	oldUnits := amendmentTestUnits(t, amendmentTestBodyWithVerify("old outcome", "Old/X", "echo old"))
	middleUnits := amendmentTestUnits(t, amendmentTestBodyWithVerify("middle outcome", "Middle/Y", "echo middle"))
	finalUnits := amendmentTestUnits(t, amendmentTestBodyWithVerify("final outcome", "Final/Z", "echo final"))
	oldDigest := unitContractDigest(oldUnits[0])
	middleDigest := unitContractDigest(middleUnits[0])
	finalDigest := unitContractDigest(finalUnits[0])
	comments := []string{
		amendmentReceiptWithVerify(1, "Old/X", oldDigest, "echo old"),
		amendmentRecord(1, "Middle/Y", oldDigest, middleDigest, "rename the first revision"),
		amendmentReceiptWithVerify(1, "Middle/Y", middleDigest, "echo middle"),
		amendmentRecord(1, "Final/Z", middleDigest, finalDigest, "rename the final revision"),
		amendmentReceiptWithVerify(1, "Final/Z", finalDigest, "echo final"),
	}

	unresolved, errs := unresolvedRebuildRequests(finalUnits, comments)
	if len(errs) != 0 {
		t.Fatalf("successive renamed amendments returned errors: %v", errs)
	}
	if len(unresolved) != 0 {
		t.Fatalf("successive renamed amendments unresolved = %v, want none", unresolved)
	}

	withoutIntermediateReceipt := []string{comments[0], comments[1], comments[3], comments[4]}
	unresolved, errs = unresolvedRebuildRequests(finalUnits, withoutIntermediateReceipt)
	if len(errs) != 0 {
		t.Fatalf("successive renamed amendments without intermediate receipt returned errors: %v", errs)
	}
	if len(unresolved) != 0 {
		t.Fatalf("successive renamed amendments without intermediate receipt unresolved = %v, want none", unresolved)
	}
}

func TestSuccessiveRenamedAmendmentsWithChangedVerifyChunks(t *testing.T) {
	oldUnits := amendmentTestUnits(t, amendmentTestBodyWithVerify("old outcome", "Old/X", "echo old"))
	middleUnits := amendmentTestUnits(t, amendmentTestBodyWithVerify("middle outcome", "Middle/Y", "echo middle"))
	finalUnits := amendmentTestUnits(t, amendmentTestBodyWithVerify("final outcome", "Final/Z", "echo final"))
	oldDigest := unitContractDigest(oldUnits[0])
	middleDigest := unitContractDigest(middleUnits[0])
	finalDigest := unitContractDigest(finalUnits[0])
	chunked := func(heading, verify, digest string) []string {
		return receiptPartsForIdentity(chunkedReceiptComments, verify, heading, digest)
	}
	comments := chunked("Old/X", "echo old", oldDigest)
	comments = append(comments, amendmentRecord(1, "Middle/Y", oldDigest, middleDigest, "rename the first revision"))
	comments = append(comments, chunked("Middle/Y", "echo middle", middleDigest)...)
	comments = append(comments, amendmentRecord(1, "Final/Z", middleDigest, finalDigest, "rename the final revision"))
	comments = append(comments, chunked("Final/Z", "echo final", finalDigest)...)

	unresolved, errs := unresolvedRebuildRequests(finalUnits, comments)
	if len(errs) != 0 {
		t.Fatalf("successive renamed chunk receipts returned errors: %v", errs)
	}
	if len(unresolved) != 0 {
		t.Fatalf("successive renamed chunk receipts unresolved = %v, want none", unresolved)
	}
}

func TestSuccessiveRenamedAmendmentsRequireExactIntermediateIdentity(t *testing.T) {
	oldUnits := amendmentTestUnits(t, amendmentTestBodyWithVerify("old outcome", "Old/X", "echo old"))
	middleUnits := amendmentTestUnits(t, amendmentTestBodyWithVerify("middle outcome", "Middle/Y", "echo middle"))
	finalUnits := amendmentTestUnits(t, amendmentTestBodyWithVerify("final outcome", "Final/Z", "echo final"))
	oldDigest := unitContractDigest(oldUnits[0])
	middleDigest := unitContractDigest(middleUnits[0])
	finalDigest := unitContractDigest(finalUnits[0])
	base := []string{
		amendmentReceiptWithVerify(1, "Old/X", oldDigest, "echo old"),
		amendmentRecord(1, "Middle/Y", oldDigest, middleDigest, "rename the first revision"),
		amendmentReceiptWithVerify(1, "Middle/Y", middleDigest, "echo middle"),
		amendmentRecord(1, "Final/Z", middleDigest, finalDigest, "rename the final revision"),
		amendmentReceiptWithVerify(1, "Final/Z", finalDigest, "echo final"),
	}

	for _, test := range []struct {
		name string
		edit func([]string) []string
	}{
		{
			name: "intermediate amendment heading",
			edit: func(comments []string) []string {
				comments[1] = strings.Replace(comments[1], "Unit heading: Middle/Y", "Unit heading: Wrong/Y", 1)
				return comments
			},
		},
		{
			name: "intermediate heading",
			edit: func(comments []string) []string {
				comments[2] = strings.Replace(comments[2], "Unit heading: Middle/Y", "Unit heading: Wrong/Y", 1)
				return comments
			},
		},
		{
			name: "intermediate amendment occurrence",
			edit: func(comments []string) []string {
				comments[1] = strings.Replace(comments[1], "Unit occurrence: 1", "Unit occurrence: 2", 1)
				return comments
			},
		},
		{
			name: "intermediate occurrence",
			edit: func(comments []string) []string {
				comments[2] = strings.Replace(comments[2], "Unit occurrence: 1", "Unit occurrence: 2", 1)
				return comments
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			comments := append([]string(nil), base...)
			_, errs := unresolvedRebuildRequests(finalUnits, test.edit(comments))
			if len(errs) == 0 {
				t.Fatalf("expected exact intermediate identity failure")
			}
		})
	}
}

func TestAmendmentHistoryRejectsBranchingAndDisconnectedSuccessors(t *testing.T) {
	oldUnits := amendmentTestUnits(t, amendmentTestBodyWithVerify("old outcome", "Old/X", "echo old"))
	middleUnits := amendmentTestUnits(t, amendmentTestBodyWithVerify("middle outcome", "Middle/Y", "echo middle"))
	finalUnits := amendmentTestUnits(t, amendmentTestBodyWithVerify("final outcome", "Final/Z", "echo final"))
	oldDigest := unitContractDigest(oldUnits[0])
	middleDigest := unitContractDigest(middleUnits[0])
	finalDigest := unitContractDigest(finalUnits[0])
	oldReceipt := amendmentReceiptWithVerify(1, "Old/X", oldDigest, "echo old")
	finalReceipt := amendmentReceiptWithVerify(1, "Final/Z", finalDigest, "echo final")

	t.Run("missing transition is disconnected", func(t *testing.T) {
		comments := []string{
			oldReceipt,
			amendmentRecord(1, "Final/Z", middleDigest, finalDigest, "skip the missing revision"),
			finalReceipt,
		}
		_, errs := unresolvedRebuildRequests(finalUnits, comments)
		if len(errs) == 0 {
			t.Fatal("expected disconnected amendment history to fail")
		}
	})

	t.Run("two successors branch the history", func(t *testing.T) {
		comments := []string{
			oldReceipt,
			amendmentRecord(1, "Middle/Y", oldDigest, middleDigest, "first path"),
			amendmentRecord(1, "Final/Z", middleDigest, finalDigest, "valid path"),
			amendmentRecord(1, "Final/Z", middleDigest, amendmentFakeDigest, "forged path"),
			finalReceipt,
		}
		_, errs := unresolvedRebuildRequests(finalUnits, comments)
		found := false
		for _, err := range errs {
			if strings.Contains(err.Error(), "branching amendment history") {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected branching amendment history error, got %v", errs)
		}
	})
}

func amendmentTestLocalBody(doneMeans, digest string) string {
	base := amendmentTestBody(doneMeans, "First unit")
	lines := strings.Split(base, "\n")
	evidence := strings.SplitN(amendmentReceipt(1, "First unit", digest), "\n", 3)[2]
	out := append([]string{}, lines[:len(lines)-1]...)
	out = append(out, "", evidence, lines[len(lines)-1])
	return strings.Join(out, "\n")
}

func TestLocalQueueAmendmentBlockUsesSameGrammar(t *testing.T) {
	newUnits := amendmentTestUnits(t, amendmentTestBody(amendmentTestNewDoneMeans, "First unit"))
	newDigest := unitContractDigest(newUnits[0])
	oldDigest := unitContractDigest(amendmentTestUnits(t, amendmentTestBody(amendmentTestOldDoneMeans, "First unit"))[0])

	fileBody := amendmentTestLocalBody(amendmentTestNewDoneMeans, newDigest) + "\n\n" +
		amendmentRecord(1, "First unit", oldDigest, newDigest, "local lane witness") + "\n"

	dir := t.TempDir()
	path := filepath.Join(dir, "demo.md")
	if err := os.WriteFile(path, []byte(fileBody), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := ValidateQueueFile(path)
	if err != nil {
		t.Fatal(err)
	}
	foundUnresolved := false
	foundMismatch := false
	for _, iss := range result.Errors {
		if strings.Contains(iss.Message, "unresolved rebuild request") {
			foundUnresolved = true
		}
		if strings.Contains(iss.Message, "unit digest mismatch") {
			foundMismatch = true
		}
	}
	if !foundUnresolved {
		t.Errorf("expected unresolved-rebuild-request error from local amendment block; errors: %v", result.Errors)
	}
	if foundMismatch {
		t.Errorf("body receipt matches current contract; unexpected digest mismatch: %v", result.Errors)
	}
}

func replanMarker(occurrence int, heading, digest string) string {
	return strings.Join([]string{
		"Re-plan required:",
		fmt.Sprintf("Unit occurrence: %d", occurrence),
		"Unit heading: " + heading,
		"Unit digest: " + digest,
		"Reason: repeated review found the contract shape inadequate",
	}, "\n")
}

func TestReplanMarkerResolvedByAmendment(t *testing.T) {
	identity := queueUnitIdentity{Occurrence: 1, Heading: "First unit"}
	oldUnits := amendmentTestUnits(t, amendmentTestBody(amendmentTestOldDoneMeans, identity.Heading))
	newUnits := amendmentTestUnits(t, amendmentTestBody(amendmentTestNewDoneMeans, identity.Heading))
	oldDigest := unitContractDigest(oldUnits[0])
	newDigest := unitContractDigest(newUnits[0])
	completedCycles := []string{
		amendmentReceipt(identity.Occurrence, identity.Heading, oldDigest),
		formatRebuildRequest(identity),
		amendmentReceipt(identity.Occurrence, identity.Heading, oldDigest),
		formatRebuildRequest(identity),
		amendmentReceipt(identity.Occurrence, identity.Heading, oldDigest),
	}
	marked := append(append([]string{}, completedCycles...), replanMarker(identity.Occurrence, identity.Heading, oldDigest))

	result := &ValidationResult{Valid: true}
	applyQueueIssues(result, "comments", oldUnits, nil, marked)
	foundMarker := false
	for _, issue := range result.Errors {
		if strings.Contains(issue.Message, "unresolved re-plan marker") {
			foundMarker = true
		}
	}
	if !foundMarker {
		t.Errorf("expected unresolved re-plan marker, got %v", result.Errors)
	}
	selectable, errs := selectableUnitIdentities(oldUnits, marked)
	if len(errs) != 0 {
		t.Fatalf("marked contract returned selection errors: %v", errs)
	}
	if len(selectable) != 0 {
		t.Errorf("marked contract selectable = %v, want none", selectable)
	}

	amended := append(append([]string{}, marked...), amendmentRecord(
		identity.Occurrence,
		identity.Heading,
		oldDigest,
		newDigest,
		"split the repeated failure policy",
	))
	unresolved, errs := unresolvedRebuildRequests(newUnits, amended)
	if len(errs) != 0 {
		t.Fatalf("amendment returned errors: %v", errs)
	}
	if len(unresolved) != 1 || unresolved[0] != identity {
		t.Errorf("amendment unresolved = %v, want the amended unit", unresolved)
	}
	selectable, errs = selectableUnitIdentities(newUnits, amended)
	if len(errs) != 0 {
		t.Fatalf("amended contract returned selection errors: %v", errs)
	}
	if len(selectable) != 1 || selectable[0] != identity {
		t.Errorf("amended contract selectable = %v, want %v", selectable, identity)
	}

	resolved := append(append([]string{}, amended...), amendmentReceipt(identity.Occurrence, identity.Heading, newDigest))
	scan := scanQueueComments(newUnits, resolved)
	if len(scan.errors) != 0 {
		t.Fatalf("fresh evidence returned errors: %v", scan.errors)
	}
	if len(scan.unresolved) != 0 {
		t.Errorf("fresh evidence unresolved = %v, want none", scan.unresolved)
	}
	if got := scan.completedRebuildCycles[identity][newDigest]; got != 0 {
		t.Errorf("new digest completed rebuild cycles = %d, want 0", got)
	}
}

func TestQueueReplanMarkerValidation(t *testing.T) {
	identity := queueUnitIdentity{Occurrence: 1, Heading: "First unit"}
	units := amendmentTestUnits(t, amendmentTestBody(amendmentTestOldDoneMeans, identity.Heading))
	digest := unitContractDigest(units[0])
	evidence := amendmentReceipt(identity.Occurrence, identity.Heading, digest)
	completedCycles := []string{
		evidence,
		formatRebuildRequest(identity),
		evidence,
		formatRebuildRequest(identity),
		evidence,
	}
	marker := replanMarker(identity.Occurrence, identity.Heading, digest)
	hasError := func(comments []string, fragment string) bool {
		t.Helper()
		for _, err := range scanQueueComments(units, comments).errors {
			if strings.Contains(err.Error(), fragment) {
				return true
			}
		}
		return false
	}

	if !hasError([]string{evidence, marker}, "requires two completed rebuild cycles") {
		t.Error("marker before two completed cycles was accepted")
	}
	if !hasError(append(append([]string{}, completedCycles...), marker, marker), "duplicate unresolved re-plan marker") {
		t.Error("duplicate unresolved marker was accepted")
	}
	malformed := strings.Replace(marker, "Unit digest: "+digest, "Unit digest: bad", 1)
	if !hasError(append(append([]string{}, completedCycles...), malformed), "malformed re-plan marker") {
		t.Error("malformed marker was accepted")
	}
	wrongDigest := strings.Repeat("9", 64)
	if !hasError(append(append([]string{}, completedCycles...), replanMarker(identity.Occurrence, identity.Heading, wrongDigest)), "outside its contract history") {
		t.Error("marker for an unrelated digest was accepted")
	}
}

func TestLocalQueueReplanMarkerMetadata(t *testing.T) {
	body := amendmentTestBody(amendmentTestOldDoneMeans, "First unit")
	units := amendmentTestUnits(t, body)
	marker := replanMarker(1, "First unit", unitContractDigest(units[0]))
	stripped, metadata := splitLocalQueueMetadataBlocks(body + "\n\n" + marker + "\n")
	if strings.Contains(stripped, "Re-plan required:") {
		t.Error("local re-plan marker remained in the queue contract body")
	}
	if len(metadata) != 1 || metadata[0] != marker {
		t.Errorf("local metadata = %q, want the exact marker", metadata)
	}
}
