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

type queueReplanMarker struct {
	identity queueUnitIdentity
	digest   string
	reason   string
}

type digestEdge struct {
	from string
	to   string
}

type queueCommentScan struct {
	unresolved             []queueUnitIdentity
	replanRequired         []queueUnitIdentity
	observed               map[queueUnitIdentity][]string
	edges                  map[queueUnitIdentity][]digestEdge
	completedRebuildCycles map[queueUnitIdentity]map[string]int
	errors                 []error
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

func parseReplanMarkerComment(comment string) (queueReplanMarker, bool, error) {
	normalized := strings.TrimSpace(strings.ReplaceAll(comment, "\r\n", "\n"))
	lines := strings.Split(normalized, "\n")
	if len(lines) == 0 || !strings.HasPrefix(lines[0], "Re-plan required") {
		return queueReplanMarker{}, false, nil
	}
	malformed := func(reason string) (queueReplanMarker, bool, error) {
		return queueReplanMarker{}, true, fmt.Errorf("malformed re-plan marker: %s", reason)
	}
	if lines[0] != "Re-plan required:" {
		return malformed("must begin with the exact line `Re-plan required:`")
	}
	if len(lines) != 5 {
		return malformed("must contain exactly Re-plan required:, Unit occurrence:, Unit heading:, Unit digest:, and Reason: lines")
	}
	identity, err := parseIdentityLines(lines[1], lines[2])
	if err != nil {
		return malformed(err.Error())
	}
	const digestPrefix = "Unit digest: "
	if !strings.HasPrefix(lines[3], digestPrefix) {
		return malformed("Unit digest: must be 64 lowercase hexadecimal characters")
	}
	digest := strings.TrimSpace(strings.TrimPrefix(lines[3], digestPrefix))
	if !unitDigestPattern.MatchString(digest) {
		return malformed("Unit digest: must be 64 lowercase hexadecimal characters")
	}
	if !strings.HasPrefix(lines[4], "Reason:") {
		return malformed("Reason: line is required")
	}
	reason := strings.TrimSpace(strings.TrimPrefix(lines[4], "Reason:"))
	if reason == "" {
		return malformed("reason must be nonempty")
	}
	return queueReplanMarker{identity: identity, digest: digest, reason: reason}, true, nil
}

func splitLocalQueueMetadataBlocks(body string) (string, []string) {
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
		metadataLines := 0
		switch strings.TrimSpace(line) {
		case "Amendment:":
			metadataLines = 6
		case "Re-plan required:":
			metadataLines = 5
		}
		if openFence == "" && metadataLines > 0 {
			block := []string{line}
			j := i + 1
			for j < len(lines) && len(block) < metadataLines && strings.TrimSpace(lines[j]) != "" {
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
	comments = mergeContinuedComments(comments)
	identities := queueUnitIdentities(units)
	validUnit := make(map[queueUnitIdentity]int, len(identities))
	for i, identity := range identities {
		validUnit[identity] = i
	}

	scan := queueCommentScan{
		observed:               make(map[queueUnitIdentity][]string),
		edges:                  make(map[queueUnitIdentity][]digestEdge),
		completedRebuildCycles: make(map[queueUnitIdentity]map[string]int),
	}

	type amendmentSighting struct {
		index  int
		record contractAmendment
		valid  bool
		err    error
	}
	var sightings []amendmentSighting
	amendmentShaped := make(map[int]bool)
	amendmentsAt := make(map[int]contractAmendment)
	lastAmendmentAt := make(map[queueUnitIdentity]int)
	for index, comment := range comments {
		record, handled, validRecord, err := parseAmendmentComment(comment, units)
		if !handled {
			continue
		}
		amendmentShaped[index] = true
		sightings = append(sightings, amendmentSighting{index: index, record: record, valid: validRecord, err: err})
		if !validRecord {
			continue
		}
		amendmentsAt[index] = record
		scan.edges[record.identity] = append(
			scan.edges[record.identity],
			digestEdge{from: record.oldDigest, to: record.newDigest},
		)
		if prev, ok := lastAmendmentAt[record.identity]; !ok || index > prev {
			lastAmendmentAt[record.identity] = index
		}
	}
	for _, sight := range sightings {
		if !sight.valid {
			scan.errors = append(scan.errors, fmt.Errorf("comment %d: %v", sight.index+1, sight.err))
		}
	}

	resolveIdentity := func(identity queueUnitIdentity, digest string) (queueUnitIdentity, bool) {
		if _, ok := validUnit[identity]; ok {
			return identity, true
		}
		candidates := make(map[queueUnitIdentity]bool)
		for _, sight := range sightings {
			if sight.valid && sight.record.oldDigest == digest {
				candidates[sight.record.identity] = true
			}
		}
		if len(candidates) != 1 {
			return queueUnitIdentity{}, false
		}
		for candidate := range candidates {
			return candidate, true
		}
		return queueUnitIdentity{}, false
	}
	digestBelongsToIdentity := func(identity queueUnitIdentity, digest string) bool {
		unitIndex, ok := validUnit[identity]
		if ok && unitContractDigest(units[unitIndex]) == digest {
			return true
		}
		for _, edge := range scan.edges[identity] {
			if edge.from == digest {
				return true
			}
		}
		return false
	}

	unresolved := make(map[queueUnitIdentity]bool)
	markers := make(map[queueUnitIdentity]queueReplanMarker)
	pendingRebuildRequests := make(map[queueUnitIdentity]int)
	var pendingOrder []queueUnitIdentity
	for commentIndex, comment := range comments {
		if amendmentShaped[commentIndex] {
			record, ok := amendmentsAt[commentIndex]
			if !ok {
				continue
			}
			if marker, marked := markers[record.identity]; marked && marker.digest == record.oldDigest {
				delete(markers, record.identity)
			}
			unresolved[record.identity] = true
			continue
		}

		marker, handled, err := parseReplanMarkerComment(comment)
		if handled {
			if err != nil {
				scan.errors = append(scan.errors, fmt.Errorf("comment %d: %w", commentIndex+1, err))
				continue
			}
			identity, ok := resolveIdentity(marker.identity, marker.digest)
			if !ok {
				scan.errors = append(scan.errors, fmt.Errorf(
					"comment %d: re-plan marker occurrence %d with heading %q does not identify exactly one queue unit",
					commentIndex+1,
					marker.identity.Occurrence,
					marker.identity.Heading,
				))
				continue
			}
			marker.identity = identity
			if !digestBelongsToIdentity(identity, marker.digest) {
				scan.errors = append(scan.errors, fmt.Errorf(
					"comment %d: re-plan marker for unit occurrence %d with heading %q names digest %s outside its contract history",
					commentIndex+1,
					identity.Occurrence,
					identity.Heading,
					marker.digest,
				))
				continue
			}
			if scan.completedRebuildCycles[identity][marker.digest] < 2 {
				scan.errors = append(scan.errors, fmt.Errorf(
					"comment %d: re-plan marker for unit occurrence %d with heading %q requires two completed rebuild cycles for digest %s",
					commentIndex+1,
					identity.Occurrence,
					identity.Heading,
					marker.digest,
				))
				continue
			}
			if existing, duplicate := markers[identity]; duplicate {
				scan.errors = append(scan.errors, fmt.Errorf(
					"comment %d: duplicate unresolved re-plan marker for unit occurrence %d with heading %q and digest %s",
					commentIndex+1,
					identity.Occurrence,
					identity.Heading,
					existing.digest,
				))
				continue
			}
			markers[identity] = marker
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
		if kind == rebuildCommentRequest {
			if _, pending := pendingRebuildRequests[identity]; !pending {
				pendingRebuildRequests[identity] = commentIndex
				pendingOrder = append(pendingOrder, identity)
			}
			continue
		}
		resolvedIdentity, ok := resolveIdentity(identity, digest)
		if !ok {
			scan.errors = append(scan.errors, fmt.Errorf(
				"comment %d: unit occurrence %d with heading %q does not identify exactly one queue unit",
				commentIndex+1,
				identity.Occurrence,
				identity.Heading,
			))
			continue
		}
		scan.observed[resolvedIdentity] = append(scan.observed[resolvedIdentity], digest)
		if requestIndex, pending := pendingRebuildRequests[identity]; pending {
			if scan.completedRebuildCycles[resolvedIdentity] == nil {
				scan.completedRebuildCycles[resolvedIdentity] = make(map[string]int)
			}
			if scan.completedRebuildCycles[resolvedIdentity][digest] >= 2 {
				scan.errors = append(scan.errors, thirdRebuildRequestError(requestIndex, resolvedIdentity, digest))
			} else {
				scan.completedRebuildCycles[resolvedIdentity][digest]++
			}
			delete(pendingRebuildRequests, identity)
		}
		if kind == rebuildCommentEvidence {
			delete(unresolved, resolvedIdentity)
		}
	}

	for _, identity := range pendingOrder {
		requestIndex, pending := pendingRebuildRequests[identity]
		if !pending {
			continue
		}
		unitIndex, ok := validUnit[identity]
		if !ok {
			scan.errors = append(scan.errors, fmt.Errorf(
				"comment %d: unit occurrence %d with heading %q does not identify exactly one queue unit",
				requestIndex+1,
				identity.Occurrence,
				identity.Heading,
			))
			continue
		}
		currentDigest := unitContractDigest(units[unitIndex])
		if scan.completedRebuildCycles[identity][currentDigest] >= 2 {
			scan.errors = append(scan.errors, thirdRebuildRequestError(requestIndex, identity, currentDigest))
			continue
		}
		unresolved[identity] = true
	}

	finalNewDigest := make(map[queueUnitIdentity]string)
	for _, sight := range sightings {
		if sight.valid && lastAmendmentAt[sight.record.identity] == sight.index {
			finalNewDigest[sight.record.identity] = sight.record.newDigest
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
	scan.replanRequired = make([]queueUnitIdentity, 0, len(markers))
	for _, identity := range identities {
		if unresolved[identity] {
			scan.unresolved = append(scan.unresolved, identity)
		}
		if _, marked := markers[identity]; marked {
			scan.replanRequired = append(scan.replanRequired, identity)
		}
	}
	return scan
}

func thirdRebuildRequestError(commentIndex int, identity queueUnitIdentity, digest string) error {
	return fmt.Errorf(
		"comment %d: unit occurrence %d with heading %q already completed two rebuild cycles for digest %s and requires re-planning",
		commentIndex+1,
		identity.Occurrence,
		identity.Heading,
		digest,
	)
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
