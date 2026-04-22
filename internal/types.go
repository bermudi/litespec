package internal

import "time"

type ArtifactState string

const (
	ArtifactBlocked ArtifactState = "BLOCKED"
	ArtifactReady   ArtifactState = "READY"
	ArtifactDone    ArtifactState = "DONE"
)

type DeltaOperation string

const (
	DeltaAdded    DeltaOperation = "ADDED"
	DeltaModified DeltaOperation = "MODIFIED"
	DeltaRemoved  DeltaOperation = "REMOVED"
	DeltaRenamed  DeltaOperation = "RENAMED"
)

type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)

type ChangeMeta struct {
	Schema    string    `yaml:"schema"`
	Created   time.Time `yaml:"created"`
	DependsOn []string  `yaml:"dependsOn,omitempty"`
}

type Change struct {
	Name      string
	Schema    string
	Created   time.Time
	Artifacts map[string]ArtifactState
}

type ArtifactInfo struct {
	ID          string
	Filename    string
	Description string
	Requires    []string
}

type Scenario struct {
	Name    string
	Content string
}

type DeltaRequirement struct {
	Operation DeltaOperation
	Name      string
	OldName   string
	Content   string
	Scenarios []Scenario
}

type DeltaSpec struct {
	Capability   string
	Requirements []DeltaRequirement
}

type ValidationIssue struct {
	Severity Severity
	Message  string
	File     string
}

type ValidationResult struct {
	Valid             bool
	Errors            []ValidationIssue
	Warnings          []ValidationIssue
	ChangesCount      int
	CapabilitiesCount int
	RequirementsCount int
	ScenariosCount    int
}

type SpecRequirement struct {
	Name      string
	Content   string
	Scenarios []Scenario
}

type Spec struct {
	Capability   string
	Purpose      string
	Requirements []SpecRequirement
}

type SkillInfo struct {
	ID          string
	Name        string
	Description string
	Template    string
}

type ToolAdapter struct {
	ID        string
	Name      string
	SkillsDir string
}
