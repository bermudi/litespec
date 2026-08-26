package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const digestTestFence = "```"

func digestTestBody() string {
	return strings.Join([]string{
		"Base: 0000000000000000000000000000000000000001",
		"Branch: litespec/demo",
		"",
		"## First unit",
		"",
		"Done means: it works",
		"",
		"Verify:",
		digestTestFence + "bash",
		"echo first",
		digestTestFence,
		"",
		"- [ ] pending",
		"",
		"## Second unit",
		"",
		"Read first: some context",
		"Depends: First unit",
		"Done means: it also works",
		"",
		"Verify:",
		digestTestFence + "bash",
		"echo second",
		digestTestFence,
		"",
		"- [ ] pending",
	}, "\n")
}

func digestTestUnits(t *testing.T) []queueUnit {
	t.Helper()
	var units []queueUnit
	for _, section := range parseQueueUnits(digestTestBody()) {
		if isUnit(section) {
			units = append(units, section)
		}
	}
	if len(units) != 2 {
		t.Fatalf("fixture should yield 2 units, got %d", len(units))
	}
	return units
}

func TestDigestQueueUnitsLocalFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "demo.md")
	if err := os.WriteFile(path, []byte(digestTestBody()), 0o644); err != nil {
		t.Fatal(err)
	}

	lines, err := DigestQueueUnits("", 0, path)
	if err != nil {
		t.Fatal(err)
	}
	units := digestTestUnits(t)
	identities := queueUnitIdentities(units)
	want := make([]UnitDigestLine, len(units))
	for i, unit := range units {
		want[i] = UnitDigestLine{
			Occurrence: identities[i].Occurrence,
			Heading:    unit.Heading,
			Digest:     unitContractDigest(unit),
		}
	}

	if len(lines) != len(want) {
		t.Fatalf("got %d lines, want %d", len(lines), len(want))
	}
	for i := range want {
		if lines[i] != want[i] {
			t.Errorf("line %d = %+v, want %+v", i, lines[i], want[i])
		}
	}
}

func TestDigestQueueUnitsGHIssueMatchesLocalCanonicalization(t *testing.T) {
	payload, err := json.Marshal(ghIssue{
		Number: 42,
		Title:  "t",
		Body:   digestTestBody(),
	})
	if err != nil {
		t.Fatal(err)
	}

	oldView := ghIssueView
	oldLook := lookPathGh
	ghIssueView = func(string, int) ([]byte, error) { return payload, nil }
	lookPathGh = func(string) (string, error) { return "/usr/bin/gh", nil }
	defer func() {
		ghIssueView = oldView
		lookPathGh = oldLook
	}()

	localPath := filepath.Join(t.TempDir(), "demo.md")
	if err := os.WriteFile(localPath, []byte(digestTestBody()), 0o644); err != nil {
		t.Fatal(err)
	}
	viaIssue, err := DigestQueueUnits("", 42, "")
	if err != nil {
		t.Fatal(err)
	}
	viaFile, err := DigestQueueUnits("", 0, localPath)
	if err != nil {
		t.Fatal(err)
	}

	if len(viaIssue) != len(viaFile) || len(viaIssue) != 2 {
		t.Fatalf("expected 2 identical digest lines per lane, got issue=%d file=%d", len(viaIssue), len(viaFile))
	}
	for i := range viaIssue {
		if viaIssue[i] != viaFile[i] {
			t.Errorf("lane mismatch at line %d: %+v vs %+v", i, viaIssue[i], viaFile[i])
		}
	}
}

func TestFormatUnitDigestLinesIsTSV(t *testing.T) {
	out := FormatUnitDigestLines([]UnitDigestLine{
		{Occurrence: 1, Heading: "First unit", Digest: "abc123"},
		{Occurrence: 2, Heading: "Second unit", Digest: "def456"},
	})
	want := "1\tFirst unit\tabc123\n2\tSecond unit\tdef456\n"
	if out != want {
		t.Errorf("formatted = %q, want %q", out, want)
	}
}
