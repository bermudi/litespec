package internal

import (
	"fmt"
	"regexp"
	"strings"
)

func ParseMainSpec(content string) (*Spec, error) {
	spec := &Spec{}
	lines := strings.Split(content, "\n")

	for i := range lines {
		lines[i] = strings.TrimSuffix(lines[i], "\r")
	}

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
	var flushErr error

	flush := func() {
		if current == nil {
			return
		}
		preamble, scenarios, err := parseScenariosFromBody(body, current.Name)
		if err != nil {
			flushErr = err
			current = nil
			body = nil
			return
		}
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
			if flushErr != nil {
				return nil, flushErr
			}
			spec.Capability = strings.TrimSpace(trimmed[1:])
			continue
		}

		if spec.Capability == "" {
			continue
		}

		if isH2(trimmed) {
			heading := strings.TrimSpace(trimmed[2:])

			if state == stateRequirements {
				return nil, fmt.Errorf("unexpected H2 section %q inside ## Requirements", trimmed)
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
				if flushErr != nil {
					return nil, flushErr
				}
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
			if flushErr != nil {
				return nil, flushErr
			}
			name := strings.TrimSpace(strings.TrimPrefix(trimmed, "### Requirement:"))
			if strings.TrimSpace(name) == "" {
				return nil, fmt.Errorf("empty requirement name")
			}
			current = &SpecRequirement{Name: name}
			body = nil
			continue
		}

		if current != nil {
			body = append(body, line)
		}
	}
	flush()
	if flushErr != nil {
		return nil, flushErr
	}

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

func parseScenariosFromBody(body []string, reqName string) (string, []Scenario, error) {
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
			if strings.TrimSpace(name) == "" {
				return "", nil, fmt.Errorf("empty scenario name in requirement %q", reqName)
			}
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

	return strings.Join(preamble, "\n"), scenarios, nil
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

var keywordRe = regexp.MustCompile(`\b(SHALL|MUST)\b`)

func stripCodeBlocks(content string) string {
	noFenced := regexp.MustCompile("(?s)```.*?```").ReplaceAllString(content, "")
	noInline := regexp.MustCompile("`[^`]*`").ReplaceAllString(noFenced, "")
	return noInline
}

func containsKeyword(content string) bool {
	return keywordRe.MatchString(stripCodeBlocks(content))
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
