package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type ResolvedDep struct {
	Name     string
	IsActive bool
}

func ListArchivedChanges(root string) ([]string, error) {
	archiveDir := ArchivePath(root)
	entries, err := os.ReadDir(archiveDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read archive directory: %w", err)
	}

	var names []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		names = append(names, ParseArchivedName(entry.Name()))
	}
	return names, nil
}

func ResolveDep(root, depName string) (ResolvedDep, bool) {
	if ChangeExists(root, depName) {
		return ResolvedDep{Name: depName, IsActive: true}, true
	}

	archived, err := ListArchivedChanges(root)
	if err != nil {
		return ResolvedDep{}, false
	}
	for _, name := range archived {
		if name == depName {
			return ResolvedDep{Name: depName, IsActive: false}, true
		}
	}

	return ResolvedDep{}, false
}

func ResolveDeps(root string, deps []string) ([]ResolvedDep, error) {
	if len(deps) == 0 {
		return nil, nil
	}

	var resolved []ResolvedDep
	for _, dep := range deps {
		r, found := ResolveDep(root, dep)
		if !found {
			return nil, fmt.Errorf("dependency %q not found", dep)
		}
		resolved = append(resolved, r)
	}
	return resolved, nil
}

func LoadDepMap(root string) (map[string][]string, error) {
	changes, err := ListChanges(root)
	if err != nil {
		return nil, err
	}

	depMap := make(map[string][]string)
	for _, ci := range changes {
		meta, err := ReadChangeMeta(root, ci.Name)
		if err != nil {
			continue
		}
		if len(meta.DependsOn) > 0 {
			depMap[ci.Name] = meta.DependsOn
		}
	}
	return depMap, nil
}

func GetDependents(root, name string) ([]string, error) {
	depMap, err := LoadDepMap(root)
	if err != nil {
		return nil, err
	}

	var dependents []string
	for changeName, deps := range depMap {
		for _, dep := range deps {
			if dep == name {
				dependents = append(dependents, changeName)
				break
			}
		}
	}
	return dependents, nil
}

func DetectCycles(depMap map[string][]string) [][]string {
	visited := make(map[string]bool)
	inStack := make(map[string]bool)
	var cycles [][]string

	var dfs func(name string, path []string)
	dfs = func(name string, path []string) {
		if inStack[name] {
			cycleStart := -1
			for i, n := range path {
				if n == name {
					cycleStart = i
					break
				}
			}
			if cycleStart >= 0 {
				cycle := append(append([]string{}, path[cycleStart:]...), name)
				cycles = append(cycles, cycle)
			}
			return
		}
		if visited[name] {
			return
		}

		visited[name] = true
		inStack[name] = true
		path = append(path, name)

		for _, dep := range depMap[name] {
			dfs(dep, path)
		}

		inStack[name] = false
	}

	for name := range depMap {
		dfs(name, nil)
	}

	return cycles
}

func TopologicalSort(changes []ChangeInfo, depMap map[string][]string) []ChangeInfo {
	inDegree := make(map[string]int)
	adj := make(map[string][]string)

	changeMap := make(map[string]ChangeInfo)
	for _, c := range changes {
		changeMap[c.Name] = c
		if _, ok := inDegree[c.Name]; !ok {
			inDegree[c.Name] = 0
		}
	}

	for name, deps := range depMap {
		if _, ok := changeMap[name]; !ok {
			continue
		}
		for _, dep := range deps {
			if _, ok := changeMap[dep]; ok {
				adj[dep] = append(adj[dep], name)
				inDegree[name]++
			}
		}
	}

	var level []string
	for name, deg := range inDegree {
		if deg == 0 {
			level = append(level, name)
		}
	}
	sort.Strings(level)

	var result []ChangeInfo
	for len(level) > 0 {
		var nextLevel []string
		for _, name := range level {
			result = append(result, changeMap[name])
			neighbors := append([]string{}, adj[name]...)
			sort.Strings(neighbors)
			for _, neighbor := range neighbors {
				inDegree[neighbor]--
				if inDegree[neighbor] == 0 {
					nextLevel = append(nextLevel, neighbor)
				}
			}
		}
		sort.Strings(nextLevel)
		level = nextLevel
	}

	for _, c := range changes {
		found := false
		for _, r := range result {
			if r.Name == c.Name {
				found = true
				break
			}
		}
		if !found {
			result = append(result, c)
		}
	}

	return result
}

func DetectOverlaps(root string, changes []ChangeInfo, depMap map[string][]string) []ValidationIssue {
	type changeTarget struct {
		name        string
		capability  string
		requirement string
		operation   DeltaOperation
	}

	var targets []changeTarget
	for _, ci := range changes {
		specsDir := ChangeSpecsPath(root, ci.Name)
		entries, err := os.ReadDir(specsDir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			capDir := filepath.Join(specsDir, entry.Name())
			files, readErr := os.ReadDir(capDir)
			if readErr != nil {
				continue
			}
			for _, f := range files {
				if filepath.Ext(f.Name()) != ".md" {
					continue
				}
				data, readErr := os.ReadFile(filepath.Join(capDir, f.Name()))
				if readErr != nil {
					continue
				}
				delta, parseErr := ParseDeltaSpec(string(data))
				if parseErr != nil {
					continue
				}
				for _, req := range delta.Requirements {
					if req.Operation == DeltaModified || req.Operation == DeltaRenamed {
						targets = append(targets, changeTarget{
							name:        ci.Name,
							capability:  entry.Name(),
							requirement: req.Name,
							operation:   req.Operation,
						})
					}
				}
			}
		}
	}

	hasDepEdge := func(a, b string) bool {
		for _, dep := range depMap[a] {
			if dep == b {
				return true
			}
		}
		for _, dep := range depMap[b] {
			if dep == a {
				return true
			}
		}
		return false
	}

	var issues []ValidationIssue
	seen := make(map[string]bool)
	for i, t1 := range targets {
		for j := i + 1; j < len(targets); j++ {
			t2 := targets[j]
			if t1.capability != t2.capability || t1.requirement != t2.requirement {
				continue
			}
			if t1.name == t2.name {
				continue
			}
			if hasDepEdge(t1.name, t2.name) {
				continue
			}
			pair := t1.name + ":" + t2.name
			if seen[pair] {
				continue
			}
			seen[pair] = true
			issues = append(issues, ValidationIssue{
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("changes %q and %q both target requirement %q in capability %q", t1.name, t2.name, t1.requirement, t1.capability),
			})
		}
	}

	return issues
}
