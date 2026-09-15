package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bermudi/litespec/v2/internal"
)

func cmdReceipt(args []string) error {
	fs := newFlagSet("receipt", printReceiptHelp)
	var (
		issueNumber   int
		queuePath     string
		heading       string
		occurrence    int
		preSHA        string
		preStatus     int
		preOutPath    string
		postSHA       string
		postStatus    int
		postOutPath   string
		rebuild       bool
		recoveredFrom string
		post          bool
		outDir        string
	)
	fs.IntVar(&issueNumber, "issue", 0, "GH issue number")
	fs.StringVar(&queuePath, "queue", "", "local queue markdown file")
	fs.StringVar(&heading, "heading", "", "exact unit heading")
	fs.IntVar(&occurrence, "occurrence", 0, "1-based occurrence among units sharing the heading")
	fs.StringVar(&preSHA, "pre-sha", "", "pre commit SHA")
	fs.IntVar(&preStatus, "pre-status", 0, "pre exit status (must be non-zero)")
	fs.StringVar(&preOutPath, "pre-out", "", "file holding the raw pre output")
	fs.StringVar(&postSHA, "post-sha", "", "post commit SHA")
	fs.IntVar(&postStatus, "post-status", 0, "post exit status (default 0)")
	fs.StringVar(&postOutPath, "post-out", "", "file holding the raw post output")
	fs.BoolVar(&rebuild, "rebuild", false, "include the rebuild routing identity")
	fs.StringVar(&recoveredFrom, "recovered-from", "", "recovery provenance receipt ID")
	fs.BoolVar(&post, "post", false, "run the printed gh issue comment commands in posting order")
	fs.StringVar(&outDir, "out", "", "directory for emitted comment files (default: current directory)")

	ok, err := parseFlagSet(fs, args)
	if !ok {
		return err
	}

	setFlags := map[string]bool{}
	fs.Visit(func(f *flag.Flag) {
		setFlags[f.Name] = true
	})

	issueSet := setFlags["issue"]
	queueSet := setFlags["queue"]
	if issueSet && issueNumber < 1 {
		return fmt.Errorf("--issue must be a positive integer, got %d", issueNumber)
	}
	if queueSet && queuePath == "" {
		return fmt.Errorf("--queue requires a non-empty path")
	}
	if issueSet && queueSet {
		return fmt.Errorf("--issue and --queue are mutually exclusive")
	}
	if !issueSet && !queueSet {
		return fmt.Errorf("one of --issue <N> or --queue <path> is required")
	}
	if heading == "" {
		return fmt.Errorf("--heading is required")
	}
	if setFlags["occurrence"] && occurrence < 1 {
		return fmt.Errorf("--occurrence must be a positive integer, got %d", occurrence)
	}
	if post && !issueSet {
		return fmt.Errorf("--post requires --issue <N>; queue mode has no issue to comment on")
	}
	if setFlags["out"] && outDir == "" {
		return fmt.Errorf("--out requires a non-empty path")
	}
	if preSHA == "" {
		return fmt.Errorf("--pre-sha is required")
	}
	if !setFlags["pre-status"] {
		return fmt.Errorf("--pre-status is required and must be non-zero: the pre run must fail because the outcome is absent")
	}
	if preOutPath == "" {
		return fmt.Errorf("--pre-out is required")
	}
	if postSHA == "" {
		return fmt.Errorf("--post-sha is required")
	}
	if postOutPath == "" {
		return fmt.Errorf("--post-out is required")
	}
	if len(fs.Args()) > 0 {
		return fmt.Errorf("unexpected positional argument: %s", fs.Args()[0])
	}

	root, err := requireProjectRootWithStaleCheck()
	if err != nil {
		return err
	}

	resolved, err := internal.ResolveQueueUnit(root, issueNumber, queuePath, heading, occurrence)
	if err != nil {
		return err
	}

	preOutput, err := os.ReadFile(preOutPath)
	if err != nil {
		return fmt.Errorf("cannot read pre output file %s: %w", preOutPath, err)
	}
	postOutput, err := os.ReadFile(postOutPath)
	if err != nil {
		return fmt.Errorf("cannot read post output file %s: %w", postOutPath, err)
	}

	if err := verifyGitAncestry(root, preSHA, postSHA); err != nil {
		return err
	}

	comments, err := internal.AssembleEvidenceV2ReceiptComments(internal.ReceiptAssemblyRequest{
		Occurrence:    resolved.Occurrence,
		Heading:       resolved.Heading,
		Verify:        resolved.Verify,
		UnitDigest:    resolved.Digest,
		RecoveredFrom: recoveredFrom,
		Rebuild:       rebuild,
		Pre:           internal.ReceiptRunEvidence{SHA: preSHA, Status: preStatus, Output: string(preOutput)},
		Post:          internal.ReceiptRunEvidence{SHA: postSHA, Status: postStatus, Output: string(postOutput)},
	})
	if err != nil {
		return err
	}

	absOut := outDir
	if absOut == "" {
		absOut = "."
	}
	absOut, err = filepath.Abs(absOut)
	if err != nil {
		return fmt.Errorf("resolve output directory %s: %w", outDir, err)
	}
	if outDir != "" {
		if err := os.MkdirAll(absOut, 0o755); err != nil {
			return fmt.Errorf("create output directory %s: %w", outDir, err)
		}
	}
	names := make([]string, len(comments))
	for i, text := range comments {
		name := filepath.Join(absOut, fmt.Sprintf("receipt-%04d.md", i+1))
		if err := os.WriteFile(name, []byte(text), 0o644); err != nil {
			return emitWriteFailure(name, err, names[:i])
		}
		names[i] = name
	}
	if issueSet {
		if !post {
			for _, name := range names {
				fmt.Println(internal.ReceiptCommentCommand(issueNumber, name))
			}
			return nil
		}
		posted, err := internal.PostReceiptComments(root, issueNumber, names)
		if err != nil {
			return err
		}
		for _, name := range posted {
			fmt.Printf("posted %s (%s)\n", name, internal.ReceiptCommentCommand(issueNumber, name))
		}
	}
	return nil
}

