package internal

import "strings"

const receiptContinuationMarker = "Receipt continues in next comment (GitHub comment size limit)."

type continuedCommentPart struct {
	text      string
	continued bool
}

type continuedComment struct {
	text  string
	parts []continuedCommentPart
}

func mergeContinuedComments(comments []string) []string {
	merged := mergeContinuedCommentRecords(comments)
	texts := make([]string, len(merged))
	for i, comment := range merged {
		texts[i] = comment.text
	}
	return texts
}

func mergeContinuedCommentRecords(comments []string) []continuedComment {
	merged := make([]continuedComment, len(comments))
	for i, comment := range comments {
		merged[i] = continuedComment{
			text: comment,
			parts: []continuedCommentPart{{
				text: comment,
			}},
		}
	}
	for i := 0; i < len(merged); i++ {
		if strings.TrimSpace(merged[i].text) == "" {
			continue
		}
		next := i + 1
		for commentEndsWithContinuationMarker(merged[i].text) {
			if next >= len(merged) || strings.TrimSpace(merged[next].text) == "" {
				break
			}
			last := len(merged[i].parts) - 1
			merged[i].parts[last].text = stripContinuationMarker(merged[i].parts[last].text)
			merged[i].parts[last].continued = true
			merged[i].parts = append(merged[i].parts, continuedCommentPart{
				text: strings.TrimSpace(merged[next].text),
			})
			merged[i].text = continuedCommentText(merged[i].parts)
			merged[next] = continuedComment{}
			next++
		}
	}
	return merged
}

func continuedCommentText(parts []continuedCommentPart) string {
	texts := make([]string, len(parts))
	for i, part := range parts {
		texts[i] = part.text
	}
	return strings.Join(texts, "\n")
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
