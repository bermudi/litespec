package internal

import (
	"fmt"
	"strings"
)

func ParseMainSpec(content string) (*Spec, error) {
	spec := &Spec{}
	lines := strings.Split(content, "\n")

	type sectionState int
	const (
		statePreamble sectionState = iota
		statePurpose
		stateRequirements
	)

	state := statePreamble
	var purposeLines []string
	var current *SpecRequirement
	var body []string

	flush := func() {
		if current == nil {
			return
		}
		preamble, scenarios := parseScenariosFromBody(body)
		current.Content = strings.TrimSpace(preamble)
		current.Scenarios = scenarios
		spec.Requirements = append(spec.Requirements, *current)
		current = nil
		body = nil
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if isH1(trimmed) && spec.Capability == "" {
			flush()
			spec.Capability = strings.TrimSpace(trimmed[1:])
			continue
		}

		if spec.Capability == "" {
			continue
		}

		if isH2(trimmed) {
			heading := strings.TrimSpace(trimmed[2:])

			if state == stateRequirements {
				continue
			}

			if heading == "Purpose" {
				if state != statePreamble {
					return nil, fmt.Errorf("## Purpose must appear before ## Requirements")
				}
				state = statePurpose
				purposeLines = nil
				continue
			}

			if heading == "Requirements" {
				flush()
				state = stateRequirements
				continue
			}

			return nil, fmt.Errorf("unexpected H2 section %q before ## Requirements; only ## Purpose is permitted", trimmed)
		}

		if state == statePurpose {
			purposeLines = append(purposeLines, line)
			continue
		}

		if isReqHeading(trimmed) {
			if state != stateRequirements {
				return nil, fmt.Errorf("requirement heading %q appears before ## Requirements section", trimmed)
			}
			flush()
			name := strings.TrimSpace(strings.TrimPrefix(trimmed, "### Requirement:"))
			current = &SpecRequirement{Name: name}
			body = nil
			continue
		}

		if current != nil {
			body = append(body, line)
		}
	}
	flush()

	if spec.Capability == "" {
		return nil, fmt.Errorf("missing capability heading (# <name>)")
	}

	if state != stateRequirements {
		return nil, fmt.Errorf("missing ## Requirements section")
	}

	if purposeLines != nil {
		spec.Purpose = strings.TrimSpace(strings.Join(purposeLines, "\n"))
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
		preamble, scenarios := parseScenariosFromBody(body)
		current.Content = strings.TrimSpace(preamble)
		current.Scenarios = scenarios
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
		Purpose:      main.Purpose,
		Requirements: make([]SpecRequirement, len(main.Requirements)),
	}
	for i, req := range main.Requirements {
		result.Requirements[i] = SpecRequirement{
			Name:      req.Name,
			Content:   req.Content,
			Scenarios: append([]Scenario(nil), req.Scenarios...),
		}
	}

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

	renamedOldNames := make(map[string]int)
	for _, r := range renamed {
		renamedOldNames[r.OldName]++
	}
	for name, count := range renamedOldNames {
		if count > 1 {
			return nil, fmt.Errorf("RENAMED: multiple renames target requirement %q", name)
		}
	}

	modifiedNames := make(map[string]int)
	for _, r := range modified {
		modifiedNames[r.Name]++
	}
	for name, count := range modifiedNames {
		if count > 1 {
			return nil, fmt.Errorf("MODIFIED: multiple deltas modify requirement %q", name)
		}
	}

	removedNames := make(map[string]int)
	for _, r := range removed {
		removedNames[r.Name]++
	}
	for name, count := range removedNames {
		if count > 1 {
			return nil, fmt.Errorf("REMOVED: multiple deltas remove requirement %q", name)
		}
	}

	addedNames := make(map[string]int)
	for _, r := range added {
		addedNames[r.Name]++
	}
	for name, count := range addedNames {
		if count > 1 {
			return nil, fmt.Errorf("ADDED: multiple deltas add requirement %q", name)
		}
	}

	allTargets := make(map[string][]string)
	for _, r := range removed {
		allTargets[r.Name] = append(allTargets[r.Name], "REMOVED")
	}
	for _, r := range modified {
		allTargets[r.Name] = append(allTargets[r.Name], "MODIFIED")
	}
	for _, r := range renamed {
		allTargets[r.OldName] = append(allTargets[r.OldName], "RENAMED")
	}
	for name, ops := range allTargets {
		if len(ops) > 1 {
			return nil, fmt.Errorf("conflicting operations on requirement %q: %s", name, strings.Join(ops, " + "))
		}
	}

	for _, r := range renamed {
		if r.OldName == r.Name {
			continue
		}
		found := false
		for i := range result.Requirements {
			if result.Requirements[i].Name == r.OldName {
				for _, existing := range result.Requirements {
					if existing.Name == r.Name {
						return nil, fmt.Errorf("RENAMED: new name %q already exists in spec", r.Name)
					}
				}
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
				result.Requirements[i].Scenarios = append([]Scenario(nil), r.Scenarios...)
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("MODIFIED: requirement %q not found", r.Name)
		}
	}

	for _, r := range added {
		for _, existing := range result.Requirements {
			if existing.Name == r.Name {
				return nil, fmt.Errorf("ADDED: requirement %q already exists in spec", r.Name)
			}
		}
		result.Requirements = append(result.Requirements, SpecRequirement{
			Name:      r.Name,
			Content:   r.Content,
			Scenarios: append([]Scenario(nil), r.Scenarios...),
		})
	}

	return result, nil
}

func SerializeSpec(spec *Spec) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n", spec.Capability)

	if spec.Purpose != "" {
		b.WriteString("\n## Purpose\n\n")
		b.WriteString(spec.Purpose)
		b.WriteString("\n")
	}

	b.WriteString("\n## Requirements\n")

	for _, req := range spec.Requirements {
		b.WriteString("\n### Requirement: ")
		b.WriteString(req.Name)
		b.WriteString("\n")
		if req.Content != "" {
			b.WriteString("\n")
			b.WriteString(req.Content)
			b.WriteString("\n")
		}
		for _, sc := range req.Scenarios {
			b.WriteString("\n#### Scenario: ")
			b.WriteString(sc.Name)
			b.WriteString("\n")
			if sc.Content != "" {
				b.WriteString("\n")
				b.WriteString(sc.Content)
				b.WriteString("\n")
			}
		}
	}

	return b.String()
}

