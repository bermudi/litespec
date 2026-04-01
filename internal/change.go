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

type PendingWrite struct {
	Capability string
	Path       string
	Dir        string
	Content    string
}

func PrepareArchiveWrites(root, name string) ([]PendingWrite, error) {
	changeSpecsDir := ChangeSpecsPath(root, name)
	entries, err := os.ReadDir(changeSpecsDir)
	if err != nil {
		return nil, fmt.Errorf("read change specs: %w", err)
	}

	var writes []PendingWrite
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		capability := entry.Name()
		capDir := filepath.Join(changeSpecsDir, capability)
		files, readErr := os.ReadDir(capDir)
		if readErr != nil {
			continue
		}

		var deltas []*DeltaSpec
		for _, f := range files {
			if filepath.Ext(f.Name()) != ".md" {
				continue
			}
			data, readErr := os.ReadFile(filepath.Join(capDir, f.Name()))
			if readErr != nil {
				return nil, fmt.Errorf("reading delta spec %s: %w", f.Name(), readErr)
			}
			delta, parseErr := ParseDeltaSpec(string(data))
			if parseErr != nil {
				return nil, fmt.Errorf("parsing delta spec %s: %w", f.Name(), parseErr)
			}
			deltas = append(deltas, delta)
		}

		if len(deltas) == 0 {
			continue
		}

		mainSpecDir := filepath.Join(SpecsPath(root), capability)
		mainSpecPath := filepath.Join(mainSpecDir, "spec.md")
		mainData, readErr := os.ReadFile(mainSpecPath)

		var mainSpec *Spec
		if readErr != nil {
			cap := deltas[0].Capability
			if cap == "" {
				cap = capability
			}
			mainSpec = &Spec{Capability: cap}
		} else {
			mainSpec, err = ParseMainSpec(string(mainData))
			if err != nil {
				return nil, fmt.Errorf("parsing main spec for %s: %w", capability, err)
			}
		}

		merged, err := MergeDelta(mainSpec, deltas)
		if err != nil {
			return nil, fmt.Errorf("merging delta for %s: %w", capability, err)
		}

		writes = append(writes, PendingWrite{
			Capability: capability,
			Path:       mainSpecPath,
			Dir:        mainSpecDir,
			Content:    SerializeSpec(merged),
		})
	}

	return writes, nil
}

func WritePendingSpecs(writes []PendingWrite) error {
	for _, w := range writes {
		if err := os.MkdirAll(w.Dir, 0o755); err != nil {
			return fmt.Errorf("creating spec directory %s: %w", w.Dir, err)
		}
		if err := os.WriteFile(w.Path, []byte(w.Content), 0o644); err != nil {
			return fmt.Errorf("writing spec %s: %w", w.Path, err)
		}
	}
	return nil
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
