package internal

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

const defaultReceiptCommentLimit = 65536

type ReceiptRunEvidence struct {
	SHA    string
	Status int
	Output string
}

type ReceiptAssemblyRequest struct {
	Occurrence    int
	Heading       string
	Verify        string
	UnitDigest    string
	RecoveredFrom string
	Rebuild       bool
	Pre           ReceiptRunEvidence
	Post          ReceiptRunEvidence
	CommentLimit  int
}

func (r ReceiptAssemblyRequest) validate() error {
	if r.Occurrence < 1 {
		return fmt.Errorf("unit occurrence must be a positive integer, got %d", r.Occurrence)
	}
	if strings.TrimSpace(r.Heading) == "" {
		return fmt.Errorf("unit heading must be nonempty")
	}
	if strings.TrimSpace(r.Verify) == "" {
		return fmt.Errorf("Verify command must be nonempty")
	}
	if !unitDigestPattern.MatchString(r.UnitDigest) {
		return fmt.Errorf("unit digest must be 64 lowercase hexadecimal characters")
	}
	if !evidenceCommitPattern.MatchString(r.Pre.SHA) {
		return fmt.Errorf("pre sha must be a full 40- or 64-character hexadecimal commit ID")
	}
	if !evidenceCommitPattern.MatchString(r.Post.SHA) {
		return fmt.Errorf("post sha must be a full 40- or 64-character hexadecimal commit ID")
	}
	if r.Pre.Status == 0 {
		return fmt.Errorf("pre exit status must be non-zero: the pre run must fail because the outcome is absent")
	}
	if r.Post.Status != 0 {
		return fmt.Errorf("post exit status must be 0, got %d", r.Post.Status)
	}
	if strings.EqualFold(r.Pre.SHA, r.Post.SHA) {
		return fmt.Errorf("pre and post shas must differ, both are %s", r.Pre.SHA)
	}
	if r.RecoveredFrom != "" && !receiptIDPattern.MatchString(r.RecoveredFrom) {
		return fmt.Errorf("Recovered from must be a complete receipt ID, got %q", r.RecoveredFrom)
	}
	return nil
}

