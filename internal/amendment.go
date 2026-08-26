package internal

import (
	"fmt"
	"strings"
)

type contractAmendment struct {
	identity  queueUnitIdentity
	oldDigest string
	newDigest string
	reason    string
}

type digestEdge struct {
	from string
	to   string
}

type queueCommentScan struct {
	unresolved []queueUnitIdentity
	observed   map[queueUnitIdentity][]string
	edges      map[queueUnitIdentity][]digestEdge
	errors     []error
}

func parseAmendmentComment(comment string, units []queueUnit) (contractAmendment, bool, bool, error) {
	normalized := strings.TrimSpace(strings.ReplaceAll(comment, "\r\n", "\n"))
	lines := strings.Split(normalized, "\n")
	if len(lines) == 0 || !strings.HasPrefix(lines[0], "Amendment:") {
		return contractAmendment{}, false, false, nil
	}
	malformed := func(reason string) (contractAmendment, bool, bool, error) {
		return contractAmendment{}, true, false, fmt.Errorf("malformed amendment: %s", reason)
	}
	if lines[0] != "Amendment:" {
		return malformed("must begin with the exact line `Amendment:`")
	}
	if len(lines) != 6 {
		return malformed("must contain exactly Amendment:, Unit occurrence:, Unit heading:, Old digest:, New digest:, and Reason: lines")
	}
	identity, err := parseIdentityLines(lines[1], lines[2])
	if err != nil {
		return malformed(err.Error())
	}
	const oldPrefix = "Old digest: "
	const newPrefix = "New digest: "
	if !strings.HasPrefix(lines[3], oldPrefix) || !strings.HasPrefix(lines[4], newPrefix) {
		return malformed("Old digest: and New digest: must be 64 lowercase hexadecimal characters")
	}
	oldDigest := strings.TrimSpace(strings.TrimPrefix(lines[3], oldPrefix))
	newDigest := strings.TrimSpace(strings.TrimPrefix(lines[4], newPrefix))
	if !unitDigestPattern.MatchString(oldDigest) || !unitDigestPattern.MatchString(newDigest) {
		return malformed("Old digest: and New digest: must be 64 lowercase hexadecimal characters")
	}
	if !strings.HasPrefix(lines[5], "Reason:") {
		return malformed("Reason: line is required")
	}
	reason := strings.TrimSpace(strings.TrimPrefix(lines[5], "Reason:"))
	if reason == "" {
		return malformed("reason must be nonempty")
	}
	if _, ok := findQueueUnit(units, identity); !ok {
		return contractAmendment{}, true, false, fmt.Errorf(
			"amendment occurrence %d with heading %q does not identify exactly one queue unit",
			identity.Occurrence,
			identity.Heading,
		)
	}
	return contractAmendment{
		identity:  identity,
		oldDigest: oldDigest,
		newDigest: newDigest,
		reason:    reason,
	}, true, true, nil
}

func splitLocalAmendmentBlocks(body string) (string, []string) {
	lines := strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n")
	openFence := ""
	var blocks []string
	kept := make([]string, 0, len(lines))
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if consumeMarkdownFenceLine(&openFence, line) {
			kept = append(kept, line)
			continue
		}
		if openFence == "" && strings.TrimSpace(line) == "Amendment:" {
			block := []string{line}
			j := i + 1
			for j < len(lines) && len(block) < 6 && strings.TrimSpace(lines[j]) != "" {
				block = append(block, lines[j])
				j++
			}
			blocks = append(blocks, strings.Join(block, "\n"))
			i = j - 1
			continue
		}
		kept = append(kept, line)
	}
	return strings.Join(kept, "\n"), blocks
}

