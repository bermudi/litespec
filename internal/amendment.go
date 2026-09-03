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

type amendmentSighting struct {
	index  int
	record contractAmendment
	valid  bool
	err    error
}

type amendmentTransitionKey struct {
	identity  queueUnitIdentity
	oldDigest string
	newDigest string
}

type digestStateKey struct {
	digest     string
	occurrence int
}

type amendmentHistory struct {
	terminalByIndex    map[int]queueUnitIdentity
	terminalByPosition map[int]queueUnitIdentity
	aliases            map[queueUnitIdentity]map[queueUnitIdentity]bool
	postStates         map[digestStateKey][]int
	oldStates          map[digestStateKey][]int
	current            map[queueUnitIdentity]string
	records            map[int]contractAmendment
	errors             []error
}

type queueCommentScan struct {
	unresolved             []queueUnitIdentity
	replanRequired         []queueUnitIdentity
	observed               map[queueUnitIdentity][]string
	edges                  map[queueUnitIdentity][]digestEdge
	completedRebuildCycles map[queueUnitIdentity]map[string]int
	errors                 []error
}

func parseAmendmentComment(comment string, _ []queueUnit) (contractAmendment, bool, bool, error) {
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

func buildAmendmentHistory(units []queueUnit, sightings []amendmentSighting) amendmentHistory {
	history := amendmentHistory{
		terminalByIndex:    make(map[int]queueUnitIdentity),
		terminalByPosition: make(map[int]queueUnitIdentity),
		aliases:            make(map[queueUnitIdentity]map[queueUnitIdentity]bool),
		postStates:         make(map[digestStateKey][]int),
		oldStates:          make(map[digestStateKey][]int),
		current:            make(map[queueUnitIdentity]string),
		records:            make(map[int]contractAmendment),
	}

	identities := queueUnitIdentities(units)
	currentByDigest := make(map[digestStateKey][]queueUnitIdentity)
	for i, identity := range identities {
		digest := unitContractDigest(units[i])
		history.current[identity] = digest
		state := digestStateKey{digest: digest, occurrence: identity.Occurrence}
		currentByDigest[state] = append(currentByDigest[state], identity)
		history.aliases[identity] = map[queueUnitIdentity]bool{identity: true}
	}

	canonicalByKey := make(map[amendmentTransitionKey]int)
	canonicalPositions := make([]int, 0, len(sightings))
	for position, sighting := range sightings {
		if !sighting.valid {
			continue
		}
		key := amendmentTransitionKey{
			identity:  sighting.record.identity,
			oldDigest: sighting.record.oldDigest,
			newDigest: sighting.record.newDigest,
		}
		if _, exists := canonicalByKey[key]; exists {
			continue
		}
		canonicalByKey[key] = position
		canonicalPositions = append(canonicalPositions, position)
		history.records[position] = sighting.record
	}

	byOldDigest := make(map[string][]int)
	byOld := make(map[string]map[amendmentTransitionKey]int)
	byNew := make(map[string]map[amendmentTransitionKey]int)
	for _, position := range canonicalPositions {
		record := history.records[position]
		key := amendmentTransitionKey{
			identity:  record.identity,
			oldDigest: record.oldDigest,
			newDigest: record.newDigest,
		}
		if byOld[record.oldDigest] == nil {
			byOld[record.oldDigest] = make(map[amendmentTransitionKey]int)
		}
		byOld[record.oldDigest][key] = position
		if byNew[record.newDigest] == nil {
			byNew[record.newDigest] = make(map[amendmentTransitionKey]int)
		}
		byNew[record.newDigest][key] = position
		byOldDigest[record.oldDigest] = append(byOldDigest[record.oldDigest], position)
	}

	reportedOldBranches := make(map[string]bool)
	reportedNewBranches := make(map[string]bool)
	for _, position := range canonicalPositions {
		record := history.records[position]
		if len(byOld[record.oldDigest]) > 1 && !reportedOldBranches[record.oldDigest] {
			reportedOldBranches[record.oldDigest] = true
			history.errors = append(history.errors, fmt.Errorf(
				"amendment comment %d: branching amendment history from digest %s at occurrence %d",
				sightings[position].index+1,
				record.oldDigest,
				record.identity.Occurrence,
			))
		}
		if len(byNew[record.newDigest]) > 1 && !reportedNewBranches[record.newDigest] {
			reportedNewBranches[record.newDigest] = true
			history.errors = append(history.errors, fmt.Errorf(
				"amendment comment %d: ambiguous amendment history reaches digest %s at occurrence %d",
				sightings[position].index+1,
				record.newDigest,
				record.identity.Occurrence,
			))
		}
	}

	successors := make(map[int][]int)
	for _, position := range canonicalPositions {
		record := history.records[position]
		candidates := make(map[int]bool)
		for _, candidate := range byOldDigest[record.newDigest] {
			if candidate == position || sightings[candidate].index <= sightings[position].index {
				continue
			}
			candidateRecord := history.records[candidate]
			if candidateRecord.identity.Occurrence != record.identity.Occurrence {
				continue
			}
			candidates[candidate] = true
		}
		for candidate := range candidates {
			successors[position] = append(successors[position], candidate)
		}
	}

	state := make(map[int]uint8)
	terminal := make(map[int]queueUnitIdentity)
	reportedCycles := make(map[int]bool)
	var resolve func(int) (queueUnitIdentity, bool)
	resolve = func(position int) (queueUnitIdentity, bool) {
		switch state[position] {
		case 1:
			if !reportedCycles[position] {
				reportedCycles[position] = true
				record := history.records[position]
				history.errors = append(history.errors, fmt.Errorf(
					"amendment comment %d: digest-linked amendment chain loops at digest %s",
					sightings[position].index+1,
					record.newDigest,
				))
			}
			return queueUnitIdentity{}, false
		case 2:
			resolved, ok := terminal[position]
			return resolved, ok
		}

		state[position] = 1
		record := history.records[position]
		next := successors[position]
		if len(next) > 1 {
			state[position] = 2
			return queueUnitIdentity{}, false
		}
		if len(next) == 1 {
			resolved, ok := resolve(next[0])
			state[position] = 2
			if ok {
				terminal[position] = resolved
			}
			return resolved, ok
		}

		if currentDigest, ok := history.current[record.identity]; ok {
			if currentDigest == record.newDigest {
				terminal[position] = record.identity
			}
			state[position] = 2
			return record.identity, currentDigest == record.newDigest
		}
		newState := digestStateKey{digest: record.newDigest, occurrence: record.identity.Occurrence}
		if len(currentByDigest[newState]) > 0 {
			history.errors = append(history.errors, fmt.Errorf(
				"amendment comment %d: New digest %s reaches the current contract, but post-amendment occurrence %d with heading %q is not that exact queue identity",
				sightings[position].index+1,
				record.newDigest,
				record.identity.Occurrence,
				record.identity.Heading,
			))
		} else {
			history.errors = append(history.errors, fmt.Errorf(
				"amendment comment %d: amendment occurrence %d with heading %q is disconnected at New digest %s; no later identity transition reaches the current queue",
				sightings[position].index+1,
				record.identity.Occurrence,
				record.identity.Heading,
				record.newDigest,
			))
		}
		state[position] = 2
		return queueUnitIdentity{}, false
	}

	for _, position := range canonicalPositions {
		resolve(position)
	}

	for _, position := range canonicalPositions {
		resolved, ok := terminal[position]
		if !ok {
			continue
		}
		record := history.records[position]
		history.terminalByPosition[position] = resolved
		history.aliases[record.identity] = addIdentityTarget(history.aliases[record.identity], resolved)
		appendDigestStatePosition(history.postStates, digestStateKey{
			digest:     record.newDigest,
			occurrence: record.identity.Occurrence,
		}, position)
		appendDigestStatePosition(history.oldStates, digestStateKey{
			digest:     record.oldDigest,
			occurrence: record.identity.Occurrence,
		}, position)
	}
	for _, sighting := range sightings {
		if !sighting.valid {
			continue
		}
		canonical := canonicalByKey[amendmentTransitionKey{
			identity:  sighting.record.identity,
			oldDigest: sighting.record.oldDigest,
			newDigest: sighting.record.newDigest,
		}]
		if resolved, ok := terminal[canonical]; ok {
			history.terminalByIndex[sighting.index] = resolved
		}
	}

	for identity, targets := range history.aliases {
		if len(targets) > 1 {
			history.errors = append(history.errors, fmt.Errorf(
				"amendment identity occurrence %d with heading %q maps ambiguously across digest-linked chains",
				identity.Occurrence,
				identity.Heading,
			))
		}
	}
	return history
}

func addIdentityTarget(targets map[queueUnitIdentity]bool, target queueUnitIdentity) map[queueUnitIdentity]bool {
	if targets == nil {
		targets = make(map[queueUnitIdentity]bool)
	}
	targets[target] = true
	return targets
}

func appendDigestStatePosition(states map[digestStateKey][]int, key digestStateKey, position int) {
	for _, existing := range states[key] {
		if existing == position {
			return
		}
	}
	states[key] = append(states[key], position)
}

func (h amendmentHistory) resolveMetadataIdentity(identity queueUnitIdentity) (queueUnitIdentity, bool) {
	targets := h.aliases[identity]
	if len(targets) != 1 {
		return queueUnitIdentity{}, false
	}
	for target := range targets {
		return target, true
	}
	return queueUnitIdentity{}, false
}

func (h amendmentHistory) resolveReceiptIdentity(identity queueUnitIdentity, digest string) (queueUnitIdentity, bool) {
	key := digestStateKey{digest: digest, occurrence: identity.Occurrence}
	if positions := h.postStates[key]; len(positions) > 0 {
		if len(positions) != 1 {
			return queueUnitIdentity{}, false
		}
		position := positions[0]
		if h.records[position].identity != identity {
			return queueUnitIdentity{}, false
		}
		return h.terminalByPosition[position], true
	}
	if positions := h.oldStates[key]; len(positions) > 0 {
		if len(positions) != 1 {
			return queueUnitIdentity{}, false
		}
		resolved := h.terminalByPosition[positions[0]]
		if current, ok := h.current[identity]; ok && current != "" && resolved != identity {
			return queueUnitIdentity{}, false
		}
		if aliased, ok := h.resolveMetadataIdentity(identity); ok && aliased != resolved {
			return queueUnitIdentity{}, false
		}
		return resolved, true
	}
	if _, ok := h.current[identity]; ok {
		return identity, true
	}
	return queueUnitIdentity{}, false
}

func (h amendmentHistory) digestBelongsToIdentity(identity queueUnitIdentity, digest string) bool {
	if current, ok := h.current[identity]; ok && current == digest {
		return true
	}
	key := digestStateKey{digest: digest, occurrence: identity.Occurrence}
	for _, position := range h.postStates[key] {
		if h.terminalByPosition[position] == identity {
			return true
		}
	}
	for _, position := range h.oldStates[key] {
		if h.terminalByPosition[position] == identity {
			return true
		}
	}
	return false
}

func scanQueueComments(units []queueUnit, comments []string) queueCommentScan {
	commentRecords := mergeContinuedCommentRecords(comments)
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

	var sightings []amendmentSighting
	amendmentShaped := make(map[int]bool)
	amendmentsAt := make(map[int]contractAmendment)
	lastValidAmendmentAt := make(map[queueUnitIdentity]int)
	lastValidAmendment := make(map[queueUnitIdentity]contractAmendment)
	for index, comment := range commentRecords {
		record, handled, validRecord, err := parseAmendmentComment(comment.text, units)
		if !handled {
			continue
		}
		amendmentShaped[index] = true
		sightings = append(sightings, amendmentSighting{index: index, record: record, valid: validRecord, err: err})
		if validRecord {
			amendmentsAt[index] = record
			if previous, exists := lastValidAmendmentAt[record.identity]; !exists || index > previous {
				lastValidAmendmentAt[record.identity] = index
				lastValidAmendment[record.identity] = record
			}
		}
	}
	for _, sight := range sightings {
		if !sight.valid {
			scan.errors = append(scan.errors, fmt.Errorf("comment %d: %v", sight.index+1, sight.err))
		}
	}

	history := buildAmendmentHistory(units, sightings)
	scan.errors = append(scan.errors, history.errors...)
	for _, sight := range sightings {
		if !sight.valid {
			continue
		}
		identity, ok := history.terminalByIndex[sight.index]
		if !ok {
			continue
		}
		scan.edges[identity] = append(scan.edges[identity], digestEdge{
			from: sight.record.oldDigest,
			to:   sight.record.newDigest,
		})
	}

	resolveIdentity := history.resolveReceiptIdentity
	digestBelongsToIdentity := history.digestBelongsToIdentity

	unresolved := make(map[queueUnitIdentity]bool)
	latestAmendmentDigest := make(map[queueUnitIdentity]string)
	markers := make(map[queueUnitIdentity]queueReplanMarker)
	pendingRebuildRequests := make(map[queueUnitIdentity]int)
	var pendingOrder []queueUnitIdentity
	for commentIndex, commentRecord := range commentRecords {
		comment := commentRecord.text
		if amendmentShaped[commentIndex] {
			record, ok := amendmentsAt[commentIndex]
			if !ok {
				continue
			}
			identity, ok := history.terminalByIndex[commentIndex]
			if !ok {
				continue
			}
			if marker, marked := markers[identity]; marked && marker.digest == record.oldDigest {
				delete(markers, identity)
			}
			latestAmendmentDigest[identity] = record.newDigest
			unresolved[identity] = true
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

		identity, kind, digest, err := parseRebuildCommentRecord(commentRecord, units)
		if err != nil {
			scan.errors = append(scan.errors, fmt.Errorf("comment %d: %w", commentIndex+1, err))
			continue
		}
		if kind == rebuildCommentOther {
			if commentContainsRawOutputChunk(commentRecord) && !commentHasValidHeadingEvidence(commentRecord, units) {
				scan.errors = append(scan.errors, fmt.Errorf("comment %d: orphan raw output chunk", commentIndex+1))
			}
			continue
		}
		if kind == rebuildCommentRequest {
			if resolved, ok := history.resolveMetadataIdentity(identity); ok {
				identity = resolved
			}
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
		if requestIndex, pending := pendingRebuildRequests[resolvedIdentity]; pending {
			if scan.completedRebuildCycles[resolvedIdentity] == nil {
				scan.completedRebuildCycles[resolvedIdentity] = make(map[string]int)
			}
			if scan.completedRebuildCycles[resolvedIdentity][digest] >= 2 {
				scan.errors = append(scan.errors, thirdRebuildRequestError(requestIndex, resolvedIdentity, digest))
			} else {
				scan.completedRebuildCycles[resolvedIdentity][digest]++
			}
			delete(pendingRebuildRequests, resolvedIdentity)
		}
		if kind == rebuildCommentEvidence {
			if amendedDigest, amended := latestAmendmentDigest[resolvedIdentity]; !amended || digest == amendedDigest {
				delete(unresolved, resolvedIdentity)
			}
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

	for identity := range lastValidAmendmentAt {
		if _, currentIdentity := validUnit[identity]; !currentIdentity {
			continue
		}
		record := lastValidAmendment[identity]
		current := history.current[identity]
		if record.newDigest != current {
			scan.errors = append(scan.errors, fmt.Errorf(
				"unit occurrence %d with heading %q: final amendment declares New digest %s but the current contract digest is %s",
				identity.Occurrence,
				identity.Heading,
				record.newDigest,
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
