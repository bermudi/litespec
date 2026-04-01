package internal

import (
	"testing"
)

func TestParseTasksMDBasic(t *testing.T) {
	content := `## Phase 1: Setup

- [x] Create directory structure
- [ ] Write config parser

## Phase 2: Implementation

- [ ] Build core logic
- [ ] Add error handling

## Phase 3: Tests

- [ ] Unit tests
`
	phases := parseTasksMD(content)
	if len(phases) != 3 {
		t.Fatalf("Phases count = %d, want 3", len(phases))
	}
	if phases[0].Name != "Phase 1: Setup" {
		t.Errorf("Phase[0].Name = %q", phases[0].Name)
	}
	if phases[0].Total != 2 {
		t.Errorf("Phase[0].Total = %d, want 2", phases[0].Total)
	}
	if phases[0].Complete != 1 {
		t.Errorf("Phase[0].Complete = %d, want 1", phases[0].Complete)
	}
	if phases[1].Total != 2 {
		t.Errorf("Phase[1].Total = %d, want 2", phases[1].Total)
	}
	if phases[1].Complete != 0 {
		t.Errorf("Phase[1].Complete = %d, want 0", phases[1].Complete)
	}
	if phases[2].Total != 1 {
		t.Errorf("Phase[2].Total = %d, want 1", phases[2].Total)
	}
}

func TestParseTasksMDTaskIDsAndState(t *testing.T) {
	content := `## Phase 1: Setup

- [x] Task one
- [ ] Task two
`
	phases := parseTasksMD(content)
	if len(phases) != 1 {
		t.Fatalf("Phases count = %d, want 1", len(phases))
	}
	tasks := phases[0].Tasks
	if len(tasks) != 2 {
		t.Fatalf("Tasks count = %d, want 2", len(tasks))
	}
	if tasks[0].ID != "Phase 1: Setup-1" {
		t.Errorf("Task[0].ID = %q", tasks[0].ID)
	}
	if !tasks[0].Done {
		t.Error("Task[0].Done = false, want true")
	}
	if tasks[0].Description != "Task one" {
		t.Errorf("Task[0].Description = %q", tasks[0].Description)
	}
	if tasks[1].Done {
		t.Error("Task[1].Done = true, want false")
	}
	if tasks[1].Description != "Task two" {
		t.Errorf("Task[1].Description = %q", tasks[1].Description)
	}
}

func TestParseTasksMDEmpty(t *testing.T) {
	phases := parseTasksMD("")
	if len(phases) != 0 {
		t.Errorf("Phases count = %d, want 0", len(phases))
	}
}

func TestParseTasksMDNoPhaseHeadings(t *testing.T) {
	content := `- [x] Some orphan task
- [ ] Another orphan task
`
	phases := parseTasksMD(content)
	if len(phases) != 0 {
		t.Errorf("Phases count = %d, want 0 (no phase headings)", len(phases))
	}
}

func TestParseTasksMDIgnoresNonTaskLines(t *testing.T) {
	content := `## Phase 1: Setup

Some description text here.
Another line of description.

- [x] Actual task
More text that is not a task.
- [ ] Second task
`
	phases := parseTasksMD(content)
	if len(phases) != 1 {
		t.Fatalf("Phases count = %d, want 1", len(phases))
	}
	if phases[0].Total != 2 {
		t.Errorf("Phase[0].Total = %d, want 2", phases[0].Total)
	}
}

func TestFindCurrentPhaseFirstIncomplete(t *testing.T) {
	phases := []PhaseJSON{
		{Name: "Phase 1", Tasks: []TaskItemJSON{
			{Done: true}, {Done: true},
		}, Complete: 2, Total: 2},
		{Name: "Phase 2", Tasks: []TaskItemJSON{
			{Done: true}, {Done: false},
		}, Complete: 1, Total: 2},
		{Name: "Phase 3", Tasks: []TaskItemJSON{
			{Done: false},
		}, Complete: 0, Total: 1},
	}

	idx := findCurrentPhase(phases)
	if idx != 1 {
		t.Errorf("currentPhase = %d, want 1 (first phase with unchecked task)", idx)
	}
}

func TestFindCurrentPhaseFirstPhaseActive(t *testing.T) {
	phases := []PhaseJSON{
		{Name: "Phase 1", Tasks: []TaskItemJSON{
			{Done: false}, {Done: false},
		}, Complete: 0, Total: 2},
		{Name: "Phase 2", Tasks: []TaskItemJSON{
			{Done: false},
		}, Complete: 0, Total: 1},
	}

	idx := findCurrentPhase(phases)
	if idx != 0 {
		t.Errorf("currentPhase = %d, want 0 (first phase has unchecked tasks)", idx)
	}
}

func TestFindCurrentPhaseAllDone(t *testing.T) {
	phases := []PhaseJSON{
		{Name: "Phase 1", Tasks: []TaskItemJSON{
			{Done: true},
		}, Complete: 1, Total: 1},
	}

	idx := findCurrentPhase(phases)
	if idx != 0 {
		t.Errorf("currentPhase = %d, want 0 (all done, defaults to first)", idx)
	}
}

func TestFindCurrentPhaseEmpty(t *testing.T) {
	idx := findCurrentPhase(nil)
	if idx != 0 {
		t.Errorf("currentPhase = %d, want 0 for nil phases", idx)
	}
}

func TestFindCurrentPhaseSkipsCompletedPhases(t *testing.T) {
	phases := []PhaseJSON{
		{Name: "Phase 1", Tasks: []TaskItemJSON{
			{Done: true}, {Done: true},
		}, Complete: 2, Total: 2},
		{Name: "Phase 2", Tasks: []TaskItemJSON{
			{Done: true}, {Done: true},
		}, Complete: 2, Total: 2},
		{Name: "Phase 3", Tasks: []TaskItemJSON{
			{Done: false}, {Done: false},
		}, Complete: 0, Total: 2},
	}

	idx := findCurrentPhase(phases)
	if idx != 2 {
		t.Errorf("currentPhase = %d, want 2 (skips completed phases)", idx)
	}
}

func TestComputeProgress(t *testing.T) {
	phases := []PhaseJSON{
		{Complete: 2, Total: 3},
		{Complete: 1, Total: 4},
	}
	p := computeProgress(phases)
	if p.Total != 7 {
		t.Errorf("Total = %d, want 7", p.Total)
	}
	if p.Complete != 3 {
		t.Errorf("Complete = %d, want 3", p.Complete)
	}
	if p.Remaining != 4 {
		t.Errorf("Remaining = %d, want 4", p.Remaining)
	}
}

func TestComputeProgressEmpty(t *testing.T) {
	p := computeProgress(nil)
	if p.Total != 0 || p.Complete != 0 || p.Remaining != 0 {
		t.Errorf("empty progress should be zero: %+v", p)
	}
}

func TestComputeProgressAllDone(t *testing.T) {
	phases := []PhaseJSON{
		{Complete: 3, Total: 3},
		{Complete: 2, Total: 2},
	}
	p := computeProgress(phases)
	if p.Remaining != 0 {
		t.Errorf("Remaining = %d, want 0", p.Remaining)
	}
}
