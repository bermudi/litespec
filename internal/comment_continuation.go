package internal

import "strings"

const receiptContinuationMarker = "Receipt continues in next comment (GitHub comment size limit)."

// mergeContinuedComments joins a comment whose last unfenced line is the
// receipt continuation marker with the immediately following comment, chaining
// across as many comments as needed. GitHub caps issue comments at 65,536
// characters; a verbatim red-green receipt that exceeds it splits at a field
// boundary and continues in the next comment. Consumed comments are blanked so
// comment indices and error numbering stay stable. A dangling marker (no
// following comment) is left in place so receipt validation reports the
// incomplete receipt. Marker text inside fenced output is raw content, never a
// marker.
func mergeContinuedComments(comments []string) []string {
	merged := make([]string, len(comments))
	copy(merged, comments)
	for i := 0; i < len(merged); i++ {
		if strings.TrimSpace(merged[i]) == "" {
			continue
		}
		for commentEndsWithContinuationMarker(merged[i]) {
			next := -1
			for j := i + 1; j < len(merged); j++ {
				if strings.TrimSpace(merged[j]) != "" {
					next = j
					break
				}
			}
			if next == -1 {
				break
			}
			merged[i] = stripContinuationMarker(merged[i]) + "\n" + strings.TrimSpace(merged[next])
			merged[next] = ""
		}
	}
	return merged
}

func commentEndsWithContinuationMarker(comment string) bool {
	_, last := lastUnfencedNonEmptyLine(comment)
	return last == receiptContinuationMarker
}

func lastUnfencedNonEmptyLine(comment string) (int, string) {
	lines := strings.Split(strings.ReplaceAll(comment, "\r\n", "\n"), "\n")
	openFence := ""
	index := -1
	last := ""
	for i, line := range lines {
		if consumeMarkdownFenceLine(&openFence, line) {
			continue
		}
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			index, last = i, trimmed
		}
	}
	return index, last
}

func stripContinuationMarker(comment string) string {
	index, last := lastUnfencedNonEmptyLine(comment)
	if last != receiptContinuationMarker {
		return comment
	}
	lines := strings.Split(strings.ReplaceAll(comment, "\r\n", "\n"), "\n")
	lines = append(lines[:index], lines[index+1:]...)
	return strings.TrimRight(strings.Join(lines, "\n"), "\n")
}
