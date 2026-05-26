package skill

import (
	"io/fs"
	"path/filepath"
	"strings"
)

var templates = map[string]string{}
var resources = map[string]map[string]string{}

func init() {
	loadTemplates()
	loadResources()
}

func loadTemplates() {
	entries, err := fs.ReadDir(templateFS, "templates")
	if err != nil {
		return
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), ".md")
		data, err := fs.ReadFile(templateFS, "templates/"+entry.Name())
		if err != nil {
			continue
		}
		templates[id] = string(data)
	}
}

func loadResources() {
	refDir := "templates/references"
	entries, err := fs.ReadDir(templateFS, refDir)
	if err != nil {
		return
	}
	for _, skillEntry := range entries {
		if !skillEntry.IsDir() {
			continue
		}
		skillID := skillEntry.Name()
		skillRefDir := refDir + "/" + skillID
		fileEntries, err := fs.ReadDir(templateFS, skillRefDir)
		if err != nil {
			continue
		}
		for _, fileEntry := range fileEntries {
			if fileEntry.IsDir() {
				continue
			}
			data, err := fs.ReadFile(templateFS, skillRefDir+"/"+fileEntry.Name())
			if err != nil {
				continue
			}
			relPath := filepath.Join("references", fileEntry.Name())
			if resources[skillID] == nil {
				resources[skillID] = map[string]string{}
			}
			resources[skillID][relPath] = string(data)
		}
	}
}

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
