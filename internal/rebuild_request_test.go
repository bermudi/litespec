package internal

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

type rebuildRoutingFixture struct {
	Units []struct {
		Heading string `json:"heading"`
		Checked bool   `json:"checked"`
	} `json:"units"`
	Events []struct {
		Kind       string `json:"kind"`
		Occurrence int    `json:"occurrence"`
		Heading    string `json:"heading"`
		Body       string `json:"body"`
		Complete   *bool  `json:"complete"`
	} `json:"events"`
	Unresolved []queueUnitIdentity `json:"unresolved"`
	Selectable []queueUnitIdentity `json:"selectable"`
	Errors     int                 `json:"errors"`
}

func TestGeneratedReviewRoutesRebuildRequests(t *testing.T) {
	root := t.TempDir()
	if err := GenerateSkills(root); err != nil {
		t.Fatalf("GenerateSkills: %v", err)
	}

	review, err := os.ReadFile(filepath.Join(root, SkillsDir, "litespec-review", "SKILL.md"))
	if err != nil {
		t.Fatalf("read generated review skill: %v", err)
	}
	build, err := os.ReadFile(filepath.Join(root, SkillsDir, "litespec-build", "SKILL.md"))
	if err != nil {
		t.Fatalf("read generated build skill: %v", err)
	}
	for name, content := range map[string]string{"review": string(review), "build": string(build)} {
		for _, obsolete := range []string{"gh issue edit --body-file", "edit the issue body remotely"} {
			if strings.Contains(content, obsolete) {
				t.Errorf("%s retains obsolete GitHub body-edit behavior %q", name, obsolete)
			}
		}
	}
	for _, required := range []string{
		"post exactly one separate comment",
		"Unit occurrence: <positive 1-based occurrence>",
		"Unit heading: <exact heading>",
		"body to be byte-for-byte unchanged",
		"create a separate clean routing metadata commit",
	} {
		if !strings.Contains(string(review), required) {
			t.Errorf("review missing append-only routing rule %q", required)
		}
	}
	for _, required := range []string{
		"checked unit is also selectable when its latest request state is unresolved",
		"One later complete receipt resolves all earlier requests for that identity",
		"leave the issue body and prior comments unchanged",
	} {
		if !strings.Contains(string(build), required) {
			t.Errorf("build missing request-selection rule %q", required)
		}
	}
	const closureRule = "The issue closes only when every unit checkbox is checked, every rebuild request is resolved, and review returns `PASS`."
	for _, path := range []string{"../AGENTS.md", "../docs/project-structure.md"} {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read closure documentation %s: %v", path, err)
		}
		if !strings.Contains(string(content), closureRule) {
			t.Errorf("%s missing closure rule %q", path, closureRule)
		}
	}

	var fixture rebuildRoutingFixture
	data, err := os.ReadFile("testdata/rebuild-routing/github.json")
	if err != nil {
		t.Fatalf("read GitHub fixture: %v", err)
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatalf("decode GitHub fixture: %v", err)
	}

	units := make([]queueUnit, 0, len(fixture.Units))
	for _, unit := range fixture.Units {
		status := "- [ ] pending"
		if unit.Checked {
			status = "- [x] done"
		}
		units = append(units, queueUnit{
			Heading: unit.Heading,
			Body: []string{
				"Done means:",
				"- [outcome] outcome",
				"Scenarios:",
				"- [outcome] TestOutcome",
				"Verify: `go test ./...`",
				status,
			},
		})
	}
	comments := make([]string, 0, len(fixture.Events))
	for _, event := range fixture.Events {
		identity := queueUnitIdentity{Occurrence: event.Occurrence, Heading: event.Heading}
		switch event.Kind {
		case "request":
			comments = append(comments, formatRebuildRequest(identity))
		case "evidence":
			complete := event.Complete == nil || *event.Complete
			comments = append(comments, fixtureEvidenceComment(identity, complete))
		case "malformed":
			comments = append(comments, event.Body)
		default:
			t.Fatalf("unknown fixture event kind %q", event.Kind)
		}
	}

	unresolved, associationErrors := unresolvedRebuildRequests(units, comments)
	if !reflect.DeepEqual(unresolved, fixture.Unresolved) {
		t.Errorf("unresolved requests = %#v, want %#v", unresolved, fixture.Unresolved)
	}
	if len(associationErrors) != fixture.Errors {
		t.Errorf("association errors = %d, want %d: %v", len(associationErrors), fixture.Errors, associationErrors)
	}
	selectable, selectionErrors := selectableUnitIdentities(units, comments)
	if !reflect.DeepEqual(selectable, fixture.Selectable) {
		t.Errorf("selectable units = %#v, want %#v", selectable, fixture.Selectable)
	}
	if len(selectionErrors) != fixture.Errors {
		t.Errorf("selection errors = %d, want %d: %v", len(selectionErrors), fixture.Errors, selectionErrors)
	}

	const queueBody = "Base: 1111111111111111111111111111111111111111\n" +
		"Branch: litespec/rebuild-fixture\n\n" +
		"## Duplicate\nDone means:\n- [outcome] outcome\nScenarios:\n- [outcome] TestOutcome\nVerify: `go test ./...`\n- [x] done\n"
	receiptUnits, receiptIssues := ValidateQueueBody(queueBody, "fixture")
	receiptResult := &ValidationResult{Valid: true}
	applyQueueIssues(
		receiptResult,
		"GitHub comments",
		receiptUnits,
		receiptIssues,
		[]string{fixtureEvidenceComment(queueUnitIdentity{Occurrence: 1, Heading: "Duplicate"}, true)},
	)
	if len(receiptResult.Errors) != 0 {
		t.Errorf("identity-bearing receipt did not satisfy queue validation: %v", receiptResult.Errors)
	}

	local, err := os.ReadFile("testdata/rebuild-routing/local.md")
	if err != nil {
		t.Fatalf("read local fixture: %v", err)
	}
	affected := []queueUnitIdentity{
		{Occurrence: 1, Heading: "Affected"},
		{Occurrence: 2, Heading: "Duplicate"},
	}
	var persisted string
	updated, err := persistLocalRebuildRouting(string(local), affected, func(body string) error {
		persisted = body
		return nil
	})
	if err != nil {
		t.Fatalf("persist local routing: %v", err)
	}
	if updated != persisted {
		t.Error("returned local queue differs from persisted queue")
	}
	if !strings.Contains(updated, "old receipt remains byte-for-byte\n- [x] done") {
		t.Error("local routing changed prior evidence or unaffected checked status")
	}
	if strings.Count(updated, "- [ ] done") != 2 || strings.Count(updated, "- [x] done") != 2 {
		t.Errorf("local routing changed the wrong statuses:\n%s", updated)
	}

	persistErr := errors.New("metadata commit failed")
	failedBody, err := persistLocalRebuildRouting(string(local), affected, func(string) error {
		return persistErr
	})
	if !errors.Is(err, persistErr) {
		t.Fatalf("local persistence error = %v, want %v", err, persistErr)
	}
	if failedBody != string(local) {
		t.Error("failed local persistence exposed an uncommitted mutation")
	}
}

