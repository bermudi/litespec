package internal

import (
	"strings"
	"testing"
)

func TestBoundaryVocabulary(t *testing.T) {
	source := "GH issue #1"
	completeRisks := "Risk cases:\n- timeout: [timeout]\n- cleanup: [cleanup]\n- non-ENOENT errors: N/A — no filesystem lookup\n- concurrency: N/A — each probe owns its state\n- optional configured dependencies: N/A — the probe is mandatory\n"
	unit := func(boundary string) string {
		body := "## Probe service\n"
		if boundary != "" {
			body += "Boundary: " + boundary + "\n"
		}
		body += "Done means:\n- [timeout] returns on timeout\n- [cleanup] removes temporary state\nScenarios:\n- [timeout] TestTimeout\n- [cleanup] TestCleanup\n"
		if boundary == "filesystem" || boundary == "process" || boundary == "network" {
			body += completeRisks
		}
		body += "Verify: `go test ./internal -run TestProbeService`\n- [ ] pending\n"
		return ownedQueue(body)
	}

	for _, boundary := range []string{"filesystem", "process", "network"} {
		t.Run("valid "+boundary+" passes", func(t *testing.T) {
			_, issues := ValidateQueueBody(unit(boundary), source)
			if len(issues) > 0 {
				t.Fatalf("expected Boundary: %s to pass, got %v", boundary, issues)
			}
		})
	}

	t.Run("omitted boundary passes", func(t *testing.T) {
		_, issues := ValidateQueueBody(unit(""), source)
		if len(issues) > 0 {
			t.Fatalf("expected omitted Boundary to pass, got %v", issues)
		}
	})

	for _, boundary := range []string{"Filesystem", "Process", "Network", "database", "fs"} {
		t.Run("unknown "+boundary+" fails", func(t *testing.T) {
			_, issues := ValidateQueueBody(unit(boundary), source)
			if !containsIssue(issues, `Boundary: must be one of "filesystem", "process", "network"`) {
				t.Fatalf("expected vocabulary error for Boundary: %s, got %v", boundary, issues)
			}
		})
	}
}

func TestBoundaryVocabularyAffectsUnitDigest(t *testing.T) {
	unit := queueUnit{
		Heading: "Probe service",
		Body: []string{
			"Boundary: filesystem",
			"Done means:",
			"- [timeout] returns on timeout",
			"Scenarios:",
			"- [timeout] TestTimeout",
			"Verify: `go test ./internal -run TestTimeout`",
		},
	}
	changedBoundary := unit
	changedBoundary.Body = append([]string(nil), unit.Body...)
	changedBoundary.Body[0] = "Boundary: process"

	baseDigest := unitContractDigest(unit)
	if got := unitContractDigest(changedBoundary); got == baseDigest {
		t.Fatal("changing the boundary value did not change the unit digest")
	}
	if !strings.HasPrefix(baseDigest, "") {
		t.Fatal("unreachable")
	}
}
