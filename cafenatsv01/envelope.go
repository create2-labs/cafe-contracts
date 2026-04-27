package cafenatsv01

import eventenvelopev01 "github.com/create2-labs/cafe-contracts/eventenvelope/v01"

// EnvelopeV01 is the common header for policy/remediation v0.1 family events. Embed or copy
// these fields into each concrete event struct for stable JSON field order across services.
type EnvelopeV01 = eventenvelopev01.Envelope

// SubjectRef is the stable subject address on the wire for event payloads.
type SubjectRef struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

func requireEnvelopeV01(h EnvelopeV01, typeWant string) error {
	if h.EventID == "" {
		return ErrEventIDRequired
	}
	if h.EventType != typeWant {
		return ErrEventType
	}
	if h.EventVersion != EventVersionV01 {
		return ErrEventVersion
	}
	return nil
}

func requireSubjectWallet(s SubjectRef) error {
	if s.Type != SubjectTypeWallet {
		return ErrSubjectType
	}
	if s.ID == "" {
		return ErrSubjectID
	}
	return nil
}

func requireSubjectPolicyInstance(s SubjectRef) error {
	if s.Type != SubjectTypePolicyInstance {
		return ErrSubjectType
	}
	if s.ID == "" {
		return ErrSubjectID
	}
	return nil
}

func requireSubjectRemediation(s SubjectRef) error {
	if s.Type != SubjectTypeRemediation {
		return ErrSubjectType
	}
	if s.ID == "" {
		return ErrSubjectID
	}
	return nil
}
