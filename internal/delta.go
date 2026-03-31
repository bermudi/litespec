package internal

import (
	"fmt"
	"strings"
)

func ParseMainSpec(content string) (*Spec, error) {
	spec := &Spec{}
	lines := strings.Split(content, "\n")

	var current *SpecRequirement
	var body []string

	flush := func() {
		if current == nil {
			return
		}
		current.Content = strings.TrimSpace(strings.Join(body, "\n"))
		spec.Requirements = append(spec.Requirements, *current)
		current = nil
		body = nil
	}

	for _, line := range lines {
		if isH1(line) && spec.Capability == "" {
			flush()
			spec.Capability = strings.TrimSpace(line[1:])
		} else if isReqHeading(line) {
			flush()
			name := strings.TrimSpace(strings.TrimPrefix(line, "### Requirement:"))
			current = &SpecRequirement{Name: name}
			body = nil
		} else if current != nil {
			body = append(body, line)
		}
	}
	flush()

	if spec.Capability == "" {
		return nil, fmt.Errorf("missing capability heading (# <name>)")
	}

	return spec, nil
}

func ParseDeltaSpec(content string) (*DeltaSpec, error) {
	delta := &DeltaSpec{}
	lines := strings.Split(content, "\n")

	opSections := map[string]DeltaOperation{
		"## ADDED Requirements":    DeltaAdded,
		"## MODIFIED Requirements": DeltaModified,
		"## REMOVED Requirements":  DeltaRemoved,
		"## RENAMED Requirements":  DeltaRenamed,
	}

	var currentOp DeltaOperation
	var current *DeltaRequirement
	var body []string

	flush := func() {
		if current == nil {
			return
		}
		current.Content = strings.TrimSpace(strings.Join(body, "\n"))
		delta.Requirements = append(delta.Requirements, *current)
		current = nil
		body = nil
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if isH1(trimmed) && delta.Capability == "" {
			flush()
			delta.Capability = strings.TrimSpace(trimmed[1:])
			continue
		}

		if op, ok := opSections[trimmed]; ok {
			flush()
			currentOp = op
			continue
		}

		if isReqHeading(trimmed) {
			flush()
			namePart := strings.TrimSpace(strings.TrimPrefix(trimmed, "### Requirement:"))
			req := DeltaRequirement{Operation: currentOp}

			if currentOp == DeltaRenamed {
				parts := splitArrow(namePart)
				if len(parts) != 2 {
					return nil, fmt.Errorf("RENAMED requirement missing '→' separator: %q", namePart)
				}
				req.OldName = strings.TrimSpace(parts[0])
				req.Name = strings.TrimSpace(parts[1])
			} else {
				req.Name = namePart
			}

			current = &req
			body = nil
			continue
		}

		if current != nil {
			body = append(body, line)
		}
	}
	flush()

	return delta, nil
}

func MergeDelta(main *Spec, deltas []*DeltaSpec) (*Spec, error) {
	result := &Spec{
		Capability:   main.Capability,
		Requirements: make([]SpecRequirement, len(main.Requirements)),
	}
	copy(result.Requirements, main.Requirements)

	var renamed, removed, modified, added []DeltaRequirement
	for _, d := range deltas {
		for _, r := range d.Requirements {
			switch r.Operation {
			case DeltaRenamed:
				renamed = append(renamed, r)
			case DeltaRemoved:
				removed = append(removed, r)
			case DeltaModified:
				modified = append(modified, r)
			case DeltaAdded:
				added = append(added, r)
			}
		}
	}

	for _, r := range renamed {
		found := false
		for i := range result.Requirements {
			if result.Requirements[i].Name == r.OldName {
				result.Requirements[i].Name = r.Name
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("RENAMED: requirement %q not found", r.OldName)
		}
	}

	for _, r := range removed {
		idx := -1
		for i, req := range result.Requirements {
			if req.Name == r.Name {
				idx = i
				break
			}
		}
		if idx == -1 {
			return nil, fmt.Errorf("REMOVED: requirement %q not found", r.Name)
		}
		result.Requirements = append(result.Requirements[:idx], result.Requirements[idx+1:]...)
	}

	for _, r := range modified {
		found := false
		for i := range result.Requirements {
			if result.Requirements[i].Name == r.Name {
				result.Requirements[i].Content = r.Content
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("MODIFIED: requirement %q not found", r.Name)
		}
	}

	for _, r := range added {
		result.Requirements = append(result.Requirements, SpecRequirement{
			Name:    r.Name,
			Content: r.Content,
		})
	}

	return result, nil
}

func SerializeSpec(spec *Spec) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n", spec.Capability)

	for _, req := range spec.Requirements {
		b.WriteString("\n### Requirement: ")
		b.WriteString(req.Name)
		b.WriteString("\n")
		if req.Content != "" {
			b.WriteString("\n")
			b.WriteString(req.Content)
			b.WriteString("\n")
		}
	}

	return b.String()
}

func isH1(line string) bool {
	return strings.HasPrefix(line, "# ") && !strings.HasPrefix(line, "## ")
}

func isReqHeading(line string) bool {
	return strings.HasPrefix(line, "### Requirement:")
}

func splitArrow(s string) []string {
	if strings.Contains(s, "→") {
		return strings.SplitN(s, "→", 2)
	}
	if strings.Contains(s, "->") {
		return strings.SplitN(s, "->", 2)
	}
	return nil
}
