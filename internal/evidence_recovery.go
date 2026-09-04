package internal

import (
	"fmt"
	"strings"
)

type evidenceReceiptObservation struct {
	identity queueUnitIdentity
	receipt  parsedEvidenceReceipt
}

type evidenceReceiptRegistry struct {
	receipts map[string]evidenceReceiptObservation
}

func newEvidenceReceiptRegistry() *evidenceReceiptRegistry {
	return &evidenceReceiptRegistry{receipts: make(map[string]evidenceReceiptObservation)}
}

func (r *evidenceReceiptRegistry) add(observation evidenceReceiptObservation) error {
	id := observation.receipt.header.receiptID
	if previous, ok := r.receipts[id]; ok {
		if !sameReceiptCanonicalFields(previous.receipt, observation.receipt) {
			return fmt.Errorf("divergent duplicate Receipt ID %s", id)
		}
		return nil
	}

	if recoveredFrom := observation.receipt.header.recoveredFrom; recoveredFrom != "" {
		previous, ok := r.receipts[recoveredFrom]
		if !ok {
			return fmt.Errorf("Recovered from %s does not identify an earlier complete receipt", recoveredFrom)
		}
		if previous.identity != observation.identity {
			return fmt.Errorf(
				"Recovered from %s identifies occurrence %d with heading %q, not occurrence %d with heading %q",
				recoveredFrom,
				previous.identity.Occurrence,
				previous.identity.Heading,
				observation.identity.Occurrence,
				observation.identity.Heading,
			)
		}
	}

	r.receipts[id] = observation
	return nil
}

func sameReceiptCanonicalFields(left, right parsedEvidenceReceipt) bool {
	leftFields := receiptCanonicalFields(left)
	rightFields := receiptCanonicalFields(right)
	if len(leftFields) != len(rightFields) {
		return false
	}
	for i := range leftFields {
		if leftFields[i] != rightFields[i] {
			return false
		}
	}
	return true
}

func completeEvidenceReceiptObservation(
	document evidenceDocument,
	identity queueUnitIdentity,
	units []queueUnit,
	source string,
) (evidenceReceiptObservation, bool) {
	unit, hasUnit := findQueueUnit(units, identity)
	if hasUnit {
		receipt, issues := parseEvidenceReceiptDocument(
			document,
			unitVerifyCommand(unit.Body),
			evidenceReceiptExpectedDigest(unit, document),
			source,
			identity.Heading,
			&identity,
		)
		if len(issues) == 0 {
			return evidenceReceiptObservation{identity: identity, receipt: receipt}, true
		}
	}

	declarations := validEvidenceReceiptDeclarations(document, source, identity.Heading, &identity)
	if len(declarations) != 1 {
		return evidenceReceiptObservation{}, false
	}
	if !hasUnit && digestMatchesAnyUnit(declarations[0].digest, units) {
		return evidenceReceiptObservation{}, false
	}
	return evidenceReceiptObservation{identity: identity, receipt: declarations[0].receipt}, true
}

func bodyEvidenceReceiptObservations(units []queueUnit) []evidenceReceiptObservation {
	identities := queueUnitIdentities(units)
	observations := make([]evidenceReceiptObservation, 0, len(units))
	for i, unit := range units {
		if !isCheckedUnit(unit.Body) {
			continue
		}
		document := evidencePayloadDocument(newEvidenceDocument(extractEvidenceText(unit.Body)))
		if observation, ok := completeEvidenceReceiptObservation(document, identities[i], units, "queue"); ok {
			observations = append(observations, observation)
		}
	}
	return observations
}

func completeEvidenceReceiptObservationsForComment(
	comment continuedComment,
	units []queueUnit,
) []evidenceReceiptObservation {
	document := newEvidenceDocumentFromComment(comment).trimSpace()
	if len(document.lines) >= 3 &&
		strings.HasPrefix(document.lines[0], "Unit occurrence:") &&
		strings.HasPrefix(document.lines[1], "Unit heading:") &&
		strings.TrimSpace(document.lines[2]) == "Evidence:" {
		identity, err := parseIdentityLines(document.lines[0], document.lines[1])
		if err != nil {
			return nil
		}
		observation, ok := completeEvidenceReceiptObservation(
			evidencePayloadDocument(document.afterLine(2)),
			identity,
			units,
			"comment",
		)
		if !ok {
			return nil
		}
		return []evidenceReceiptObservation{observation}
	}

	identities := queueUnitIdentities(units)
	observations := make([]evidenceReceiptObservation, 0, 1)
	for i, unit := range units {
		if !commentNamesUnit(comment.text, unit.Heading) {
			continue
		}
		document, ok := commentEvidenceDocument(comment, unit.Heading)
		if !ok {
			continue
		}
		observation, ok := completeEvidenceReceiptObservation(document, identities[i], units, "comment")
		if ok {
			observations = append(observations, observation)
			break
		}
	}
	return observations
}