func TestCompletedRebuildCyclesPerDigest(t *testing.T) {
	identity := queueUnitIdentity{Occurrence: 1, Heading: "Duplicate"}
	units := []queueUnit{{
		Heading: identity.Heading,
		Body: []string{
			"Done means:",
			"- [outcome] outcome",
			"Scenarios:",
			"- [outcome] TestOutcome",
			"Verify: `go test ./...`",
			"- [x] done",
		},
	}}
	digest := unitContractDigest(units[0])
	comments := []string{
		formatRebuildRequest(identity),
		formatRebuildRequest(identity),
		fixtureEvidenceComment(identity, true),
		formatRebuildRequest(identity),
		fixtureEvidenceComment(identity, true),
	}

	scan := scanQueueComments(units, comments)
	if len(scan.errors) != 0 {
		t.Fatalf("unexpected errors: %v", scan.errors)
	}
	if got := scan.completedRebuildCycles[identity][digest]; got != 2 {
		t.Errorf("completed rebuild cycles = %d, want 2", got)
	}
	if len(scan.unresolved) != 0 {
		t.Errorf("unresolved = %v, want none", scan.unresolved)
	}
}

func TestThirdRebuildRequestRequiresReplan(t *testing.T) {
	identity := queueUnitIdentity{Occurrence: 1, Heading: "Duplicate"}
	const queueBody = "Base: 1111111111111111111111111111111111111111\n" +
		"Branch: litespec/rebuild-fixture\n\n" +
		"## Duplicate\nDone means:\n- [outcome] outcome\nScenarios:\n- [outcome] TestOutcome\nVerify: `go test ./...`\n- [x] done\n"
	units, unitIssues := ValidateQueueBody(queueBody, "fixture")
	comments := []string{
		formatRebuildRequest(identity),
		fixtureEvidenceComment(identity, true),
		formatRebuildRequest(identity),
		fixtureEvidenceComment(identity, true),
		formatRebuildRequest(identity),
	}

	result := &ValidationResult{Valid: true}
	applyQueueIssues(result, "GitHub comments", units, unitIssues, comments)
	foundRequiresReplan := false
	for _, issue := range result.Errors {
		if strings.Contains(issue.Message, "requires re-planning") {
			foundRequiresReplan = true
		}
		if strings.Contains(issue.Message, "unresolved rebuild request") {
			t.Errorf("third request must not leave the unit rebuildable: %s", issue.Message)
		}
	}
	if !foundRequiresReplan {
		t.Errorf("expected third request validation error requiring re-planning, got %v", result.Errors)
	}

	selectable, errs := selectableUnitIdentities(units, comments)
	if len(errs) == 0 {
		t.Fatal("expected third request selection error")
	}
	if len(selectable) != 0 {
		t.Errorf("selectable = %v, want none after a third request", selectable)
	}
}

func fixtureEvidenceComment(identity queueUnitIdentity, complete bool) string {
	contractUnit := queueUnit{
		Heading: identity.Heading,
		Body: []string{
			"Done means:",
			"- [outcome] outcome",
			"Scenarios:",
			"- [outcome] TestOutcome",
			"Verify: `go test ./...`",
		},
	}
	body := "Unit occurrence: " + strconv.Itoa(identity.Occurrence) +
		"\nUnit heading: " + identity.Heading +
		"\nEvidence:\ngo test ./...\n" +
		"unit digest: " + unitContractDigest(contractUnit) + "\n" +
		"pre sha: 1111111111111111111111111111111111111111\n" +
		"pre exit status: 1\n```\nFAIL\n```\n" +
		"Pre-evidence scope: this command exited 1 at 1111111111111111111111111111111111111111; nothing else is inferred.\n" +
		"post sha: 2222222222222222222222222222222222222222\n" +
		"post exit status: 0\n```\nPASS\n```\n" +
		"Post-evidence scope: this command exited 0 at 2222222222222222222222222222222222222222; nothing else is inferred."
	if !complete {
		return strings.Replace(body, "post exit status: 0", "post exit status: 1", 1)
	}
	return body
}
