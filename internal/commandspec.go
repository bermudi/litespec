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
		},
	},
	{
		Name:        "new",
		Description: "Create a new change",
		Flags: []FlagSpec{
			{Name: "--json", Description: "Output as JSON", TakesValue: false},
		},
		Positional: &PositionalSpec{
			Description: "change name",
		},
	},
	{
		Name:        "patch",
		Description: "Create a patch-mode change (delta-only)",
		Flags: []FlagSpec{
			{Name: "--json", Description: "Output as JSON", TakesValue: false},
		},
		Positional: &PositionalSpec{
			Description: "change name and capability",
		},
	},
	{
		Name:        "list",
		Description: "List specs or changes",
		Flags: []FlagSpec{
			{Name: "--specs", Description: "List specs instead of changes", TakesValue: false},
			{Name: "--changes", Description: "List changes (default)", TakesValue: false},
			{Name: "--decisions", Description: "List architectural decision records", TakesValue: false},
			{Name: "--sort", Description: "Sort by 'recent', 'name', 'deps', or 'number'", TakesValue: true, Values: []Completion{
				{"recent", "Sort by last modified"},
				{"name", "Sort alphabetically"},
				{"deps", "Sort by dependency order"},
				{"number", "Sort by decision number"},
			}},
			{Name: "--status", Description: "Filter decisions by status (requires --decisions)", TakesValue: true, Values: []Completion{
				{"proposed", "Proposed decisions"},
				{"accepted", "Accepted decisions"},
				{"superseded", "Superseded decisions"},
			}},
			{Name: "--json", Description: "Output as JSON", TakesValue: false},
		},
	},
	{
		Name:        "status",
		Description: "Show artifact states",
		Flags: []FlagSpec{
			{Name: "--json", Description: "Output as JSON", TakesValue: false},
		},
		Positional: &PositionalSpec{
			Description: "change name",
			Resolver:    completeChangeNames,
		},
	},
	{
		Name:        "validate",
		Description: "Validate changes and specs",
		Flags: []FlagSpec{
			{Name: "--all", Description: "Validate all changes, specs, and decisions", TakesValue: false},
			{Name: "--changes", Description: "Validate all changes only", TakesValue: false},
			{Name: "--specs", Description: "Validate all specs only", TakesValue: false},
			{Name: "--decisions", Description: "Validate all decisions only", TakesValue: false},
			{Name: "--strict", Description: "Treat warnings as errors", TakesValue: false},
			{Name: "--json", Description: "Output as JSON", TakesValue: false},
			{Name: "--type", Description: "Disambiguate name: change|spec|decision", TakesValue: true, Values: []Completion{
				{"change", "Disambiguate as change"},
				{"spec", "Disambiguate as spec"},
				{"decision", "Disambiguate as decision"},
			}},
		},
	},
	{
		Name:        "instructions",
		Description: "Get artifact instructions",
		Flags: []FlagSpec{
			{Name: "--json", Description: "Output as JSON", TakesValue: false},
		},
		Positional: &PositionalSpec{
			Description: "artifact ID",
			Resolver:    func(root string) []Completion { return completeArtifactIDs() },
		},
	},
	{
		Name:        "archive",
		Description: "Apply deltas and archive change",
		Flags: []FlagSpec{
			{Name: "--allow-incomplete", Description: "Archive even with incomplete tasks or unarchived dependencies", TakesValue: false},
		},
		Positional: &PositionalSpec{
			Description: "change name",
			Resolver:    completeChangeNames,
		},
	},
	{
		Name:        "preview",
		Description: "Preview what archive would do to canon specs",
		Flags: []FlagSpec{
			{Name: "--json", Description: "Output as JSON", TakesValue: false},
		},
		Positional: &PositionalSpec{
			Description: "change name",
			Resolver:    completeChangeNames,
		},
	},
	{
		Name:        "view",
		Description: "Dashboard overview with dependency graph",
		Flags: []FlagSpec{
			{Name: "--json", Description: "Output as JSON", TakesValue: false},
		},
	},
	{
		Name:        "decide",
		Description: "Create a new architectural decision record",
		Positional: &PositionalSpec{
			Description: "decision slug",
		},
	},
	{
		Name:        "import",
		Description: "Import OpenSpec project to litespec",
		Flags: []FlagSpec{
			{Name: "--source", Description: "Source OpenSpec project directory", TakesValue: true},
			{Name: "--dry-run", Description: "Preview import without making changes", TakesValue: false},
			{Name: "--force", Description: "Overwrite existing files in target", TakesValue: false},
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
		},
	},
	{
		Name:        "upgrade",
		Description: "Check for and install the latest version",
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
