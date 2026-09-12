package internal

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

var ghIssueComment = func(root string, number int, bodyFile string) ([]byte, error) {
	cmd := exec.Command("gh", "issue", "comment", strconv.Itoa(number), "--body-file", bodyFile)
	cmd.Dir = root
	return cmd.CombinedOutput()
}

// ReceiptCommentCommand is the exact gh command that posts one receipt
// comment file; the receipt command prints it and the poster runs it.
func ReceiptCommentCommand(number int, bodyFile string) string {
	return fmt.Sprintf("gh issue comment %d --body-file %s", number, bodyFile)
}

// PostReceiptComments posts numbered receipt comment files in strict posting
// order and returns the posted file names. A failing gh invocation stops the
// chain immediately with a visible error naming the failing command, the
// already-posted parts, and the unposted files; nothing is retried or
// reordered.
func PostReceiptComments(root string, issueNumber int, files []string) ([]string, error) {
	posted := make([]string, 0, len(files))
	for i, file := range files {
		command := ReceiptCommentCommand(issueNumber, file)
		out, err := ghIssueComment(root, issueNumber, file)
		if err != nil {
			return posted, fmt.Errorf("%s failed: %v: %s (posted: %s; not posted: %s)",
				command, err, strings.TrimSpace(string(out)),
				receiptFileList(posted), receiptFileList(files[i:]))
		}
		posted = append(posted, file)
	}
	return posted, nil
}

func receiptFileList(files []string) string {
	if len(files) == 0 {
		return "(none)"
	}
	return strings.Join(files, ", ")
}
