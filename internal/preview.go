package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type PreviewOperation struct {
	Type        string `json:"type"`
	Requirement string `json:"requirement"`
	OldName     string `json:"oldName,omitempty"`
}

type PreviewCapability struct {
	Name       string             `json:"name"`
	IsNew      bool               `json:"isNew"`
	Operations []PreviewOperation `json:"operations"`
}

type PreviewTotals struct {
	Capabilities int `json:"capabilities"`
	Added        int `json:"added"`
	Modified     int `json:"modified"`
	Removed      int `json:"removed"`
	Renamed      int `json:"renamed"`
}

type PreviewResult struct {
	Capabilities []PreviewCapability `json:"capabilities"`
	Totals       PreviewTotals       `json:"totals"`
}

func ComputePreviewResult(writes []PendingWrite, root string) (*PreviewResult, error) {
	if len(writes) == 0 {
		return &PreviewResult{}, nil
	}

	result := &PreviewResult{}
	for _, w := range writes {
		merged, err := ParseMainSpec(w.Content)
		if err != nil {
			return nil, fmt.Errorf("parsing merged spec for %s: %w", w.Capability, err)
		}

		mainSpecPath := filepath.Join(CanonPath(root), w.Capability, "spec.md")
		isNew := false
		var current *Spec
		mainData, err := os.ReadFile(mainSpecPath)
		if err != nil {
			isNew = true
			current = &Spec{Capability: w.Capability}
		} else {
			current, err = ParseMainSpec(string(mainData))
			if err != nil {
				return nil, fmt.Errorf("parsing current spec for %s: %w", w.Capability, err)
			}
		}

		ops := computeOperations(current, merged)
		if len(ops) == 0 && !isNew {
			continue
		}

		cap := PreviewCapability{
			Name:       w.Capability,
			IsNew:      isNew,
			Operations: ops,
		}
		result.Capabilities = append(result.Capabilities, cap)

		for _, op := range ops {
			result.Totals.Capabilities = len(result.Capabilities)
			switch op.Type {
			case "ADDED":
				result.Totals.Added++
			case "MODIFIED":
				result.Totals.Modified++
			case "REMOVED":
				result.Totals.Removed++
			case "RENAMED":
				result.Totals.Renamed++
			}
		}
	}

	result.Totals.Capabilities = len(result.Capabilities)
	return result, nil
}

func computeOperations(current, merged *Spec) []PreviewOperation {
	currentReqs := make(map[string]SpecRequirement, len(current.Requirements))
	for _, r := range current.Requirements {
		currentReqs[r.Name] = r
	}

	mergedReqs := make(map[string]SpecRequirement, len(merged.Requirements))
	for _, r := range merged.Requirements {
		mergedReqs[r.Name] = r
	}

	var ops []PreviewOperation

	// Detect RENAMED: merged name not in current, scan for a current name not in merged
	renamedFrom := make(map[string]string) // mergedName -> oldName
	currentUsed := make(map[string]bool)
	for _, mr := range merged.Requirements {
		if _, inCurrent := currentReqs[mr.Name]; !inCurrent {
			// This merged requirement wasn't in current — check if it's a rename
			for _, cr := range current.Requirements {
				if currentUsed[cr.Name] {
					continue
				}
				if _, inMerged := mergedReqs[cr.Name]; !inMerged {
					// cr.Name exists in current but not merged — candidate for rename source
					if normalizeContent(cr.Content) == normalizeContent(mr.Content) {
						renamedFrom[mr.Name] = cr.Name
						currentUsed[cr.Name] = true
						break
					}
				}
			}
		}
	}

	// Build operation list
	mergedSeen := make(map[string]bool)
	for _, mr := range merged.Requirements {
		mergedSeen[mr.Name] = true
		if oldName, ok := renamedFrom[mr.Name]; ok {
			ops = append(ops, PreviewOperation{
				Type:        "RENAMED",
				Requirement: mr.Name,
				OldName:     oldName,
			})
			continue
		}
		cr, inCurrent := currentReqs[mr.Name]
		if !inCurrent {
			ops = append(ops, PreviewOperation{
				Type:        "ADDED",
				Requirement: mr.Name,
			})
			continue
		}
		if normalizeContent(cr.Content) != normalizeContent(mr.Content) {
			ops = append(ops, PreviewOperation{
				Type:        "MODIFIED",
				Requirement: mr.Name,
			})
		}
	}

	// Detect REMOVED: current requirements not found in merged and not renamed
	for _, cr := range current.Requirements {
		if mergedSeen[cr.Name] {
			continue
		}
		if currentUsed[cr.Name] {
			continue
		}
		ops = append(ops, PreviewOperation{
			Type:        "REMOVED",
			Requirement: cr.Name,
		})
	}

	sortOperations(ops)
	return ops
}

func normalizeContent(s string) string {
	return strings.TrimSpace(s)
}

// Merge order: RENAMED first, then REMOVED, then MODIFIED, then ADDED.
func sortOperations(ops []PreviewOperation) {
	order := map[string]int{
		"RENAMED":  0,
		"REMOVED":  1,
		"MODIFIED": 2,
		"ADDED":    3,
	}
	sort.SliceStable(ops, func(i, j int) bool {
		return order[ops[i].Type] < order[ops[j].Type]
	})
}

func FormatPreviewText(result *PreviewResult) string {
	if len(result.Capabilities) == 0 {
		return "No changes to preview"
	}

	var b strings.Builder
	for _, cap := range result.Capabilities {
		status := "MODIFIED"
		if cap.IsNew {
			status = "NEW SPEC"
		}
		fmt.Fprintf(&b, "▸ %s (%s)\n", cap.Name, status)
		for _, op := range cap.Operations {
			switch op.Type {
			case "ADDED":
				fmt.Fprintf(&b, "  + ADDED: %s\n", op.Requirement)
			case "MODIFIED":
				fmt.Fprintf(&b, "  ~ MODIFIED: %s\n", op.Requirement)
			case "REMOVED":
				fmt.Fprintf(&b, "  - REMOVED: %s\n", op.Requirement)
			case "RENAMED":
				fmt.Fprintf(&b, "  → RENAMED: %s → %s\n", op.OldName, op.Requirement)
			}
		}
	}

	b.WriteString("═══════════════════════════════════════════════════════════\n")

	t := result.Totals
	capWord := "capabilities"
	if t.Capabilities == 1 {
		capWord = "capability"
	}
	fmt.Fprintf(&b, "%d %s affected • %d added • %d modified • %d removed • %d renamed\n",
		t.Capabilities, capWord, t.Added, t.Modified, t.Removed, t.Renamed)

	return b.String()
}

func FormatPreviewJSON(result *PreviewResult) ([]byte, error) {
	return json.MarshalIndent(result, "", "  ")
}
