package importer

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	OpenSpecDir        = "openspec"
	OpenSpecCanonDir   = "specs"
	OpenSpecChangesDir = "changes"
	OpenSpecArchiveDir = "archive"
	OpenSpecMetaFile   = ".openspec.yaml"
	LiteSpecMetaFile   = ".litespec.yaml"
)

type OpenSpecMeta struct {
	Schema    string   `yaml:"schema"`
	Created   string   `yaml:"created"`
	DependsOn []string `yaml:"dependsOn,omitempty"`
}

type ImportStats struct {
	CanonSpecs        int
	ActiveChanges     int
	Archives          int
	CanonSpecNames    []string
	ActiveChangeNames []string
	ArchiveNames      []string
	SkippedFiles      []string
	Warnings          []string
}

func DetectOpenSpecProject(source string) bool {
	specsDir := filepath.Join(source, OpenSpecDir, OpenSpecCanonDir)
	changesDir := filepath.Join(source, OpenSpecDir, OpenSpecChangesDir)

	if _, err := os.Stat(specsDir); err == nil {
		return true
	}
	if _, err := os.Stat(changesDir); err == nil {
		return true
	}
	return false
}

var knownSkippedItems = map[string]string{
	"config.yaml": "config.yaml (no litespec equivalent)",
	"project.md":  "project.md (no litespec equivalent)",
	"AGENTS.md":   "AGENTS.md (no litespec equivalent)",
}

func scanOpenSpecRoot(source string, stats *ImportStats) {
	root := filepath.Join(source, OpenSpecDir)
	entries, err := os.ReadDir(root)
	if err != nil {
		return
	}
	for _, entry := range entries {
		name := entry.Name()
		if name == OpenSpecCanonDir || name == OpenSpecChangesDir {
			continue
		}
		if msg, ok := knownSkippedItems[name]; ok {
			stats.Warnings = append(stats.Warnings, fmt.Sprintf("skipped %s", msg))
			stats.SkippedFiles = append(stats.SkippedFiles, filepath.Join(root, name))
		}
		if entry.IsDir() && name == "explorations" {
			stats.Warnings = append(stats.Warnings, "skipped explorations/ (no litespec equivalent)")
			stats.SkippedFiles = append(stats.SkippedFiles, filepath.Join(root, name))
		}
	}
}

func ImportOpenSpecProject(source, target string) (*ImportStats, error) {
	stats := &ImportStats{}

	if !DetectOpenSpecProject(source) {
		return nil, fmt.Errorf("no OpenSpec project found at %s", source)
	}

	scanOpenSpecRoot(source, stats)

	openSpecCanon := filepath.Join(source, OpenSpecDir, OpenSpecCanonDir)
	if _, err := os.Stat(openSpecCanon); err == nil {
		if err := copyCanonSpecs(openSpecCanon, target, stats, false); err != nil {
			return nil, fmt.Errorf("copy canon specs: %w", err)
		}
	}

	openSpecChanges := filepath.Join(source, OpenSpecDir, OpenSpecChangesDir)
	if _, err := os.Stat(openSpecChanges); err == nil {
		if err := migrateChanges(openSpecChanges, target, stats, false); err != nil {
			return nil, fmt.Errorf("migrate changes: %w", err)
		}
	}

	return stats, nil
}

func PreviewImport(source string) (*ImportStats, error) {
	stats := &ImportStats{}

	if !DetectOpenSpecProject(source) {
		return nil, fmt.Errorf("no OpenSpec project found at %s", source)
	}

	scanOpenSpecRoot(source, stats)

	openSpecCanon := filepath.Join(source, OpenSpecDir, OpenSpecCanonDir)
	if _, err := os.Stat(openSpecCanon); err == nil {
		if err := copyCanonSpecs(openSpecCanon, "", stats, true); err != nil {
			return nil, fmt.Errorf("preview canon specs: %w", err)
		}
	}

	openSpecChanges := filepath.Join(source, OpenSpecDir, OpenSpecChangesDir)
	if _, err := os.Stat(openSpecChanges); err == nil {
		if err := migrateChanges(openSpecChanges, "", stats, true); err != nil {
			return nil, fmt.Errorf("preview changes: %w", err)
		}
	}

	return stats, nil
}

