package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type ghIssue struct {
	Number   int    `json:"number"`
	Title    string `json:"title"`
	Body     string `json:"body"`
	URL      string `json:"url"`
	Comments []struct {
		Body string `json:"body"`
	} `json:"comments"`
}

type queueUnit struct {
	Heading string
	Body    []string
	Depends []string
}

func ValidateGHIssueQueues(root string) (*ValidationResult, error) {
	result := &ValidationResult{Valid: true}

	if _, err := lookPathGh("gh"); err != nil {
		result.Warnings = append(result.Warnings, ValidationIssue{
			Severity:     SeverityWarning,
			Message:      "gh not available — issue queue not validated",
			StrictExempt: true,
		})
		return result, nil
	}

	out, err := ghIssueList(root)
	if err != nil {
		result.Warnings = append(result.Warnings, ValidationIssue{
			Severity:     SeverityWarning,
			Message:      fmt.Sprintf("gh issue list failed — issue queue not validated: %v", err),
			StrictExempt: true,
		})
		return result, nil
	}

	if len(out) == 0 {
		result.Warnings = append(result.Warnings, ValidationIssue{
			Severity:     SeverityWarning,
			Message:      "gh issue list returned empty output — issue queue not validated",
			StrictExempt: true,
		})
	} else {
		var issues []ghIssue
		if err := json.Unmarshal(out, &issues); err != nil {
			result.Warnings = append(result.Warnings, ValidationIssue{
				Severity:     SeverityWarning,
				Message:      fmt.Sprintf("parse gh issue list output: %v — issue queue not validated", err),
				StrictExempt: true,
			})
			return result, nil
		}

		for _, issue := range issues {
			source := fmt.Sprintf("GH issue #%d", issue.Number)
			units, unitIssues := ValidateQueueBody(issue.Body, source)
			result.UnitsCount += len(units)
			var commentBodies []string
			for _, c := range issue.Comments {
				commentBodies = append(commentBodies, c.Body)
			}
			applyQueueIssues(result, "GitHub comments", units, unitIssues, commentBodies)
		}
	}

	result.Valid = len(result.Errors) == 0
	return result, nil
}

func parseQueueUnits(body string) []queueUnit {
	lines := strings.Split(body, "\n")
	var units []queueUnit
	var current *queueUnit
	openFence := ""

	for _, line := range lines {
		if consumeMarkdownFenceLine(&openFence, line) {
			if current != nil {
				current.Body = append(current.Body, line)
			}
			continue
		}
		if strings.HasPrefix(line, "## ") {
			if current != nil {
				units = append(units, *current)
			}
			heading := strings.TrimSpace(line[3:])
			current = &queueUnit{Heading: heading}
			continue
		}
		if current != nil {
			current.Body = append(current.Body, line)
		}
	}

	if current != nil {
		units = append(units, *current)
	}

	return units
}

var lookPathBash = exec.LookPath
var lookPathGh = exec.LookPath
var queueBasePattern = regexp.MustCompile(`^[0-9a-fA-F]{40}([0-9a-fA-F]{24})?$`)
var queueBranchPattern = regexp.MustCompile(`^litespec/[a-z0-9]+(?:-[a-z0-9]+)*$`)
var vacuousVerifyPattern = regexp.MustCompile(`^(?:true|:|exit[ \t]+0)[ \t]*;?[ \t]*(?:#.*)?$`)

func isObviouslyVacuous(command string) bool {
	lines := strings.Split(command, "\n")
	var meaningful []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(strings.TrimLeft(line, " \t"), "#") {
			continue
		}
		meaningful = append(meaningful, trimmed)
	}
	if len(meaningful) == 0 {
		return true
	}
	if len(meaningful) != 1 {
		return false
	}
	return vacuousVerifyPattern.MatchString(meaningful[0])
}

var placeholderSet = map[string]bool{
	"-": true, "--": true, "n/a": true, "na": true,
	"none": true, "tbd": true, "todo": true, "null": true, "nil": true,
}

var identifiedQueueEntryPattern = regexp.MustCompile(`^[-*][ \t]+\[([^][ \t]+)\](?:[ \t]+(.*))?$`)

func queueUnitFieldLines(body []string, prefix string) ([]string, bool) {
	openFence := ""
	for i, line := range body {
		if consumeMarkdownFenceLine(&openFence, line) {
			continue
		}
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, prefix) {
			continue
		}
		var values []string
		if rest := strings.TrimSpace(strings.TrimPrefix(trimmed, prefix)); rest != "" {
			values = append(values, rest)
		}
		for _, nextLine := range body[i+1:] {
			next := strings.TrimSpace(nextLine)
			if next == "" {
				continue
			}
			if isQueueUnitFieldLine(next) || isCheckboxLine(next) {
				break
			}
			values = append(values, next)
		}
		return values, true
	}
	return nil, false
}

