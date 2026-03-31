package internal

var Artifacts = []ArtifactInfo{
	{
		ID:          "proposal",
		Filename:    "proposal.md",
		Description: "Why and what — the motivation, scope, and approach for this change",
		Requires:    []string{},
	},
	{
		ID:          "specs",
		Filename:    "specs",
		Description: "Delta specifications — ADDED/MODIFIED/REMOVED/RENAMED requirements",
		Requires:    []string{"proposal"},
	},
	{
		ID:          "design",
		Filename:    "design.md",
		Description: "How — technical approach, architecture decisions, data flow, file changes",
		Requires:    []string{"proposal"},
	},
	{
		ID:          "tasks",
		Filename:    "tasks.md",
		Description: "What to do — phased implementation checklist",
		Requires:    []string{"proposal", "specs", "design"},
	},
}

func GetArtifact(id string) *ArtifactInfo {
	for i := range Artifacts {
		if Artifacts[i].ID == id {
			return &Artifacts[i]
		}
	}
	return nil
}
