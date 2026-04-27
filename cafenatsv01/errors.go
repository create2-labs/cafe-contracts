package cafenatsv01

import "errors"

// Validation errors for the cafenatsv01 wire bundle. Use errors.Is in handlers.
var (
	ErrEventIDRequired  = errors.New("cafenatsv01 v0.1: event_id is required")
	ErrEventType        = errors.New("cafenatsv01 v0.1: event_type mismatch")
	ErrEventVersion     = errors.New("cafenatsv01 v0.1: event_version must be v0.1")
	ErrProducer         = errors.New("cafenatsv01 v0.1: producer is not valid for this event")
	ErrSubjectType      = errors.New("cafenatsv01 v0.1: subject.type is invalid for this event")
	ErrSubjectID        = errors.New("cafenatsv01 v0.1: subject.id is required")
	ErrPayloadObs       = errors.New("cafenatsv01 v0.1: nested observation is invalid or missing")
	ErrSelectionRequest = errors.New("cafenatsv01 v0.1: selection_request is invalid")
	ErrPayloadInvalid   = errors.New("cafenatsv01 v0.1: payload is invalid or incomplete")
)