func isQueueUnitFieldLine(line string) bool {
	for _, prefix := range []string{
		"Read first:", "Constraints:", "Depends:", "Boundary:", "Done means:",
		"Scenarios:", "Risk cases:", "Verify:", "Evidence:",
	} {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}
	return false
}

func queueUnitFieldCount(body []string, prefix string) int {
	count := 0
	openFence := ""
	for _, line := range body {
		if consumeMarkdownFenceLine(&openFence, line) {
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(line), prefix) {
			count++
		}
	}
	return count
}

func validateUnitScenarioMapping(unit queueUnit, source string) []ValidationIssue {
	doneLines, doneFound := queueUnitFieldLines(unit.Body, "Done means:")
	if !doneFound {
		return nil
	}

	var issues []ValidationIssue
	fail := func(message string) {
		issues = append(issues, ValidationIssue{
			Severity: SeverityError,
			Message:  fmt.Sprintf("%s: unit %q %s", source, unit.Heading, message),
			File:     source,
		})
	}

	if queueUnitFieldCount(unit.Body, "Done means:") > 1 {
		fail("has duplicate Done means: fields (only one allowed)")
	}

	clauseIDs := make(map[string]bool)
	for _, line := range doneLines {
		match := identifiedQueueEntryPattern.FindStringSubmatch(line)
		if match == nil || strings.TrimSpace(match[2]) == "" {
			fail("must contain a nonempty identified Done means clause bullet")
			continue
		}
		id := match[1]
		if clauseIDs[id] {
			fail(fmt.Sprintf("has duplicate Done means clause ID %q", id))
			continue
		}
		clauseIDs[id] = true
	}
	if len(doneLines) == 0 {
		fail("must contain at least one identified Done means clause")
	}

	scenarioLines, scenariosFound := queueUnitFieldLines(unit.Body, "Scenarios:")
	if !scenariosFound {
		fail("missing Scenarios: mapping")
		return issues
	}
	if queueUnitFieldCount(unit.Body, "Scenarios:") > 1 {
		fail("has duplicate Scenarios: fields (only one allowed)")
	}
	if len(scenarioLines) == 0 {
		fail("Scenarios: mapping must be nonempty")
		return issues
	}

	mapped := make(map[string]bool)
	for _, line := range scenarioLines {
		match := identifiedQueueEntryPattern.FindStringSubmatch(line)
		if match == nil {
			fail("Scenarios: must contain identified mapping bullets")
			continue
		}
		id := match[1]
		if strings.TrimSpace(match[2]) == "" {
			fail(fmt.Sprintf("scenario mapping for clause ID %q must name a test scenario", id))
			continue
		}
		if !clauseIDs[id] {
			fail(fmt.Sprintf("scenario mapping references unknown Done means clause ID %q", id))
			continue
		}
		mapped[id] = true
	}
	for id := range clauseIDs {
		if !mapped[id] {
			fail(fmt.Sprintf("Done means clause ID %q has no scenario mapping", id))
		}
	}
	return issues
}

func isPlaceholderValue(s string) bool {
	return placeholderSet[strings.ToLower(strings.TrimSpace(s))]
}

func hasBulletAfter(body []string, idx int) bool {
	for j := idx + 1; j < len(body); j++ {
		t := strings.TrimSpace(body[j])
		if t == "" {
			continue
		}
		if strings.HasPrefix(t, "```") {
			break
		}
		if strings.HasPrefix(t, "Done means:") || strings.HasPrefix(t, "Verify:") || strings.HasPrefix(t, "Depends:") || strings.HasPrefix(t, "Read first:") || strings.HasPrefix(t, "Constraints:") || isCheckboxLine(t) {
			break
		}
		if strings.HasPrefix(t, "- ") || strings.HasPrefix(t, "* ") {
			return true
		}
		break
	}
	return false
}

func validateOptionalField(body []string, fieldName string, count int, idx int, rest string, source, heading string) []ValidationIssue {
	if count > 1 {
		return []ValidationIssue{{
			Severity: SeverityError,
			Message:  fmt.Sprintf("%s: unit %q has duplicate %s: (only one allowed)", source, heading, fieldName),
			File:     source,
		}}
	}
	if count == 1 {
		trimmed := strings.TrimSpace(rest)
		if isPlaceholderValue(trimmed) {
			return []ValidationIssue{{
				Severity: SeverityError,
				Message:  fmt.Sprintf("%s: unit %q %s: must be nonempty — omit rather than placeholder %q", source, heading, fieldName, trimmed),
				File:     source,
			}}
		}
		if trimmed == "" && !hasBulletAfter(body, idx) {
			return []ValidationIssue{{
				Severity: SeverityError,
				Message:  fmt.Sprintf("%s: unit %q %s: must be nonempty — omit rather than placeholder", source, heading, fieldName),
				File:     source,
			}}
		}
	}
	return nil
}