// AssembleEvidenceReceiptComments turns one red-green run into validator-clean
// receipt comment texts. It refuses invalid run evidence before producing any
// content, splits oversized content only at legal boundaries, and self-parses
// the result through the existing evidence grammar before returning.
func AssembleEvidenceReceiptComments(req ReceiptAssemblyRequest) ([]string, error) {
	if err := req.validate(); err != nil {
		return nil, err
	}
	limit := req.CommentLimit
	if limit <= 0 {
		limit = defaultReceiptCommentLimit
	}
	identity := queueUnitIdentity{Occurrence: req.Occurrence, Heading: req.Heading}
	receipt := parsedEvidenceReceipt{
		header: evidenceReceiptHeader{
			protocol:        evidenceProtocolV1,
			digestAlgorithm: digestAlgorithmV1,
			recoveredFrom:   req.RecoveredFrom,
			versioned:       true,
		},
		identity:   &identity,
		heading:    req.Heading,
		verify:     req.Verify,
		digest:     req.UnitDigest,
		preSHA:     req.Pre.SHA,
		preStatus:  strconv.Itoa(req.Pre.Status),
		preOutput:  receiptOutputField(req.Pre.Output),
		preScope:   receiptScopeLine("Pre", req.Pre.Status, req.Pre.SHA),
		postSHA:    req.Post.SHA,
		postStatus: "0",
		postOutput: receiptOutputField(req.Post.Output),
		postScope:  receiptScopeLine("Post", 0, req.Post.SHA),
	}
	receiptID := receiptIDForCanonicalReceipt(receipt)
	receipt.header.receiptID = receiptID

	head := []string{}
	if req.Rebuild {
		head = append(head,
			fmt.Sprintf("Unit occurrence: %d", req.Occurrence),
			"Unit heading: "+req.Heading,
		)
	} else {
		head = append(head, "## "+req.Heading)
	}
	head = append(head,
		"Evidence:",
		"Protocol: "+evidenceProtocolV1,
		"Digest algorithm: "+digestAlgorithmV1,
		"Receipt ID: "+receiptID,
	)
	if req.RecoveredFrom != "" {
		head = append(head, "Recovered from: "+req.RecoveredFrom)
	}
	head = append(head, strings.Split(req.Verify, "\n")...)
	head = append(head,
		"unit digest: "+req.UnitDigest,
		"pre sha: "+req.Pre.SHA,
		fmt.Sprintf("pre exit status: %d", req.Pre.Status),
	)

	identityPrefix := []string{
		"Protocol: " + evidenceProtocolV1,
		"Digest algorithm: " + digestAlgorithmV1,
		"Receipt ID: " + receiptID,
		fmt.Sprintf("Unit occurrence: %d", req.Occurrence),
		"Unit heading: " + req.Heading,
	}
	mid := []string{
		"post sha: " + req.Post.SHA,
		"post exit status: 0",
	}
	preScope := []string{receipt.preScope}
	postScope := []string{receipt.postScope}

	markerRoom := len(receiptContinuationMarker) + 1
	overhead := receiptLinesOverhead

	comments := make([][]string, 0, 2)
	var cur []string
	markerUsed := false
	closeComment := func() {
		cur = append(cur, receiptContinuationMarker)
		comments = append(comments, cur)
		cur = nil
		markerUsed = true
	}
	flushFinal := func() {
		comments = append(comments, cur)
		cur = nil
	}

	cur = append(cur, head...)
	if overhead(cur)+markerRoom > limit {
		return nil, fmt.Errorf("receipt header exceeds the comment size limit %d", limit)
	}

	preDelim := receiptFenceDelimiter(receipt.preOutput)
	plainPreFence := receiptFenceLines(receipt.preOutput, preDelim)
	if overhead(cur)+1+overhead(plainPreFence)+1+overhead(preScope)+markerRoom <= limit {
		cur = append(cur, plainPreFence...)
		cur = append(cur, preScope...)
	} else {
		chunks, err := receiptChunkPlan(
			func(number, total int, payload string) []string {
				return receiptChunkLines("pre", number, total, receiptID, identity, req.UnitDigest, preDelim, payload)
			},
			receipt.preOutput, cur, preScope, true, limit, markerRoom,
		)
		if err != nil {
			return nil, err
		}
		for k, payload := range chunks {
			if k > 0 {
				closeComment()
			}
			cur = append(cur, receiptChunkLines("pre", k+1, len(chunks), receiptID, identity, req.UnitDigest, preDelim, payload)...)
		}
		cur = append(cur, preScope...)
	}

	if markerUsed {
		closeComment()
		cur = append(cur, identityPrefix...)
	} else if overhead(cur)+1+overhead(mid)+markerRoom > limit {
		closeComment()
		cur = append(cur, identityPrefix...)
	}
	cur = append(cur, mid...)

	postDelim := receiptFenceDelimiter(receipt.postOutput)
	plainPostFence := receiptFenceLines(receipt.postOutput, postDelim)
	if overhead(cur)+1+overhead(plainPostFence)+1+overhead(postScope) <= limit {
		cur = append(cur, plainPostFence...)
		cur = append(cur, postScope...)
		flushFinal()
	} else {
		chunks, err := receiptChunkPlan(
			func(number, total int, payload string) []string {
				return receiptChunkLines("post", number, total, receiptID, identity, req.UnitDigest, postDelim, payload)
			},
			receipt.postOutput, cur, postScope, false, limit, markerRoom,
		)
		if err != nil {
			return nil, err
		}
		for k, payload := range chunks {
			if k > 0 {
				closeComment()
			}
			cur = append(cur, receiptChunkLines("post", k+1, len(chunks), receiptID, identity, req.UnitDigest, postDelim, payload)...)
		}
		cur = append(cur, postScope...)
		flushFinal()
	}

	texts := make([]string, 0, len(comments))
	for _, comment := range comments {
		texts = append(texts, strings.Join(comment, "\n"))
	}
	if err := receiptSelfValidate(texts, receipt, req.Verify, identity); err != nil {
		return nil, err
	}
	return texts, nil
}

func receiptOutputField(output string) string {
	if strings.TrimSpace(output) == "" {
		return "<no output>"
	}
	return output
}

func receiptScopeLine(phase string, status int, sha string) string {
	return fmt.Sprintf("%s-evidence scope: this command exited %d at %s; nothing else is inferred.", phase, status, sha)
}

func receiptLinesOverhead(lines []string) int {
	total := 0
	for i, line := range lines {
		if i > 0 {
			total++
		}
		total += len(line)
	}
	return total
}

func receiptFenceLines(payload, delim string) []string {
	payloadLines := strings.Split(payload, "\n")
	lines := make([]string, 0, len(payloadLines)+2)
	lines = append(lines, delim)
	lines = append(lines, payloadLines...)
	return append(lines, delim)
}

