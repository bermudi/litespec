package internal

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

type ChangeStatusJSON struct {
	ChangeName string               `json:"changeName"`
	SchemaName string               `json:"schemaName"`
	IsComplete bool                 `json:"isComplete"`
	Artifacts  []ArtifactStatusJSON `json:"artifacts"`
}

type ArtifactStatusJSON struct {
	ID          string   `json:"id"`
	OutputPath  string   `json:"outputPath"`
	Status      string   `json:"status"`
	MissingDeps []string `json:"missingDeps,omitempty"`
}

type ArtifactInstructionsJSON struct {
	ChangeName   string               `json:"changeName"`
	ArtifactID   string               `json:"artifactId"`
	SchemaName   string               `json:"schemaName"`
	ChangeDir    string               `json:"changeDir"`
	OutputPath   string               `json:"outputPath"`
	Description  string               `json:"description"`
	Instruction  string               `json:"instruction"`
	Template     string               `json:"template"`
	Dependencies []DependencyInfoJSON `json:"dependencies"`
	Unlocks      []string             `json:"unlocks"`
}

type DependencyInfoJSON struct {
	ID          string `json:"id"`
	Done        bool   `json:"done"`
	Path        string `json:"path"`
	Description string `json:"description"`
}

type ProgressJSON struct {
	Total     int `json:"total"`
	Complete  int `json:"complete"`
	Remaining int `json:"remaining"`
}

type PhaseJSON struct {
	Name     string         `json:"name"`
	Tasks    []TaskItemJSON `json:"tasks"`
	Complete int            `json:"complete"`
	Total    int            `json:"total"`
}

type TaskItemJSON struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Done        bool   `json:"done"`
}

type ChangeListItemJSON struct {
	Name string `json:"name"`
}

type ValidationResultJSON struct {
	Valid    bool                  `json:"valid"`
	Errors   []ValidationIssueJSON `json:"errors"`
	Warnings []ValidationIssueJSON `json:"warnings"`
	Summary  ValidationSummaryJSON `json:"summary"`
}

type ValidationIssueJSON struct {
	Severity string `json:"severity"`
	Message  string `json:"message"`
	File     string `json:"file"`
}

type ValidationSummaryJSON struct {
	Total   int `json:"total"`
	Invalid int `json:"invalid"`
}

func BuildValidationResultJSON(r *ValidationResult) ValidationResultJSON {
	errors := make([]ValidationIssueJSON, len(r.Errors))
	for i, e := range r.Errors {
		errors[i] = ValidationIssueJSON{Severity: "error", Message: e.Message, File: e.File}
	}
	warnings := make([]ValidationIssueJSON, len(r.Warnings))
	for i, w := range r.Warnings {
		warnings[i] = ValidationIssueJSON{Severity: "warning", Message: w.Message, File: w.File}
	}
	return ValidationResultJSON{
		Valid:    r.Valid,
		Errors:   errors,
		Warnings: warnings,
		Summary:  ValidationSummaryJSON{Total: len(errors) + len(warnings), Invalid: len(errors)},
	}
}

