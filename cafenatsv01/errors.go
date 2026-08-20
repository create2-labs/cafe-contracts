package cafenatsv01

import "errors"

// Validation errors for the cafenatsv01 wire bundle. Use errors.Is in handlers.
var (
	ErrEventIDRequired        = errors.New("cafenatsv01 v0.1: event_id is required")
	ErrEventType              = errors.New("cafenatsv01 v0.1: event_type mismatch")
	ErrEventVersion           = errors.New("cafenatsv01 v0.1: event_version must be v0.1")
	ErrProducer               = errors.New("cafenatsv01 v0.1: producer is not valid for this event")
	ErrSubjectType            = errors.New("cafenatsv01 v0.1: subject.type is invalid for this event")
	ErrSubjectID              = errors.New("cafenatsv01 v0.1: subject.id is required")
	ErrPayloadInvalid         = errors.New("cafenatsv01 v0.1: payload is invalid or incomplete")
	ErrCryptoPolicyIDRequired = errors.New("cafenatsv01: crypto_policy_id is required")
	// ErrLegacyAssessmentField is returned when assessment payload v0.2 still carries
	// selection_request or couche-B / former selection fields.
	ErrLegacyAssessmentField = errors.New("cafenatsv01: assessment v0.2 rejects selection_request and layer-B fields")
)
