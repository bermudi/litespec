package internal

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var ghIssueEdit = func(root string, number int, bodyFile string) ([]byte, error) {
	cmd := exec.Command("gh", "issue", "edit", strconv.Itoa(number), "--body-file", bodyFile)
	cmd.Dir = root
	return cmd.CombinedOutput()
}

// IssueCheckResult reports the unit whose status checkbox the managed tick
// flipped.
type IssueCheckResult struct {
	Number     int
	Heading    string
	Occurrence int
}

// IssueCheckTicksOneUnit fetches the issue body through the existing gh read
// seam, flips exactly one unchecked status checkbox for the resolved unit,
// and writes the body back via gh issue edit. The written body must differ
// from the fetched body by exactly that one checkbox flip with the
// Base:/Branch: ownership lines byte-unchanged; any other delta, an
// ambiguous or unknown heading, an already-checked target, or a gh failure
// is a refusal that issues no write.
func IssueCheckTicksOneUnit(root string, issueNumber int, heading string, occurrence int) (IssueCheckResult, error) {
	body, source, err := queueBody(root, issueNumber, "")
	if err != nil {
		return IssueCheckResult{}, err
	}
	edited, result, err := tickUnitCheckbox(body, source, heading, occurrence)
	if err != nil {
		return IssueCheckResult{}, err
	}
	if err := writeIssueBody(root, issueNumber, edited); err != nil {
		return IssueCheckResult{}, err
	}
	result.Number = issueNumber
	return result, nil
}

// tickUnitCheckbox returns the body with exactly one unchecked status
// checkbox flipped to checked for the resolved unit. Every refusal happens
// before any write: the caller writes only when this returns nil.
func tickUnitCheckbox(body, source, heading string, occurrence int) (string, IssueCheckResult, error) {
	result := IssueCheckResult{Heading: heading}
	units, err := queueUnitsFromBody(body, source)
	if err != nil {
		return "", result, err
	}
	identities := queueUnitIdentities(units)
	index, err := resolveQueueUnitIndex(units, identities, source, heading, occurrence)
	if err != nil {
		return "", result, err
	}
	result.Occurrence = identities[index].Occurrence

	lines := strings.Split(body, "\n")
	start, end, ok := queueUnitLineRange(lines, index)
	if !ok {
		return "", result, fmt.Errorf("unit %q (occurrence %d) in %s could not be located in the body", heading, result.Occurrence, source)
	}
	flip := -1
	unchecked := 0
	openFence := ""
	for i := start + 1; i < end; i++ {
		if consumeMarkdownFenceLine(&openFence, lines[i]) {
			continue
		}
		if !isCheckboxLine(lines[i]) || isCheckedLine(strings.TrimSpace(lines[i])) {
			continue
		}
		unchecked++
		flip = i
	}
	if unchecked == 0 {
		if isCheckedUnit(units[index].Body) {
			return "", result, fmt.Errorf("unit %q (occurrence %d) in %s is already checked", heading, result.Occurrence, source)
		}
		return "", result, fmt.Errorf("unit %q (occurrence %d) in %s has no status checkbox", heading, result.Occurrence, source)
	}
	if unchecked > 1 {
		return "", result, fmt.Errorf("unit %q (occurrence %d) in %s has %d unchecked status checkboxes; exactly one is required", heading, result.Occurrence, source, unchecked)
	}

	edited := make([]string, len(lines))
	copy(edited, lines)
	flipped := strings.Replace(edited[flip], "- [ ]", "- [x]", 1)
	if flipped == edited[flip] {
		return "", result, fmt.Errorf("unit %q (occurrence %d) in %s: %q carries no unchecked checkbox to flip", heading, result.Occurrence, source, edited[flip])
	}
	edited[flip] = flipped
	if err := verifySingleCheckboxFlip(lines, edited, source); err != nil {
		return "", result, err
	}
	if err := verifyOwnershipLinesUnchanged(lines, edited, source); err != nil {
		return "", result, err
	}
	return strings.Join(edited, "\n"), result, nil
}

// queueUnitLineRange locates the [start, end) line range of the index-th
// unit section, using the same document order, fence handling, and isUnit
// filter as parseQueueUnits and queueUnitsFromBody.
func queueUnitLineRange(lines []string, index int) (int, int, bool) {
	seen := 0
	openFence := ""
	i := 0
	for i < len(lines) {
		if consumeMarkdownFenceLine(&openFence, lines[i]) {
			i++
			continue
		}
		if !strings.HasPrefix(lines[i], "## ") {
			i++
			continue
		}
		j := i + 1
		for j < len(lines) {
			if consumeMarkdownFenceLine(&openFence, lines[j]) {
				j++
				continue
			}
			if strings.HasPrefix(lines[j], "## ") {
				break
			}
			j++
		}
		if isUnit(queueUnit{Heading: strings.TrimSpace(lines[i][3:]), Body: lines[i+1 : j]}) {
			if seen == index {
				return i, j, true
			}
			seen++
		}
		i = j
	}
	return 0, 0, false
}

// verifySingleCheckboxFlip guards the managed edit: before and after must
// differ in exactly one line, and that line must be the same line with its
// unchecked checkbox flipped to checked.
func verifySingleCheckboxFlip(before, after []string, source string) error {
	if len(before) != len(after) {
		return fmt.Errorf("managed tick would change the line count of %s; refusing", source)
	}
	flipped := 0
	for i := range before {
		if before[i] == after[i] {
			continue
		}
		flipped++
		want := strings.Replace(before[i], "- [ ]", "- [x]", 1)
		if after[i] != want {
			return fmt.Errorf("managed tick would change line %d of %s to %q; only the checkbox flip to %q is permitted", i+1, source, after[i], want)
		}
	}
	if flipped != 1 {
		return fmt.Errorf("managed tick would change %d lines of %s; exactly one checkbox flip is permitted", flipped, source)
	}
	return nil
}

// queueOwnershipLines returns the Base:/Branch: ownership lines appearing
// before the first unit heading, in order.
func queueOwnershipLines(lines []string) []string {
	var out []string
	openFence := ""
	for _, line := range lines {
		if consumeMarkdownFenceLine(&openFence, line) {
			continue
		}
		if strings.HasPrefix(line, "## ") {
			return out
		}
		if strings.HasPrefix(line, "Base:") || strings.HasPrefix(line, "Branch:") {
			out = append(out, line)
		}
	}
	return out
}

func verifyOwnershipLinesUnchanged(before, after []string, source string) error {
	beforeLines := queueOwnershipLines(before)
	afterLines := queueOwnershipLines(after)
	if strings.Join(beforeLines, "\n") != strings.Join(afterLines, "\n") {
		return fmt.Errorf("managed tick would change the ownership lines of %s; refusing", source)
	}
	return nil
}

// writeIssueBody writes the edited body back through gh issue edit. A gh
// failure is a visible refusal naming the command; nothing is retried.
func writeIssueBody(root string, issueNumber int, body string) error {
	file, err := os.CreateTemp("", "litespec-issue-body-*.md")
	if err != nil {
		return fmt.Errorf("create body file: %w", err)
	}
	name := file.Name()
	defer os.Remove(name)
	if _, err := file.WriteString(body); err != nil {
		file.Close()
		return fmt.Errorf("write body file: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("write body file: %w", err)
	}
	command := fmt.Sprintf("gh issue edit %d --body-file %s", issueNumber, name)
	out, err := ghIssueEdit(root, issueNumber, name)
	if err != nil {
		return fmt.Errorf("%s failed: %v: %s", command, err, strings.TrimSpace(string(out)))
	}
	return nil
}