func MarshalJSON(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

func artifactStateToString(s ArtifactState) string {
	switch s {
	case ArtifactBlocked:
		return "blocked"
	case ArtifactReady:
		return "ready"
	case ArtifactDone:
		return "done"
	default:
		return "blocked"
	}
}

func ArtifactInstructionID(artifactID string) string {
	switch artifactID {
	case "proposal":
		return "artifact-proposal"
	case "specs":
		return "artifact-specs"
	case "design":
		return "artifact-design"
	case "tasks":
		return "artifact-tasks"
	default:
		return "artifact-proposal"
	}
}

func BuildChangeStatusJSON(change *Change) ChangeStatusJSON {
	var artifacts []ArtifactStatusJSON
	allDone := true

	for _, info := range Artifacts {
		state := change.Artifacts[info.ID]
		if state != ArtifactDone {
			allDone = false
		}

		var missing []string
		for _, req := range info.Requires {
			if change.Artifacts[req] != ArtifactDone {
				missing = append(missing, req)
			}
		}

		artifacts = append(artifacts, ArtifactStatusJSON{
			ID:          info.ID,
			OutputPath:  info.Filename,
			Status:      artifactStateToString(state),
			MissingDeps: missing,
		})
	}

	return ChangeStatusJSON{
		ChangeName: change.Name,
		SchemaName: change.Schema,
		IsComplete: allDone,
		Artifacts:  artifacts,
	}
}

func BuildArtifactInstructionsJSON(root, changeName, artifactID string) (*ArtifactInstructionsJSON, error) {
	change, err := LoadChangeContext(root, changeName)
	if err != nil {
		return nil, err
	}

	info := GetArtifact(artifactID)
	if info == nil {
		return nil, fmt.Errorf("artifact %q not found", artifactID)
	}

	changeDir := ChangePath(root, changeName)
	instruction := GetSkillTemplate(ArtifactInstructionID(artifactID))
	template := GetSkillTemplate("propose")

	var deps []DependencyInfoJSON
	for _, reqID := range info.Requires {
		reqInfo := GetArtifact(reqID)
		if reqInfo == nil {
			continue
		}
		deps = append(deps, DependencyInfoJSON{
			ID:          reqInfo.ID,
			Done:        change.Artifacts[reqID] == ArtifactDone,
			Path:        filepath.Join(changeDir, reqInfo.Filename),
			Description: reqInfo.Description,
		})
	}

	var unlocks []string
	for _, candidate := range Artifacts {
		for _, req := range candidate.Requires {
			if req == artifactID {
				unlocks = append(unlocks, candidate.ID)
				break
			}
		}
	}

	return &ArtifactInstructionsJSON{
		ChangeName:   changeName,
		ArtifactID:   artifactID,
		SchemaName:   change.Schema,
		ChangeDir:    changeDir,
		OutputPath:   info.Filename,
		Description:  info.Description,
		Instruction:  instruction,
		Template:     template,
		Dependencies: deps,
		Unlocks:      unlocks,
	}, nil
}

func parseTasksMD(content string) []PhaseJSON {
	var phases []PhaseJSON
	var current *PhaseJSON

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "## Phase") {
			name := strings.TrimSpace(strings.TrimPrefix(line, "##"))
			phases = append(phases, PhaseJSON{Name: name})
			current = &phases[len(phases)-1]
			continue
		}

		if current == nil {
			continue
		}

		if strings.HasPrefix(line, "- [x] ") {
			desc := strings.TrimPrefix(line, "- [x] ")
			current.Tasks = append(current.Tasks, TaskItemJSON{
				ID:          fmt.Sprintf("%s-%d", current.Name, len(current.Tasks)+1),
				Description: strings.TrimSpace(desc),
				Done:        true,
			})
		} else if strings.HasPrefix(line, "- [ ] ") {
			desc := strings.TrimPrefix(line, "- [ ] ")
			current.Tasks = append(current.Tasks, TaskItemJSON{
				ID:          fmt.Sprintf("%s-%d", current.Name, len(current.Tasks)+1),
				Description: strings.TrimSpace(desc),
				Done:        false,
			})
		}
	}

	for i := range phases {
		complete := 0
		for _, t := range phases[i].Tasks {
			if t.Done {
				complete++
			}
		}
		phases[i].Complete = complete
		phases[i].Total = len(phases[i].Tasks)
	}

	return phases
}

func computeProgress(phases []PhaseJSON) ProgressJSON {
	total := 0
	complete := 0
	for _, p := range phases {
		total += p.Total
		complete += p.Complete
	}
	return ProgressJSON{
		Total:     total,
		Complete:  complete,
		Remaining: total - complete,
	}
}

func findCurrentPhase(phases []PhaseJSON) int {
	for i, p := range phases {
		for _, t := range p.Tasks {
			if !t.Done {
				return i
			}
		}
	}
	return 0
}
