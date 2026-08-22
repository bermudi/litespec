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
			Body:    []string{"Done means: outcome", "Verify: `go test ./...`", status},
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
	if strings.Count(updated, "- [ ] pending rebuild") != 2 || strings.Count(updated, "- [x] done") != 2 {
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

func fixtureEvidenceComment(identity queueUnitIdentity, complete bool) string {
	body := "Unit occurrence: " + strconv.Itoa(identity.Occurrence) +
		"\nUnit heading: " + identity.Heading +
		"\nEvidence:\ngo test ./...\n" +
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