var evidenceCommitPattern = regexp.MustCompile(`^[0-9a-fA-F]{40}(?:[0-9a-fA-F]{24})?$`)
var preEvidenceScopePattern = regexp.MustCompile(`^Pre-evidence scope: this command exited (-?\d+) at ([0-9a-fA-F]{40}(?:[0-9a-fA-F]{24})?); nothing else is inferred\.$`)
var postEvidenceScopePattern = regexp.MustCompile(`^Post-evidence scope: this command exited (-?\d+) at ([0-9a-fA-F]{40}(?:[0-9a-fA-F]{24})?); nothing else is inferred\.$`)

func isCheckedLine(trimmed string) bool {
	for _, cb := range []string{"- [x]", "- [X]"} {
		if trimmed == cb || strings.HasPrefix(trimmed, cb+" ") {
			return true
		}
	}
	return false
}

func isCheckedUnit(body []string) bool {
	openFence := ""
	for _, line := range body {
		if consumeMarkdownFenceLine(&openFence, line) {
			continue
		}
		if isCheckedLine(strings.TrimSpace(line)) {
			return true
		}
	}
	return false
}

func unitVerifyCommand(body []string) string {
	for i, line := range body {
		if !strings.HasPrefix(line, "Verify:") {
			continue
		}
		rest := strings.TrimSpace(line[len("Verify:"):])
		firstBacktick := strings.Index(rest, "`")
		lastBacktick := strings.LastIndex(rest, "`")
		inline := ""
		if firstBacktick >= 0 && lastBacktick > firstBacktick {
			inline = strings.TrimSpace(rest[firstBacktick+1 : lastBacktick])
		}
		for j := i + 1; j < len(body); j++ {
			trimmed := strings.TrimSpace(body[j])
			if trimmed == "" {
				continue
			}
			delimiter := fenceDelimiter(trimmed)
			if delimiter == "" {
				return inline
			}
			var blockLines []string
			for k := j + 1; k < len(body); k++ {
				if strings.TrimSpace(body[k]) == delimiter {
					return strings.Join(blockLines, "\n")
				}
				blockLines = append(blockLines, body[k])
			}
			break
		}
		return inline
	}
	return ""
}

func extractEvidenceText(body []string) string {
	seenVerify := false
	verifyFence := ""
	for i, line := range body {
		trimmed := strings.TrimSpace(line)
		if !seenVerify {
			if strings.HasPrefix(line, "Verify:") {
				seenVerify = true
			}
			continue
		}
		if consumeMarkdownFenceLine(&verifyFence, line) {
			continue
		}
		if isCheckboxLine(trimmed) {
			break
		}
		if !strings.HasPrefix(trimmed, "Evidence:") {
			continue
		}
		var parts []string
		if rest := strings.TrimSpace(strings.TrimPrefix(trimmed, "Evidence:")); rest != "" {
			parts = append(parts, rest)
		}
		receiptFence := ""
		for j := i + 1; j < len(body); j++ {
			t := strings.TrimSpace(body[j])
			if consumeMarkdownFenceLine(&receiptFence, body[j]) {
				parts = append(parts, body[j])
				continue
			}
			if isCheckboxLine(t) {
				break
			}
			parts = append(parts, body[j])
		}
		return strings.Join(parts, "\n")
	}
	return ""
}

type evidenceCursor struct {
	lines []string
	at    int
}

