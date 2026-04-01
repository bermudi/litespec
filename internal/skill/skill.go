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
