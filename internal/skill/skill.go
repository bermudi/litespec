package skill

var templates = map[string]string{}

func Register(id string, template string) {
	templates[id] = template
}

func Get(id string) string {
	return templates[id]
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