// receiptFenceDelimiter returns a fence delimiter strictly longer than any
// delimiter line in the payload.
func receiptFenceDelimiter(payload string) string {
	maxRun := 2
	for _, line := range strings.Split(payload, "\n") {
		if run := len(fenceDelimiter(line)); run > maxRun {
			maxRun = run
		}
	}
	return strings.Repeat("`", maxRun+1)
}

func receiptChunkLines(phase string, number, total int, receiptID string, identity queueUnitIdentity, unitDigest, delim, payload string) []string {
	lines := []string{
		"Raw output chunk:",
		"Protocol: " + evidenceProtocolV1,
		"Digest algorithm: " + digestAlgorithmV1,
		"Receipt ID: " + receiptID,
		"Output: " + phase,
		fmt.Sprintf("Chunk: %d/%d", number, total),
		fmt.Sprintf("Unit occurrence: %d", identity.Occurrence),
		"Unit heading: " + identity.Heading,
		"unit digest: " + unitDigest,
	}
	return append(lines, receiptFenceLines(payload, delim)...)
}

// receiptChunkPlan sizes one chunked output so every chunk comment fits the
// comment limit: the first chunk rides after prior, later chunks stand alone,
// and trailing joins the final chunk. closeAfterTrailing marks that the final
// chunk comment still ends with a continuation marker. It shrinks the payload
// budget until an exact size check passes and refuses when nothing fits.
func receiptChunkPlan(
	build func(number, total int, payload string) []string,
	output string,
	prior, trailing []string,
	closeAfterTrailing bool,
	limit, markerRoom int,
) ([]string, error) {
	commentSize := func(k, total int, payload string) int {
		record := build(k, total, payload)
		size := receiptLinesOverhead(record)
		if k == 1 {
			size += receiptLinesOverhead(prior) + 1
		} else {
			size += markerRoom
		}
		if k == total {
			size += receiptLinesOverhead(trailing) + 1
			if closeAfterTrailing {
				size += markerRoom
			}
		} else {
			size += markerRoom
		}
		return size
	}
	budget := limit - 1
	for attempt := 0; attempt < 32; attempt++ {
		chunks := receiptSplitPayload(output, budget)
		total := len(chunks)
		fits := total >= 2
		for k := range chunks {
			if commentSize(k+1, total, chunks[k]) > limit {
				fits = false
				break
			}
		}
		if fits {
			return chunks, nil
		}
		if budget <= 1 {
			break
		}
		budget -= budget/4 + 1
		if budget < 1 {
			budget = 1
		}
	}
	return nil, fmt.Errorf("raw output cannot fit within the comment size limit %d", limit)
}

// receiptSplitPayload cuts the payload into consecutive byte-accurate pieces
// of at most budget bytes, preferring newline boundaries and otherwise
// backing off to a rune boundary.
func receiptSplitPayload(payload string, budget int) []string {
	var chunks []string
	remaining := payload
	for len(remaining) > budget {
		cut := budget
		if newline := strings.LastIndexByte(remaining[:budget], '\n'); newline > 0 {
			cut = newline + 1
		} else {
			for cut > 1 && !utf8.RuneStart(remaining[cut]) {
				cut--
			}
		}
		chunks = append(chunks, remaining[:cut])
		remaining = remaining[cut:]
	}
	return append(chunks, remaining)
}

func receiptSelfValidate(comments []string, expected parsedEvidenceReceipt, verify string, identity queueUnitIdentity) error {
	merged := mergeContinuedCommentRecords(comments)
	if len(merged) == 0 {
		return fmt.Errorf("assembly produced no receipt comments")
	}
	document := newEvidenceDocumentFromComment(merged[0])
	parsed, issues := parseEvidenceReceiptDocument(
		evidencePayloadDocument(document),
		verify,
		expected.digest,
		"receipt assembly",
		expected.heading,
		&identity,
	)
	if len(issues) > 0 {
		return fmt.Errorf("assembled receipt failed self-validation: %s", issues[0].Message)
	}
	if parsed.preOutput != expected.preOutput || parsed.postOutput != expected.postOutput {
		return fmt.Errorf("assembled receipt does not reconstruct the raw outputs verbatim")
	}
	if receiptIDForCanonicalReceipt(parsed) != expected.header.receiptID {
		return fmt.Errorf("assembled receipt ID does not match the validator derivation")
	}
	return nil
}