func newEvidenceCursor(text string) *evidenceCursor {
	normalized := strings.ReplaceAll(text, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	return &evidenceCursor{lines: lines}
}

func (c *evidenceCursor) consumeExactLines(value string) bool {
	wanted := strings.Split(strings.ReplaceAll(value, "\r\n", "\n"), "\n")
	if c.at+len(wanted) > len(c.lines) {
		return false
	}
	for i, line := range wanted {
		if c.lines[c.at+i] != line {
			return false
		}
	}
	c.at += len(wanted)
	return true
}

func (c *evidenceCursor) consumeField(label string) (string, bool) {
	if c.at >= len(c.lines) {
		return "", false
	}
	prefix := label + ":"
	if !strings.HasPrefix(c.lines[c.at], prefix) {
		return "", false
	}
	value := strings.TrimSpace(strings.TrimPrefix(c.lines[c.at], prefix))
	c.at++
	return value, true
}

func fenceDelimiter(line string) string {
	trimmed := strings.TrimSpace(line)
	count := 0
	for count < len(trimmed) && trimmed[count] == '`' {
		count++
	}
	if count < 3 {
		return ""
	}
	return trimmed[:count]
}

func consumeMarkdownFenceLine(open *string, line string) bool {
	trimmed := strings.TrimSpace(line)
	if *open != "" {
		if trimmed == *open {
			*open = ""
		}
		return true
	}
	delimiter := fenceDelimiter(trimmed)
	if delimiter == "" {
		return false
	}
	*open = delimiter
	return true
}

func (c *evidenceCursor) consumeFence() (string, bool) {
	if c.at >= len(c.lines) {
		return "", false
	}
	delimiter := fenceDelimiter(c.lines[c.at])
	if delimiter == "" {
		return "", false
	}
	c.at++
	start := c.at
	for c.at < len(c.lines) {
		if strings.TrimSpace(c.lines[c.at]) == delimiter {
			output := strings.Join(c.lines[start:c.at], "\n")
			c.at++
			return output, true
		}
		c.at++
	}
	return "", false
}

func evidencePayload(text, verifyCmd string) string {
	normalized := strings.ReplaceAll(text, "\r\n", "\n")
	if strings.HasPrefix(strings.TrimLeft(normalized, "\n"), verifyCmd+"\n") {
		return normalized
	}
	lines := strings.Split(normalized, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) != "Evidence:" {
			continue
		}
		return strings.Join(lines[i+1:], "\n")
	}
	return text
}

func evidenceReceiptIssues(text, verifyCmd, expectedDigest, source, heading string) []ValidationIssue {
	if strings.TrimSpace(text) == "" {
		return []ValidationIssue{{
			Severity: SeverityError,
			Message:  fmt.Sprintf("%s: checked unit %q missing Evidence receipt", source, heading),
			File:     source,
		}}
	}

	var issues []ValidationIssue
	fail := func(reason string) {
		issues = append(issues, ValidationIssue{
			Severity: SeverityError,
			Message:  fmt.Sprintf("%s: checked unit %q Evidence receipt %s", source, heading, reason),
			File:     source,
		})
	}

	cursor := newEvidenceCursor(evidencePayload(text, verifyCmd))
	if verifyCmd == "" || !cursor.consumeExactLines(verifyCmd) {
		fail("must quote the Verify command verbatim")
		return issues
	}

	digestText, ok := cursor.consumeField("unit digest")
	if !ok {
		fail("must include a `unit digest:` field between the Verify command and pre sha")
		return issues
	}
	if !unitDigestPattern.MatchString(digestText) {
		fail("unit digest must be 64 lowercase hexadecimal characters")
	} else if digestText != expectedDigest {
		fail(fmt.Sprintf("unit digest mismatch: expected %s, actual %s — recompute with litespec or amend the contract", expectedDigest, digestText))
	}

	preSHA, ok := cursor.consumeField("pre sha")
	if !ok {
		fail("fields must appear in order beginning with pre sha")
		return issues
	}
	if !evidenceCommitPattern.MatchString(preSHA) {
		fail("pre sha must be a full 40- or 64-character hexadecimal commit ID")
	}

	preStatusText, ok := cursor.consumeField("pre exit status")
	if !ok {
		fail("fields must appear in order: pre exit status must follow pre sha")
		return issues
	}
	preStatus, err := strconv.Atoi(preStatusText)
	if err != nil {
		fail("pre exit status must be an integer")
	} else if preStatus == 0 {
		fail("pre exit status must be non-zero")
	}

	preOutput, ok := cursor.consumeFence()
	if !ok || strings.TrimSpace(preOutput) == "" {
		fail("must include pre raw command output, or `<no output>`, in a fenced block")
		return issues
	}

	if cursor.at >= len(cursor.lines) {
		fail("must include a matching Pre-evidence scope line")
		return issues
	}
	preScope := preEvidenceScopePattern.FindStringSubmatch(cursor.lines[cursor.at])
	cursor.at++
	if preScope == nil {
		fail("must include a matching Pre-evidence scope line")
	} else {
		if preScope[1] != preStatusText {
			fail("pre scope line status must match pre exit status")
		}
		if !strings.EqualFold(preScope[2], preSHA) {
			fail("pre scope line sha must match pre sha")
		}
	}

	postSHA, ok := cursor.consumeField("post sha")
	if !ok {
		fail("fields must appear in order: post sha must follow the pre scope line")
		return issues
	}
	if !evidenceCommitPattern.MatchString(postSHA) {
		fail("post sha must be a full 40- or 64-character hexadecimal commit ID")
	}
	if evidenceCommitPattern.MatchString(preSHA) && strings.EqualFold(preSHA, postSHA) {
		fail("pre and post sha must differ")
	}

	postStatusText, ok := cursor.consumeField("post exit status")
	if !ok {
		fail("fields must appear in order: post exit status must follow post sha")
		return issues
	}
	if postStatusText != "0" {
		fail("post exit status must be 0")
	}

	postOutput, ok := cursor.consumeFence()
	if !ok || strings.TrimSpace(postOutput) == "" {
		fail("must include post raw command output, or `<no output>`, in a fenced block")
		return issues
	}

	if cursor.at >= len(cursor.lines) {
		fail("must include a matching Post-evidence scope line")
		return issues
	}
	postScope := postEvidenceScopePattern.FindStringSubmatch(cursor.lines[cursor.at])
	cursor.at++
	if postScope == nil {
		fail("must include a matching Post-evidence scope line")
	} else {
		if postScope[1] != postStatusText {
			fail("post scope line status must match post exit status")
		}
		if !strings.EqualFold(postScope[2], postSHA) {
			fail("post scope line sha must match post sha")
		}
	}

	for cursor.at < len(cursor.lines) && strings.TrimSpace(cursor.lines[cursor.at]) == "" {
		cursor.at++
	}
	if cursor.at != len(cursor.lines) {
		fail("fields must appear in order with no unexpected trailing content")
	}
	return issues
}