func copyCanonSpecs(source, target string, stats *ImportStats, dryRun bool) error {
	if dryRun {
		entries, err := os.ReadDir(source)
		if err != nil {
			return fmt.Errorf("read canon directory: %w", err)
		}
		for _, entry := range entries {
			if entry.IsDir() {
				stats.CanonSpecs++
				stats.CanonSpecNames = append(stats.CanonSpecNames, entry.Name())
			}
		}
		return nil
	}

	targetCanon := filepath.Join(target, "specs", "canon")

	entries, err := os.ReadDir(source)
	if err != nil {
		return fmt.Errorf("read canon directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			stats.SkippedFiles = append(stats.SkippedFiles, filepath.Join(source, entry.Name()))
			continue
		}

		capability := entry.Name()
		srcDir := filepath.Join(source, capability)
		dstDir := filepath.Join(targetCanon, capability)

		if err := os.MkdirAll(dstDir, 0755); err != nil {
			return fmt.Errorf("create canon directory %s: %w", capability, err)
		}

		srcSpec := filepath.Join(srcDir, "spec.md")
		dstSpec := filepath.Join(dstDir, "spec.md")

		data, err := os.ReadFile(srcSpec)
		if err != nil {
			stats.Warnings = append(stats.Warnings, fmt.Sprintf("skipped %s: no spec.md", capability))
			continue
		}

		normalized := normalizeH1(string(data))
		if err := os.WriteFile(dstSpec, []byte(normalized), 0644); err != nil {
			return fmt.Errorf("write canon spec %s: %w", capability, err)
		}

		stats.CanonSpecs++
	}

	return nil
}

func migrateChanges(source, target string, stats *ImportStats, dryRun bool) error {
	entries, err := os.ReadDir(source)
	if err != nil {
		return fmt.Errorf("read changes directory: %w", err)
	}

	for _, entry := range entries {
		name := entry.Name()

		if name == OpenSpecArchiveDir {
			archiveDir := filepath.Join(source, OpenSpecArchiveDir)
			if err := migrateArchives(archiveDir, target, stats, dryRun); err != nil {
				return fmt.Errorf("migrate archives: %w", err)
			}
			continue
		}

		if !entry.IsDir() {
			if name == "IMPLEMENTATION_ORDER.md" {
				stats.SkippedFiles = append(stats.SkippedFiles, filepath.Join(source, name))
				stats.Warnings = append(stats.Warnings, "skipped IMPLEMENTATION_ORDER.md (no litespec equivalent)")
			} else {
				stats.SkippedFiles = append(stats.SkippedFiles, filepath.Join(source, name))
				stats.Warnings = append(stats.Warnings, fmt.Sprintf("skipped loose file: %s", name))
			}
			continue
		}

		if dryRun {
			stats.ActiveChanges++
			stats.ActiveChangeNames = append(stats.ActiveChangeNames, name)
			continue
		}

		srcChangeDir := filepath.Join(source, name)
		dstChangeDir := filepath.Join(target, "specs", "changes", name)

		if err := migrateActiveChange(srcChangeDir, dstChangeDir, stats); err != nil {
			return fmt.Errorf("migrate change %s: %w", name, err)
		}

		stats.ActiveChanges++
	}

	return nil
}

func migrateActiveChange(source, target string, stats *ImportStats) error {
	if err := os.MkdirAll(target, 0755); err != nil {
		return fmt.Errorf("create change directory: %w", err)
	}

	entries, err := os.ReadDir(source)
	if err != nil {
		return fmt.Errorf("read change directory: %w", err)
	}

	for _, entry := range entries {
		name := entry.Name()

		if name == "specs" {
			srcSpecsDir := filepath.Join(source, "specs")
			dstSpecsDir := filepath.Join(target, "specs")
			if err := copyDir(srcSpecsDir, dstSpecsDir); err != nil {
				return fmt.Errorf("copy specs: %w", err)
			}
			continue
		}

		if name == OpenSpecMetaFile {
			if err := convertMetaFile(source, target, stats); err != nil {
				return fmt.Errorf("convert metadata: %w", err)
			}
			continue
		}

		if name == "tasks.md" {
			srcTasks := filepath.Join(source, "tasks.md")
			dstTasks := filepath.Join(target, "tasks.md")
			if err := migrateTasksFile(srcTasks, dstTasks); err != nil {
				return fmt.Errorf("migrate tasks: %w", err)
			}
			continue
		}

		if entry.IsDir() {
			srcDir := filepath.Join(source, name)
			dstDir := filepath.Join(target, name)
			if err := copyDir(srcDir, dstDir); err != nil {
				return fmt.Errorf("copy directory %s: %w", name, err)
			}
			continue
		}

		srcFile := filepath.Join(source, name)
		dstFile := filepath.Join(target, name)
		if err := copyFile(srcFile, dstFile); err != nil {
			return fmt.Errorf("copy file %s: %w", name, err)
		}
	}

	return nil
}

