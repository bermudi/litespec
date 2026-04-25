package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func artifactExists(root, changeName string, info ArtifactInfo) bool {
	path := filepath.Join(ChangePath(root, changeName), info.Filename)
	if info.ID == "specs" {
		entries, err := os.ReadDir(path)
		if err != nil {
			return false
		}
		for _, e := range entries {
			if e.IsDir() {
				if hasMarkdownFiles(filepath.Join(path, e.Name())) {
					return true
				}
			}
		}
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

func hasMarkdownFiles(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".md" {
			return true
		}
	}
	return false
}

func LoadArtifactStates(root, changeName string) (map[string]ArtifactState, error) {
	if IsPatchMode(root, changeName) {
		return map[string]ArtifactState{"specs": ArtifactDone}, nil
	}

	states := make(map[string]ArtifactState)

	for _, art := range Artifacts {
		states[art.ID] = ArtifactBlocked
	}

	for i := 0; i < len(Artifacts); i++ {
		changed := false
		for _, art := range Artifacts {
			if states[art.ID] != ArtifactBlocked {
				continue
			}

			if artifactExists(root, changeName, art) {
				states[art.ID] = ArtifactDone
				changed = true
				continue
			}

			allDone := true
			for _, req := range art.Requires {
				if states[req] != ArtifactDone {
					allDone = false
					break
				}
			}
			if allDone {
				states[art.ID] = ArtifactReady
				changed = true
			}
		}
		if !changed {
			break
		}
	}

	return states, nil
}

func GetReadyArtifacts(states map[string]ArtifactState) []string {
	var ready []string
	for _, art := range Artifacts {
		if states[art.ID] == ArtifactReady {
			ready = append(ready, art.ID)
		}
	}
	return ready
}

func GetNextArtifact(states map[string]ArtifactState) string {
	for _, art := range Artifacts {
		if states[art.ID] == ArtifactReady {
			return art.ID
		}
	}
	return ""
}

func LoadChangeContext(root, changeName string) (*Change, error) {
	changeDir := ChangePath(root, changeName)
	if _, err := os.Stat(changeDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("change %q not found", changeName)
	}

	metaPath := filepath.Join(changeDir, MetaFileName)
	var meta ChangeMeta

	data, err := os.ReadFile(metaPath)
	if err == nil {
		if err := yaml.Unmarshal(data, &meta); err != nil {
			return nil, fmt.Errorf("invalid metadata file: %w", err)
		}
	}

	states, err := LoadArtifactStates(root, changeName)
	if err != nil {
		return nil, err
	}

	return &Change{
		Name:      changeName,
		Schema:    meta.Schema,
		Created:   meta.Created,
		Artifacts: states,
	}, nil
}