func validateCheckedUnitEvidence(unit queueUnit, source string) []ValidationIssue {
	if !isCheckedUnit(unit.Body) {
		return nil
	}
	return evidenceReceiptIssues(extractEvidenceText(unit.Body), unitVerifyCommand(unit.Body), unitContractDigest(unit), source, unit.Heading)
}

func commentSatisfiesEvidence(heading, verifyCmd, expectedDigest string, comments []string) bool {
	_, ok := matchingEvidenceComment(heading, verifyCmd, expectedDigest, comments, nil)
	return ok
}

func matchingEvidenceComment(heading, verifyCmd, expectedDigest string, comments []string, used map[int]bool) (int, bool) {
	for i, c := range comments {
		if used[i] {
			continue
		}
		if !commentNamesUnit(c, heading) {
			continue
		}
		evidenceText := c
		if afterHeading, ok := commentTextAfterUnitHeading(c, heading); ok {
			evidenceText = afterHeading
		}
		if len(evidenceReceiptIssues(evidenceText, verifyCmd, expectedDigest, "comment", heading)) == 0 {
			return i, true
		}
	}
	return 0, false
}

func commentTextAfterUnitHeading(comment, heading string) (string, bool) {
	lines := strings.Split(strings.ReplaceAll(comment, "\r\n", "\n"), "\n")
	openFence := ""
	for i, line := range lines {
		if consumeMarkdownFenceLine(&openFence, line) {
			continue
		}
		trimmed := strings.TrimSpace(line)
		if trimmed == heading || trimmed == "## "+heading {
			return strings.Join(lines[i+1:], "\n"), true
		}
	}
	return "", false
}

func commentNamesUnit(comment, heading string) bool {
	openFence := ""
	for _, line := range strings.Split(strings.ReplaceAll(comment, "\r\n", "\n"), "\n") {
		trimmed := strings.TrimSpace(line)
		if consumeMarkdownFenceLine(&openFence, line) {
			continue
		}
		if trimmed == "Evidence:" || strings.HasPrefix(trimmed, "Evidence: ") {
			return false
		}
		if trimmed == heading || trimmed == "## "+heading {
			return true
		}
	}
	return false
}

func applyQueueIssues(result *ValidationResult, commentSource string, units []queueUnit, unitIssues []ValidationIssue, comments []string) {
	usedComments := make(map[int]bool)
	commentEvidence := make(map[int]bool)
	identities := queueUnitIdentities(units)
	for unitIndex, unit := range units {
		if !isCheckedUnit(unit.Body) {
			continue
		}
		if len(validateCheckedUnitEvidence(unit, "queue")) == 0 {
			continue
		}
		commentIndex, ok := matchingEvidenceCommentForUnit(
			identities[unitIndex],
			unit,
			unitContractDigest(unit),
			units,
			comments,
			usedComments,
		)
		if !ok {
			continue
		}
		usedComments[commentIndex] = true
		commentEvidence[unitIndex] = true
	}

	for _, iss := range unitIssues {
		if strings.Contains(iss.Message, "Evidence receipt") && commentEvidence[iss.queueUnitIndex] {
			continue
		}
		if iss.Severity == SeverityWarning {
			result.Warnings = append(result.Warnings, iss)
		} else {
			result.Errors = append(result.Errors, iss)
		}
	}

	unresolved, requestErrors := unresolvedRebuildRequests(units, comments)
	for _, err := range requestErrors {
		result.Errors = append(result.Errors, ValidationIssue{
			Severity: SeverityError,
			Message:  fmt.Sprintf("queue rebuild routing: %v", err),
			File:     commentSource,
		})
	}
	for _, identity := range unresolved {
		result.Errors = append(result.Errors, ValidationIssue{
			Severity: SeverityError,
			Message: fmt.Sprintf(
				"unit occurrence %d with heading %q has an unresolved rebuild request",
				identity.Occurrence,
				identity.Heading,
			),
			File: commentSource,
		})
	}
}

