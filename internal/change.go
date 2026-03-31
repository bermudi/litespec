package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

func InitProject(root string) error {
	dirs := []string{
		SpecsPath(root),
		ChangesPath(root),
		ArchivePath(root),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create directory %s: %w", dir, err)
		}
	}
	return nil
}

func CreateChange(root, name string) error {
	changeDir := ChangePath(root, name)
	if _, err := os.Stat(changeDir); err == nil {
		return fmt.Errorf("change %q already exists", name)
	}

	if err := os.MkdirAll(ChangeSpecsPath(root, name), 0o755); err != nil {
		return fmt.Errorf("create change directory: %w", err)
	}

	meta := ChangeMeta{
		Schema:  "spec-driven",
		Created: time.Now().UTC().Truncate(time.Second),
	}

	data, err := yaml.Marshal(&meta)
	if err != nil {
		return fmt.Errorf("marshal change metadata: %w", err)
	}

	metaPath := filepath.Join(changeDir, MetaFileName)
	if err := os.WriteFile(metaPath, data, 0o644); err != nil {
		return fmt.Errorf("write change metadata: %w", err)
	}

	return nil
}

func ListChanges(root string) ([]string, error) {
	changesDir := ChangesPath(root)
	entries, err := os.ReadDir(changesDir)
	if err != nil {
		return nil, fmt.Errorf("read changes directory: %w", err)
	}

	var names []string
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == ArchiveDirName {
			continue
		}
		names = append(names, entry.Name())
	}
	return names, nil
}

func ListSpecs(root string) ([]string, error) {
	specsDir := SpecsPath(root)
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return nil, fmt.Errorf("read specs directory: %w", err)
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	return names, nil
}

func ArchiveChange(root, name string) error {
	changeDir := ChangePath(root, name)
	if _, err := os.Stat(changeDir); err != nil {
		return fmt.Errorf("change %q does not exist", name)
	}

	archivedName := time.Now().Format("2006-01-02") + "-" + name
	dest := filepath.Join(ArchivePath(root), archivedName)

	if err := os.Rename(changeDir, dest); err != nil {
		return fmt.Errorf("archive change: %w", err)
	}

	return nil
}

func ChangeExists(root, name string) bool {
	_, err := os.Stat(ChangePath(root, name))
	return err == nil
}

func ReadChangeMeta(root, name string) (*ChangeMeta, error) {
	metaPath := filepath.Join(ChangePath(root, name), MetaFileName)
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, fmt.Errorf("read change metadata: %w", err)
	}

	var meta ChangeMeta
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("parse change metadata: %w", err)
	}

	return &meta, nil
}