func migrateArchives(source, target string, stats *ImportStats, dryRun bool) error {
	entries, err := os.ReadDir(source)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read archive directory: %w", err)
	}

	if dryRun {
		for _, entry := range entries {
			if entry.IsDir() {
				stats.Archives++
				stats.ArchiveNames = append(stats.ArchiveNames, entry.Name())
			}
		}
		return nil
	}

	targetArchive := filepath.Join(target, "specs", "changes", "archive")
	if err := os.MkdirAll(targetArchive, 0755); err != nil {
		return fmt.Errorf("create archive directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		srcArchiveDir := filepath.Join(source, name)
		dstArchiveDir := filepath.Join(targetArchive, name)

		if err := migrateArchive(srcArchiveDir, dstArchiveDir, stats); err != nil {
			return fmt.Errorf("migrate archive %s: %w", name, err)
		}

		stats.Archives++
	}

	return nil
}

func migrateArchive(source, target string, stats *ImportStats) error {
	if err := os.MkdirAll(target, 0755); err != nil {
		return fmt.Errorf("create archive directory: %w", err)
	}

	entries, err := os.ReadDir(source)
	if err != nil {
		return fmt.Errorf("read archive: %w", err)
	}

	for _, entry := range entries {
		name := entry.Name()

		if name == "specs" {
			stats.Warnings = append(stats.Warnings, fmt.Sprintf("stripped specs/ from archive %s", filepath.Base(source)))
			continue
		}

		if name == OpenSpecMetaFile {
			if err := convertMetaFile(source, target, stats); err != nil {
				return fmt.Errorf("convert archive metadata: %w", err)
			}
			continue
		}

		if name == "tasks.md" {
			srcTasks := filepath.Join(source, "tasks.md")
			dstTasks := filepath.Join(target, "tasks.md")
			if err := migrateTasksFile(srcTasks, dstTasks); err != nil {
				return fmt.Errorf("migrate archive tasks: %w", err)
			}
			continue
		}

		if entry.IsDir() {
			srcDir := filepath.Join(source, name)
			dstDir := filepath.Join(target, name)
			if err := copyDir(srcDir, dstDir); err != nil {
				return fmt.Errorf("copy archive directory %s: %w", name, err)
			}
			continue
		}

		srcFile := filepath.Join(source, name)
		dstFile := filepath.Join(target, name)
		if err := copyFile(srcFile, dstFile); err != nil {
			return fmt.Errorf("copy archive file %s: %w", name, err)
		}
	}

	if _, err := os.Stat(filepath.Join(target, LiteSpecMetaFile)); os.IsNotExist(err) {
		if err := synthesizeArchiveMeta(source, target); err != nil {
			return fmt.Errorf("synthesize metadata: %w", err)
		}
	}

	return nil
}

var h1SuffixRe = regexp.MustCompile(`^#\s+(.+?)\s+Specification\s*$`)

func normalizeH1(content string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if match := h1SuffixRe.FindStringSubmatch(line); match != nil {
			lines[i] = "# " + match[1]
			break
		}
		if strings.HasPrefix(line, "# ") && !strings.Contains(line, " Specification") {
			break
		}
	}
	return strings.Join(lines, "\n")
}

var phaseLabelRe = regexp.MustCompile(`^##\s+(\d+)\.\s+(.+)$`)

func normalizeTasksPhases(content string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if match := phaseLabelRe.FindStringSubmatch(line); match != nil {
			lines[i] = fmt.Sprintf("## Phase %s: %s", match[1], match[2])
		}
	}
	return strings.Join(lines, "\n")
}

func migrateTasksFile(source, target string) error {
	data, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("read tasks: %w", err)
	}

	normalized := normalizeTasksPhases(string(data))
	if err := os.WriteFile(target, []byte(normalized), 0644); err != nil {
		return fmt.Errorf("write tasks: %w", err)
	}

	return nil
}