func matchingEvidenceCommentForUnit(
	identity queueUnitIdentity,
	unit queueUnit,
	expectedDigest string,
	units []queueUnit,
	comments []string,
	used map[int]bool,
) (int, bool) {
	for i, comment := range comments {
		if used[i] {
			continue
		}
		commentIdentity, kind, _, err := parseRebuildComment(comment, units)
		if err == nil && kind == rebuildCommentEvidence && commentIdentity == identity {
			return i, true
		}
	}
	return matchingEvidenceComment(unit.Heading, unitVerifyCommand(unit.Body), expectedDigest, comments, used)
}

var ghIssueView = func(root string, number int) ([]byte, error) {
	cmd := exec.Command("gh", "issue", "view", strconv.Itoa(number),
		"--json", "number,title,body,url,comments")
	cmd.Dir = root
	return cmd.Output()
}

var ghIssueList = func(root string) ([]byte, error) {
	cmd := exec.Command("gh", "issue", "list",
		"--label", "litespec",
		"--state", "open",
		"--json", "number,title,body,url,comments",
		"--limit", "10000",
	)
	cmd.Dir = root
	return cmd.Output()
}

func lintVerifyShell(block string, source string, unitHeading string) []ValidationIssue {
	// Blank is "Verify block is empty" (existing contract, tested); isObviouslyVacuous handles comment-only and single true/: /exit 0.
	if strings.TrimSpace(block) == "" {
		return []ValidationIssue{{
			Severity: SeverityError,
			Message:  fmt.Sprintf("%s: unit %q Verify block is empty", source, unitHeading),
			File:     source,
		}}
	}

	if isObviouslyVacuous(block) {
		return []ValidationIssue{{
			Severity: SeverityError,
			Message:  fmt.Sprintf("%s: unit %q Verify command is obviously vacuous; assert the unit outcome", source, unitHeading),
			File:     source,
		}}
	}

	bashPath, err := lookPathBash("bash")
	if err != nil {
		return []ValidationIssue{{
			Severity: SeverityWarning,
			Message:  fmt.Sprintf("%s: unit %q Verify block not syntax-checked (bash unavailable)", source, unitHeading),
			File:     source,
		}}
	}

	cmd := exec.Command(bashPath, "-n", "-c", block)
	out, err := cmd.CombinedOutput()
	if err != nil {
		trimmed := strings.TrimSpace(string(out))
		return []ValidationIssue{{
			Severity: SeverityError,
			Message:  fmt.Sprintf("%s: unit %q Verify shell syntax error: %s", source, unitHeading, trimmed),
			File:     source,
		}}
	}

	return nil
}

func isUnit(unit queueUnit) bool {
	openFence := ""
	for _, line := range unit.Body {
		if consumeMarkdownFenceLine(&openFence, line) {
			continue
		}
		if strings.HasPrefix(line, "Done means:") || strings.HasPrefix(line, "Verify:") {
			return true
		}
	}
	return false
}

func isCheckboxLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	for _, checkbox := range []string{"- [ ]", "- [x]", "- [X]"} {
		if trimmed == checkbox || strings.HasPrefix(trimmed, checkbox+" ") {
			return true
		}
	}
	return false
}

func parseDepends(body []string) []string {
	var deps []string
	openFence := ""
	for _, line := range body {
		if consumeMarkdownFenceLine(&openFence, line) {
			continue
		}
		if !strings.HasPrefix(line, "Depends:") {
			continue
		}
		rest := strings.TrimSpace(line[len("Depends:"):])
		for _, part := range strings.Split(rest, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				deps = append(deps, part)
			}
		}
		break
	}
	return deps
}

