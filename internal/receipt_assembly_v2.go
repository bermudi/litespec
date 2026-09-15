package internal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	receiptV2TotalBudget = 8192
	receiptV2HeadBudget  = 2048
	receiptV2TailBudget  = 3072
)

// AssembleEvidenceV2ReceiptComments turns one red-green run into a single
// bounded evidence/v2 receipt comment. It refuses invalid run evidence before
// producing any content, excerpts oversized outputs head/marker/tail beside
// the full output's byte count and SHA-256, shrinks excerpt budgets
// deterministically until the comment fits the fixed receipt budget, and
// self-parses the result through the existing evidence grammar before
// returning.
func AssembleEvidenceV2ReceiptComments(req ReceiptAssemblyRequest) ([]string, error) {
	if err := req.validate(); err != nil {
		return nil, err
	}
	preOutput, preBytes, preOutputSHA := receiptV2OutputMeta(req.Pre.Output)
	postOutput, postBytes, postOutputSHA := receiptV2OutputMeta(req.Post.Output)
	identity := queueUnitIdentity{Occurrence: req.Occurrence, Heading: req.Heading}
	mirror := parsedEvidenceReceipt{
		header: evidenceReceiptHeader{
			protocol:        evidenceProtocolV2,
			digestAlgorithm: digestAlgorithmV1,
			recoveredFrom:   req.RecoveredFrom,
			versioned:       true,
		},
		identity:      &identity,
		heading:       req.Heading,
		verify:        req.Verify,
		digest:        req.UnitDigest,
		preSHA:        req.Pre.SHA,
		preStatus:     strconv.Itoa(req.Pre.Status),
		preBytes:      preBytes,
		preOutputSHA:  preOutputSHA,
		preScope:      receiptScopeLine("Pre", req.Pre.Status, req.Pre.SHA),
		postSHA:       req.Post.SHA,
		postStatus:    "0",
		postBytes:     postBytes,
		postOutputSHA: postOutputSHA,
		postScope:     receiptScopeLine("Post", 0, req.Post.SHA),
	}
	headBudget, tailBudget := receiptV2HeadBudget, receiptV2TailBudget
	for {
		pre := receiptV2ExcerptFor(preOutput, headBudget, tailBudget)
		post := receiptV2ExcerptFor(postOutput, headBudget, tailBudget)
		mirror.preOutput = pre.payload
		mirror.postOutput = post.payload
		mirror.header.receiptID = receiptIDForCanonicalReceipt(mirror)
		lines := receiptV2CommentLines(req, mirror.header.receiptID, pre, post, preBytes, preOutputSHA, postBytes, postOutputSHA)
		comment := strings.Join(lines, "\n")
		if len(comment) <= receiptV2TotalBudget {
			if err := receiptV2SelfValidate(comment, mirror, req.Verify, identity); err != nil {
				return nil, err
			}
			return []string{comment}, nil
		}
		if headBudget == 0 && tailBudget == 0 {
			return nil, fmt.Errorf(
				"receipt identity and status fields exceed the %d-byte receipt budget even with empty excerpts (oversized field: %s)",
				receiptV2TotalBudget, receiptV2OversizedField(req),
			)
		}
		headBudget /= 2
		tailBudget /= 2
	}
}

type receiptV2Excerpt struct {
	payload string
	elided  int
}

func receiptV2ExcerptFor(output string, headBudget, tailBudget int) receiptV2Excerpt {
	if len(output) <= headBudget+tailBudget {
		return receiptV2Excerpt{payload: output}
	}
	head := output[:receiptV2HeadCut(output, headBudget)]
	tail := output[receiptV2TailStart(output, tailBudget):]
	elided := len(output) - len(head) - len(tail)
	marker := fmt.Sprintf("... %d bytes elided ...", elided)
	if tail == "" {
		return receiptV2Excerpt{payload: head + marker, elided: elided}
	}
	return receiptV2Excerpt{payload: head + marker + "\n" + tail, elided: elided}
}

// receiptV2HeadCut returns the byte length of the head excerpt: at most
// budget bytes, backed off to a rune boundary and then to the last line
// boundary so the elision marker always opens its own line.
func receiptV2HeadCut(output string, budget int) int {
	if budget >= len(output) {
		return len(output)
	}
	cut := budget
	for cut > 0 && !utf8.RuneStart(output[cut]) {
		cut--
	}
	if newline := strings.LastIndexByte(output[:cut], '\n'); newline >= 0 {
		return newline + 1
	}
	return 0

}