func parseScenariosFromBody(body []string) (string, []Scenario) {
	var scenarios []Scenario
	var preamble []string
	var currentScenario *Scenario
	var scenarioBody []string

	flushScenario := func() {
		if currentScenario == nil {
			return
		}
		currentScenario.Content = strings.TrimSpace(strings.Join(scenarioBody, "\n"))
		scenarios = append(scenarios, *currentScenario)
		currentScenario = nil
		scenarioBody = nil
	}

	for _, line := range body {
		if isScenarioHeading(line) {
			flushScenario()
			name := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "#### Scenario:"))
			currentScenario = &Scenario{Name: name}
			scenarioBody = nil
			continue
		}
		if currentScenario != nil {
			scenarioBody = append(scenarioBody, line)
		} else {
			preamble = append(preamble, line)
		}
	}
	flushScenario()

	return strings.Join(preamble, "\n"), scenarios
}

func isH1(line string) bool {
	return strings.HasPrefix(line, "# ") && !strings.HasPrefix(line, "## ")
}

func isH2(line string) bool {
	return strings.HasPrefix(line, "## ") && !strings.HasPrefix(line, "### ")
}

func isReqHeading(line string) bool {
	return strings.HasPrefix(line, "### Requirement:")
}

func isScenarioHeading(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, "#### Scenario:")
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