func validateQueueOwnership(body string, source string) []ValidationIssue {
	type ownershipLine struct {
		value         string
		beforeHeading bool
	}

	var bases []ownershipLine
	var branches []ownershipLine
	beforeHeading := true
	openFence := ""
	for _, line := range strings.Split(body, "\n") {
		if consumeMarkdownFenceLine(&openFence, line) {
			continue
		}
		if strings.HasPrefix(line, "## ") {
			beforeHeading = false
		}
		if strings.HasPrefix(line, "Base:") {
			bases = append(bases, ownershipLine{
				value:         strings.TrimSpace(strings.TrimPrefix(line, "Base:")),
				beforeHeading: beforeHeading,
			})
		}
		if strings.HasPrefix(line, "Branch:") {
			branches = append(branches, ownershipLine{
				value:         strings.TrimSpace(strings.TrimPrefix(line, "Branch:")),
				beforeHeading: beforeHeading,
			})
		}
	}

	var issues []ValidationIssue
	if len(bases) != 1 {
		issues = append(issues, ValidationIssue{
			Severity: SeverityError,
			Message:  fmt.Sprintf("%s: expected exactly one Base: ownership line in the queue", source),
			File:     source,
		})
	} else if !bases[0].beforeHeading {
		issues = append(issues, ValidationIssue{
			Severity: SeverityError,
			Message:  fmt.Sprintf("%s: Base: ownership line must appear before the first ## heading", source),
			File:     source,
		})
	} else if !queueBasePattern.MatchString(bases[0].value) {
		issues = append(issues, ValidationIssue{
			Severity: SeverityError,
			Message:  fmt.Sprintf("%s: Base: must contain a full 40- or 64-character hexadecimal commit ID", source),
			File:     source,
		})
	}

	if len(branches) != 1 {
		issues = append(issues, ValidationIssue{
			Severity: SeverityError,
			Message:  fmt.Sprintf("%s: expected exactly one Branch: ownership line in the queue", source),
			File:     source,
		})
	} else if !branches[0].beforeHeading {
		issues = append(issues, ValidationIssue{
			Severity: SeverityError,
			Message:  fmt.Sprintf("%s: Branch: ownership line must appear before the first ## heading", source),
			File:     source,
		})
	} else if !queueBranchPattern.MatchString(branches[0].value) {
		issues = append(issues, ValidationIssue{
			Severity: SeverityError,
			Message:  fmt.Sprintf("%s: Branch: must match litespec/<kebab-change-name>", source),
			File:     source,
		})
	}

	return issues
}

func ValidateQueueBody(body string, source string) ([]queueUnit, []ValidationIssue) {
	all := parseQueueUnits(body)
	units := make([]queueUnit, 0, len(all))
	for _, u := range all {
		if isUnit(u) {
			units = append(units, u)
		}
	}
	issues := validateQueueOwnership(body, source)

	for unitIndex, unit := range units {
		if strings.TrimSpace(unit.Heading) == "" {
			issues = append(issues, ValidationIssue{
				Severity: SeverityError,
				Message:  fmt.Sprintf("%s: empty unit heading", source),
				File:     source,
			})
		}

		doneFound := false
		verifyFound := false
		checkboxFound := false
		inlineVerify := false
		inlineVerifyContent := ""
		hasFencedBlock := false
		var verifyBlock string
		fencedBlock := ""
		constraintsCount := 0
		readFirstCount := 0
		constraintsIdx := -1
		readFirstIdx := -1
		constraintsRest := ""
		readFirstRest := ""

		for i, line := range unit.Body {
			if consumeMarkdownFenceLine(&fencedBlock, line) {
				continue
			}
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "Constraints:") {
				constraintsCount++
				if constraintsIdx == -1 {
					constraintsIdx = i
					constraintsRest = strings.TrimSpace(strings.TrimPrefix(trimmed, "Constraints:"))
				}
			}
			if strings.HasPrefix(trimmed, "Read first:") {
				readFirstCount++
				if readFirstIdx == -1 {
					readFirstIdx = i
					readFirstRest = strings.TrimSpace(strings.TrimPrefix(trimmed, "Read first:"))
				}
			}
			if strings.HasPrefix(line, "Done means:") {
				doneFound = true
			}
			if !verifyFound && strings.HasPrefix(line, "Verify:") {
				verifyFound = true
				rest := strings.TrimSpace(line[len("Verify:"):])
				firstBacktick := strings.Index(rest, "`")
				lastBacktick := strings.LastIndex(rest, "`")
				if firstBacktick >= 0 && lastBacktick > firstBacktick {
					span := strings.TrimSpace(rest[firstBacktick+1 : lastBacktick])
					if span != "" {
						inlineVerify = true
						inlineVerifyContent = span
					}
				}
				for j := i + 1; j < len(unit.Body); j++ {
					delimiter := fenceDelimiter(unit.Body[j])
					if delimiter != "" {
						var blockLines []string
						for k := j + 1; k < len(unit.Body); k++ {
							if strings.TrimSpace(unit.Body[k]) == delimiter {
								hasFencedBlock = true
								break
							}
							blockLines = append(blockLines, unit.Body[k])
						}
						if hasFencedBlock {
							verifyBlock = strings.Join(blockLines, "\n")
						}
						break
					}
				}
			}
			if isCheckboxLine(line) {
				checkboxFound = true
			}
		}

		issues = append(issues, validateOptionalField(unit.Body, "Constraints", constraintsCount, constraintsIdx, constraintsRest, source, unit.Heading)...)
		issues = append(issues, validateOptionalField(unit.Body, "Read first", readFirstCount, readFirstIdx, readFirstRest, source, unit.Heading)...)
		issues = append(issues, validateUnitScenarioMapping(unit, source)...)

		if !doneFound {
			issues = append(issues, ValidationIssue{
				Severity: SeverityError,
				Message:  fmt.Sprintf("%s: unit %q missing Done means:", source, unit.Heading),
				File:     source,
			})
		}
		if !verifyFound {
			issues = append(issues, ValidationIssue{
				Severity: SeverityError,
				Message:  fmt.Sprintf("%s: unit %q missing Verify:", source, unit.Heading),
				File:     source,
			})
		} else if !inlineVerify && !hasFencedBlock {
			issues = append(issues, ValidationIssue{
				Severity: SeverityError,
				Message:  fmt.Sprintf("%s: unit %q Verify: not followed by a command or fenced code block", source, unit.Heading),
				File:     source,
			})
		} else if hasFencedBlock {
			issues = append(issues, lintVerifyShell(verifyBlock, source, unit.Heading)...)
		} else if inlineVerify {
			if isObviouslyVacuous(inlineVerifyContent) {
				issues = append(issues, ValidationIssue{
					Severity: SeverityError,
					Message:  fmt.Sprintf("%s: unit %q Verify command is obviously vacuous; assert the unit outcome", source, unit.Heading),
					File:     source,
				})
			}
		}
		if !checkboxFound {
			issues = append(issues, ValidationIssue{
				Severity: SeverityError,
				Message:  fmt.Sprintf("%s: unit %q missing checkbox", source, unit.Heading),
				File:     source,
			})
		}
		evidenceIssues := validateCheckedUnitEvidence(unit, source)
		for i := range evidenceIssues {
			evidenceIssues[i].queueUnitIndex = unitIndex
		}
		issues = append(issues, evidenceIssues...)
	}

	headings := make(map[string]bool, len(units))
	for _, u := range units {
		headings[u.Heading] = true
	}

	for i := range units {
		deps := parseDepends(units[i].Body)
		units[i].Depends = deps
		seen := make(map[string]bool, len(deps))
		for _, dep := range deps {
			if seen[dep] {
				continue
			}
			seen[dep] = true
			if !headings[dep] {
				issues = append(issues, ValidationIssue{
					Severity: SeverityError,
					Message:  fmt.Sprintf("%s: unit %q depends on non-existent unit %q", source, units[i].Heading, dep),
					File:     source,
				})
			}
		}
	}

	return units, issues
}