func convertMetaFile(source, target string, stats *ImportStats) error {
	srcMeta := filepath.Join(source, OpenSpecMetaFile)
	dstMeta := filepath.Join(target, LiteSpecMetaFile)

	data, err := os.ReadFile(srcMeta)
	if err != nil {
		return fmt.Errorf("read openspec metadata: %w", err)
	}

	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("parse openspec metadata: %w", err)
	}

	unsupportedFields := []string{"provides", "requires", "touches", "parent"}
	var dropped []string
	for _, field := range unsupportedFields {
		if _, ok := raw[field]; ok {
			dropped = append(dropped, field)
		}
	}
	if len(dropped) > 0 {
		stats.Warnings = append(stats.Warnings, fmt.Sprintf("skipped unsupported fields in %s: %s", filepath.Base(source), strings.Join(dropped, ", ")))
	}

	var openMeta OpenSpecMeta
	if err := yaml.Unmarshal(data, &openMeta); err != nil {
		return fmt.Errorf("parse openspec metadata: %w", err)
	}

	created, err := parseCreatedTime(openMeta.Created)
	if err != nil {
		created = time.Now().UTC().Truncate(time.Second)
		stats.Warnings = append(stats.Warnings, fmt.Sprintf("could not parse date %q in %s, using current time", openMeta.Created, filepath.Base(source)))
	}

	liteMeta := struct {
		Schema    string   `yaml:"schema"`
		Created   string   `yaml:"created"`
		DependsOn []string `yaml:"dependsOn,omitempty"`
	}{
		Schema:    openMeta.Schema,
		Created:   created.Format(time.RFC3339),
		DependsOn: openMeta.DependsOn,
	}

	out, err := yaml.Marshal(&liteMeta)
	if err != nil {
		return fmt.Errorf("marshal litespec metadata: %w", err)
	}

	if err := os.WriteFile(dstMeta, out, 0644); err != nil {
		return fmt.Errorf("write litespec metadata: %w", err)
	}

	return nil
}

var dateOnlyRe = regexp.MustCompile(`^(\d{4})-(\d{2})-(\d{2})$`)
var quotedDateRe = regexp.MustCompile(`^["'](\d{4}-\d{2}-\d{2})["']$`)

func parseCreatedTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, fmt.Errorf("empty date")
	}

	if match := quotedDateRe.FindStringSubmatch(s); match != nil {
		s = match[1]
	}

	if match := dateOnlyRe.FindStringSubmatch(s); match != nil {
		return time.Parse("2006-01-02", s)
	}

	return time.Parse(time.RFC3339, s)
}

var archiveDateRe = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})-(.+)$`)

func synthesizeArchiveMeta(source, target string) error {
	dirName := filepath.Base(source)

	var created time.Time
	var err error

	if match := archiveDateRe.FindStringSubmatch(dirName); match != nil {
		created, err = time.Parse("2006-01-02", match[1])
		if err != nil {
			created = time.Now().UTC().Truncate(time.Second)
		}
	} else {
		created = time.Now().UTC().Truncate(time.Second)
	}

	liteMeta := struct {
		Schema  string `yaml:"schema"`
		Created string `yaml:"created"`
	}{
		Schema:  "spec-driven",
		Created: created.Format(time.RFC3339),
	}

	out, err := yaml.Marshal(&liteMeta)
	if err != nil {
		return fmt.Errorf("marshal synthesized metadata: %w", err)
	}

	dstMeta := filepath.Join(target, LiteSpecMetaFile)
	if err := os.WriteFile(dstMeta, out, 0644); err != nil {
		return fmt.Errorf("write synthesized metadata: %w", err)
	}

	return nil
}

func copyDir(source, target string) error {
	if err := os.MkdirAll(target, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	entries, err := os.ReadDir(source)
	if err != nil {
		return fmt.Errorf("read directory: %w", err)
	}

	for _, entry := range entries {
		src := filepath.Join(source, entry.Name())
		dst := filepath.Join(target, entry.Name())

		if entry.IsDir() {
			if err := copyDir(src, dst); err != nil {
				return err
			}
		} else {
			if err := copyFile(src, dst); err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(source, target string) error {
	data, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	if err := os.WriteFile(target, data, 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

func CheckConflicts(source, target string) ([]string, error) {
	var conflicts []string

	targetCanon := filepath.Join(target, "specs", "canon")
	if _, err := os.Stat(targetCanon); err == nil {
		conflicts = append(conflicts, targetCanon)
	}

	targetChanges := filepath.Join(target, "specs", "changes")
	if _, err := os.Stat(targetChanges); err == nil {
		entries, err := os.ReadDir(targetChanges)
		if err != nil {
			return nil, fmt.Errorf("read changes: %w", err)
		}
		for _, entry := range entries {
			if entry.Name() != "archive" {
				conflicts = append(conflicts, filepath.Join(targetChanges, entry.Name()))
			}
		}
	}

	return conflicts, nil
}
