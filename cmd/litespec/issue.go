package main

import (
	"flag"
	"fmt"

	"github.com/bermudi/litespec/v2/internal"
)

func cmdIssue(args []string) error {
	if len(args) == 0 {
		printIssueHelp()
		return nil
	}
	switch args[0] {
	case "check":
		return cmdIssueCheck(args[1:])
	case "--help", "-h":
		printIssueHelp()
		return nil
	default:
		printIssueHelp()
		return fmt.Errorf("unknown issue subcommand: %s (valid: check)", args[0])
	}
}

func cmdIssueCheck(args []string) error {
	fs := newFlagSet("issue check", printIssueCheckHelp)
	var (
		issueNumber int
		heading     string
		occurrence  int
	)
	fs.IntVar(&issueNumber, "issue", 0, "GH issue number")
	fs.StringVar(&heading, "heading", "", "exact unit heading")
	fs.IntVar(&occurrence, "occurrence", 0, "1-based occurrence among units sharing the heading")

	ok, err := parseFlagSet(fs, args)
	if !ok {
		return err
	}

	if issueNumber < 1 {
		return fmt.Errorf("--issue must be a positive integer, got %d", issueNumber)
	}
	if heading == "" {
		return fmt.Errorf("--heading is required")
	}
	setFlags := map[string]bool{}
	fs.Visit(func(f *flag.Flag) {
		setFlags[f.Name] = true
	})
	if setFlags["occurrence"] && occurrence < 1 {
		return fmt.Errorf("--occurrence must be a positive integer, got %d", occurrence)
	}
	if len(fs.Args()) > 0 {
		return fmt.Errorf("unexpected positional argument: %s", fs.Args()[0])
	}

	root, err := requireProjectRootWithStaleCheck()
	if err != nil {
		return err
	}

	result, err := internal.IssueCheckTicksOneUnit(root, issueNumber, heading, occurrence)
	if err != nil {
		return err
	}
	fmt.Printf("checked %q (occurrence %d) in GH issue #%d\n", result.Heading, result.Occurrence, result.Number)
	return nil
}

func printIssueHelp() {
	fmt.Print(`Usage: litespec issue <subcommand> [flags]

Managed GH issue body operations. Deliberately minimal: there is no general
managed editor, only the single-checkbox tick.

Subcommands:
  check    Tick exactly one unit checkbox (see: litespec issue check --help)
`)
}

func printIssueCheckHelp() {
	fmt.Print(`Usage: litespec issue check --issue <N> --heading "<heading>" [--occurrence <K>]

Tick exactly one unit checkbox in a GH issue body. The unit is resolved by
exact heading and positive same-heading occurrence, with the same identity
semantics as litespec digest. The command writes the body back only when the
result differs from the fetched body by exactly one "- [ ]" to "- [x]" flip
with the Base:/Branch: ownership lines byte-unchanged. Any other body delta,
an ambiguous or unknown heading, an out-of-range occurrence, an
already-checked target, or a gh failure is a visible refusal that issues no
write. There is deliberately no offline queue lane and no broader body
management.

Flags:
  --issue <N>       GH issue number (requires gh)
  --heading <text>  Exact unit heading
  --occurrence <K>  1-based occurrence among units sharing the heading
`)
}