func ValidateGHIssueByNumber(root string, number int) (*ValidationResult, error) {
	result := &ValidationResult{Valid: true}

	if _, err := lookPathGh("gh"); err != nil {
		return nil, fmt.Errorf("gh not available")
	}

	out, err := ghIssueView(root, number)
	if err != nil {
		return nil, fmt.Errorf("gh issue view %d failed: %w", number, err)
	}

	var issue ghIssue
	if err := json.Unmarshal(out, &issue); err != nil {
		return nil, fmt.Errorf("parse gh issue: %w", err)
	}

	source := fmt.Sprintf("GH issue #%d", issue.Number)
	units, unitIssues := ValidateQueueBody(issue.Body, source)
	result.UnitsCount += len(units)
	var commentBodies []string
	for _, c := range issue.Comments {
		commentBodies = append(commentBodies, c.Body)
	}
	applyQueueIssues(result, "GitHub comments", units, unitIssues, commentBodies)
	result.Valid = len(result.Errors) == 0
	return result, nil
}

func ValidateQueueFile(path string) (*ValidationResult, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	stripped, amendmentBlocks := splitLocalAmendmentBlocks(string(raw))
	result := &ValidationResult{Valid: true}
	source := fmt.Sprintf("queue file %s", path)
	units, unitIssues := ValidateQueueBody(stripped, source)
	result.UnitsCount += len(units)
	applyQueueIssues(result, source, units, unitIssues, amendmentBlocks)
	result.Valid = len(result.Errors) == 0
	return result, nil
}

func ValidateLocalQueues(root string) (*ValidationResult, error) {
	result := &ValidationResult{Valid: true}

	dir := filepath.Join(root, "specs", "queues")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		sub, err := ValidateQueueFile(path)
		if err != nil {
			return nil, err
		}
		result.Errors = append(result.Errors, sub.Errors...)
		result.Warnings = append(result.Warnings, sub.Warnings...)
		result.UnitsCount += sub.UnitsCount
	}

	result.Valid = len(result.Errors) == 0
	return result, nil
}
