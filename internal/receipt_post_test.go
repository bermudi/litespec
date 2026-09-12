package internal

import (
	"errors"
	"strings"
	"testing"
)

func TestReceiptPostsCommentsOnlyWhenAsked(t *testing.T) {
	t.Run("opt-in flag posts the numbered comments in order", func(t *testing.T) {
		var calls []string
		original := ghIssueComment
		ghIssueComment = func(root string, number int, bodyFile string) ([]byte, error) {
			calls = append(calls, ReceiptCommentCommand(number, bodyFile))
			if root == "" {
				t.Error("posting must run gh inside the project root")
			}
			return []byte("https://example.invalid/issues/42#comment-1\n"), nil
		}
		defer func() { ghIssueComment = original }()

		files := []string{"receipt-0001.md", "receipt-0002.md", "receipt-0003.md"}
		posted, err := PostReceiptComments("/repo", 42, files)
		if err != nil {
			t.Fatalf("posting failed: %v", err)
		}
		wantCalls := []string{
			"gh issue comment 42 --body-file receipt-0001.md",
			"gh issue comment 42 --body-file receipt-0002.md",
			"gh issue comment 42 --body-file receipt-0003.md",
		}
		if strings.Join(calls, "\n") != strings.Join(wantCalls, "\n") {
			t.Fatalf("gh calls = %v, want the printed commands in posting order %v", calls, wantCalls)
		}
		if strings.Join(posted, "\n") != strings.Join(files, "\n") {
			t.Fatalf("posted = %v, want every posted comment reported %v", posted, files)
		}
	})

	t.Run("mid-chain gh failure stops visibly with posted and unposted parts named", func(t *testing.T) {
		var calls []string
		original := ghIssueComment
		ghIssueComment = func(root string, number int, bodyFile string) ([]byte, error) {
			calls = append(calls, ReceiptCommentCommand(number, bodyFile))
			if bodyFile == "receipt-0002.md" {
				return []byte("gh: server error\n"), errors.New("exit status 1")
			}
			return []byte("https://example.invalid/issues/42#comment-1\n"), nil
		}
		defer func() { ghIssueComment = original }()

		posted, err := PostReceiptComments("/repo", 42, []string{"receipt-0001.md", "receipt-0002.md", "receipt-0003.md"})
		if err == nil {
			t.Fatalf("mid-chain gh failure must stop the chain, got posted=%v", posted)
		}
		for _, want := range []string{
			"gh issue comment 42 --body-file receipt-0002.md",
			"receipt-0001.md",
			"receipt-0003.md",
		} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("error must name %q: %v", want, err)
			}
		}
		wantCalls := []string{
			"gh issue comment 42 --body-file receipt-0001.md",
			"gh issue comment 42 --body-file receipt-0002.md",
		}
		if strings.Join(calls, "\n") != strings.Join(wantCalls, "\n") {
			t.Fatalf("gh calls = %v, want exactly %v with no retry and no reordering", calls, wantCalls)
		}
		if len(posted) != 1 || posted[0] != "receipt-0001.md" {
			t.Fatalf("posted = %v, want only the successfully posted receipt-0001.md", posted)
		}
	})
}