func scanQueueComments(units []queueUnit, comments []string) queueCommentScan {
	identities := queueUnitIdentities(units)
	validUnit := make(map[queueUnitIdentity]int, len(identities))
	for i, identity := range identities {
		validUnit[identity] = i
	}

	scan := queueCommentScan{
		observed: make(map[queueUnitIdentity][]string),
		edges:    make(map[queueUnitIdentity][]digestEdge),
	}

	type amendmentSighting struct {
		index  int
		record contractAmendment
		valid  bool
		err    error
	}
	var sightings []amendmentSighting
	amendmentShaped := make(map[int]bool)
	for index, comment := range comments {
		record, handled, validRecord, err := parseAmendmentComment(comment, units)
		if !handled {
			continue
		}
		amendmentShaped[index] = true
		sightings = append(sightings, amendmentSighting{index: index, record: record, valid: validRecord, err: err})
	}

	unresolved := make(map[queueUnitIdentity]bool)
	oldDigests := make(map[string]bool)
	lastAmendmentAt := make(map[queueUnitIdentity]int)
	for _, sight := range sightings {
		if !sight.valid {
			scan.errors = append(scan.errors, fmt.Errorf("comment %d: %v", sight.index+1, sight.err))
			continue
		}
		unresolved[sight.record.identity] = true
		oldDigests[sight.record.oldDigest] = true
		scan.edges[sight.record.identity] = append(
			scan.edges[sight.record.identity],
			digestEdge{from: sight.record.oldDigest, to: sight.record.newDigest},
		)
		if prev, ok := lastAmendmentAt[sight.record.identity]; !ok || sight.index > prev {
			lastAmendmentAt[sight.record.identity] = sight.index
		}
	}
	finalNewDigest := make(map[queueUnitIdentity]string)
	for _, sight := range sightings {
		if sight.valid && lastAmendmentAt[sight.record.identity] == sight.index {
			finalNewDigest[sight.record.identity] = sight.record.newDigest
		}
	}

	for commentIndex, comment := range comments {
		if amendmentShaped[commentIndex] {
			continue
		}
		identity, kind, digest, err := parseRebuildComment(comment, units)
		if err != nil {
			scan.errors = append(scan.errors, fmt.Errorf("comment %d: %w", commentIndex+1, err))
			continue
		}
		if kind == rebuildCommentOther {
			continue
		}
		if _, ok := validUnit[identity]; !ok {
			if kind != rebuildCommentRequest && digest != "" && oldDigests[digest] {
				for _, sight := range sightings {
					if sight.valid && sight.record.oldDigest == digest {
						scan.observed[sight.record.identity] = append(scan.observed[sight.record.identity], digest)
					}
				}
				continue
			}
			scan.errors = append(scan.errors, fmt.Errorf(
				"comment %d: unit occurrence %d with heading %q does not identify exactly one queue unit",
				commentIndex+1,
				identity.Occurrence,
				identity.Heading,
			))
			continue
		}
		if kind == rebuildCommentRequest {
			unresolved[identity] = true
			continue
		}
		scan.observed[identity] = append(scan.observed[identity], digest)
		if kind == rebuildCommentEvidence {
			delete(unresolved, identity)
		}
	}

	for identity, newDigest := range finalNewDigest {
		unit, ok := findQueueUnit(units, identity)
		if !ok {
			continue
		}
		current := unitContractDigest(unit)
		if newDigest != current {
			scan.errors = append(scan.errors, fmt.Errorf(
				"unit occurrence %d with heading %q: final amendment declares New digest %s but the current contract digest is %s",
				identity.Occurrence,
				identity.Heading,
				newDigest,
				current,
			))
		}
	}
	for i, unit := range units {
		chainIssues := digestChainIssues(identities[i], scan.observed[identities[i]], scan.edges[identities[i]], unitContractDigest(unit))
		scan.errors = append(scan.errors, chainIssues...)
	}

	scan.unresolved = make([]queueUnitIdentity, 0, len(unresolved))
	for _, identity := range identities {
		if unresolved[identity] {
			scan.unresolved = append(scan.unresolved, identity)
		}
	}
	return scan
}

func digestChainIssues(identity queueUnitIdentity, observed []string, edges []digestEdge, current string) []error {
	adjacent := make(map[string][]string, len(edges))
	for _, edge := range edges {
		adjacent[edge.from] = append(adjacent[edge.from], edge.to)
	}
	reaches := func(from, to string) bool {
		if from == to {
			return true
		}
		seen := map[string]bool{from: true}
		queue := []string{from}
		for len(queue) > 0 {
			node := queue[0]
			queue = queue[1:]
			for _, next := range adjacent[node] {
				if next == to {
					return true
				}
				if !seen[next] {
					seen[next] = true
					queue = append(queue, next)
				}
			}
		}
		return false
	}

	var errs []error
	prev := ""
	for _, digest := range observed {
		if prev != "" && prev != digest && !reaches(prev, digest) {
			errs = append(errs, fmt.Errorf(
				"unit occurrence %d with heading %q: observed receipt digests %s -> %s are not bridged by an amendment; witness the contract edit before rebuilding",
				identity.Occurrence,
				identity.Heading,
				prev,
				digest,
			))
		}
		prev = digest
	}
	if prev != "" && prev != current && !reaches(prev, current) {
		errs = append(errs, fmt.Errorf(
			"unit occurrence %d with heading %q: last receipt digest %s does not chain to the current contract digest %s through an amendment",
			identity.Occurrence,
			identity.Heading,
			prev,
			current,
		))
	}
	return errs
}
