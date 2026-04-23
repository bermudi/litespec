package skill

var templates = map[string]string{}
var resources = map[string]map[string]string{}

func Register(id string, template string) {
	templates[id] = template
}

func RegisterResource(skillID, relPath, content string) {
	if resources[skillID] == nil {
		resources[skillID] = map[string]string{}
	}
	resources[skillID][relPath] = content
}

func Get(id string) string {
	return templates[id]
}

func GetResources(skillID string) map[string]string {
	return resources[skillID]
}

func All() map[string]string {
	return templates
}

func ValidateSkillTemplates(skillIDs []string) []string {
	missing := make([]string, 0)
	for _, id := range skillIDs {
		if templates[id] == "" {
			missing = append(missing, id)
		}
	}
	return missing
}