// emitWriteFailure reports a failed comment-file write and removes the
// already-written files so no partial receipt set is left behind. A cleanup
// failure surfaces alongside the write failure.
func emitWriteFailure(name string, writeErr error, written []string) error {
	failure := fmt.Errorf("write %s: %w", name, writeErr)
	if len(written) == 0 {
		return failure
	}
	var cleanupFailures []string
	for _, file := range written {
		if err := os.Remove(file); err != nil {
			cleanupFailures = append(cleanupFailures, fmt.Sprintf("%s: %v", file, err))
		}
	}
	if len(cleanupFailures) > 0 {
		return fmt.Errorf("%w; cleanup also failed, partial files may remain: %s", failure, strings.Join(cleanupFailures, "; "))
	}
	return fmt.Errorf("%v; removed %d already-written comment file(s) so no partial set is left", failure, len(written))
}

// verifyGitAncestry refuses unless preSHA is an ancestor of postSHA in the
// project repository. Git failures surface as visible errors.
func verifyGitAncestry(root, preSHA, postSHA string) error {
	cmd := exec.Command("git", "merge-base", "--is-ancestor", preSHA, postSHA)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return fmt.Errorf("pre commit %s is not an ancestor of post commit %s", preSHA, postSHA)
	}
	return fmt.Errorf("git merge-base --is-ancestor %s %s failed: %v: %s", preSHA, postSHA, err, strings.TrimSpace(string(out)))
}

func printReceiptHelp() {
	fmt.Print(`Usage: litespec receipt --issue <N> | --queue <path> --heading "<heading>" [flags]

Assemble one validator-clean evidence receipt for a resolved queue unit and
emit a single bounded evidence/v2 comment file (receipt-0001.md). Outputs
are excerpted head/marker/tail beside the full output's byte count and
SHA-256 so one receipt always fits one comment; oversized output never
chunks and the receipt never continues. Issue mode also prints the exact
gh issue comment command; queue mode emits the file only. By default the
command never posts and never ticks checkboxes. Opt-in --post runs the
printed gh issue comment command itself, reporting the posted comment; a
failing gh invocation stops with a visible error naming the failing
command and the unposted file, with no retry. Posting never ticks
checkboxes.

The unit is resolved by exact heading and positive same-heading occurrence,
with the same identity semantics as litespec digest. Ambiguous, unknown, or
out-of-range headings, unreadable output files, ancestry violations, and
identity/status fields that exceed the fixed receipt budget are refused
before any file is written; the assembled receipt self-parses through the
existing evidence grammar. Posted receipts are independently
re-verified by ` + "`litespec validate --issue <N>`" + ` (and the default
labeled-issue scan): every versioned Receipt ID is recomputed from the
posted comment's own fields, and a mismatch fails validation. Emitted file paths and printed
commands are absolute, so posting and pasted commands work from any
directory. A failed comment-file write removes any already-written file so
no partial set is left behind.

Flags:
  --issue <N>            Fetch the GH issue by number (requires gh)
  --queue <path>         Read a local specs/queues/<name>.md file
  --heading <text>       Exact unit heading
  --occurrence <K>       1-based occurrence among units sharing the heading
  --pre-sha <sha>        Pre commit SHA (full hexadecimal)
  --pre-status <n>       Pre exit status (must be non-zero)
  --pre-out <file>       File holding the raw pre output
  --post-sha <sha>       Post commit SHA (full hexadecimal)
  --post-status <n>      Post exit status (default 0)
  --post-out <file>      File holding the raw post output
  --rebuild              Include the rebuild routing identity
  --recovered-from <id>  Recovery provenance receipt ID
  --post                 Run the printed gh issue comment commands in order
  --out <dir>            Directory for the emitted comment files (default:
                        current directory; created when missing)
`)
}