// receiptV2TailStart returns the byte offset of the tail excerpt: at most
// budget bytes from the end, advanced to a rune boundary, never mid-rune.
func receiptV2TailStart(output string, budget int) int {
	if budget >= len(output) {
		return 0
	}
	start := len(output) - budget
	for start < len(output) && !utf8.RuneStart(output[start]) {
		start++
	}
	return start
}

// receiptV2OutputMeta normalizes empty output to the literal `<no output>`
// fence payload and records the full output's byte count and SHA-256; the
// empty-output shape declares zero bytes and the empty-string hash.
func receiptV2OutputMeta(original string) (normalized string, byteCount string, outputSHA string) {
	normalized = receiptOutputField(original)
	if strings.TrimSpace(original) == "" {
		sum := sha256.Sum256(nil)
		return normalized, "0", hex.EncodeToString(sum[:])
	}
	sum := sha256.Sum256([]byte(normalized))
	return normalized, strconv.Itoa(len(normalized)), hex.EncodeToString(sum[:])
}

func receiptV2CommentLines(req ReceiptAssemblyRequest, receiptID string, pre, post receiptV2Excerpt, preBytes, preOutputSHA, postBytes, postOutputSHA string) []string {
	lines := []string{}
	if req.Rebuild {
		lines = append(lines,
			fmt.Sprintf("Unit occurrence: %d", req.Occurrence),
			"Unit heading: "+req.Heading,
		)
	} else {
		lines = append(lines, "## "+req.Heading)
	}
	lines = append(lines,
		"Evidence:",
		"Protocol: "+evidenceProtocolV2,
		"Digest algorithm: "+digestAlgorithmV1,
		"Receipt ID: "+receiptID,
	)
	if req.RecoveredFrom != "" {
		lines = append(lines, "Recovered from: "+req.RecoveredFrom)
	}
	lines = append(lines, strings.Split(req.Verify, "\n")...)
	lines = append(lines,
		"unit digest: "+req.UnitDigest,
		"pre sha: "+req.Pre.SHA,
		fmt.Sprintf("pre exit status: %d", req.Pre.Status),
	)
	lines = append(lines, receiptFenceLines(pre.payload, receiptFenceDelimiter(pre.payload))...)
	lines = append(lines,
		"pre bytes: "+preBytes,
		"pre output sha256: "+preOutputSHA,
		receiptScopeLine("Pre", req.Pre.Status, req.Pre.SHA),
		"post sha: "+req.Post.SHA,
		"post exit status: 0",
	)
	lines = append(lines, receiptFenceLines(post.payload, receiptFenceDelimiter(post.payload))...)
	lines = append(lines,
		"post bytes: "+postBytes,
		"post output sha256: "+postOutputSHA,
		receiptScopeLine("Post", 0, req.Post.SHA),
	)
	return lines
}

func receiptV2OversizedField(req ReceiptAssemblyRequest) string {
	fields := []struct {
		name string
		size int
	}{
		{"Verify command", receiptLinesOverhead(strings.Split(req.Verify, "\n"))},
		{"Unit heading", len(req.Heading)},
		{"Recovered from", len(req.RecoveredFrom)},
		{
			"scope lines",
			len(receiptScopeLine("Pre", req.Pre.Status, req.Pre.SHA)) + len(receiptScopeLine("Post", 0, req.Post.SHA)),
		},
	}
	oversized := fields[0]
	for _, field := range fields[1:] {
		if field.size > oversized.size {
			oversized = field
		}
	}
	return oversized.name
}

func receiptV2SelfValidate(comment string, expected parsedEvidenceReceipt, verify string, identity queueUnitIdentity) error {
	document := newEvidenceDocument(comment)
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
		return fmt.Errorf("assembled receipt does not carry the computed excerpts verbatim")
	}
	if parsed.preBytes != expected.preBytes || parsed.preOutputSHA != expected.preOutputSHA ||
		parsed.postBytes != expected.postBytes || parsed.postOutputSHA != expected.postOutputSHA {
		return fmt.Errorf("assembled receipt does not carry the declared output metadata")
	}
	if receiptIDForCanonicalReceipt(parsed) != expected.header.receiptID {
		return fmt.Errorf("assembled receipt ID does not match the validator derivation")
	}
	return nil
}
