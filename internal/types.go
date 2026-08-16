package internal

type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)

type Scenario struct {
	Name    string
	Content string
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
	CapabilitiesCount int
	RequirementsCount int
	ScenariosCount    int
	DecisionsCount    int
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
