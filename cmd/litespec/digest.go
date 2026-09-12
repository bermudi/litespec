package main

import (
	"flag"
	"fmt"

	"github.com/bermudi/litespec/v2/internal"
)

func cmdDigest(args []string) error {
	fs := newFlagSet("digest", printDigestHelp)
	var issueNumber int
	var queuePath string
	var heading string
	fs.IntVar(&issueNumber, "issue", 0, "print unit digests for a single GH issue by number")
	fs.StringVar(&queuePath, "queue", "", "print unit digests for a single local queue markdown file")
	fs.StringVar(&heading, "heading", "", "print only the digest lines whose unit heading equals this text exactly")

	ok, err := parseFlagSet(fs, args)
	if !ok {
		return err
	}

	issueSet := false
	queueSet := false
	headingSet := false
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "issue":
			issueSet = true
		case "queue":
			queueSet = true
		case "heading":
			headingSet = true
		}
	})

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
	if len(fs.Args()) > 0 {
		return fmt.Errorf("unexpected positional argument: %s", fs.Args()[0])
	}

	root, err := requireProjectRootWithStaleCheck()
	if err != nil {
		return err
	}

	lines, err := internal.DigestQueueUnits(root, issueNumber, queuePath)
	if err != nil {
		return err
	}

	if headingSet {
		filtered := make([]internal.UnitDigestLine, 0, len(lines))
		for _, line := range lines {
			if line.Heading == heading {
				filtered = append(filtered, line)
			}
		}
		if len(filtered) == 0 {
			return fmt.Errorf("no unit with heading %q", heading)
		}
		lines = filtered
	}

	fmt.Print(internal.FormatUnitDigestLines(lines))
	return nil
}

func printDigestHelp() {
	fmt.Print(`Usage: litespec digest --issue <N> | --queue <path>

Print each queue unit's identity (occurrence and heading) and its expected
contract digest, one tab-separated line per unit. Paste the digest into an
evidence receipt without transformation. With --heading, print only the
lines whose unit heading equals the given text exactly; every occurrence
of a duplicated heading is listed with its own occurrence number.

Options:
  --issue <N>     Fetch the GH issue by number (requires gh)
  --queue <path>  Read a local specs/queues/<name>.md file
  --heading <t>   Print only lines whose unit heading matches exactly
`)
}
