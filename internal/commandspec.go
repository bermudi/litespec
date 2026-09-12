package internal

type FlagSpec struct {
	Name        string
	Description string
	TakesValue  bool
	Values      []Completion
	ValuesFunc  func(root string) []Completion
}

type PositionalSpec struct {
	Description string
	Resolver    func(root string) []Completion
}

type CommandSpec struct {
	Name        string
	Description string
	Hidden      bool
	Flags       []FlagSpec
	Positional  *PositionalSpec
}

var CommandSpecs = []CommandSpec{
	{
		Name:        "init",
		Description: "Initialize project structure",
		Flags: []FlagSpec{
			{
				Name:        "--tools",
				Description: "Tool IDs (comma-separated)",
				TakesValue:  true,
				ValuesFunc:  func(root string) []Completion { return completeToolIDs() },
			},
			{Name: "--json", Description: "Output as JSON", TakesValue: false},
			{Name: "--minimal", Description: "Minimal output", TakesValue: false},
		},
	},
	{
		Name:        "validate",
		Description: "Validate specs, decisions, and queues",
		Flags: []FlagSpec{
			{Name: "--all", Description: "Validate all specs, decisions, and queues", TakesValue: false},
			{Name: "--specs", Description: "Validate all specs only", TakesValue: false},
			{Name: "--decisions", Description: "Validate all decisions only", TakesValue: false},
			{Name: "--issue", Description: "Fetch and validate a single GH issue by number", TakesValue: true},
			{Name: "--queue", Description: "Validate a single local queue markdown file", TakesValue: true},
			{Name: "--strict", Description: "Treat warnings as errors", TakesValue: false},
			{Name: "--json", Description: "Output as JSON", TakesValue: false},
			{Name: "--minimal", Description: "Minimal output", TakesValue: false},
			{Name: "--type", Description: "Disambiguate name: spec|decision", TakesValue: true, Values: []Completion{
				{"spec", "Disambiguate as spec"},
				{"decision", "Disambiguate as decision"},
			}},
		},
	},
	{
		Name:        "digest",
		Description: "Print expected unit contract digests for a queue",
		Flags: []FlagSpec{
			{Name: "--issue", Description: "GH issue number", TakesValue: true},
			{Name: "--queue", Description: "Local queue markdown file", TakesValue: true},
		},
	},
	{
		Name:        "receipt",
		Description: "Assemble validator-clean evidence receipt comment files",
		Flags: []FlagSpec{
			{Name: "--issue", Description: "GH issue number", TakesValue: true},
			{Name: "--queue", Description: "Local queue markdown file", TakesValue: true},
			{Name: "--heading", Description: "Exact unit heading", TakesValue: true},
			{Name: "--occurrence", Description: "1-based same-heading occurrence", TakesValue: true},
			{Name: "--pre-sha", Description: "Pre commit SHA", TakesValue: true},
			{Name: "--pre-status", Description: "Pre exit status (non-zero)", TakesValue: true},
			{Name: "--pre-out", Description: "File with raw pre output", TakesValue: true},
			{Name: "--post-sha", Description: "Post commit SHA", TakesValue: true},
			{Name: "--post-status", Description: "Post exit status (default 0)", TakesValue: true},
			{Name: "--post-out", Description: "File with raw post output", TakesValue: true},
			{Name: "--rebuild", Description: "Include rebuild routing identity", TakesValue: false},
			{Name: "--recovered-from", Description: "Recovery provenance receipt ID", TakesValue: true},
			{Name: "--post", Description: "Run the printed gh issue comment commands in order", TakesValue: false},
		},
	},
	{
		Name:        "issue",
		Description: "Managed GH issue body operations",
		Positional: &PositionalSpec{
			Description: "subcommand",
			Resolver:    func(root string) []Completion { return completeIssueSubcommands() },
		},
		Flags: []FlagSpec{
			{Name: "--issue", Description: "GH issue number", TakesValue: true},
			{Name: "--heading", Description: "Exact unit heading", TakesValue: true},
			{Name: "--occurrence", Description: "1-based same-heading occurrence", TakesValue: true},
		},
	},
	{
		Name:        "view",
		Description: "Dashboard overview",
		Flags: []FlagSpec{
			{Name: "--json", Description: "Output as JSON", TakesValue: false},
			{Name: "--minimal", Description: "Minimal output", TakesValue: false},
		},
	},
	{
		Name:        "update",
		Description: "Regenerate skills and adapters",
		Flags: []FlagSpec{
			{
				Name:        "--tools",
				Description: "Tool IDs (comma-separated)",
				TakesValue:  true,
				ValuesFunc:  func(root string) []Completion { return completeToolIDs() },
			},
			{Name: "--json", Description: "Output as JSON", TakesValue: false},
			{Name: "--minimal", Description: "Minimal output", TakesValue: false},
		},
	},
	{
		Name:        "upgrade",
		Description: "Check for and install the latest version",
		Flags: []FlagSpec{
			{Name: "--json", Description: "Output as JSON", TakesValue: false},
			{Name: "--minimal", Description: "Minimal output", TakesValue: false},
		},
	},
	{
		Name:        "completion",
		Description: "Generate shell completion script",
		Positional: &PositionalSpec{
			Description: "shell name",
			Resolver:    func(root string) []Completion { return completeShells() },
		},
	},
	{
		Name:   "__complete",
		Hidden: true,
	},
}
